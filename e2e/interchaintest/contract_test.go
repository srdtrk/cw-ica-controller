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

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"

	mysuite "github.com/srdtrk/cw-ica-controller/interchaintest/v2/testsuite"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types"
	callbackcounter "github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/callback-counter"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/icacontroller"
)

type ContractTestSuite struct {
	mysuite.TestSuite

	Contract   *types.IcaContract
	IcaAddress string
	// CallbackContractAddress is the address of the callback counter contract
	CallbackContractAddress string
}

// SetupSuite calls the underlying TestSuite's SetupSuite method and initializes an empty contract
func (s *ContractTestSuite) SetupSuite(ctx context.Context, chainSpecs []*interchaintest.ChainSpec) {
	s.TestSuite.SetupSuite(ctx, chainSpecs)

	// Initialize an empty contract so that we can use the methods of the contract
	s.Contract = types.NewIcaContract(types.Contract{})
}

// SetupContractTestSuite starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
// sets up the contract and does the channel handshake for the contract test suite.
func (s *ContractTestSuite) SetupContractTestSuite(ctx context.Context, encoding string) {
	s.SetupSuite(ctx, chainSpecs)

	codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/callback_counter.wasm")
	s.Require().NoError(err)

	s.CallbackContractAddress, err = s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), codeId, callbackcounter.InstantiateMsg, true)
	s.Require().NoError(err)

	codeId, err = s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	// Instantiate the contract with channel:
	instantiateMsg := icacontroller.InstantiateMsg{
		Owner: nil,
		ChannelOpenInitOptions: &icacontroller.ChannelOpenInitOptions{
			ConnectionId:             s.ChainAConnID,
			CounterpartyConnectionId: s.ChainBConnID,
			CounterpartyPortId:       nil,
			TxEncoding:               &encoding,
		},
		SendCallbacksTo: &s.CallbackContractAddress,
	}

	err = s.Contract.Instantiate(ctx, s.UserA.KeyName(), s.ChainA, codeId, instantiateMsg, "--gas", "500000")
	s.Require().NoError(err)

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	contractState, err := types.QueryAnyMsg[icacontroller.ContractState](
		ctx, &s.Contract.Contract,
		icacontroller.GetContractStateRequest,
	)
	s.Require().NoError(err)

	ownershipResponse, err := types.QueryAnyMsg[icacontroller.OwnershipResponse](ctx, &s.Contract.Contract, icacontroller.OwnershipRequest)
	s.Require().NoError(err)

	s.IcaAddress = contractState.IcaInfo.IcaAddress
	s.Contract.SetIcaAddress(s.IcaAddress)

	s.Require().Equal(s.UserA.FormattedAddress(), ownershipResponse.Owner)
	s.Require().Nil(ownershipResponse.PendingOwner)
	s.Require().Nil(ownershipResponse.PendingExpiry)
}

func TestWithContractTestSuite(t *testing.T) {
	suite.Run(t, new(ContractTestSuite))
}

func (s *ContractTestSuite) TestIcaContractChannelHandshake() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, icatypes.EncodingProto3JSON)
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
		contractChannelState, err := types.QueryAnyMsg[icacontroller.ContractChannelState](ctx, &s.Contract.Contract, icacontroller.GetChannelRequest)
		s.Require().NoError(err)

		s.T().Logf("contract's channel store after handshake: %s", toJSONString(contractChannelState))

		s.Require().Equal(wasmdChannel.State, contractChannelState.ChannelStatus)
		s.Require().Equal(wasmdChannel.Version, contractChannelState.Channel.Version)
		s.Require().Equal(wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionID)
		s.Require().Equal(wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelID)
		s.Require().Equal(wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortID)
		s.Require().Equal(wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelID)
		s.Require().Equal(wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortID)
		s.Require().Equal(wasmdChannel.Ordering, contractChannelState.Channel.Order)

		// Check contract state
		contractState, err := types.QueryAnyMsg[icacontroller.ContractState](
			ctx, &s.Contract.Contract,
			icacontroller.GetContractStateRequest,
		)
		s.Require().NoError(err)

		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelID)
		s.Require().Equal(false, contractState.AllowChannelOpenInit)
	})
}

func (s *ContractTestSuite) TestIcaRelayerInstantiatedChannelHandshake() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, and creates the ibc clients and connections.
	s.SetupSuite(ctx, chainSpecs)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser := s.UserA

	var err error
	// Upload and Instantiate the contract on wasmd:
	codeId, err := wasmd.StoreContract(ctx, wasmdUser.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	contractAddr, err := wasmd.InstantiateContract(ctx, wasmdUser.KeyName(), codeId, "{}", true)
	s.Require().NoError(err)

	contract := types.Contract{
		Address: contractAddr,
		CodeID:  codeId,
		Chain:   wasmd,
	}

	s.Contract = types.NewIcaContract(contract)

	contractState, err := types.QueryAnyMsg[icacontroller.ContractState](
		ctx, &s.Contract.Contract,
		icacontroller.GetContractStateRequest,
	)
	s.Require().NoError(err)
	s.Require().Equal(true, contractState.AllowChannelOpenInit)

	version := fmt.Sprintf(`{"version":"%s","controller_connection_id":"%s","host_connection_id":"%s","address":"","encoding":"%s","tx_type":"%s"}`, icatypes.Version, s.ChainAConnID, s.ChainBConnID, icatypes.EncodingProtobuf, icatypes.TxTypeSDKMultiMsg)
	err = s.Relayer.CreateChannel(ctx, s.ExecRep, s.PathName, ibc.CreateChannelOptions{
		SourcePortName: s.Contract.Port(),
		DestPortName:   icatypes.HostPortID,
		Order:          ibc.Ordered,
		// cannot use an empty version here, see README
		Version: version,
	})
	s.Require().NoError(err)

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	contractState, err = types.QueryAnyMsg[icacontroller.ContractState](
		ctx, &s.Contract.Contract,
		icacontroller.GetContractStateRequest,
	)
	s.Require().NoError(err)
	s.Require().Equal(false, contractState.AllowChannelOpenInit)

	s.IcaAddress = contractState.IcaInfo.IcaAddress
	s.Contract.SetIcaAddress(s.IcaAddress)

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
		contractChannelState, err := types.QueryAnyMsg[icacontroller.ContractChannelState](ctx, &s.Contract.Contract, icacontroller.GetChannelRequest)
		s.Require().NoError(err)

		s.T().Logf("contract's channel store after handshake: %s", toJSONString(contractChannelState))

		s.Require().Equal(wasmdChannel.State, contractChannelState.ChannelStatus)
		s.Require().Equal(wasmdChannel.Version, contractChannelState.Channel.Version)
		s.Require().Equal(wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionID)
		s.Require().Equal(wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelID)
		s.Require().Equal(wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortID)
		s.Require().Equal(wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelID)
		s.Require().Equal(wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortID)
		s.Require().Equal(wasmdChannel.Ordering, contractChannelState.Channel.Order)

		// Check contract state
		contractState, err = types.QueryAnyMsg[icacontroller.ContractState](
			ctx, &s.Contract.Contract,
			icacontroller.GetContractStateRequest,
		)
		s.Require().NoError(err)

		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelID)
		s.Require().Equal(false, contractState.AllowChannelOpenInit)
	})
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
			ChannelOpenInitOptions: &icacontroller.ChannelOpenInitOptions{
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
			ChannelOpenInitOptions: &icacontroller.ChannelOpenInitOptions{
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
		err = testutil.WaitForBlocks(ctx, 8, s.ChainA, s.ChainB)
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
		contractChannelState, err := types.QueryAnyMsg[icacontroller.ContractChannelState](ctx, &s.Contract.Contract, icacontroller.GetChannelRequest)
		s.Require().NoError(err)

		s.T().Logf("contract's channel store after handshake: %s", toJSONString(contractChannelState))

		s.Require().Equal(wasmdChannel.State, contractChannelState.ChannelStatus)
		s.Require().Equal(wasmdChannel.Version, contractChannelState.Channel.Version)
		s.Require().Equal(wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionID)
		s.Require().Equal(wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelID)
		s.Require().Equal(wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortID)
		s.Require().Equal(wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelID)
		s.Require().Equal(wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortID)
		s.Require().Equal(wasmdChannel.Ordering, contractChannelState.Channel.Order)

		// Check contract state
		contractState, err := types.QueryAnyMsg[icacontroller.ContractState](
			ctx, &s.Contract.Contract,
			icacontroller.GetContractStateRequest,
		)
		s.Require().NoError(err)

		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelID)
		s.Require().Equal(false, contractState.AllowChannelOpenInit)
	})
}

func (s *ContractTestSuite) TestIcaContractExecutionProto3JsonEncoding() {
	s.IcaContractExecutionTestWithEncoding(icatypes.EncodingProto3JSON)
}

func (s *ContractTestSuite) TestIcaContractExecutionProtobufEncoding() {
	s.IcaContractExecutionTestWithEncoding(icatypes.EncodingProtobuf)
}

func (s *ContractTestSuite) IcaContractExecutionTestWithEncoding(encoding string) {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, encoding)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser := s.UserA

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaAddress)

	s.Run(fmt.Sprintf("TestSendCustomIcaMessagesSuccess-%s", encoding), func() {
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
			InitialDeposit: sdk.NewCoins(sdk.NewCoin(simd.Config().Denom, sdkmath.NewInt(5_000))),
			Proposer:       s.IcaAddress,
		}

		// Create deposit message:
		depositMsg := &govtypes.MsgDeposit{
			ProposalId: 1,
			Depositor:  s.IcaAddress,
			Amount:     sdk.NewCoins(sdk.NewCoin(simd.Config().Denom, sdkmath.NewInt(10_000_000))),
		}

		intialBalance, err := simd.GetBalance(ctx, s.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)

		// Execute the contract:
		sendCustomIcaMsg := icacontroller.NewExecuteMsg_SendCustomIcaMessages_FromProto(
			simd.Config().EncodingConfig.Codec,
			[]proto.Message{proposalMsg, depositMsg},
			encoding, nil, nil,
		)
		err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCustomIcaMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, &s.Contract.Contract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(1), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		// Check if the proposal was created:
		proposal, err := simd.QueryProposal(ctx, "1")
		s.Require().NoError(err)
		s.Require().Equal(simd.Config().Denom, proposal.TotalDeposit[0].Denom)
		s.Require().Equal(fmt.Sprint(10_000_000+5_000), proposal.TotalDeposit[0].Amount)
		// We do not check title and description of the proposal because this is a legacy proposal.

		postBalance, err := simd.GetBalance(ctx, s.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(intialBalance.Sub(sdkmath.NewInt(10_000_000+5_000)), postBalance)
	})

	s.Run(fmt.Sprintf("TestSendCosmosMsgsSuccess-%s", encoding), func() {
		intialBalance, err := simd.GetBalance(ctx, s.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)

		validator, err := simd.Validators[0].KeyBech32(ctx, "validator", "val")
		s.Require().NoError(err)

		// Stake some tokens through CosmosMsgs:
		stakeCosmosMsg := icacontroller.ContractCosmosMsg{
			Staking: &icacontroller.StakingCosmosMsg{
				Delegate: &icacontroller.StakingDelegateCosmosMsg{
					Validator: validator,
					Amount: icacontroller.Coin{
						Denom:  simd.Config().Denom,
						Amount: "10000000",
					},
				},
			},
		}
		// Vote on the proposal through CosmosMsgs:
		voteCosmosMsg := icacontroller.ContractCosmosMsg{
			Gov: &icacontroller.GovCosmosMsg{
				Vote: &icacontroller.GovVoteCosmosMsg{
					ProposalID: 1,
					Vote:       "yes",
				},
			},
		}

		// Execute the contract:
		sendCosmosMsgsExecMsg := icacontroller.ExecuteMsg{
			SendCosmosMsgs: &icacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []icacontroller.ContractCosmosMsg{stakeCosmosMsg, voteCosmosMsg},
			},
		}
		err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, &s.Contract.Contract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(2), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		// Check if the delegation was successful:
		postBalance, err := simd.GetBalance(ctx, s.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(intialBalance.Sub(sdkmath.NewInt(10_000_000)), postBalance)

		delegationsQuerier := mysuite.NewGRPCQuerier[stakingtypes.QueryDelegationResponse](s.T(), simd, "/cosmos.staking.v1beta1.Query/Delegation")

		delRequest := stakingtypes.QueryDelegationRequest{
			DelegatorAddr: s.IcaAddress,
			ValidatorAddr: validator,
		}
		delResp, err := delegationsQuerier.GRPCQuery(ctx, &delRequest)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(10_000_000), delResp.DelegationResponse.Balance.Amount)

		// Check if the vote was successful:
		votesQuerier := mysuite.NewGRPCQuerier[govtypes.QueryVoteResponse](s.T(), simd, "/cosmos.gov.v1beta1.Query/Vote")

		voteRequest := govtypes.QueryVoteRequest{
			ProposalId: 1,
			Voter:      s.IcaAddress,
		}
		voteResp, err := votesQuerier.GRPCQuery(ctx, &voteRequest)
		s.Require().NoError(err)
		s.Require().Len(voteResp.Vote.Options, 1)
		s.Require().Equal(govtypes.OptionYes, voteResp.Vote.Options[0].Option)
		s.Require().Equal(sdkmath.LegacyNewDec(1), voteResp.Vote.Options[0].Weight)
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
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, &s.Contract.Contract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)
		s.Require().Equal(uint64(2), callbackCounter.Success)
		s.Require().Equal(uint64(1), callbackCounter.Error)
		s.Require().Equal(uint64(0), callbackCounter.Timeout)
	})
}

func (s *ContractTestSuite) TestSendCosmosMsgsProto3JsonEncoding() {
	s.SendCosmosMsgsTestWithEncoding(icatypes.EncodingProto3JSON)
}

func (s *ContractTestSuite) TestSendCosmosMsgsProtobufEncoding() {
	s.SendCosmosMsgsTestWithEncoding(icatypes.EncodingProtobuf)
}

// SendCosmosMsgsTestWithEncoding tests some more CosmosMsgs that are not covered by the IcaContractExecutionTestWithEncoding.
// The following CosmosMsgs are tested here:
//
// - Bank::Send
// - Stargate
// - VoteWeighted
// - SetWithdrawAddress
func (s *ContractTestSuite) SendCosmosMsgsTestWithEncoding(encoding string) {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, encoding)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser := s.UserA
	simdUser := s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaAddress)

	s.Run(fmt.Sprintf("TestStargate-%s", encoding), func() {
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
			InitialDeposit: sdk.NewCoins(sdk.NewCoin(simd.Config().Denom, sdkmath.NewInt(5_000))),
			Proposer:       s.IcaAddress,
		}

		// Create deposit message:
		depositMsg := &govtypes.MsgDeposit{
			ProposalId: 1,
			Depositor:  s.IcaAddress,
			Amount:     sdk.NewCoins(sdk.NewCoin(simd.Config().Denom, sdkmath.NewInt(10_000_000))),
		}

		initialBalance, err := simd.GetBalance(ctx, s.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)

		if encoding == icatypes.EncodingProtobuf {
			// Execute the contract:
			sendStargateMsg := icacontroller.NewExecuteMsg_SendCosmosMsgs_FromProto(
				[]proto.Message{proposalMsg, depositMsg}, nil, nil,
			)
			err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendStargateMsg)
			s.Require().NoError(err)
		} else if encoding == icatypes.EncodingProto3JSON {
			sendCustomIcaMsg := icacontroller.NewExecuteMsg_SendCustomIcaMessages_FromProto(
				simd.Config().EncodingConfig.Codec,
				[]proto.Message{proposalMsg, depositMsg},
				icatypes.EncodingProto3JSON, nil, nil,
			)
			err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCustomIcaMsg)
			s.Require().NoError(err)
		}

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, &s.Contract.Contract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(1), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		// Check if the proposal was created:
		proposal, err := simd.QueryProposal(ctx, "1")
		s.Require().NoError(err)
		s.Require().Equal(simd.Config().Denom, proposal.TotalDeposit[0].Denom)
		s.Require().Equal(fmt.Sprint(10_000_000+5_000), proposal.TotalDeposit[0].Amount)
		// We do not check title and description of the proposal because this is a legacy proposal.

		postBalance, err := simd.GetBalance(ctx, s.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(initialBalance.Sub(sdkmath.NewInt(10_000_000+5_000)), postBalance)
	})

	s.Run(fmt.Sprintf("TestDelegateAndVoteWeighted-%s", encoding), func() {
		intialBalance, err := simd.GetBalance(ctx, s.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)

		validator, err := simd.Validators[0].KeyBech32(ctx, "validator", "val")
		s.Require().NoError(err)

		// Stake some tokens through CosmosMsgs:
		stakeCosmosMsg := icacontroller.ContractCosmosMsg{
			Staking: &icacontroller.StakingCosmosMsg{
				Delegate: &icacontroller.StakingDelegateCosmosMsg{
					Validator: validator,
					Amount: icacontroller.Coin{
						Denom:  simd.Config().Denom,
						Amount: "10000000",
					},
				},
			},
		}
		// Vote on the proposal through CosmosMsgs:
		voteCosmosMsg := icacontroller.ContractCosmosMsg{
			Gov: &icacontroller.GovCosmosMsg{
				VoteWeighted: &icacontroller.GovVoteWeightedCosmosMsg{
					ProposalID: 1,
					Options: []icacontroller.GovVoteWeightedOption{
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

		// Execute the contract:
		sendCosmosMsgsExecMsg := icacontroller.ExecuteMsg{
			SendCosmosMsgs: &icacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []icacontroller.ContractCosmosMsg{stakeCosmosMsg, voteCosmosMsg},
			},
		}
		err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, &s.Contract.Contract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(2), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		// Check if the delegation was successful:
		postBalance, err := simd.GetBalance(ctx, s.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(intialBalance.Sub(sdkmath.NewInt(10_000_000)), postBalance)

		delegationsQuerier := mysuite.NewGRPCQuerier[stakingtypes.QueryDelegationResponse](s.T(), simd, "/cosmos.staking.v1beta1.Query/Delegation")

		delRequest := stakingtypes.QueryDelegationRequest{
			DelegatorAddr: s.IcaAddress,
			ValidatorAddr: validator,
		}
		delResp, err := delegationsQuerier.GRPCQuery(ctx, &delRequest)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(10_000_000), delResp.DelegationResponse.Balance.Amount)

		// Check if the vote was successful:
		votesQuerier := mysuite.NewGRPCQuerier[govtypes.QueryVoteResponse](s.T(), simd, "/cosmos.gov.v1beta1.Query/Vote")

		voteRequest := govtypes.QueryVoteRequest{
			ProposalId: 1,
			Voter:      s.IcaAddress,
		}
		voteResp, err := votesQuerier.GRPCQuery(ctx, &voteRequest)
		s.Require().NoError(err)
		s.Require().Len(voteResp.Vote.Options, 2)
		s.Require().Equal(govtypes.OptionYes, voteResp.Vote.Options[0].Option)
		expWeight, err := sdkmath.LegacyNewDecFromStr("0.5")
		s.Require().NoError(err)
		s.Require().Equal(expWeight, voteResp.Vote.Options[0].Weight)
		s.Require().Equal(govtypes.OptionAbstain, voteResp.Vote.Options[1].Option)
		s.Require().Equal(expWeight, voteResp.Vote.Options[1].Weight)
	})

	s.Run(fmt.Sprintf("TestSendAndSetWithdrawAddress-%s", encoding), func() {
		initialBalance, err := simd.GetBalance(ctx, s.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)

		// Send some tokens to the simdUser from the ICA address
		sendMsg := icacontroller.ContractCosmosMsg{
			Bank: &icacontroller.BankCosmosMsg{
				Send: &icacontroller.BankSendCosmosMsg{
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
		setWithdrawAddressMsg := icacontroller.ContractCosmosMsg{
			Distribution: &icacontroller.DistributionCosmosMsg{
				SetWithdrawAddress: &icacontroller.DistributionSetWithdrawAddressCosmosMsg{
					Address: simdUser.FormattedAddress(),
				},
			},
		}

		// Execute the contract:
		sendCosmosMsgsExecMsg := icacontroller.ExecuteMsg{
			SendCosmosMsgs: &icacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []icacontroller.ContractCosmosMsg{sendMsg, setWithdrawAddressMsg},
			},
		}
		err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, &s.Contract.Contract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)
		s.Require().Equal(uint64(3), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		// Check if the send was successful:
		postBalance, err := simd.GetBalance(ctx, s.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(1_000_000), initialBalance.Sub(postBalance))
	})
}

func (s *ContractTestSuite) TestIcaContractTimeoutPacket() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, icatypes.EncodingProto3JSON)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser, _ := s.UserA, s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaAddress)

	contractState, err := types.QueryAnyMsg[icacontroller.ContractState](
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

		timeout := uint64(3)
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

		// Wait until timeout acknoledgement is received:
		err = testutil.WaitForBlocks(ctx, 2, wasmd, simd)
		s.Require().NoError(err)

		// Flush to make sure the channel is closed in simd:
		err = s.Relayer.Flush(ctx, s.ExecRep, s.PathName, contractState.IcaInfo.ChannelID)
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
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, &s.Contract.Contract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)
		s.Require().Equal(uint64(0), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)
		s.Require().Equal(uint64(1), callbackCounter.Timeout)

		// Check if contract channel state was updated:
		contractChannelState, err := types.QueryAnyMsg[icacontroller.ContractChannelState](ctx, &s.Contract.Contract, icacontroller.GetChannelRequest)
		s.Require().NoError(err)
		s.Require().Equal(channeltypes.CLOSED.String(), contractChannelState.ChannelStatus)
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
		contractChannelState, err := types.QueryAnyMsg[icacontroller.ContractChannelState](ctx, &s.Contract.Contract, icacontroller.GetChannelRequest)
		s.Require().NoError(err)
		s.Require().Equal(channeltypes.OPEN.String(), contractChannelState.ChannelStatus)
		// The version string is wrapped by the fee middleware. We we cannot check it directly here.
		// s.Require().Equal(wasmdChannel.Version, contractChannelState.Channel.Version)
		s.Require().Equal(wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionID)
		s.Require().Equal(wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelID)
		s.Require().Equal(wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortID)
		s.Require().Equal(wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelID)
		s.Require().Equal(wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortID)
		s.Require().Equal(wasmdChannel.Ordering, contractChannelState.Channel.Order)

		contractState, err := types.QueryAnyMsg[icacontroller.ContractState](
			ctx, &s.Contract.Contract,
			icacontroller.GetContractStateRequest,
		)
		s.Require().NoError(err)
		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelID)
		s.Require().Equal(s.IcaAddress, contractState.IcaInfo.IcaAddress)

		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, &s.Contract.Contract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(0), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)
		s.Require().Equal(uint64(1), callbackCounter.Timeout)
	})

	s.Run("TestSendCustomIcaMessagesAfterReopen", func() {
		// Send custom ICA message through the contract:
		sendMsg := &banktypes.MsgSend{
			FromAddress: s.IcaAddress,
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

		icaBalance, err := simd.GetBalance(ctx, s.IcaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(1000000000-100), icaBalance)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, &s.Contract.Contract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(1), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)
		s.Require().Equal(uint64(1), callbackCounter.Timeout)
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
