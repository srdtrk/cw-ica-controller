package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types"

	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos/wasm"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

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
						// unfortunately, the latest wasmd doesn't work with interchaintest at the moment
						Version: "v0.32.1", // FOR LOCAL IMAGE USE: Docker Image Tag
					},
				},
				Bin:           "wasmd",
				Bech32Prefix:  "wasm",
				Denom:         "gos",
				GasPrices:     "0.00gos",
				GasAdjustment: 1.3,
				// cannot run wasmd commands without wasm encoding
				EncodingConfig:         wasm.WasmEncoding(),
				TrustingPeriod:         "508h",
				NoHostMount:            false,
				UsingNewGenesisCommand: false,
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

	relayer := interchaintest.NewBuiltinRelayerFactory(
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
		AddRelayer(relayer, relayerName).
		AddLink(interchaintest.InterchainLink{
			Chain1:  wasmd,
			Chain2:  simd,
			Relayer: relayer,
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
	err = relayer.GeneratePath(ctx, eRep, wasmd.Config().ChainID, simd.Config().ChainID, pathName)
	require.NoError(t, err)

	// Create new clients
	err = relayer.CreateClients(ctx, eRep, pathName, ibc.CreateClientOptions{TrustingPeriod: "330h"})
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 2, wasmd, simd)
	require.NoError(t, err)

	// Create a new connection
	err = relayer.CreateConnections(ctx, eRep, pathName)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 2, wasmd, simd)
	require.NoError(t, err)

	// Query for the newly created connection in wasmd
	connections, err := relayer.GetConnections(ctx, eRep, wasmd.Config().ChainID)
	require.NoError(t, err)
	require.Equal(t, 1, len(connections))
	wasmdConnection := connections[0]
	require.Equal(t, "connection-0", wasmdConnection.ID)

	// Query for the newly created connection in simd
	connections, err = relayer.GetConnections(ctx, eRep, simd.Config().ChainID)
	require.NoError(t, err)
	// localhost is always a connection in new version of ibc-go
	require.Equal(t, 2, len(connections))
	simdConnection := connections[0]
	require.Equal(t, "connection-0", simdConnection.ID)

	// Start the relayer and set the cleanup function.
	err = relayer.StartRelayer(ctx, eRep, pathName)
	require.NoError(t, err)

	t.Cleanup(
		func() {
			err := relayer.StopRelayer(ctx, eRep)
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

	// Create Channel between wasmd contract and simd
	version := fmt.Sprintf(`{"version":"ics27-1","controller_connection_id":"%s","host_connection_id":"%s","address":"","encoding":"json","tx_type":"sdk_multi_msg"}`, wasmdConnection.ID, simdConnection.ID)
	println("version: ", version)
	err = relayer.CreateChannel(ctx, eRep, pathName, ibc.CreateChannelOptions{
		SourcePortName: contractPort,
		DestPortName:   icatypes.HostPortID,
		Order:          ibc.Ordered,
		// asking the contract to generate the version by passing an empty string
		Version: version,
	})
	require.NoError(t, err)

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
	require.NoError(t, err)

	// Test if the handshake was successful
	wasmdChannels, err := relayer.GetChannels(ctx, eRep, wasmd.Config().ChainID)
	require.NoError(t, err)
	require.Equal(t, 1, len(wasmdChannels))

	wasmdChannel := wasmdChannels[0]
	println("wasmd channel: ", wasmdChannel)
	require.Equal(t, contractPort, wasmdChannel.PortID)
	require.Equal(t, icatypes.HostPortID, wasmdChannel.Counterparty.PortID)
	require.Equal(t, channeltypes.OPEN.String(), wasmdChannel.State)

	simdChannels, err := relayer.GetChannels(ctx, eRep, simd.Config().ChainID)
	require.NoError(t, err)
	require.Equal(t, 1, len(simdChannels))

	simdChannel := simdChannels[0]
	println("simd channel: ", simdChannel)
	require.Equal(t, icatypes.HostPortID, simdChannel.PortID)
	require.Equal(t, contractPort, simdChannel.Counterparty.PortID)
	// require.Equal(t, channeltypes.OPEN.String(), channel.State)
}
