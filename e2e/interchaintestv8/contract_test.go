package main

import (
	"context"
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
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	"github.com/srdtrk/go-codegen/e2esuite/v8/e2esuite"
	"github.com/srdtrk/go-codegen/e2esuite/v8/types/callbackcounter"
	"github.com/srdtrk/go-codegen/e2esuite/v8/types/cwicacontroller"
)

type ContractTestSuite struct {
	e2esuite.TestSuite

	// Contract is the representation of the ICA controller contract
	Contract *cwicacontroller.Contract
	// CallbackCounterContract is the representation of the callback counter contract
	CallbackCounterContract *callbackcounter.Contract

	// IcaContractToAddrMap is a map of ICA contract address to the address of ICA
	IcaContractToAddrMap map[string]string

	// this line is used by go-codegen # suite/contract
}

// SetupSuite calls the underlying TestSuite's SetupSuite method
func (s *ContractTestSuite) SetupSuite(ctx context.Context) {
	s.TestSuite.SetupSuite(ctx)

	s.IcaContractToAddrMap = make(map[string]string)
}

// SetupContractTestSuite starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
// sets up the contract and does the channel handshake for the contract test suite.
func (s *ContractTestSuite) SetupContractTestSuite(ctx context.Context, ordering cwicacontroller.IbcOrder) {
	s.SetupSuite(ctx)

	codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/callback_counter.wasm")
	s.Require().NoError(err)

	s.CallbackCounterContract, err = callbackcounter.Instantiate(ctx, s.UserA.KeyName(), codeId, "", s.ChainA, callbackcounter.InstantiateMsg{})
	s.Require().NoError(err)

	codeId, err = s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	// Instantiate the contract with channel:
	instantiateMsg := cwicacontroller.InstantiateMsg{
		Owner: nil,
		ChannelOpenInitOptions: cwicacontroller.ChannelOpenInitOptions{
			ConnectionId:             ibctesting.FirstConnectionID,
			CounterpartyConnectionId: ibctesting.FirstConnectionID,
			CounterpartyPortId:       nil,
			ChannelOrdering:          &ordering,
		},
		SendCallbacksTo: &s.CallbackCounterContract.Address,
	}

	s.Contract, err = cwicacontroller.Instantiate(ctx, s.UserA.KeyName(), codeId, "", s.ChainA, instantiateMsg, "--gas", "500000")
	s.Require().NoError(err)

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	contractState, err := s.Contract.QueryClient().GetContractState(ctx, &cwicacontroller.QueryMsg_GetContractState{})
	s.Require().NoError(err)

	ownershipResponse, err := s.Contract.QueryClient().Ownership(ctx, &cwicacontroller.QueryMsg_Ownership{})
	s.Require().NoError(err)
	s.Require().NotEmpty(contractState.IcaInfo.IcaAddress)

	s.IcaContractToAddrMap[s.Contract.Address] = contractState.IcaInfo.IcaAddress

	s.Require().Equal(s.UserA.FormattedAddress(), *ownershipResponse.Owner)
	s.Require().Nil(ownershipResponse.PendingOwner)
	s.Require().Nil(ownershipResponse.PendingExpiry)
}

func TestWithContractTestSuite(t *testing.T) {
	suite.Run(t, new(ContractTestSuite))
}

func (s *ContractTestSuite) TestIcaContractChannelHandshake_Ordered_Protobuf() {
	s.IcaContractChannelHandshakeTest_WithOrdering(cwicacontroller.IbcOrder_OrderOrdered)
}

func (s *ContractTestSuite) TestIcaContractChannelHandshake_Unordered_Protobuf() {
	s.IcaContractChannelHandshakeTest_WithOrdering(cwicacontroller.IbcOrder_OrderUnordered)
}

func (s *ContractTestSuite) IcaContractChannelHandshakeTest_WithOrdering(ordering cwicacontroller.IbcOrder) {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, ordering)
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
		contractChannelState, err := s.Contract.QueryClient().GetChannel(ctx, &cwicacontroller.QueryMsg_GetChannel{})
		s.Require().NoError(err)
		s.Require().Equal(wasmdChannel.State, string(contractChannelState.ChannelStatus))
		s.Require().Equal(wasmdChannel.Version, contractChannelState.Channel.Version)
		s.Require().Equal(wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionId)
		s.Require().Equal(wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelId)
		s.Require().Equal(wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortId)
		s.Require().Equal(wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelId)
		s.Require().Equal(wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortId)
		s.Require().Equal(wasmdChannel.Ordering, string(contractChannelState.Channel.Order))

		// Check contract state
		contractState, err := s.Contract.QueryClient().GetContractState(ctx, &cwicacontroller.QueryMsg_GetContractState{})
		s.Require().NoError(err)
		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelId)
	})
}

// This tests that the relayer cannot create a channel with the contract's port.
func (s *ContractTestSuite) TestIcaRelayerInstantiatedChannelHandshake() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, and creates the ibc clients and connections.
	s.SetupSuite(ctx)

	wasmd := s.ChainA
	wasmdUser := s.UserA

	// Upload and Instantiate the contract on wasmd:
	codeId, err := wasmd.StoreContract(ctx, wasmdUser.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	// Instantiate the contract with incorrect counterparty connection id:
	instantiateMsg := cwicacontroller.InstantiateMsg{
		Owner: nil,
		ChannelOpenInitOptions: cwicacontroller.ChannelOpenInitOptions{
			ConnectionId:             ibctesting.FirstConnectionID,
			CounterpartyConnectionId: "connection-123",
			CounterpartyPortId:       nil,
		},
		SendCallbacksTo: nil,
	}

	s.Contract, err = cwicacontroller.Instantiate(ctx, wasmdUser.KeyName(), codeId, "", wasmd, instantiateMsg, "--gas", "500000")
	s.Require().NoError(err)

	version := fmt.Sprintf(`{"version":"%s","controller_connection_id":"%s","host_connection_id":"%s","address":"","encoding":"%s","tx_type":"%s"}`, icatypes.Version, ibctesting.FirstConnectionID, ibctesting.FirstConnectionID, icatypes.EncodingProtobuf, icatypes.TxTypeSDKMultiMsg)
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

	s.SetupSuite(ctx)

	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser := s.UserA

	codeId, err := wasmd.StoreContract(ctx, wasmdUser.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	s.Run("TestChannelHandshakeFail: invalid connection id", func() {
		// Instantiate the contract with channel:
		instantiateMsg := cwicacontroller.InstantiateMsg{
			Owner: nil,
			ChannelOpenInitOptions: cwicacontroller.ChannelOpenInitOptions{
				ConnectionId:             "invalid",
				CounterpartyConnectionId: ibctesting.FirstConnectionID,
				CounterpartyPortId:       nil,
			},
			SendCallbacksTo: nil,
		}

		_, err = cwicacontroller.Instantiate(ctx, wasmdUser.KeyName(), codeId, "", wasmd, instantiateMsg, "--gas", "500000")
		s.Require().ErrorContains(err, "submessages: invalid connection hop ID")
	})

	s.Run("TestChannelHandshakeFail: invalid counterparty connection id", func() {
		// Instantiate the contract with channel:
		instantiateMsg := cwicacontroller.InstantiateMsg{
			Owner: nil,
			ChannelOpenInitOptions: cwicacontroller.ChannelOpenInitOptions{
				ConnectionId:             ibctesting.FirstConnectionID,
				CounterpartyConnectionId: "connection-123",
				CounterpartyPortId:       nil,
			},
			SendCallbacksTo: nil,
		}

		s.Contract, err = cwicacontroller.Instantiate(ctx, wasmdUser.KeyName(), codeId, "", wasmd, instantiateMsg, "--gas", "500000")
		s.Require().NoError(err)
	})

	s.Run("TestChannelHandshakeSuccessAfterFail", func() {
		createChannelMsg := cwicacontroller.ExecuteMsg{
			CreateChannel: &cwicacontroller.ExecuteMsg_CreateChannel{
				ChannelOpenInitOptions: &cwicacontroller.ChannelOpenInitOptions{
					ConnectionId:             ibctesting.FirstConnectionID,
					CounterpartyConnectionId: ibctesting.FirstConnectionID,
					CounterpartyPortId:       nil,
				},
			},
		}

		_, err = s.Contract.Execute(ctx, wasmdUser.KeyName(), createChannelMsg, "--gas", "500000")
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
		contractChannelState, err := s.Contract.QueryClient().GetChannel(ctx, &cwicacontroller.QueryMsg_GetChannel{})
		s.Require().NoError(err)
		s.Require().Equal(wasmdChannel.State, string(contractChannelState.ChannelStatus))
		s.Require().Equal(wasmdChannel.Version, contractChannelState.Channel.Version)
		s.Require().Equal(wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionId)
		s.Require().Equal(wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelId)
		s.Require().Equal(wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortId)
		s.Require().Equal(wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelId)
		s.Require().Equal(wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortId)
		s.Require().Equal(wasmdChannel.Ordering, string(contractChannelState.Channel.Order))

		// Check contract state
		contractState, err := s.Contract.QueryClient().GetContractState(ctx, &cwicacontroller.QueryMsg_GetContractState{})
		s.Require().NoError(err)
		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelId)
	})
}

func (s *ContractTestSuite) TestIcaContractExecution_Ordered_Protobuf() {
	s.IcaContractExecutionTestWithOrdering(cwicacontroller.IbcOrder_OrderOrdered)
}

func (s *ContractTestSuite) TestIcaContractExecution_Unordered_Protobuf() {
	s.IcaContractExecutionTestWithOrdering(cwicacontroller.IbcOrder_OrderUnordered)
}

func (s *ContractTestSuite) IcaContractExecutionTestWithOrdering(ordering cwicacontroller.IbcOrder) {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, ordering)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser, simdUser := s.UserA, s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaContractToAddrMap[s.Contract.Address])

	s.Run("TestStargateMsgSuccess", func() {
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
			s.IcaContractToAddrMap[s.Contract.Address], "e2e", "e2e", "e2e", false,
		)
		s.Require().NoError(err)

		intialBalance, err := simd.GetBalance(ctx, s.IcaContractToAddrMap[s.Contract.Address], simd.Config().Denom)
		s.Require().NoError(err)

		// Execute the contract:
		stargateExecMsg := cwicacontroller.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{proposalMsg}, nil, nil,
		)
		_, err = s.Contract.Execute(ctx, wasmdUser.KeyName(), stargateExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(1), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))

		// Check if the proposal was created:
		proposalResp, err := e2esuite.GRPCQuery[govv1.QueryProposalResponse](ctx, simd, &govv1.QueryProposalRequest{
			ProposalId: 1,
		})
		s.Require().NoError(err)
		s.Require().Equal("e2e", proposalResp.Proposal.Title)

		postBalance, err := simd.GetBalance(ctx, s.IcaContractToAddrMap[s.Contract.Address], simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(intialBalance.Sub(sdkmath.NewInt(10_000_000)), postBalance)
	})

	s.Run("TestSendCosmosMsgsSuccess", func() {
		intialBalance, err := simd.GetBalance(ctx, s.IcaContractToAddrMap[s.Contract.Address], simd.Config().Denom)
		s.Require().NoError(err)

		validator, err := simd.Validators[0].KeyBech32(ctx, "validator", "val")
		s.Require().NoError(err)

		// Stake some tokens through CosmosMsgs:
		stakeCosmosMsg := cwicacontroller.CosmosMsg_for_Empty{
			Staking: &cwicacontroller.CosmosMsg_for_Empty_Staking{
				Delegate: &cwicacontroller.StakingMsg_Delegate{
					Validator: validator,
					Amount: cwicacontroller.Coin{
						Denom:  simd.Config().Denom,
						Amount: "10000000",
					},
				},
			},
		}
		// Vote on the proposal through CosmosMsgs:
		voteCosmosMsg := cwicacontroller.CosmosMsg_for_Empty{
			Gov: &cwicacontroller.CosmosMsg_for_Empty_Gov{
				Vote: &cwicacontroller.GovMsg_Vote{
					ProposalId: 1,
					Vote:       "yes",
				},
			},
		}

		// Execute the contract:
		sendCosmosMsgsExecMsg := cwicacontroller.ExecuteMsg{
			SendCosmosMsgs: &cwicacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []cwicacontroller.CosmosMsg_for_Empty{stakeCosmosMsg, voteCosmosMsg},
			},
		}
		_, err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(2), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))

		// Check if the delegation was successful:
		postBalance, err := simd.GetBalance(ctx, s.IcaContractToAddrMap[s.Contract.Address], simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(intialBalance.Sub(sdkmath.NewInt(10_000_000)), postBalance)

		delResp, err := e2esuite.GRPCQuery[stakingtypes.QueryDelegationResponse](ctx, simd, &stakingtypes.QueryDelegationRequest{
			DelegatorAddr: s.IcaContractToAddrMap[s.Contract.Address],
			ValidatorAddr: validator,
		})
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(10_000_000), delResp.DelegationResponse.Balance.Amount)

		// Check if the vote was successful:
		voteResp, err := e2esuite.GRPCQuery[govv1.QueryVoteResponse](ctx, simd, &govv1.QueryVoteRequest{
			ProposalId: 1,
			Voter:      s.IcaContractToAddrMap[s.Contract.Address],
		})
		s.Require().NoError(err)
		s.Require().Len(voteResp.Vote.Options, 1)
		s.Require().Equal(govv1.OptionYes, voteResp.Vote.Options[0].Option)
		s.Require().Equal(sdkmath.LegacyNewDec(1).String(), voteResp.Vote.Options[0].Weight)
	})

	s.Run("TestIcaError", func() {
		// Test erroneous callback:
		// Send incorrect custom ICA messages through the contract:
		badSendMsg := cwicacontroller.CosmosMsg_for_Empty{
			Bank: &cwicacontroller.CosmosMsg_for_Empty_Bank{
				Send: &cwicacontroller.BankMsg_Send{
					ToAddress: simdUser.FormattedAddress(),
					Amount: []cwicacontroller.Coin{
						{
							Denom:  "INVALID_DENOM",
							Amount: "1",
						},
					},
				},
			},
		}

		// Execute the contract:
		badMsg := cwicacontroller.ExecuteMsg{
			SendCosmosMsgs: &cwicacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []cwicacontroller.CosmosMsg_for_Empty{badSendMsg},
			},
		}
		_, err := s.Contract.Execute(ctx, wasmdUser.KeyName(), badMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(2), len(callbackCounter.Success))
		s.Require().Equal(int(1), len(callbackCounter.Error))
		s.Require().Equal(int(0), len(callbackCounter.Timeout))
	})
}

func (s *ContractTestSuite) TestSendCosmosMsgs_Ordered_Protobuf() {
	s.SendCosmosMsgsTestWithOrdering(cwicacontroller.IbcOrder_OrderOrdered)
}

func (s *ContractTestSuite) TestSendCosmosMsgs_Unordered_Protobuf() {
	s.SendCosmosMsgsTestWithOrdering(cwicacontroller.IbcOrder_OrderUnordered)
}

// SendCosmosMsgsTestWithOrdering tests some more CosmosMsgs that are not covered by the IcaContractExecutionTestWithOrdering.
// The following CosmosMsgs are tested here:
//
// - Bank::Send
// - Stargate
// - VoteWeighted
// - FundCommunityPool
// - SetWithdrawAddress
func (s *ContractTestSuite) SendCosmosMsgsTestWithOrdering(ordering cwicacontroller.IbcOrder) {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, ordering)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser := s.UserA
	simdUser := s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaContractToAddrMap[s.Contract.Address])

	s.Run("TestStargate", func() {
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
			s.IcaContractToAddrMap[s.Contract.Address], "e2e", "e2e", "e2e", false,
		)
		s.Require().NoError(err)

		initialBalance, err := simd.GetBalance(ctx, s.IcaContractToAddrMap[s.Contract.Address], simd.Config().Denom)
		s.Require().NoError(err)

		// Execute the contract:
		sendStargateMsg := cwicacontroller.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{proposalMsg}, nil, nil,
		)
		_, err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendStargateMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(1), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))

		// Check if the proposal was created:
		proposalResp, err := e2esuite.GRPCQuery[govv1.QueryProposalResponse](ctx, simd, &govv1.QueryProposalRequest{
			ProposalId: 1,
		})
		s.Require().NoError(err)
		s.Require().Equal("e2e", proposalResp.Proposal.Title)

		postBalance, err := simd.GetBalance(ctx, s.IcaContractToAddrMap[s.Contract.Address], simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(initialBalance.Sub(sdkmath.NewInt(10_000_000)), postBalance)
	})

	s.Run("TestDelegateAndVoteWeightedAndCommunityPool", func() {
		intialBalance, err := simd.GetBalance(ctx, s.IcaContractToAddrMap[s.Contract.Address], simd.Config().Denom)
		s.Require().NoError(err)

		validator, err := simd.Validators[0].KeyBech32(ctx, "validator", "val")
		s.Require().NoError(err)

		// Stake some tokens through CosmosMsgs:
		stakeCosmosMsg := cwicacontroller.CosmosMsg_for_Empty{
			Staking: &cwicacontroller.CosmosMsg_for_Empty_Staking{
				Delegate: &cwicacontroller.StakingMsg_Delegate{
					Validator: validator,
					Amount: cwicacontroller.Coin{
						Denom:  simd.Config().Denom,
						Amount: "10000000",
					},
				},
			},
		}
		// Vote on the proposal through CosmosMsgs:
		voteCosmosMsg := cwicacontroller.CosmosMsg_for_Empty{
			Gov: &cwicacontroller.CosmosMsg_for_Empty_Gov{
				VoteWeighted: &cwicacontroller.GovMsg_VoteWeighted{
					ProposalId: 1,
					Options: []cwicacontroller.WeightedVoteOption{
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
		fundPoolCosmosMsg := cwicacontroller.CosmosMsg_for_Empty{
			Distribution: &cwicacontroller.CosmosMsg_for_Empty_Distribution{
				FundCommunityPool: &cwicacontroller.DistributionMsg_FundCommunityPool{
					Amount: []cwicacontroller.Coin{{
						Denom:  simd.Config().Denom,
						Amount: "10000000",
					}},
				},
			},
		}

		// Execute the contract:
		sendCosmosMsgsExecMsg := cwicacontroller.ExecuteMsg{
			SendCosmosMsgs: &cwicacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []cwicacontroller.CosmosMsg_for_Empty{stakeCosmosMsg, voteCosmosMsg, fundPoolCosmosMsg},
			},
		}
		_, err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(2), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))

		// Check if the delegation was successful:
		postBalance, err := simd.GetBalance(ctx, s.IcaContractToAddrMap[s.Contract.Address], simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(intialBalance.Sub(sdkmath.NewInt(20_000_000)), postBalance)

		delResp, err := e2esuite.GRPCQuery[stakingtypes.QueryDelegationResponse](ctx, simd, &stakingtypes.QueryDelegationRequest{
			DelegatorAddr: s.IcaContractToAddrMap[s.Contract.Address],
			ValidatorAddr: validator,
		})
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(10_000_000), delResp.DelegationResponse.Balance.Amount)

		// Check if the vote was successful:
		voteResp, err := e2esuite.GRPCQuery[govv1.QueryVoteResponse](ctx, simd, &govv1.QueryVoteRequest{
			ProposalId: 1,
			Voter:      s.IcaContractToAddrMap[s.Contract.Address],
		})
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

	s.Run("TestSendAndSetWithdrawAddress", func() {
		initialBalance, err := simd.GetBalance(ctx, s.IcaContractToAddrMap[s.Contract.Address], simd.Config().Denom)
		s.Require().NoError(err)

		// Send some tokens to the simdUser from the ICA address
		sendMsg := cwicacontroller.CosmosMsg_for_Empty{
			Bank: &cwicacontroller.CosmosMsg_for_Empty_Bank{
				Send: &cwicacontroller.BankMsg_Send{
					ToAddress: simdUser.FormattedAddress(),
					Amount: []cwicacontroller.Coin{
						{
							Denom:  simd.Config().Denom,
							Amount: "1000000",
						},
					},
				},
			},
		}

		// Set the withdraw address to the simdUser
		setWithdrawAddressMsg := cwicacontroller.CosmosMsg_for_Empty{
			Distribution: &cwicacontroller.CosmosMsg_for_Empty_Distribution{
				SetWithdrawAddress: &cwicacontroller.DistributionMsg_SetWithdrawAddress{
					Address: simdUser.FormattedAddress(),
				},
			},
		}

		// Execute the contract:
		sendCosmosMsgsExecMsg := cwicacontroller.ExecuteMsg{
			SendCosmosMsgs: &cwicacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []cwicacontroller.CosmosMsg_for_Empty{sendMsg, setWithdrawAddressMsg},
			},
		}
		_, err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(3), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))

		// Check if the send was successful:
		postBalance, err := simd.GetBalance(ctx, s.IcaContractToAddrMap[s.Contract.Address], simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(1_000_000), initialBalance.Sub(postBalance))
	})
}

func (s *ContractTestSuite) TestSendCosmosMsgs_WithQueries() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, cwicacontroller.IbcOrder_OrderUnordered)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser := s.UserA
	simdUser := s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaContractToAddrMap[s.Contract.Address])

	s.Run("BankQuery_Balance", func() {
		balanceQueryMsg := cwicacontroller.ExecuteMsg{
			SendCosmosMsgs: &cwicacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []cwicacontroller.CosmosMsg_for_Empty{},
				Queries: []cwicacontroller.QueryRequest_for_Empty{
					{
						Bank: &cwicacontroller.QueryRequest_for_Empty_Bank{
							Balance: &cwicacontroller.BankQuery_Balance{
								Address: simdUser.FormattedAddress(),
								Denom:   simd.Config().Denom,
							},
						},
					},
				},
			},
		}

		expBalance, err := simd.GetBalance(ctx, simdUser.FormattedAddress(), simd.Config().Denom)
		s.Require().NoError(err)

		_, err = s.Contract.Execute(ctx, wasmdUser.KeyName(), balanceQueryMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(1), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))

		s.Run("test unmarshaling ica acknowledgement", func() {
			icaAck := &sdk.TxMsgData{}
			s.Require().True(s.Run("unmarshal ica response", func() {
				err := proto.Unmarshal(callbackCounter.Success[0].OnAcknowledgementPacketCallback.IcaAcknowledgement.Result.Unwrap(), icaAck)
				s.Require().NoError(err)
				s.Require().Len(icaAck.GetMsgResponses(), 1)
			}))

			queryTxResp := &icahosttypes.MsgModuleQuerySafeResponse{}
			s.Require().True(s.Run("unmarshal MsgModuleQuerySafeResponse", func() {
				err := proto.Unmarshal(icaAck.MsgResponses[0].Value, queryTxResp)
				s.Require().NoError(err)
				s.Require().Len(queryTxResp.Responses, 1)
			}))

			balanceResp := &banktypes.QueryBalanceResponse{}
			s.Require().True(s.Run("unmarshal and verify bank query response", func() {
				err := proto.Unmarshal(queryTxResp.Responses[0], balanceResp)
				s.Require().NoError(err)
				s.Require().Equal(simd.Config().Denom, balanceResp.Balance.Denom)
				s.Require().Equal(expBalance.Int64(), balanceResp.Balance.Amount.Int64())
			}))
		})

		s.Run("verify query result", func() {
			s.Require().Nil(callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Error)
			s.Require().NotNil(callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success)
			s.Require().Len(callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses, 1)
			s.Require().Equal(simd.Config().Denom, callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Bank.Balance.Amount.Denom)
			s.Require().Equal(expBalance.String(), string(callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Bank.Balance.Amount.Amount))
		})
	})
}

func (s *ContractTestSuite) TestIcaContractTimeoutPacket_Ordered_Protobuf() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, cwicacontroller.IbcOrder_OrderOrdered)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser, _ := s.UserA, s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaContractToAddrMap[s.Contract.Address])

	contractState, err := s.Contract.QueryClient().GetContractState(ctx, &cwicacontroller.QueryMsg_GetContractState{})
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
		stargateExecMsg := cwicacontroller.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{}, nil, &timeout,
		)
		_, err = s.Contract.Execute(ctx, wasmdUser.KeyName(), stargateExecMsg)
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
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(0), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))
		s.Require().Equal(int(1), len(callbackCounter.Timeout))

		// Check if contract channel state was updated:
		contractChannelState, err := s.Contract.QueryClient().GetChannel(ctx, &cwicacontroller.QueryMsg_GetChannel{})
		s.Require().NoError(err)
		s.Require().Equal(cwicacontroller.ChannelStatus_StateClosed, contractChannelState.ChannelStatus)
	})

	s.Run("TestChannelReopening", func() {
		// Reopen the channel:
		createChannelMsg := cwicacontroller.ExecuteMsg{
			CreateChannel: &cwicacontroller.ExecuteMsg_CreateChannel{
				ChannelOpenInitOptions: nil,
			},
		}

		_, err := s.Contract.Execute(ctx, wasmdUser.KeyName(), createChannelMsg, "--gas", "500000")
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
		contractChannelState, err := s.Contract.QueryClient().GetChannel(ctx, &cwicacontroller.QueryMsg_GetChannel{})
		s.Require().NoError(err)
		s.Require().Equal(cwicacontroller.ChannelStatus_StateOpen, contractChannelState.ChannelStatus)
		s.Require().Equal(wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionId)
		s.Require().Equal(wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelId)
		s.Require().Equal(wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortId)
		s.Require().Equal(wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelId)
		s.Require().Equal(wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortId)
		s.Require().Equal(wasmdChannel.Ordering, string(contractChannelState.Channel.Order))

		contractState, err := s.Contract.QueryClient().GetContractState(ctx, &cwicacontroller.QueryMsg_GetContractState{})
		s.Require().NoError(err)
		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelId)
		s.Require().Equal(s.IcaContractToAddrMap[s.Contract.Address], contractState.IcaInfo.IcaAddress)

		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(0), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))
		s.Require().Equal(int(1), len(callbackCounter.Timeout))
	})

	s.Run("TestSendCustomIcaMessagesAfterReopen", func() {
		// Send custom ICA message through the contract:
		sendMsg := &banktypes.MsgSend{
			FromAddress: s.IcaContractToAddrMap[s.Contract.Address],
			ToAddress:   s.UserB.FormattedAddress(),
			Amount:      sdk.NewCoins(sdk.NewCoin(simd.Config().Denom, sdkmath.NewInt(100))),
		}

		// Execute the contract:
		sendCustomIcaMsg := cwicacontroller.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{sendMsg}, nil, nil,
		)
		_, err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCustomIcaMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 10, wasmd, simd)
		s.Require().NoError(err)

		icaBalance, err := simd.GetBalance(ctx, s.IcaContractToAddrMap[s.Contract.Address], simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(1000000000-100), icaBalance)

		// Check if contract callbacks were executed:
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(1), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))
		s.Require().Equal(int(1), len(callbackCounter.Timeout))
	})
}

func (s *ContractTestSuite) TestIcaContractTimeoutPacket_Unordered_Protobuf() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, cwicacontroller.IbcOrder_OrderUnordered)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser, _ := s.UserA, s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaContractToAddrMap[s.Contract.Address])

	contractState, err := s.Contract.QueryClient().GetContractState(ctx, &cwicacontroller.QueryMsg_GetContractState{})
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
		sendCustomIcaMsg := cwicacontroller.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{}, nil, &timeout,
		)
		_, err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCustomIcaMsg)
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
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(0), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))
		s.Require().Equal(int(1), len(callbackCounter.Timeout))

		// Check if contract channel state is still open:
		contractChannelState, err := s.Contract.QueryClient().GetChannel(ctx, &cwicacontroller.QueryMsg_GetChannel{})
		s.Require().NoError(err)
		s.Require().Equal(cwicacontroller.ChannelStatus_StateOpen, contractChannelState.ChannelStatus)
	})

	s.Run("TestSendCustomIcaMessagesAfterTimeout", func() {
		// Send custom ICA message through the contract:
		sendMsg := &banktypes.MsgSend{
			FromAddress: s.IcaContractToAddrMap[s.Contract.Address],
			ToAddress:   s.UserB.FormattedAddress(),
			Amount:      sdk.NewCoins(sdk.NewCoin(simd.Config().Denom, sdkmath.NewInt(100))),
		}

		// Execute the contract:
		sendCustomIcaMsg := cwicacontroller.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{sendMsg}, nil, nil,
		)
		_, err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCustomIcaMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		icaBalance, err := simd.GetBalance(ctx, s.IcaContractToAddrMap[s.Contract.Address], simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(1000000000-100), icaBalance)

		// Check if contract callbacks were executed:
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(1), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))
		s.Require().Equal(int(1), len(callbackCounter.Timeout))
	})
}

func (s *ContractTestSuite) TestMigrateOrderedToUnordered() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, cwicacontroller.IbcOrder_OrderOrdered)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser, _ := s.UserA, s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaContractToAddrMap[s.Contract.Address])

	var simdChannelsLen int
	s.Run("TestCloseChannel", func() {
		// Close the channel:
		closeChannelMsg := cwicacontroller.ExecuteMsg{
			CloseChannel: &cwicacontroller.ExecuteMsg_CloseChannel{},
		}
		_, err := s.Contract.Execute(ctx, wasmdUser.KeyName(), closeChannelMsg)
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
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(0), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))
		s.Require().Equal(int(0), len(callbackCounter.Timeout))

		// Check if contract channel state was updated:
		contractChannelState, err := s.Contract.QueryClient().GetChannel(ctx, &cwicacontroller.QueryMsg_GetChannel{})
		s.Require().NoError(err)
		s.Require().Equal(cwicacontroller.ChannelStatus_StateClosed, contractChannelState.ChannelStatus)
	})

	s.Run("TestChannelReopening", func() {
		// Reopen the channel:
		ordering := cwicacontroller.IbcOrder_OrderUnordered

		createChannelMsg := cwicacontroller.ExecuteMsg{
			CreateChannel: &cwicacontroller.ExecuteMsg_CreateChannel{
				ChannelOpenInitOptions: &cwicacontroller.ChannelOpenInitOptions{
					ConnectionId:             ibctesting.FirstConnectionID,
					CounterpartyConnectionId: ibctesting.FirstConnectionID,
					CounterpartyPortId:       nil,
					ChannelOrdering:          &ordering,
				},
			},
		}

		_, err := s.Contract.Execute(ctx, wasmdUser.KeyName(), createChannelMsg, "--gas", "500000")
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
		contractChannelState, err := s.Contract.QueryClient().GetChannel(ctx, &cwicacontroller.QueryMsg_GetChannel{})
		s.Require().NoError(err)
		s.Require().Equal(cwicacontroller.ChannelStatus_StateOpen, contractChannelState.ChannelStatus)
		s.Require().Equal(wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionId)
		s.Require().Equal(wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelId)
		s.Require().Equal(wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortId)
		s.Require().Equal(wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelId)
		s.Require().Equal(wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortId)
		s.Require().Equal(wasmdChannel.Ordering, string(contractChannelState.Channel.Order))

		contractState, err := s.Contract.QueryClient().GetContractState(ctx, &cwicacontroller.QueryMsg_GetContractState{})
		s.Require().NoError(err)
		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelId)
		s.Require().Equal(s.IcaContractToAddrMap[s.Contract.Address], contractState.IcaInfo.IcaAddress)

		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(0), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))
		s.Require().Equal(int(0), len(callbackCounter.Timeout))
	})

	s.Run("TestSendCustomIcaMessagesAfterReopen", func() {
		// Send custom ICA message through the contract:
		sendMsg := &banktypes.MsgSend{
			FromAddress: s.IcaContractToAddrMap[s.Contract.Address],
			ToAddress:   s.UserB.FormattedAddress(),
			Amount:      sdk.NewCoins(sdk.NewCoin(simd.Config().Denom, sdkmath.NewInt(100))),
		}

		// Execute the contract:
		sendCustomIcaMsg := cwicacontroller.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{sendMsg}, nil, nil,
		)
		_, err := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCustomIcaMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 7, wasmd, simd)
		s.Require().NoError(err)

		icaBalance, err := simd.GetBalance(ctx, s.IcaContractToAddrMap[s.Contract.Address], simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(1000000000-100), icaBalance)

		// Check if contract callbacks were executed:
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(1), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))
		s.Require().Equal(int(0), len(callbackCounter.Timeout))
	})
}

func (s *ContractTestSuite) TestCloseChannel_Protobuf_Unordered() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, cwicacontroller.IbcOrder_OrderUnordered)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser, _ := s.UserA, s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaContractToAddrMap[s.Contract.Address])

	s.Run("TestCloseChannel", func() {
		// Close the channel:
		closeChannelMsg := cwicacontroller.ExecuteMsg{
			CloseChannel: &cwicacontroller.ExecuteMsg_CloseChannel{},
		}
		_, err := s.Contract.Execute(ctx, wasmdUser.KeyName(), closeChannelMsg)
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
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(0), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))
		s.Require().Equal(int(0), len(callbackCounter.Timeout))

		// Check if contract channel state was updated:
		contractChannelState, err := s.Contract.QueryClient().GetChannel(ctx, &cwicacontroller.QueryMsg_GetChannel{})
		s.Require().NoError(err)
		s.Require().Equal(cwicacontroller.ChannelStatus_StateClosed, contractChannelState.ChannelStatus)
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
