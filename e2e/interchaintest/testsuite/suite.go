package testsuite

import (
	"context"

	dockerclient "github.com/docker/docker/client"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
)

type TestSuite struct {
	suite.Suite

	ChainA       *cosmos.CosmosChain
	ChainB       *cosmos.CosmosChain
	UserA        ibc.Wallet
	UserB        ibc.Wallet
	ChainAConnID string
	ChainBConnID string
	dockerClient *dockerclient.Client
	Relayer      ibc.Relayer
	network      string
	logger       *zap.Logger
	ExecRep      *testreporter.RelayerExecReporter
	PathName     string
}

// SetupSuite sets up the chains, relayer, user accounts, clients, and connections
func (s *TestSuite) SetupSuite(ctx context.Context, chainSpecs []*interchaintest.ChainSpec) {
	if len(chainSpecs) != 2 {
		panic("ContractTestSuite requires exactly 2 chain specs")
	}

	t := s.T()

	s.logger = zaptest.NewLogger(t)
	s.dockerClient, s.network = interchaintest.DockerSetup(t)

	cf := interchaintest.NewBuiltinChainFactory(s.logger, chainSpecs)

	chains, err := cf.Chains(t.Name())
	s.Require().NoError(err)
	s.ChainA = chains[0].(*cosmos.CosmosChain)
	s.ChainB = chains[1].(*cosmos.CosmosChain)

	// docker run -it --rm --entrypoint echo ghcr.io/cosmos/relayer "$(id -u):$(id -g)"
	customRelayerImage := relayer.CustomDockerImage("ghcr.io/cosmos/relayer", "", "100:1000")

	s.Relayer = interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		relayer.RelayerOptionExtraStartFlags{Flags: []string{"-p", "events", "-b", "100"}},
		customRelayerImage,
	).Build(t, s.dockerClient, s.network)

	s.ExecRep = testreporter.NewNopReporter().RelayerExecReporter(t)

	s.PathName = s.ChainA.Config().Name + "-" + s.ChainB.Config().Name

	ic := interchaintest.NewInterchain().
		AddChain(s.ChainA).
		AddChain(s.ChainB).
		AddRelayer(s.Relayer, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  s.ChainA,
			Chain2:  s.ChainB,
			Relayer: s.Relayer,
			Path:    s.PathName,
		})

	s.Require().NoError(ic.Build(ctx, s.ExecRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           s.dockerClient,
		NetworkID:        s.network,
		SkipPathCreation: true,
	}))

	// Fund a user account on ChainA and ChainB
	const userFunds = int64(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, s.ChainA, s.ChainB)
	s.UserA = users[0]
	s.UserB = users[1]

	// Generate a new IBC path
	err = s.Relayer.GeneratePath(ctx, s.ExecRep, s.ChainA.Config().ChainID, s.ChainB.Config().ChainID, s.PathName)
	s.Require().NoError(err)

	// Create new clients
	err = s.Relayer.CreateClients(ctx, s.ExecRep, s.PathName, ibc.CreateClientOptions{TrustingPeriod: "330h"})
	s.Require().NoError(err)

	err = testutil.WaitForBlocks(ctx, 2, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	// Create a new connection
	err = s.Relayer.CreateConnections(ctx, s.ExecRep, s.PathName)
	s.Require().NoError(err)

	err = testutil.WaitForBlocks(ctx, 2, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	// Query for the newly created connection in wasmd
	connections, err := s.Relayer.GetConnections(ctx, s.ExecRep, s.ChainA.Config().ChainID)
	s.Require().NoError(err)
	// localhost is always a connection since ibc-go v7.1+
	s.Require().Equal(2, len(connections))
	wasmdConnection := connections[0]
	s.Require().NotEqual("connection-localhost", wasmdConnection.ID)
	s.ChainAConnID = wasmdConnection.ID

	// Query for the newly created connection in simd
	connections, err = s.Relayer.GetConnections(ctx, s.ExecRep, s.ChainB.Config().ChainID)
	s.Require().NoError(err)
	// localhost is always a connection since ibc-go v7.1+
	s.Require().Equal(2, len(connections))
	simdConnection := connections[0]
	s.Require().NotEqual("connection-localhost", simdConnection.ID)
	s.ChainBConnID = simdConnection.ID

	// Start the relayer and set the cleanup function.
	err = s.Relayer.StartRelayer(ctx, s.ExecRep, s.PathName)
	s.Require().NoError(err)

	t.Cleanup(
		func() {
			err := s.Relayer.StopRelayer(ctx, s.ExecRep)
			if err != nil {
				t.Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)
}
