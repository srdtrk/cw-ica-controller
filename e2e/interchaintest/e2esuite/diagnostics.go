package e2esuite

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
)

const (
	testLabel = "ibc-test"

	dockerInspectFileName = "docker-inspect.json"
	e2eDir                = "e2e"
	defaultFilePerm       = 0o750
)

// collect can be used in `t.Cleanup` and will copy all the of the container logs and relevant files
// into e2e/<test-suite>/<test-name>.log. These log files will be uploaded to GH upon test failure.
func collect(t *testing.T, dc *dockerclient.Client, debugModeEnabled bool, chainNames ...string) {
	t.Helper()

	if !debugModeEnabled {
		// when we are not forcing log collection, we only upload upon test failing.
		if !t.Failed() {
			t.Logf("test passed, not uploading logs")
			return
		}
	}

	t.Logf("writing logs for test: %s", t.Name())

	ctx := context.TODO()
	e2eDir, err := getE2EDir(t)
	if err != nil {
		t.Logf("failed finding log directory: %s", err)
		return
	}

	logsDir := fmt.Sprintf("%s/diagnostics", e2eDir)

	if err := os.MkdirAll(fmt.Sprintf("%s/%s", logsDir, t.Name()), defaultFilePerm); err != nil {
		t.Logf("failed creating logs directory in test cleanup: %s", err)
		return
	}

	testContainers, err := GetTestContainers(ctx, t, dc)
	if err != nil {
		t.Logf("failed listing containers during test cleanup: %s", err)
		return
	}

	for _, container := range testContainers {
		containerName := getContainerName(t, container)
		containerDir := fmt.Sprintf("%s/%s/%s", logsDir, t.Name(), containerName)
		if err := os.MkdirAll(containerDir, defaultFilePerm); err != nil {
			t.Logf("failed creating logs directory for container %s: %s", containerDir, err)
			continue
		}

		logsBz, err := getContainerLogs(ctx, dc, container.ID)
		if err != nil {
			t.Logf("failed reading logs in test cleanup: %s", err)
			continue
		}

		logFile := fmt.Sprintf("%s/%s.log", containerDir, containerName)
		if err := os.WriteFile(logFile, logsBz, defaultFilePerm); err != nil {
			continue
		}

		t.Logf("successfully wrote log file %s", logFile)

		var diagnosticFiles []string
		for _, chainName := range chainNames {
			diagnosticFiles = append(diagnosticFiles, chainDiagnosticAbsoluteFilePaths(chainName)...)
		}
		diagnosticFiles = append(diagnosticFiles, relayerDiagnosticAbsoluteFilePaths()...)

		for _, absoluteFilePathInContainer := range diagnosticFiles {
			localFilePath := path.Join(containerDir, path.Base(absoluteFilePathInContainer))
			if err := fetchAndWriteDiagnosticsFile(ctx, dc, container.ID, localFilePath, absoluteFilePathInContainer); err != nil {
				continue
			}
			t.Logf("successfully wrote diagnostics file %s", absoluteFilePathInContainer)
		}

		localFilePath := path.Join(containerDir, dockerInspectFileName)
		if err := fetchAndWriteDockerInspectOutput(ctx, dc, container.ID, localFilePath); err != nil {
			continue
		}
		t.Logf("successfully wrote docker inspect output")
	}
}

// getContainerName returns a either the ID of the container or stripped down human-readable
// version of the name if the name is non-empty.
//
// Note: You should still always use the ID  when interacting with the docker client.
func getContainerName(t *testing.T, container dockertypes.Container) string {
	t.Helper()
	// container will always have an id, by may not have a name.
	containerName := container.ID
	if len(container.Names) > 0 {
		containerName = container.Names[0]
		// remove the test name from the container as the folder structure will provide this
		// information already.
		containerName = strings.TrimRight(containerName, "-"+t.Name())
		containerName = strings.TrimLeft(containerName, "/")
	}
	return containerName
}

// fetchAndWriteDiagnosticsFile fetches the contents of a single file from the given container id and writes
// the contents of the file to a local path provided.
func fetchAndWriteDiagnosticsFile(ctx context.Context, dc *dockerclient.Client, containerID, localPath, absoluteFilePathInContainer string) error {
	fileBz, err := getFileContentsFromContainer(ctx, dc, containerID, absoluteFilePathInContainer)
	if err != nil {
		return err
	}

	return os.WriteFile(localPath, fileBz, defaultFilePerm)
}

// fetchAndWriteDockerInspectOutput writes the contents of docker inspect to the specified file.
func fetchAndWriteDockerInspectOutput(ctx context.Context, dc *dockerclient.Client, containerID, localPath string) error {
	containerJSON, err := dc.ContainerInspect(ctx, containerID)
	if err != nil {
		return err
	}

	fileBz, err := json.MarshalIndent(containerJSON, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(localPath, fileBz, defaultFilePerm)
}

// chainDiagnosticAbsoluteFilePaths returns a slice of absolute file paths (in the containers) which are the files that should be
// copied locally when fetching diagnostics.
func chainDiagnosticAbsoluteFilePaths(chainName string) []string {
	return []string{
		fmt.Sprintf("/var/cosmos-chain/%s/config/genesis.json", chainName),
		fmt.Sprintf("/var/cosmos-chain/%s/config/app.toml", chainName),
		fmt.Sprintf("/var/cosmos-chain/%s/config/config.toml", chainName),
		fmt.Sprintf("/var/cosmos-chain/%s/config/client.toml", chainName),
	}
}

// relayerDiagnosticAbsoluteFilePaths returns a slice of absolute file paths (in the containers) which are the files that should be
// copied locally when fetching diagnostics.
func relayerDiagnosticAbsoluteFilePaths() []string {
	return []string{
		"/home/hermes/.hermes/config.toml",
	}
}

// getE2EDir finds the e2e directory above the test.
func getE2EDir(t *testing.T) (string, error) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	const maxAttempts = 100
	count := 0
	for ; !strings.HasSuffix(wd, e2eDir) || count > maxAttempts; wd = path.Dir(wd) {
		count++
	}

	// arbitrary value to avoid getting stuck in an infinite loop if this is called
	// in a context where the e2e directory does not exist.
	if count > maxAttempts {
		return "", fmt.Errorf("unable to find e2e directory after %d tries", maxAttempts)
	}

	return wd, nil
}

// GetTestContainers returns all docker containers that have been created by interchain test.
func GetTestContainers(ctx context.Context, t *testing.T, dc *dockerclient.Client) ([]dockertypes.Container, error) {
	t.Helper()

	testContainers, err := dc.ContainerList(ctx, dockertypes.ContainerListOptions{
		All: true,
		Filters: filters.NewArgs(
			// see: https://github.com/strangelove-ventures/interchaintest/blob/0bdc194c2aa11aa32479f32b19e1c50304301981/internal/dockerutil/setup.go#L31-L36
			// for the label needed to identify test containers.
			filters.Arg("label", testLabel+"="+t.Name()),
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed listing containers: %s", err)
	}

	return testContainers, nil
}

// getContainerLogs returns the logs of a container as a byte array.
func getContainerLogs(ctx context.Context, dc *dockerclient.Client, containerName string) ([]byte, error) {
	readCloser, err := dc.ContainerLogs(ctx, containerName, dockertypes.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed reading logs in test cleanup: %s", err)
	}
	return io.ReadAll(readCloser)
}

// getFileContentsFromContainer reads the contents of a specific file from a container.
func getFileContentsFromContainer(ctx context.Context, dc *dockerclient.Client, containerID, absolutePath string) ([]byte, error) {
	readCloser, _, err := dc.CopyFromContainer(ctx, containerID, absolutePath)
	if err != nil {
		return nil, fmt.Errorf("copying from container: %w", err)
	}

	defer readCloser.Close()

	fileName := path.Base(absolutePath)
	tr := tar.NewReader(readCloser)

	hdr, err := tr.Next()
	if err != nil {
		return nil, err
	}

	if hdr.Name != fileName {
		return nil, fmt.Errorf("expected to find %s but found %s", fileName, hdr.Name)
	}

	return io.ReadAll(tr)
}
