package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	mysuite "github.com/srdtrk/cw-ica-controller/interchaintest/v2/testsuite"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types"

	sdkmath "cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos/wasm"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/stretchr/testify/suite"
)

type ContractTestSuite struct {
	mysuite.TestSuite
}

func TestWithContractTestSuite(t *testing.T) {
	suite.Run(t, new(ContractTestSuite))
}

func (s *ContractTestSuite) TestIcaControllerContract() {
	// Parallel indicates that this test is safe for parallel execution.
	// This is true since this is the only test in this file.
	t := s.T()
	// t.Parallel()

	ctx := context.Background()

	chainSpecs := []*interchaintest.ChainSpec{
		// -- WASMD --
		{
			ChainConfig: ibc.ChainConfig{
				Type:    "cosmos",
				Name:    "wasmd",
				ChainID: "wasmd-1",
				Images: []ibc.DockerImage{
					{
						Repository: "cosmwasm/wasmd", // FOR LOCAL IMAGE USE: Docker Image Name
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
	}

	// This starts the chains, relayer, creates the user accounts, and creates the ibc clients and connections.
	s.SetupSuite(ctx, chainSpecs)

	relayer := s.Relayer
	wasmd := s.ChainA
	simd := s.ChainB
	wasmdUser := s.UserA
	simdUser := s.UserB

	// Start the relayer and set the cleanup function.
	err := relayer.StartRelayer(ctx, s.ExecRep, s.PathName)
	s.Require().NoError(err)

	t.Cleanup(
		func() {
			err := relayer.StopRelayer(ctx, s.ExecRep)
			if err != nil {
				t.Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)

	// Upload and Instantiate the contract on wasmd:
	codeId, err := wasmd.StoreContract(ctx, wasmdUser.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)
	contractAddr, err := wasmd.InstantiateContract(ctx, wasmdUser.KeyName(), codeId, types.NewInstantiateMsg(nil), true)
	s.Require().NoError(err)

	contractPort := "wasm." + contractAddr

	// Test channel handshake between wasmd contract and simd:

	// Query for the newly created connection in wasmd
	wamsdConnections, err := s.Relayer.GetConnections(ctx, s.ExecRep, s.ChainA.Config().ChainID)
	s.Require().NoError(err)
	wasmdConnection := wamsdConnections[0]

	// Query for the newly created connection in simd
	simdConnections, err := s.Relayer.GetConnections(ctx, s.ExecRep, s.ChainB.Config().ChainID)
	s.Require().NoError(err)
	simdConnection := simdConnections[0]

	version := fmt.Sprintf(`{"version":"ics27-1","controller_connection_id":"%s","host_connection_id":"%s","address":"","encoding":"json","tx_type":"sdk_multi_msg"}`, wasmdConnection.ID, simdConnection.ID)
	err = relayer.CreateChannel(ctx, s.ExecRep, s.PathName, ibc.CreateChannelOptions{
		SourcePortName: contractPort,
		DestPortName:   icatypes.HostPortID,
		Order:          ibc.Ordered,
		// asking the contract to generate the version by passing an empty string
		Version: version,
	})
	s.Require().NoError(err)

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
	s.Require().NoError(err)

	// Test if the handshake was successful
	wasmdChannels, err := relayer.GetChannels(ctx, s.ExecRep, wasmd.Config().ChainID)
	s.Require().NoError(err)
	s.Require().Equal(1, len(wasmdChannels))

	wasmdChannel := wasmdChannels[0]
	t.Logf("wasmd channel: %s", toJSONString(wasmdChannel))
	s.Require().Equal(contractPort, wasmdChannel.PortID)
	s.Require().Equal(icatypes.HostPortID, wasmdChannel.Counterparty.PortID)
	s.Require().Equal(channeltypes.OPEN.String(), wasmdChannel.State)

	simdChannels, err := relayer.GetChannels(ctx, s.ExecRep, simd.Config().ChainID)
	s.Require().NoError(err)
	// I don't know why sometimes an extra channel is created in simd.
	// this is not related to the localhost connection, and is a failed
	// clone of the successful channel at index 0. I will log it for now.
	s.Require().Greater(len(simdChannels), 0)
	if len(simdChannels) > 1 {
		t.Logf("extra simd channel detected: %s", toJSONString(simdChannels[1]))
	}

	simdChannel := simdChannels[0]
	t.Logf("simd channel state: %s", toJSONString(simdChannel.State))
	s.Require().Equal(icatypes.HostPortID, simdChannel.PortID)
	s.Require().Equal(contractPort, simdChannel.Counterparty.PortID)
	s.Require().Equal(channeltypes.OPEN.String(), simdChannel.State)

	// Check contract's channel state
	queryResp := types.QueryResponse{}
	err = wasmd.QueryContract(ctx, contractAddr, types.NewGetChannelQueryMsg(), &queryResp)
	s.Require().NoError(err)

	contractChannelState, err := queryResp.GetChannelState()
	s.Require().NoError(err)

	t.Logf("contract's channel store after handshake: %s", toJSONString(contractChannelState))

	s.Require().Equal(wasmdChannel.State, contractChannelState.ChannelStatus)
	s.Require().Equal(wasmdChannel.Version, contractChannelState.Channel.Version)
	s.Require().Equal(wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionID)
	s.Require().Equal(wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelID)
	s.Require().Equal(wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortID)
	s.Require().Equal(wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelID)
	s.Require().Equal(wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortID)
	s.Require().Equal(wasmdChannel.Ordering, contractChannelState.Channel.Order)

	// Check contract state
	err = wasmd.QueryContract(ctx, contractAddr, types.NewGetContractStateQueryMsg(), &queryResp)
	s.Require().NoError(err)
	contractState, err := queryResp.GetContractState()
	s.Require().NoError(err)

	s.Require().Equal(wasmdUser.FormattedAddress(), contractState.Admin)
	s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelID)

	icaAddress := contractState.IcaInfo.IcaAddress
	t.Logf("ICA address after handshake: %s", icaAddress)

	// Fund the ICA address:
	err = simd.SendFunds(ctx, simdUser.KeyName(), ibc.WalletAmount{
		Address: icaAddress,
		Denom:   simd.Config().Denom,
		Amount:  1000000000,
	})
	s.Require().NoError(err)

	// wait for 2 blocks for the funds to be received
	err = testutil.WaitForBlocks(ctx, 2, wasmd, simd)
	s.Require().NoError(err)

	// Take predefined action on the ICA through the contract:
	_, err = wasmd.ExecuteContract(ctx, wasmdUser.KeyName(), contractAddr, types.NewSendPredefinedActionMsg(simdUser.FormattedAddress()))
	s.Require().NoError(err)

	err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
	s.Require().NoError(err)

	icaBalance, err := simd.GetBalance(ctx, icaAddress, simd.Config().Denom)
	s.Require().NoError(err)
	s.Require().Equal(int64(1000000000-100), icaBalance)

	// Check if contract callbacks were executed:
	err = wasmd.QueryContract(ctx, contractAddr, types.NewGetCallbackCounterQueryMsg(), &queryResp)
	s.Require().NoError(err)

	callbackCounter, err := queryResp.GetCallbackCounter()
	s.Require().NoError(err)

	s.Require().Equal(uint64(1), callbackCounter.Success)
	s.Require().Equal(uint64(0), callbackCounter.Error)

	// Send custom ICA messages through the contract:
	// Let's create a governance proposal on simd and deposit some funds to it.

	testProposal := govtypes.TextProposal{
		Title:       "IBC Gov Proposal",
		Description: "tokens for all!",
	}
	protoAny, err := codectypes.NewAnyWithValue(&testProposal)
	s.Require().NoError(err)
	proposalMsg := &govtypes.MsgSubmitProposal{
		Content:        protoAny,
		InitialDeposit: sdk.NewCoins(sdk.NewCoin(simd.Config().Denom, sdkmath.NewInt(5000))),
		Proposer:       icaAddress,
	}

	// Create deposit message:
	depositMsg := &govtypes.MsgDeposit{
		ProposalId: 1,
		Depositor:  icaAddress,
		Amount:     sdk.NewCoins(sdk.NewCoin(simd.Config().Denom, sdkmath.NewInt(10000000))),
	}

	// Create string message:
	customMsg := types.NewSendCustomIcaMessagesMsg(wasmd.Config().EncodingConfig.Codec, []sdk.Msg{proposalMsg, depositMsg}, nil, nil)

	// Execute them:
	_, err = wasmd.ExecuteContract(ctx, wasmdUser.KeyName(), contractAddr, customMsg)
	s.Require().NoError(err)

	err = testutil.WaitForBlocks(ctx, 4, wasmd, simd)
	s.Require().NoError(err)

	// Check if contract callbacks were executed:
	err = wasmd.QueryContract(ctx, contractAddr, types.NewGetCallbackCounterQueryMsg(), &queryResp)
	s.Require().NoError(err)

	callbackCounter, err = queryResp.GetCallbackCounter()
	s.Require().NoError(err)

	s.Require().Equal(uint64(2), callbackCounter.Success)
	s.Require().Equal(uint64(0), callbackCounter.Error)

	// Check if the proposal was created:
	proposal, err := simd.QueryProposal(ctx, "1")
	s.Require().NoError(err)
	s.Require().Equal(simd.Config().Denom, proposal.TotalDeposit[0].Denom)
	s.Require().Equal(fmt.Sprint(10000000+5000), proposal.TotalDeposit[0].Amount)
	// We do not check title and description of the proposal because this is a legacy proposal.
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
