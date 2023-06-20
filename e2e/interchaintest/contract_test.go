package main

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	chantypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestIcaControllerContract(t *testing.T) {
	// Parallel indicates that this test is safe for parallel execution.
	// This is true since this is the only test in this file.
	t.Parallel()

	client, network := interchaintest.DockerSetup(t)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()

	// Get both chains
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name: "icad",
			ChainConfig: ibc.ChainConfig{
				Images:                 []ibc.DockerImage{{Repository: "ghcr.io/cosmos/ibc-go-icad", Version: "v0.5.0"}},
				UsingNewGenesisCommand: true,
			},
		},
		{
			Name: "icad",
			ChainConfig: ibc.ChainConfig{
				Images:                 []ibc.DockerImage{{Repository: "ghcr.io/cosmos/ibc-go-icad", Version: "v0.5.0"}},
				UsingNewGenesisCommand: true,
			},
		},
	})
}