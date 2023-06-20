package main

import (
	"context"
	"testing"

	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types"
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
		// -- WASMD --
		{
			ChainConfig: ibc.ChainConfig{
				Type:    "cosmos",
				Name:    "wasmd",
				ChainID: "wasmd-1",
				Images: []ibc.DockerImage{
					{
						Repository: "cosmwasm/wasmd", // FOR LOCAL IMAGE USE: Docker Image Name
						Version:    "v0.40.2",        // FOR LOCAL IMAGE USE: Docker Image Tag
					},
				},
				Bin:                    "wasmd",
				Bech32Prefix:           "wasm",
				Denom:                  "stake",
				GasPrices:              "0.00stake",
				GasAdjustment:          1.3,
				TrustingPeriod:         "508h",
				NoHostMount:            false,
				UsingNewGenesisCommand: true,
			},
		},
		// -- IBC-GO --
		{
			ChainConfig: ibc.ChainConfig{
				Type:    "cosmos",
				Name:    "ibc-go-simd",
				ChainID: "simd-1",
				Images: []ibc.DockerImage{
					{
						Repository: "ghcr.io/cosmos/ibc-go-simd", // FOR LOCAL IMAGE USE: Docker Image Name
						Version:    "pr-3796",                    // FOR LOCAL IMAGE USE: Docker Image Tag
					},
				},
				Bin:                    "simd",
				Bech32Prefix:           "cosmos",
				Denom:                  "gos",
				GasPrices:              "0.00gos",
				GasAdjustment:          1.3,
				TrustingPeriod:         "508h",
				NoHostMount:            false,
				UsingNewGenesisCommand: true,
			},
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	wasmd, simd := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	// Get a relayer instance
	customRelayer := relayer.CustomDockerImage("damiannolan/rly", "", "100:1000")

	r := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		relayer.RelayerOptionExtraStartFlags{Flags: []string{"-p", "events", "-b", "100"}},
		customRelayer,
	).Build(t, client, network)

	// Build the network; spin up the chains and configure the relayer
	const (
		pathName    = "wasmd-simd"
		relayerName = "relayer"
	)

	ic := interchaintest.NewInterchain().
		AddChain(wasmd).
		AddChain(simd).
		AddRelayer(r, relayerName).
		AddLink(interchaintest.InterchainLink{
			Chain1:  wasmd,
			Chain2:  simd,
			Relayer: r,
			Path:    pathName,
		})

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: true,
	}))

	// Fund a user account on wasmd and simd
	const userFunds = int64(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, wasmd, simd)
	wasmdUser := users[0]
	// simdUser := users[1]

	// Generate a new IBC path
	err = r.GeneratePath(ctx, eRep, wasmd.Config().ChainID, simd.Config().ChainID, pathName)
	require.NoError(t, err)

	// Create new clients
	err = r.CreateClients(ctx, eRep, pathName, ibc.CreateClientOptions{TrustingPeriod: "330h"})
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 2, wasmd, simd)
	require.NoError(t, err)

	// Create a new connection
	err = r.CreateConnections(ctx, eRep, pathName)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 2, wasmd, simd)
	require.NoError(t, err)

	// Query for the newly created connection
	connections, err := r.GetConnections(ctx, eRep, wasmd.Config().ChainID)
	require.NoError(t, err)
	require.Equal(t, 1, len(connections))

	// Start the relayer and set the cleanup function.
	err = r.StartRelayer(ctx, eRep, pathName)
	require.NoError(t, err)

	t.Cleanup(
		func() {
			err := r.StopRelayer(ctx, eRep)
			if err != nil {
				t.Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)

	// Upload and Instantiate the contract on wasmd:
	codeId, err := wasmd.StoreContract(ctx, wasmdUser.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	require.NoError(t, err)
	contractAddr, err := wasmd.InstantiateContract(ctx, wasmdUser.KeyName(), codeId, types.NewInstantiateMsg(nil), true)
	require.NoError(t, err)

	contractPort := "wasm." + contractAddr

	// (WIP) Tests to be continued:
	println(contractPort)
}