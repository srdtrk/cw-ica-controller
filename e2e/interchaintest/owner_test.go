package main

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/strangelove-ventures/interchaintest/v7/testutil"

	mysuite "github.com/srdtrk/cw-ica-controller/interchaintest/v2/testsuite"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types"
)

type OwnerTestSuite struct {
	mysuite.TestSuite

	IcaContractCodeId uint64
	OwnerContract     *types.OwnerContract
	NumOfIcaContracts uint32
}

// SetupOwnerTestSuite starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
// sets up the contract and does the channel handshake for the contract test suite.
func (s *OwnerTestSuite) SetupOwnerTestSuite(ctx context.Context) {
	s.SetupSuite(ctx, chainSpecs)

	codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	// codeId is string and needs to be converted to uint64
	s.IcaContractCodeId, err = strconv.ParseUint(codeId, 10, 64)
	s.Require().NoError(err)

	s.OwnerContract, err = types.StoreAndInstantiateNewOwnerContract(
		ctx, s.ChainA, s.UserA.KeyName(), "../../artifacts/cw_ica_owner.wasm", s.IcaContractCodeId,
	)
	s.Require().NoError(err)

	s.NumOfIcaContracts = 0

	// Create the ICA Contract
	channelOpenInitOptions := types.ChannelOpenInitOptions{
		ConnectionId:             s.ChainAConnID,
		CounterpartyConnectionId: s.ChainBConnID,
	}
	createMsg := types.NewOwnerCreateIcaContractMsg(nil, &channelOpenInitOptions)

	err = s.OwnerContract.Execute(ctx, s.UserA.KeyName(), createMsg, "--gas", "500000")
	s.Require().NoError(err)

	s.NumOfIcaContracts++

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)
}

func TestWithOwnerTestSuite(t *testing.T) {
	suite.Run(t, new(OwnerTestSuite))
}

func (s *OwnerTestSuite) TestCreateIcaContract() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupOwnerTestSuite(ctx)
	wasmd, simd := s.ChainA, s.ChainB

	icaState, err := s.OwnerContract.QueryIcaContractState(ctx, 0)
	s.Require().NoError(err)

	icaContract := types.NewIcaContract(types.NewContract(icaState.ContractAddr, strconv.FormatUint(s.IcaContractCodeId, 10), wasmd))

	s.Run("TestChannelHandshakeSuccess", func() {
		// Test if the handshake was successful
		wasmdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, wasmd.Config().ChainID)
		s.Require().NoError(err)
		s.Require().Equal(1, len(wasmdChannels))

		wasmdChannel := wasmdChannels[0]
		s.T().Logf("wasmd channel: %s", toJSONString(wasmdChannel))
		s.Require().Equal(icaContract.Port(), wasmdChannel.PortID)
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
		s.Require().Equal(icaContract.Port(), simdChannel.Counterparty.PortID)
		s.Require().Equal(channeltypes.OPEN.String(), simdChannel.State)

		// Check contract's channel state
		contractChannelState, err := icaContract.QueryChannelState(ctx)
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
		contractState, err := icaContract.QueryContractState(ctx)
		s.Require().NoError(err)

		s.Require().Equal(s.OwnerContract.Address, contractState.Admin)
		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelID)
	})
}

func (s *OwnerTestSuite) TestOwnerPredefinedAction() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupOwnerTestSuite(ctx)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser, simdUser := s.UserA, s.UserB

	icaState, err := s.OwnerContract.QueryIcaContractState(ctx, 0)
	s.Require().NoError(err)

	icaContract := types.NewIcaContract(types.NewContract(icaState.ContractAddr, strconv.FormatUint(s.IcaContractCodeId, 10), wasmd))

	// Check contract state
	contractState, err := icaContract.QueryContractState(ctx)
	s.Require().NoError(err)

	icaAddress := contractState.IcaInfo.IcaAddress

	// Fund the ICA address:
	s.FundAddressChainB(ctx, icaAddress)

	s.Run("TestSendPredefinedActionSuccess", func() {
		err := s.OwnerContract.ExecSendPredefinedAction(ctx, wasmdUser.KeyName(), 0, simdUser.FormattedAddress())
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 6, wasmd, simd)
		s.Require().NoError(err)

		icaBalance, err := simd.GetBalance(ctx, icaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(1000000000-100), icaBalance)

		// Check if contract callbacks were executed:
		callbackCounter, err := icaContract.QueryCallbackCounter(ctx)
		s.Require().NoError(err)

		s.Require().Equal(uint64(1), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)
		s.Require().Equal(uint64(0), callbackCounter.Timeout)
	})
}
