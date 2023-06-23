package main

import (
	"context"
	"encoding/json"
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
						Version: "v0.40.2", // FOR LOCAL IMAGE USE: Docker Image Tag
					},
				},
				Bin:           "wasmd",
				Bech32Prefix:  "wasm",
				Denom:         "stake",
				GasPrices:     "0.00stake",
				GasAdjustment: 1.3,
				// cannot run wasmd commands without wasm encoding
				EncodingConfig:         wasm.WasmEncoding(),
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
				Denom:                  "stake",
				GasPrices:              "0.00stake",
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
	simdUser := users[1]

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
	t.Logf("wasmd channel: %s", toJSONString(wasmdChannel))
	require.Equal(t, contractPort, wasmdChannel.PortID)
	require.Equal(t, icatypes.HostPortID, wasmdChannel.Counterparty.PortID)
	require.Equal(t, channeltypes.OPEN.String(), wasmdChannel.State)

	simdChannels, err := relayer.GetChannels(ctx, eRep, simd.Config().ChainID)
	require.NoError(t, err)
	// I don't know why sometimes an extra channel is created in simd.
	// this is not related to the localhost connection, and is a failed
	// clone of the successful channel at index 0. I will log it for now.
	require.Greater(t, len(simdChannels), 0)
	if len(simdChannels) > 1 {
		t.Logf("extra simd channel detected: %s", toJSONString(simdChannels[1]))
	}

	simdChannel := simdChannels[0]
	t.Logf("simd channel state: %s", toJSONString(simdChannel.State))
	require.Equal(t, icatypes.HostPortID, simdChannel.PortID)
	require.Equal(t, contractPort, simdChannel.Counterparty.PortID)
	require.Equal(t, channeltypes.OPEN.String(), simdChannel.State)

	// Check contract's channel state
	queryResp := types.QueryResponse{}
	err = wasmd.QueryContract(ctx, contractAddr, types.NewGetChannelQueryMsg(), &queryResp)
	require.NoError(t, err)

	contractChannelState, err := queryResp.GetChannelState()
	require.NoError(t, err)

	t.Logf("contract's channel store after handshake: %s", toJSONString(contractChannelState))

	require.Equal(t, wasmdChannel.State, contractChannelState.ChannelStatus)
	require.Equal(t, wasmdChannel.Version, contractChannelState.Channel.Version)
	require.Equal(t, wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionID)
	require.Equal(t, wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelID)
	require.Equal(t, wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortID)
	require.Equal(t, wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelID)
	require.Equal(t, wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortID)
	require.Equal(t, wasmdChannel.Ordering, contractChannelState.Channel.Order)

	// Check contract state
	err = wasmd.QueryContract(ctx, contractAddr, types.NewGetContractStateQueryMsg(), &queryResp)
	require.NoError(t, err)
	contractState, err := queryResp.GetContractState()
	require.NoError(t, err)
	
	require.Equal(t, wasmdUser.FormattedAddress(), contractState.Admin)
	require.Equal(t, wasmdChannel.ChannelID, contractState.IcaInfo.ChannelID)

	icaAddress := contractState.IcaInfo.IcaAddress
	t.Logf("ICA address after handshake: %s", icaAddress)

	// Fund the ICA address:
	err = simd.SendFunds(ctx, simdUser.KeyName(), ibc.WalletAmount{
		Address: icaAddress,
		Denom:  simd.Config().Denom,
		Amount: 1000000000,
	})
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 3, wasmd, simd)
	require.NoError(t, err)
	
	// Take predefined action on the ICA through the contract:
	_, err = wasmd.ExecuteContract(ctx, wasmdUser.KeyName(), contractAddr, types.NewSendPredefinedActionMsg(simdUser.FormattedAddress()))
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
	require.NoError(t, err)

	icaBalance, err := simd.GetBalance(ctx, icaAddress, simd.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, int64(1000000000 - 100), icaBalance)

	// Check if contract callbacks were executed:
	err = wasmd.QueryContract(ctx, contractAddr, types.NewGetCallbackCounterQueryMsg(), &queryResp)
	require.NoError(t, err)

	callbackCounter, err := queryResp.GetCallbackCounter()
	require.NoError(t, err)

	require.Equal(t, uint64(1), callbackCounter.Success)
	require.Equal(t, uint64(0), callbackCounter.Error)

	// Send custom ICA messages through the contract:
	// Let's create a governance proposal on simd and deposit some funds to it.
}

// toJSONString returns a string representation of the given value
// by marshaling it to JSON. It panics if marshaling fails.
func toJSONString(v any) string {
	bz, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(bz)
}
