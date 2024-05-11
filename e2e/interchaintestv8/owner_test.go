package main

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	"github.com/srdtrk/go-codegen/e2esuite/v8/e2esuite"
	"github.com/srdtrk/go-codegen/e2esuite/v8/types"
	"github.com/srdtrk/go-codegen/e2esuite/v8/types/cwicacontroller"
	"github.com/srdtrk/go-codegen/e2esuite/v8/types/cwicaowner"
)

type OwnerTestSuite struct {
	e2esuite.TestSuite

	// ChainAConnID is the connection id of chain A
	ChainAConnID string
	// ChainBConnID is the connection id of chain B
	ChainBConnID string

	IcaContractCodeId int64
	OwnerContract     *cwicaowner.Contract
	NumOfIcaContracts uint32
}

// SetupOwnerTestSuite starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
// sets up the contract and does the channel handshake for the contract test suite.
func (s *OwnerTestSuite) SetupOwnerTestSuite(ctx context.Context) {
	s.SetupSuite(ctx)

	codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	// codeId is string and needs to be converted to uint64
	s.IcaContractCodeId, err = strconv.ParseInt(codeId, 10, 64)
	s.Require().NoError(err)

	codeId, err = s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/cw_ica_owner.wasm")
	s.Require().NoError(err)

	instantiateMsg := cwicaowner.InstantiateMsg{IcaControllerCodeId: int(s.IcaContractCodeId)}
	s.OwnerContract, err = cwicaowner.Instantiate(ctx, s.UserA.KeyName(), codeId, "", s.ChainA, instantiateMsg)
	s.Require().NoError(err)

	s.NumOfIcaContracts = 0
	s.ChainAConnID, s.ChainBConnID = "connection-0", "connection-0"

	// Create the ICA Contract
	createMsg := cwicaowner.ExecuteMsg{
		CreateIcaContract: &cwicaowner.ExecuteMsg_CreateIcaContract{
			Salt: nil,
			ChannelOpenInitOptions: cwicaowner.ChannelOpenInitOptions{
				ConnectionId:             s.ChainAConnID,
				CounterpartyConnectionId: s.ChainBConnID,
			},
		},
	}

	_, err = s.OwnerContract.Execute(ctx, s.UserA.KeyName(), createMsg, "--gas", "500000")
	s.Require().NoError(err)

	s.NumOfIcaContracts++

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)
}

func TestWithOwnerTestSuite(t *testing.T) {
	suite.Run(t, new(OwnerTestSuite))
}

func (s *OwnerTestSuite) TestOwnerCreateIcaContract() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupOwnerTestSuite(ctx)
	wasmd, simd := s.ChainA, s.ChainB

	icaState, err := s.OwnerContract.QueryClient().GetIcaContractState(ctx, &cwicaowner.QueryMsg_GetIcaContractState{IcaId: 0})
	s.Require().NoError(err)
	s.Require().NotNil(icaState.IcaState)

	icaContract := types.Contract[cwicacontroller.InstantiateMsg, cwicacontroller.ExecuteMsg, cwicacontroller.QueryMsg, cwicacontroller.QueryClient]{
		Address: string(icaState.ContractAddr),
		CodeID:  strconv.FormatInt(s.IcaContractCodeId, 10),
		Chain:   wasmd,
	}

	icaQc, err := cwicacontroller.NewQueryClient(wasmd.GetHostGRPCAddress(), icaContract.Address)
	s.Require().NoError(err)
	icaContract.SetQueryClient(icaQc)

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
		contractChannelState, err := icaContract.QueryClient().GetChannel(ctx, &cwicacontroller.QueryMsg_GetChannel{})
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
		contractState, err := icaContract.QueryClient().GetContractState(ctx, &cwicacontroller.QueryMsg_GetContractState{})
		s.Require().NoError(err)
		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelId)

		// Check contract's ownership
		ownershipResponse, err := icaContract.QueryClient().Ownership(ctx, &cwicacontroller.QueryMsg_Ownership{})
		s.Require().NoError(err)
		s.Require().Equal(s.OwnerContract.Address, *ownershipResponse.Owner)
		s.Require().Nil(ownershipResponse.PendingOwner)
		s.Require().Nil(ownershipResponse.PendingExpiry)
	})
}

func (s *OwnerTestSuite) TestOwnerPredefinedAction() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupOwnerTestSuite(ctx)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser, simdUser := s.UserA, s.UserB

	icaState, err := s.OwnerContract.QueryClient().GetIcaContractState(ctx, &cwicaowner.QueryMsg_GetIcaContractState{IcaId: 0})
	s.Require().NoError(err)

	icaContract := types.Contract[cwicacontroller.InstantiateMsg, cwicacontroller.ExecuteMsg, cwicacontroller.QueryMsg, cwicacontroller.QueryClient]{
		Address: string(icaState.ContractAddr),
		CodeID:  strconv.FormatInt(s.IcaContractCodeId, 10),
		Chain:   wasmd,
	}

	icaQc, err := cwicacontroller.NewQueryClient(wasmd.GetHostGRPCAddress(), icaContract.Address)
	s.Require().NoError(err)
	icaContract.SetQueryClient(icaQc)

	// Check contract state
	contractState, err := icaContract.QueryClient().GetContractState(ctx, &cwicacontroller.QueryMsg_GetContractState{})
	s.Require().NoError(err)

	icaAddress := contractState.IcaInfo.IcaAddress

	// Fund the ICA address:
	s.FundAddressChainB(ctx, icaAddress)

	s.Run("TestSendPredefinedActionSuccess", func() {
		execPredefinedActionMsg := cwicaowner.ExecuteMsg{
			SendPredefinedAction: &cwicaowner.ExecuteMsg_SendPredefinedAction{
				IcaId:     0,
				ToAddress: simdUser.FormattedAddress(),
			},
		}
		_, err := s.OwnerContract.Execute(ctx, wasmdUser.KeyName(), execPredefinedActionMsg, "--gas", "500000")
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 6, wasmd, simd)
		s.Require().NoError(err)

		icaBalance, err := simd.GetBalance(ctx, icaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(1000000000-100), icaBalance)
	})
}
