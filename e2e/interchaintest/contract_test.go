package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	controllertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	mysuite "github.com/srdtrk/cw-ica-controller/interchaintest/v2/testsuite"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types"
	callbackcounter "github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/callback-counter"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/icacontroller"
)

type ContractTestSuite struct {
	mysuite.TestSuite

	Contract *types.IcaContract
	// CallbackCounterContract is the address of the callback counter contract
	CallbackCounterContract *types.Contract
}

// SetupSuite calls the underlying TestSuite's SetupSuite method and initializes an empty contract
func (s *ContractTestSuite) SetupSuite(ctx context.Context, chainSpecs []*interchaintest.ChainSpec) {
	s.TestSuite.SetupSuite(ctx, chainSpecs)

	// Initialize an empty contract so that we can use the methods of the contract
	s.Contract = types.NewIcaContract(types.Contract{})
}

// SetupContractTestSuite starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
// sets up the contract and does the channel handshake for the contract test suite.
func (s *ContractTestSuite) SetupContractTestSuite(ctx context.Context, encoding icacontroller.TxEncoding, ordering icacontroller.IbcOrder) {
	s.SetupSuite(ctx, chainSpecs)

	codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/callback_counter.wasm")
	s.Require().NoError(err)

	callbackAddress, err := s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), codeId, callbackcounter.InstantiateMsg, true)
	s.Require().NoError(err)

	s.CallbackCounterContract = types.NewContract(callbackAddress, codeId, s.ChainA)

	codeId, err = s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	// Instantiate the contract with channel:
	instantiateMsg := icacontroller.InstantiateMsg{
		Owner: nil,
		ChannelOpenInitOptions: icacontroller.ChannelOpenInitOptions{
			ConnectionId:             s.ChainAConnID,
			CounterpartyConnectionId: s.ChainBConnID,
			CounterpartyPortId:       nil,
			TxEncoding:               &encoding,
			ChannelOrdering:          &ordering,
		},
		SendCallbacksTo: &callbackAddress,
	}

	err = s.Contract.Instantiate(ctx, s.UserA.KeyName(), s.ChainA, codeId, instantiateMsg, "--gas", "500000")
	s.Require().NoError(err)

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	contractState, err := types.QueryAnyMsg[icacontroller.State_2](
		ctx, &s.Contract.Contract,
		icacontroller.GetContractStateRequest,
	)
	s.Require().NoError(err)

	ownershipResponse, err := types.QueryAnyMsg[icacontroller.Ownership_for_String](ctx, &s.Contract.Contract, icacontroller.OwnershipRequest)
	s.Require().NoError(err)

	s.Require().NotEmpty(contractState.IcaInfo.IcaAddress)
	s.Contract.SetIcaAddress(contractState.IcaInfo.IcaAddress)

	s.Require().Equal(s.UserA.FormattedAddress(), *ownershipResponse.Owner)
	s.Require().Nil(ownershipResponse.PendingOwner)
	s.Require().Nil(ownershipResponse.PendingExpiry)
}

func TestWithContractTestSuite(t *testing.T) {
	suite.Run(t, new(ContractTestSuite))
}

func (s *ContractTestSuite) TestIcaContractChannelHandshake_Ordered_Protobuf() {
	s.IcaContractChannelHandshakeTest_WithEncodingAndOrdering(icacontroller.TxEncoding_Proto3_Value, icacontroller.IbcOrder_OrderOrdered)
}

func (s *ContractTestSuite) TestIcaContractChannelHandshake_Ordered_Proto3Json() {
	s.IcaContractChannelHandshakeTest_WithEncodingAndOrdering(icacontroller.TxEncoding_Proto3Json_Value, icacontroller.IbcOrder_OrderOrdered)
}

func (s *ContractTestSuite) TestIcaContractChannelHandshake_Unordered_Protobuf() {
	s.IcaContractChannelHandshakeTest_WithEncodingAndOrdering(icacontroller.TxEncoding_Proto3_Value, icacontroller.IbcOrder_OrderUnordered)
}

func (s *ContractTestSuite) TestIcaContractChannelHandshake_Unordered_Proto3Json() {
	s.IcaContractChannelHandshakeTest_WithEncodingAndOrdering(icacontroller.TxEncoding_Proto3Json_Value, icacontroller.IbcOrder_OrderUnordered)
}

func (s *ContractTestSuite) IcaContractChannelHandshakeTest_WithEncodingAndOrdering(encoding icacontroller.TxEncoding, ordering icacontroller.IbcOrder) {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, encoding, ordering)
	wasmd, simd := s.ChainA, s.ChainB

	s.Run("TestChannelHandshakeSuccess", func() {
		// Test if the handshake was successful
		wasmdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, wasmd.Config().ChainID)
		s.Require().NoError(err)
		s.Require().Equal(1, len(wasmdChannels))

		wasmdChannel := wasmdChannels[0]
		s.T().Logf("wasmd channel: %s", toJSONString(wasmdChannel))
		s.Require().Equal(s.Contract.Port(), wasmdChannel.PortID)
		s.Require().Equal(icatypes.HostPortID, wasmdChannel.Counterparty.PortID)
		s.Require().Equal(channeltypes.OPEN.String(), wasmdChannel.State)
		s.Require().Equal(string(ordering), wasmdChannel.Ordering)

		simdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, simd.Config().ChainID)
		s.Require().NoError(err)
		// I don't know why sometimes an extra channel is created in simd.
		// this is not related to the localhost connection, and is a failed
		// clone of the successful channel at index 0. I will log it for now.
		s.Require().Greater(len(simdChannels), 0)
		if len(simdChannels) > 1 {
			s.T().Logf("extra simd channels detected: %s", toJSONString(simdChannels))
		}

		simdChannel := simdChannels[0]
		s.T().Logf("simd channel state: %s", toJSONString(simdChannel.State))
		s.Require().Equal(icatypes.HostPortID, simdChannel.PortID)
		s.Require().Equal(s.Contract.Port(), simdChannel.Counterparty.PortID)
		s.Require().Equal(channeltypes.OPEN.String(), simdChannel.State)
		s.Require().Equal(string(ordering), simdChannel.Ordering)

		// Check contract's channel state
		contractChannelState, err := types.QueryAnyMsg[icacontroller.State](ctx, &s.Contract.Contract, icacontroller.GetChannelRequest)
		s.Require().NoError(err)

		s.T().Logf("contract's channel store after handshake: %s", toJSONString(contractChannelState))

		s.Require().Equal(wasmdChannel.State, string(contractChannelState.ChannelStatus))
		s.Require().Equal(wasmdChannel.Version, contractChannelState.Channel.Version)
		s.Require().Equal(wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionId)
		s.Require().Equal(wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelId)
		s.Require().Equal(wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortId)
		s.Require().Equal(wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelId)
		s.Require().Equal(wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortId)
		s.Require().Equal(wasmdChannel.Ordering, string(contractChannelState.Channel.Order))

		// Check contract state
		contractState, err := types.QueryAnyMsg[icacontroller.State_2](
			ctx, &s.Contract.Contract,
			icacontroller.GetContractStateRequest,
		)
		s.Require().NoError(err)

		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelId)
	})
}

// This tests that the relayer cannot create a channel with the contract's port.
func (s *ContractTestSuite) TestIcaRelayerInstantiatedChannelHandshake() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, and creates the ibc clients and connections.
	s.SetupSuite(ctx, chainSpecs)
	wasmd := s.ChainA
	wasmdUser := s.UserA

	var err error
	// Upload and Instantiate the contract on wasmd:
	codeId, err := wasmd.StoreContract(ctx, wasmdUser.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	// Instantiate the contract with incorrect counterparty connection id:
	instantiateMsg := icacontroller.InstantiateMsg{
		Owner: nil,
		ChannelOpenInitOptions: icacontroller.ChannelOpenInitOptions{
			ConnectionId:             s.ChainAConnID,
			CounterpartyConnectionId: "connection-123",
			CounterpartyPortId:       nil,
			TxEncoding:               nil,
		},
		SendCallbacksTo: nil,
	}

	err = s.Contract.Instantiate(ctx, wasmdUser.KeyName(), wasmd, codeId, instantiateMsg, "--gas", "500000")
	s.Require().NoError(err)

	version := fmt.Sprintf(`{"version":"%s","controller_connection_id":"%s","host_connection_id":"%s","address":"","encoding":"%s","tx_type":"%s"}`, icatypes.Version, s.ChainAConnID, s.ChainBConnID, icatypes.EncodingProtobuf, icatypes.TxTypeSDKMultiMsg)
	err = s.Relayer.CreateChannel(ctx, s.ExecRep, s.PathName, ibc.CreateChannelOptions{
		SourcePortName: s.Contract.Port(),
		DestPortName:   icatypes.HostPortID,
		Order:          ibc.Ordered,
		Version:        version,
	})
	s.Require().Error(err)
}

func (s *ContractTestSuite) TestRecoveredIcaContractInstantiatedChannelHandshake() {
	ctx := context.Background()

	s.SetupSuite(ctx, chainSpecs)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser := s.UserA

	codeId, err := wasmd.StoreContract(ctx, wasmdUser.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	s.Run("TestChannelHandshakeFail: invalid connection id", func() {
		// Instantiate the contract with channel:
		instantiateMsg := icacontroller.InstantiateMsg{
			Owner: nil,
			ChannelOpenInitOptions: icacontroller.ChannelOpenInitOptions{
				ConnectionId:             "invalid",
				CounterpartyConnectionId: s.ChainBConnID,
				CounterpartyPortId:       nil,
				TxEncoding:               nil,
			},
			SendCallbacksTo: nil,
		}

		err = s.Contract.Instantiate(ctx, wasmdUser.KeyName(), wasmd, codeId, instantiateMsg, "--gas", "500000")
		s.Require().ErrorContains(err, "submessages: invalid connection hop ID")
	})

	s.Run("TestChannelHandshakeFail: invalid counterparty connection id", func() {
		// Instantiate the contract with channel:
		instantiateMsg := icacontroller.InstantiateMsg{
			Owner: nil,
			ChannelOpenInitOptions: icacontroller.ChannelOpenInitOptions{
				ConnectionId:             s.ChainAConnID,
				CounterpartyConnectionId: "connection-123",
				CounterpartyPortId:       nil,
				TxEncoding:               nil,
			},
			SendCallbacksTo: nil,
		}

		err = s.Contract.Instantiate(ctx, wasmdUser.KeyName(), wasmd, codeId, instantiateMsg, "--gas", "500000")
		s.Require().NoError(err)
	})

	s.Run("TestChannelHandshakeSuccessAfterFail", func() {
		createChannelMsg := icacontroller.ExecuteMsg{
			CreateChannel: &icacontroller.ExecuteMsg_CreateChannel{
				ChannelOpenInitOptions: &icacontroller.ChannelOpenInitOptions{
					ConnectionId:             s.ChainAConnID,
					CounterpartyConnectionId: s.ChainBConnID,
					CounterpartyPortId:       nil,
					TxEncoding:               nil,
				},
			},
		}

		err = s.Contract.Execute(ctx, wasmdUser.KeyName(), createChannelMsg, "--gas", "500000")
		s.Require().NoError(err)

		// Wait for the channel to get set up
		err = testutil.WaitForBlocks(ctx, 9, s.ChainA, s.ChainB)
		s.Require().NoError(err)

		// Test if the handshake was successful
		wasmdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, wasmd.Config().ChainID)
		s.Require().NoError(err)
		s.Require().Equal(2, len(wasmdChannels))

		wasmdChannel := wasmdChannels[1]
		s.T().Logf("wasmd channel: %s", toJSONString(wasmdChannel))
		s.Require().Equal(s.Contract.Port(), wasmdChannel.PortID)
		s.Require().Equal(icatypes.HostPortID, wasmdChannel.Counterparty.PortID)
		s.Require().Equal(channeltypes.OPEN.String(), wasmdChannel.State)

		simdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, simd.Config().ChainID)
		s.Require().NoError(err)
		// I don't know why sometimes an extra channel is created in simd.
		// this is not related to the localhost connection, and is a failed
		// clone of the successful channel at index 0. I will log it for now.
		s.Require().Greater(len(simdChannels), 0)
		if len(simdChannels) > 1 {
			s.T().Logf("extra simd channels detected: %s", toJSONString(simdChannels))
		}

		simdChannel := simdChannels[0]
		s.T().Logf("simd channel state: %s", toJSONString(simdChannel.State))
		s.Require().Equal(icatypes.HostPortID, simdChannel.PortID)
		s.Require().Equal(s.Contract.Port(), simdChannel.Counterparty.PortID)
		s.Require().Equal(channeltypes.OPEN.String(), simdChannel.State)

		// Check contract's channel state
		contractChannelState, err := types.QueryAnyMsg[icacontroller.State](ctx, &s.Contract.Contract, icacontroller.GetChannelRequest)
		s.Require().NoError(err)

		s.T().Logf("contract's channel store after handshake: %s", toJSONString(contractChannelState))

		s.Require().Equal(wasmdChannel.State, string(contractChannelState.ChannelStatus))
		s.Require().Equal(wasmdChannel.Version, contractChannelState.Channel.Version)
		s.Require().Equal(wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionId)
		s.Require().Equal(wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelId)
		s.Require().Equal(wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortId)
		s.Require().Equal(wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelId)
		s.Require().Equal(wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortId)
		s.Require().Equal(wasmdChannel.Ordering, string(contractChannelState.Channel.Order))

		// Check contract state
		contractState, err := types.QueryAnyMsg[icacontroller.State_2](
			ctx, &s.Contract.Contract,
			icacontroller.GetContractStateRequest,
		)
		s.Require().NoError(err)

		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelId)
	})
}

func (s *ContractTestSuite) TestIcaContractExecution_Ordered_Proto3Json() {
	s.IcaContractExecutionTestWithEncoding(icacontroller.TxEncoding_Proto3Json_Value, icacontroller.IbcOrder_OrderOrdered)
}

func (s *ContractTestSuite) TestIcaContractExecution_Unordered_Protobuf() {
	s.IcaContractExecutionTestWithEncoding(icacontroller.TxEncoding_Proto3_Value, icacontroller.IbcOrder_OrderUnordered)
}

func (s *ContractTestSuite) IcaContractExecutionTestWithEncoding(encoding icacontroller.TxEncoding, ordering icacontroller.IbcOrder) {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, encoding, ordering)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser := s.UserA

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.Contract.IcaAddress)

	s.Run(fmt.Sprintf("TestSendCustomIcaMessagesSuccess-%s", encoding), func() {
		// Send custom ICA messages through the contract:
		// Let's create a governance proposal on simd and deposit some funds to it.
		govAddress := s.GetModuleAddress(ctx, simd, govtypes.ModuleName)

		testProposal := &controllertypes.MsgUpdateParams{
			Signer: govAddress,
			Params: controllertypes.Params{
				ControllerEnabled: false,
			},
		}

		proposalMsg, err := govv1.NewMsgSubmitProposal(
			[]sdk.Msg{testProposal},
			sdk.NewCoins(sdk.NewCoin(simd.Config().Denom, sdkmath.NewInt(10_000_000))),
			s.Contract.IcaAddress, "e2e", "e2e", "e2e", false,
		)
		s.Require().NoError(err)

		intialBalance, err := simd.GetBalance(ctx, s.Contract.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)

		// Execute the contract:
		sendCustomIcaMsg := icacontroller.NewExecuteMsg_SendCustomIcaMessages_FromProto(
			simd.Config().EncodingConfig.Codec,
			[]proto.Message{proposalMsg},
			string(encoding), nil, nil,
		)
		err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCustomIcaMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(1), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		// Check if the proposal was created:
		proposalResp, err := mysuite.GRPCQuery[govv1.QueryProposalResponse](ctx, simd, &govv1.QueryProposalRequest{
			ProposalId: 1,
		})
		s.Require().NoError(err)
		s.Require().Equal("e2e", proposalResp.Proposal.Title)

		postBalance, err := simd.GetBalance(ctx, s.Contract.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(intialBalance.Sub(sdkmath.NewInt(10_000_000)), postBalance)
	})

	s.Run(fmt.Sprintf("TestSendCosmosMsgsSuccess-%s", encoding), func() {
		intialBalance, err := simd.GetBalance(ctx, s.Contract.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)

		validator, err := simd.Validators[0].KeyBech32(ctx, "validator", "val")
		s.Require().NoError(err)

		// Stake some tokens through CosmosMsgs:
		stakeCosmosMsg := icacontroller.CosmosMsg_for_Empty{
			Staking: &icacontroller.CosmosMsg_for_Empty_Staking{
				Delegate: &icacontroller.StakingMsg_Delegate{
					Validator: validator,
					Amount: icacontroller.Coin{
						Denom:  simd.Config().Denom,
						Amount: "10000000",
					},
				},
			},
		}
		// Vote on the proposal through CosmosMsgs:
		voteCosmosMsg := icacontroller.CosmosMsg_for_Empty{
			Gov: &icacontroller.CosmosMsg_for_Empty_Gov{
				Vote: &icacontroller.GovMsg_Vote{
					ProposalId: 1,
					Vote:       "yes",
				},
			},
		}

		// Execute the contract:
		sendCosmosMsgsExecMsg := icacontroller.ExecuteMsg{
			SendCosmosMsgs: &icacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []icacontroller.CosmosMsg_for_Empty{stakeCosmosMsg, voteCosmosMsg},
			},
		}
		err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(2), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		// Check if the delegation was successful:
		postBalance, err := simd.GetBalance(ctx, s.Contract.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(intialBalance.Sub(sdkmath.NewInt(10_000_000)), postBalance)

		delRequest := stakingtypes.QueryDelegationRequest{
			DelegatorAddr: s.Contract.IcaAddress,
			ValidatorAddr: validator,
		}
		delResp, err := mysuite.GRPCQuery[stakingtypes.QueryDelegationResponse](ctx, simd, &delRequest)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(10_000_000), delResp.DelegationResponse.Balance.Amount)

		// Check if the vote was successful:
		voteRequest := govv1.QueryVoteRequest{
			ProposalId: 1,
			Voter:      s.Contract.IcaAddress,
		}
		voteResp, err := mysuite.GRPCQuery[govv1.QueryVoteResponse](ctx, simd, &voteRequest)
		s.Require().NoError(err)
		s.Require().Len(voteResp.Vote.Options, 1)
		s.Require().Equal(govv1.OptionYes, voteResp.Vote.Options[0].Option)
		s.Require().Equal(sdkmath.LegacyNewDec(1).String(), voteResp.Vote.Options[0].Weight)
	})

	s.Run(fmt.Sprintf("TestSendCustomIcaMessagesError-%s", encoding), func() {
		// Test erroneous callback:
		// Send incorrect custom ICA messages through the contract:
		badMessage := base64.StdEncoding.EncodeToString([]byte("bad message"))
		badCustomMsg := `{"send_custom_ica_messages":{"messages":"` + badMessage + `"}}`

		// Execute the contract:
		err := s.Contract.ExecAnyMsg(ctx, wasmdUser.KeyName(), badCustomMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)
		s.Require().Equal(uint64(2), callbackCounter.Success)
		s.Require().Equal(uint64(1), callbackCounter.Error)
		s.Require().Equal(uint64(0), callbackCounter.Timeout)
	})
}

func (s *ContractTestSuite) TestSendCosmosMsgs_Ordered_Proto3Json() {
	s.SendCosmosMsgsTestWithEncoding(icacontroller.TxEncoding_Proto3Json_Value, icacontroller.IbcOrder_OrderOrdered)
}

func (s *ContractTestSuite) TestSendCosmosMsgs_Unordered_Protobuf() {
	s.SendCosmosMsgsTestWithEncoding(icacontroller.TxEncoding_Proto3_Value, icacontroller.IbcOrder_OrderUnordered)
}

// SendCosmosMsgsTestWithEncoding tests some more CosmosMsgs that are not covered by the IcaContractExecutionTestWithEncoding.
// The following CosmosMsgs are tested here:
//
// - Bank::Send
// - Stargate
// - VoteWeighted
// - FundCommunityPool
// - SetWithdrawAddress
func (s *ContractTestSuite) SendCosmosMsgsTestWithEncoding(encoding icacontroller.TxEncoding, ordering icacontroller.IbcOrder) {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, encoding, ordering)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser := s.UserA
	simdUser := s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.Contract.IcaAddress)

	s.Run(fmt.Sprintf("TestStargate-%s", encoding), func() {
		// Send custom ICA messages through the contract:
		// Let's create a governance proposal on simd and deposit some funds to it.
		govAddress := s.GetModuleAddress(ctx, simd, govtypes.ModuleName)

		testProposal := controllertypes.MsgUpdateParams{
			Signer: govAddress,
			Params: controllertypes.Params{
				ControllerEnabled: false,
			},
		}

		proposalMsg, err := govv1.NewMsgSubmitProposal(
			[]sdk.Msg{&testProposal},
			sdk.NewCoins(sdk.NewCoin(simd.Config().Denom, sdkmath.NewInt(10_000_000))),
			s.Contract.IcaAddress, "e2e", "e2e", "e2e", false,
		)
		s.Require().NoError(err)

		initialBalance, err := simd.GetBalance(ctx, s.Contract.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)

		if encoding == icatypes.EncodingProtobuf {
			// Execute the contract:
			sendStargateMsg := icacontroller.NewExecuteMsg_SendCosmosMsgs_FromProto(
				[]proto.Message{proposalMsg}, nil, nil,
			)
			err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendStargateMsg)
			s.Require().NoError(err)
		} else if encoding == icatypes.EncodingProto3JSON {
			sendCustomIcaMsg := icacontroller.NewExecuteMsg_SendCustomIcaMessages_FromProto(
				simd.Config().EncodingConfig.Codec,
				[]proto.Message{proposalMsg},
				icatypes.EncodingProto3JSON, nil, nil,
			)
			err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCustomIcaMsg)
			s.Require().NoError(err)
		}

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(1), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		// Check if the proposal was created:
		proposalResp, err := mysuite.GRPCQuery[govv1.QueryProposalResponse](ctx, simd, &govv1.QueryProposalRequest{
			ProposalId: 1,
		})
		s.Require().NoError(err)
		s.Require().Equal("e2e", proposalResp.Proposal.Title)

		postBalance, err := simd.GetBalance(ctx, s.Contract.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(initialBalance.Sub(sdkmath.NewInt(10_000_000)), postBalance)
	})

	s.Run(fmt.Sprintf("TestDelegateAndVoteWeightedAndCommunityPool-%s", encoding), func() {
		intialBalance, err := simd.GetBalance(ctx, s.Contract.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)

		validator, err := simd.Validators[0].KeyBech32(ctx, "validator", "val")
		s.Require().NoError(err)

		// Stake some tokens through CosmosMsgs:
		stakeCosmosMsg := icacontroller.CosmosMsg_for_Empty{
			Staking: &icacontroller.CosmosMsg_for_Empty_Staking{
				Delegate: &icacontroller.StakingMsg_Delegate{
					Validator: validator,
					Amount: icacontroller.Coin{
						Denom:  simd.Config().Denom,
						Amount: "10000000",
					},
				},
			},
		}
		// Vote on the proposal through CosmosMsgs:
		voteCosmosMsg := icacontroller.CosmosMsg_for_Empty{
			Gov: &icacontroller.CosmosMsg_for_Empty_Gov{
				VoteWeighted: &icacontroller.GovMsg_VoteWeighted{
					ProposalId: 1,
					Options: []icacontroller.WeightedVoteOption{
						{
							Option: "yes",
							Weight: "0.5",
						},
						{
							Option: "abstain",
							Weight: "0.5",
						},
					},
				},
			},
		}

		// Fund the community pool through CosmosMsgs:
		fundPoolCosmosMsg := icacontroller.CosmosMsg_for_Empty{
			Distribution: &icacontroller.CosmosMsg_for_Empty_Distribution{
				FundCommunityPool: &icacontroller.DistributionMsg_FundCommunityPool{
					Amount: []icacontroller.Coin{{
						Denom:  simd.Config().Denom,
						Amount: "10000000",
					}},
				},
			},
		}

		// Execute the contract:
		sendCosmosMsgsExecMsg := icacontroller.ExecuteMsg{
			SendCosmosMsgs: &icacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []icacontroller.CosmosMsg_for_Empty{stakeCosmosMsg, voteCosmosMsg, fundPoolCosmosMsg},
			},
		}
		err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(2), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		// Check if the delegation was successful:
		postBalance, err := simd.GetBalance(ctx, s.Contract.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(intialBalance.Sub(sdkmath.NewInt(20_000_000)), postBalance)

		delRequest := stakingtypes.QueryDelegationRequest{
			DelegatorAddr: s.Contract.IcaAddress,
			ValidatorAddr: validator,
		}
		delResp, err := mysuite.GRPCQuery[stakingtypes.QueryDelegationResponse](ctx, simd, &delRequest)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(10_000_000), delResp.DelegationResponse.Balance.Amount)

		// Check if the vote was successful:
		voteRequest := govv1.QueryVoteRequest{
			ProposalId: 1,
			Voter:      s.Contract.IcaAddress,
		}
		voteResp, err := mysuite.GRPCQuery[govv1.QueryVoteResponse](ctx, simd, &voteRequest)
		s.Require().NoError(err)
		s.Require().Len(voteResp.Vote.Options, 2)

		expWeight, err := sdkmath.LegacyNewDecFromStr("0.5")
		s.Require().NoError(err)
		actualWeight, err := sdkmath.LegacyNewDecFromStr(voteResp.Vote.Options[0].Weight)
		s.Require().NoError(err)
		actualWeight2, err := sdkmath.LegacyNewDecFromStr(voteResp.Vote.Options[1].Weight)
		s.Require().NoError(err)

		s.Require().Equal(govv1.OptionYes, voteResp.Vote.Options[0].Option)
		s.Require().True(expWeight.Equal(actualWeight))
		s.Require().Equal(govv1.OptionAbstain, voteResp.Vote.Options[1].Option)
		s.Require().True(expWeight.Equal(actualWeight2))
	})

	s.Run(fmt.Sprintf("TestSendAndSetWithdrawAddress-%s", encoding), func() {
		initialBalance, err := simd.GetBalance(ctx, s.Contract.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)

		// Send some tokens to the simdUser from the ICA address
		sendMsg := icacontroller.CosmosMsg_for_Empty{
			Bank: &icacontroller.CosmosMsg_for_Empty_Bank{
				Send: &icacontroller.BankMsg_Send{
					ToAddress: simdUser.FormattedAddress(),
					Amount: []icacontroller.Coin{
						{
							Denom:  simd.Config().Denom,
							Amount: "1000000",
						},
					},
				},
			},
		}

		// Set the withdraw address to the simdUser
		setWithdrawAddressMsg := icacontroller.CosmosMsg_for_Empty{
			Distribution: &icacontroller.CosmosMsg_for_Empty_Distribution{
				SetWithdrawAddress: &icacontroller.DistributionMsg_SetWithdrawAddress{
					Address: simdUser.FormattedAddress(),
				},
			},
		}

		// Execute the contract:
		sendCosmosMsgsExecMsg := icacontroller.ExecuteMsg{
			SendCosmosMsgs: &icacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []icacontroller.CosmosMsg_for_Empty{sendMsg, setWithdrawAddressMsg},
			},
		}
		err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)
		s.Require().Equal(uint64(3), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		// Check if the send was successful:
		postBalance, err := simd.GetBalance(ctx, s.Contract.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(1_000_000), initialBalance.Sub(postBalance))
	})
}

func (s *ContractTestSuite) TestIcaContractTimeoutPacket_Ordered_Proto3Json() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, icacontroller.TxEncoding_Proto3Json_Value, icacontroller.IbcOrder_OrderOrdered)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser, _ := s.UserA, s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.Contract.IcaAddress)

	contractState, err := types.QueryAnyMsg[icacontroller.State_2](
		ctx, &s.Contract.Contract,
		icacontroller.GetContractStateRequest,
	)
	s.Require().NoError(err)

	var simdChannelsLen int
	s.Run("TestTimeout", func() {
		// We will send a message to the host that will timeout after 3 seconds.
		// You cannot use 0 seconds because block timestamp will be greater than the timeout timestamp which is not allowed.
		// Host will not be able to respond to this message in time.

		// Stop the relayer so that the host cannot respond to the message:
		err := s.Relayer.StopRelayer(ctx, s.ExecRep)
		s.Require().NoError(err)

		time.Sleep(5 * time.Second)

		timeout := int(3)
		// Execute the contract:
		sendCustomIcaMsg := icacontroller.NewExecuteMsg_SendCustomIcaMessages_FromProto(
			simd.Config().EncodingConfig.Codec, []proto.Message{},
			icatypes.EncodingProto3JSON, nil, &timeout,
		)
		err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCustomIcaMsg)
		s.Require().NoError(err)

		// Wait until timeout:
		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		err = s.Relayer.StartRelayer(ctx, s.ExecRep)
		s.Require().NoError(err)

		// Wait until timeout packet is received:
		err = testutil.WaitForBlocks(ctx, 2, wasmd, simd)
		s.Require().NoError(err)

		// Flush to make sure the channel is closed in simd:
		err = s.Relayer.Flush(ctx, s.ExecRep, s.PathName, contractState.IcaInfo.ChannelId)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 2, wasmd, simd)
		s.Require().NoError(err)

		// Check if channel was closed:
		wasmdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, wasmd.Config().ChainID)
		s.Require().NoError(err)
		s.Require().Equal(1, len(wasmdChannels))
		s.Require().Equal(channeltypes.CLOSED.String(), wasmdChannels[0].State)

		simdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, simd.Config().ChainID)
		s.Require().NoError(err)
		// sometimes there is a redundant channel for unknown reasons
		simdChannelsLen = len(simdChannels)
		s.Require().Greater(simdChannelsLen, 0)
		s.Require().Equal(channeltypes.CLOSED.String(), simdChannels[0].State)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)
		s.Require().Equal(uint64(0), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)
		s.Require().Equal(uint64(1), callbackCounter.Timeout)

		// Check if contract channel state was updated:
		contractChannelState, err := types.QueryAnyMsg[icacontroller.State](ctx, &s.Contract.Contract, icacontroller.GetChannelRequest)
		s.Require().NoError(err)
		s.Require().Equal(icacontroller.Status_StateClosed_Value, contractChannelState.ChannelStatus)
	})

	s.Run("TestChannelReopening", func() {
		// Reopen the channel:
		createChannelMsg := icacontroller.ExecuteMsg{
			CreateChannel: &icacontroller.ExecuteMsg_CreateChannel{
				ChannelOpenInitOptions: nil,
			},
		}

		err := s.Contract.Execute(ctx, wasmdUser.KeyName(), createChannelMsg, "--gas", "500000")
		s.Require().NoError(err)

		// Wait for the channel to get set up
		err = testutil.WaitForBlocks(ctx, 10, s.ChainA, s.ChainB)
		s.Require().NoError(err)

		// Check if a new channel was opened in simd
		simdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, simd.Config().ChainID)
		s.Require().NoError(err)
		// An extra channel may be created in simd for unknown reasons.
		s.Require().Greater(len(simdChannels), simdChannelsLen)
		s.Require().Equal(channeltypes.OPEN.String(), simdChannels[simdChannelsLen].State)
		simdChannelsLen = len(simdChannels)

		// Check if a new channel was opened in wasmd:
		wasmdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, wasmd.Config().ChainID)
		s.Require().NoError(err)
		s.Require().Equal(2, len(wasmdChannels))
		wasmdChannel := wasmdChannels[1]
		s.Require().Equal(channeltypes.OPEN.String(), wasmdChannel.State)

		// Check if contract channel state was updated:
		contractChannelState, err := types.QueryAnyMsg[icacontroller.State](ctx, &s.Contract.Contract, icacontroller.GetChannelRequest)
		s.Require().NoError(err)
		s.Require().Equal(icacontroller.Status_StateOpen_Value, contractChannelState.ChannelStatus)
		s.Require().Equal(wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionId)
		s.Require().Equal(wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelId)
		s.Require().Equal(wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortId)
		s.Require().Equal(wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelId)
		s.Require().Equal(wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortId)
		s.Require().Equal(wasmdChannel.Ordering, string(contractChannelState.Channel.Order))

		contractState, err := types.QueryAnyMsg[icacontroller.State_2](
			ctx, &s.Contract.Contract,
			icacontroller.GetContractStateRequest,
		)
		s.Require().NoError(err)
		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelId)
		s.Require().Equal(s.Contract.IcaAddress, contractState.IcaInfo.IcaAddress)

		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(0), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)
		s.Require().Equal(uint64(1), callbackCounter.Timeout)
	})

	s.Run("TestSendCustomIcaMessagesAfterReopen", func() {
		// Send custom ICA message through the contract:
		sendMsg := &banktypes.MsgSend{
			FromAddress: s.Contract.IcaAddress,
			ToAddress:   s.UserB.FormattedAddress(),
			Amount:      sdk.NewCoins(sdk.NewCoin(simd.Config().Denom, sdkmath.NewInt(100))),
		}

		// Execute the contract:
		sendCustomIcaMsg := icacontroller.NewExecuteMsg_SendCustomIcaMessages_FromProto(
			simd.Config().EncodingConfig.Codec,
			[]proto.Message{sendMsg},
			icatypes.EncodingProto3JSON, nil, nil,
		)
		err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCustomIcaMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 10, wasmd, simd)
		s.Require().NoError(err)

		icaBalance, err := simd.GetBalance(ctx, s.Contract.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(1000000000-100), icaBalance)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(1), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)
		s.Require().Equal(uint64(1), callbackCounter.Timeout)
	})
}

func (s *ContractTestSuite) TestIcaContractTimeoutPacket_Unordered_Protobuf() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, icacontroller.TxEncoding_Proto3_Value, icacontroller.IbcOrder_OrderUnordered)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser, _ := s.UserA, s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.Contract.IcaAddress)

	contractState, err := types.QueryAnyMsg[icacontroller.State_2](
		ctx, &s.Contract.Contract,
		icacontroller.GetContractStateRequest,
	)
	s.Require().NoError(err)

	var simdChannelsLen int
	s.Run("TestTimeout", func() {
		// We will send a message to the host that will timeout after 3 seconds.
		// You cannot use 0 seconds because block timestamp will be greater than the timeout timestamp which is not allowed.
		// Host will not be able to respond to this message in time.

		// Stop the relayer so that the host cannot respond to the message:
		err := s.Relayer.StopRelayer(ctx, s.ExecRep)
		s.Require().NoError(err)

		time.Sleep(5 * time.Second)

		timeout := int(3)
		// Execute the contract:
		sendCustomIcaMsg := icacontroller.NewExecuteMsg_SendCustomIcaMessages_FromProto(
			simd.Config().EncodingConfig.Codec, []proto.Message{},
			icatypes.EncodingProto3JSON, nil, &timeout,
		)
		err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCustomIcaMsg)
		s.Require().NoError(err)

		// Wait until timeout:
		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		err = s.Relayer.StartRelayer(ctx, s.ExecRep)
		s.Require().NoError(err)

		// Wait until timeout packet is received:
		err = testutil.WaitForBlocks(ctx, 2, wasmd, simd)
		s.Require().NoError(err)

		// Flush to make sure the channel is closed in simd:
		err = s.Relayer.Flush(ctx, s.ExecRep, s.PathName, contractState.IcaInfo.ChannelId)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 2, wasmd, simd)
		s.Require().NoError(err)

		// Check if channel is stil open:
		wasmdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, wasmd.Config().ChainID)
		s.Require().NoError(err)
		s.Require().Equal(1, len(wasmdChannels))
		s.Require().Equal(channeltypes.OPEN.String(), wasmdChannels[0].State)

		simdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, simd.Config().ChainID)
		s.Require().NoError(err)
		// sometimes there is a redundant channel for unknown reasons
		simdChannelsLen = len(simdChannels)
		s.Require().Greater(simdChannelsLen, 0)
		s.Require().Equal(channeltypes.OPEN.String(), simdChannels[0].State)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)
		s.Require().Equal(uint64(0), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)
		s.Require().Equal(uint64(1), callbackCounter.Timeout)

		// Check if contract channel state is still open:
		contractChannelState, err := types.QueryAnyMsg[icacontroller.State](ctx, &s.Contract.Contract, icacontroller.GetChannelRequest)
		s.Require().NoError(err)
		s.Require().Equal(icacontroller.Status_StateOpen_Value, contractChannelState.ChannelStatus)
	})

	s.Run("TestSendCustomIcaMessagesAfterTimeout", func() {
		// Send custom ICA message through the contract:
		sendMsg := &banktypes.MsgSend{
			FromAddress: s.Contract.IcaAddress,
			ToAddress:   s.UserB.FormattedAddress(),
			Amount:      sdk.NewCoins(sdk.NewCoin(simd.Config().Denom, sdkmath.NewInt(100))),
		}

		// Execute the contract:
		sendCustomIcaMsg := icacontroller.NewExecuteMsg_SendCustomIcaMessages_FromProto(
			simd.Config().EncodingConfig.Codec,
			[]proto.Message{sendMsg},
			icatypes.EncodingProtobuf, nil, nil,
		)
		err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCustomIcaMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		icaBalance, err := simd.GetBalance(ctx, s.Contract.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(1000000000-100), icaBalance)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(1), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)
		s.Require().Equal(uint64(1), callbackCounter.Timeout)
	})
}

func (s *ContractTestSuite) TestMigrateOrderedToUnordered() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, icacontroller.TxEncoding_Proto3Json_Value, icacontroller.IbcOrder_OrderOrdered)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser, _ := s.UserA, s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.Contract.IcaAddress)

	var simdChannelsLen int
	s.Run("TestCloseChannel", func() {
		// Close the channel:
		closeChannelMsg := icacontroller.ExecuteMsg{
			CloseChannel: &icacontroller.ExecuteMsg_CloseChannel{},
		}
		err := s.Contract.Execute(ctx, wasmdUser.KeyName(), closeChannelMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if channel was closed:
		wasmdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, wasmd.Config().ChainID)
		s.Require().NoError(err)
		s.Require().Equal(1, len(wasmdChannels))
		s.Require().Equal(channeltypes.CLOSED.String(), wasmdChannels[0].State)

		simdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, simd.Config().ChainID)
		s.Require().NoError(err)
		// sometimes there is a redundant channel for unknown reasons
		simdChannelsLen = len(simdChannels)
		s.Require().Greater(simdChannelsLen, 0)
		s.Require().Equal(channeltypes.CLOSED.String(), simdChannels[0].State)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)
		s.Require().Equal(uint64(0), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)
		s.Require().Equal(uint64(0), callbackCounter.Timeout)

		// Check if contract channel state was updated:
		contractChannelState, err := types.QueryAnyMsg[icacontroller.State](ctx, &s.Contract.Contract, icacontroller.GetChannelRequest)
		s.Require().NoError(err)
		s.Require().Equal(icacontroller.Status_StateClosed_Value, contractChannelState.ChannelStatus)
	})

	s.Run("TestChannelReopening", func() {
		// Reopen the channel:
		var (
			encoding = icacontroller.TxEncoding_Proto3_Value
			ordering = icacontroller.IbcOrder_OrderUnordered
		)

		createChannelMsg := icacontroller.ExecuteMsg{
			CreateChannel: &icacontroller.ExecuteMsg_CreateChannel{
				ChannelOpenInitOptions: &icacontroller.ChannelOpenInitOptions{
					ConnectionId:             s.ChainAConnID,
					CounterpartyConnectionId: s.ChainBConnID,
					CounterpartyPortId:       nil,
					TxEncoding:               &encoding,
					ChannelOrdering:          &ordering,
				},
			},
		}

		err := s.Contract.Execute(ctx, wasmdUser.KeyName(), createChannelMsg, "--gas", "500000")
		s.Require().NoError(err)

		// Wait for the channel to get set up
		err = testutil.WaitForBlocks(ctx, 8, s.ChainA, s.ChainB)
		s.Require().NoError(err)

		// Check if a new channel was opened in simd
		simdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, simd.Config().ChainID)
		s.Require().NoError(err)
		// An extra channel may be created in simd for unknown reasons.
		s.Require().Greater(len(simdChannels), simdChannelsLen)
		s.Require().Equal(channeltypes.OPEN.String(), simdChannels[simdChannelsLen].State)
		s.Require().Equal(channeltypes.UNORDERED.String(), simdChannels[simdChannelsLen].Ordering)
		simdChannelsLen = len(simdChannels)

		// Check if a new channel was opened in wasmd:
		wasmdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, wasmd.Config().ChainID)
		s.Require().NoError(err)
		s.Require().Equal(2, len(wasmdChannels))
		wasmdChannel := wasmdChannels[1]
		s.Require().Equal(channeltypes.OPEN.String(), wasmdChannel.State)
		s.Require().Equal(channeltypes.UNORDERED.String(), wasmdChannel.Ordering)

		// Check if contract channel state was updated:
		contractChannelState, err := types.QueryAnyMsg[icacontroller.State](ctx, &s.Contract.Contract, icacontroller.GetChannelRequest)
		s.Require().NoError(err)
		s.Require().Equal(icacontroller.Status_StateOpen_Value, contractChannelState.ChannelStatus)
		s.Require().Equal(wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionId)
		s.Require().Equal(wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelId)
		s.Require().Equal(wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortId)
		s.Require().Equal(wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelId)
		s.Require().Equal(wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortId)
		s.Require().Equal(wasmdChannel.Ordering, string(contractChannelState.Channel.Order))

		contractState, err := types.QueryAnyMsg[icacontroller.State_2](
			ctx, &s.Contract.Contract,
			icacontroller.GetContractStateRequest,
		)
		s.Require().NoError(err)
		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelId)
		s.Require().Equal(s.Contract.IcaAddress, contractState.IcaInfo.IcaAddress)

		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(0), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)
		s.Require().Equal(uint64(0), callbackCounter.Timeout)
	})

	s.Run("TestSendCustomIcaMessagesAfterReopen", func() {
		// Send custom ICA message through the contract:
		sendMsg := &banktypes.MsgSend{
			FromAddress: s.Contract.IcaAddress,
			ToAddress:   s.UserB.FormattedAddress(),
			Amount:      sdk.NewCoins(sdk.NewCoin(simd.Config().Denom, sdkmath.NewInt(100))),
		}

		// Execute the contract:
		sendCustomIcaMsg := icacontroller.NewExecuteMsg_SendCustomIcaMessages_FromProto(
			simd.Config().EncodingConfig.Codec,
			[]proto.Message{sendMsg},
			icatypes.EncodingProtobuf, nil, nil,
		)
		err := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCustomIcaMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 7, wasmd, simd)
		s.Require().NoError(err)

		icaBalance, err := simd.GetBalance(ctx, s.Contract.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(1000000000-100), icaBalance)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(1), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)
		s.Require().Equal(uint64(0), callbackCounter.Timeout)
	})
}

func (s *ContractTestSuite) TestCloseChannel_Protobuf_Unordered() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, icacontroller.TxEncoding_Proto3_Value, icacontroller.IbcOrder_OrderUnordered)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser, _ := s.UserA, s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.Contract.IcaAddress)

	s.Run("TestCloseChannel", func() {
		// Close the channel:
		closeChannelMsg := icacontroller.ExecuteMsg{
			CloseChannel: &icacontroller.ExecuteMsg_CloseChannel{},
		}
		err := s.Contract.Execute(ctx, wasmdUser.KeyName(), closeChannelMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if channel was closed:
		wasmdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, wasmd.Config().ChainID)
		s.Require().NoError(err)
		s.Require().Equal(1, len(wasmdChannels))
		s.Require().Equal(channeltypes.CLOSED.String(), wasmdChannels[0].State)

		simdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, simd.Config().ChainID)
		s.Require().NoError(err)
		// sometimes there is a redundant channel for unknown reasons
		simdChannelsLen := len(simdChannels)
		s.Require().Greater(simdChannelsLen, 0)
		s.Require().Equal(channeltypes.CLOSED.String(), simdChannels[0].State)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)
		s.Require().Equal(uint64(0), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)
		s.Require().Equal(uint64(0), callbackCounter.Timeout)

		// Check if contract channel state was updated:
		contractChannelState, err := types.QueryAnyMsg[icacontroller.State](ctx, &s.Contract.Contract, icacontroller.GetChannelRequest)
		s.Require().NoError(err)
		s.Require().Equal(icacontroller.Status_StateClosed_Value, contractChannelState.ChannelStatus)
	})
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
