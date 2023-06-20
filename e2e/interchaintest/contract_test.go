package main

import (
	"context"
	"testing"

	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
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
				Images:                 []ibc.DockerImage{{Repository: "ghcr.io/srdtrk/wasmd", Version: "jsonica"}},
				UsingNewGenesisCommand: true,
			},
		},
		{
			Name: "icad",
			ChainConfig: ibc.ChainConfig{
				Images:                 []ibc.DockerImage{{Repository: "ghcr.io/srdtrk/wasmd", Version: "jsonica"}},
				UsingNewGenesisCommand: true,
			},
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	chain1, chain2 := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	// Get a relayer instance
	r := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		relayer.RelayerOptionExtraStartFlags{Flags: []string{"-p", "events", "-b", "100"}},
	).Build(t, client, network)
}