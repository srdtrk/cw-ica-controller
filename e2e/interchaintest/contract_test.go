package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	mysuite "github.com/srdtrk/cw-ica-controller/interchaintest/v2/testsuite"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types"

	sdkmath "cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

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
	// t.Parallel()

	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, and creates the ibc clients and connections.
	s.SetupSuite(ctx, chainSpecs)

	relayer := s.Relayer
	wasmd := s.ChainA
	simd := s.ChainB
	wasmdUser := s.UserA
	simdUser := s.UserB
	wasmdConnectionID := s.ChainAConnID
	simdConnectionID := s.ChainBConnID

	// Upload and Instantiate the contract on wasmd:
	contract, err := types.StoreAndInstantiateNewContract(ctx, wasmd, wasmdUser.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	contractPort := contract.Port()

	// icaAddress variable is assigned in the handshake test
	var icaAddress string
	s.Run("TestChannelHandshakeSuccess", func() {
		version := fmt.Sprintf(`{"version":"ics27-1","controller_connection_id":"%s","host_connection_id":"%s","address":"","encoding":"proto3json","tx_type":"sdk_multi_msg"}`, wasmdConnectionID, simdConnectionID)
		err = relayer.CreateChannel(ctx, s.ExecRep, s.PathName, ibc.CreateChannelOptions{
			SourcePortName: contractPort,
			DestPortName:   icatypes.HostPortID,
			Order:          ibc.Ordered,
			// cannot use an empty version here, see README
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
		s.T().Logf("wasmd channel: %s", toJSONString(wasmdChannel))
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
			s.T().Logf("extra simd channels detected: %s", toJSONString(simdChannels))
		}

		simdChannel := simdChannels[0]
		s.T().Logf("simd channel state: %s", toJSONString(simdChannel.State))
		s.Require().Equal(icatypes.HostPortID, simdChannel.PortID)
		s.Require().Equal(contractPort, simdChannel.Counterparty.PortID)
		s.Require().Equal(channeltypes.OPEN.String(), simdChannel.State)

		// Check contract's channel state
		contractChannelState, err := contract.QueryChannelState(ctx)
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
		contractState, err := contract.QueryContractState(ctx)
		s.Require().NoError(err)

		s.Require().Equal(wasmdUser.FormattedAddress(), contractState.Admin)
		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelID)

		icaAddress = contractState.IcaInfo.IcaAddress
		s.T().Logf("ICA address after handshake: %s", icaAddress)
	})

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

	s.Run("TestSendPredefinedActionSuccess", func() {
		err = contract.ExecPredefinedAction(ctx, wasmdUser.KeyName(), simdUser.FormattedAddress())
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 6, wasmd, simd)
		s.Require().NoError(err)

		icaBalance, err := simd.GetBalance(ctx, icaAddress, simd.Config().Denom)
		s.Require().NoError(err)
		s.Require().Equal(int64(1000000000-100), icaBalance)

		// Check if contract callbacks were executed:
		callbackCounter, err := contract.QueryCallbackCounter(ctx)
		s.Require().NoError(err)

		s.Require().Equal(uint64(1), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)
		s.Require().Equal(uint64(0), callbackCounter.Timeout)
	})

	s.Run("TestSendCustomIcaMessagesSuccess", func() {
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

		// Execute the contract:
		err = contract.ExecCustomIcaMessages(ctx, wasmdUser.KeyName(), []sdk.Msg{proposalMsg, depositMsg}, nil, nil)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := contract.QueryCallbackCounter(ctx)
		s.Require().NoError(err)

		s.Require().Equal(uint64(2), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		// Check if the proposal was created:
		proposal, err := simd.QueryProposal(ctx, "1")
		s.Require().NoError(err)
		s.Require().Equal(simd.Config().Denom, proposal.TotalDeposit[0].Denom)
		s.Require().Equal(fmt.Sprint(10000000+5000), proposal.TotalDeposit[0].Amount)
		// We do not check title and description of the proposal because this is a legacy proposal.
	})

	s.Run("TestSendCustomIcaMessagesError", func() {
		// Test erroneous callback:
		// Send incorrect custom ICA messages through the contract:
		badMessage := base64.StdEncoding.EncodeToString([]byte("bad message"))
		badCustomMsg := `{"send_custom_ica_messages":{"messages":["` + badMessage + `"]}}`

		// Execute the contract:
		err = contract.Execute(ctx, wasmdUser.KeyName(), badCustomMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := contract.QueryCallbackCounter(ctx)
		s.Require().NoError(err)
		s.Require().Equal(uint64(2), callbackCounter.Success)
		s.Require().Equal(uint64(1), callbackCounter.Error)
		s.Require().Equal(uint64(0), callbackCounter.Timeout)
	})

	s.Run("TestTimeout", func() {
		// We will send a message to the host that will timeout after 1 seconds.
		// You cannot use 0 seconds because block timestamp will be greater than the timeout timestamp which is not allowed.
		// Host will not be able to respond to this message in time.

		timeout := uint64(1)
		customMsg := fmt.Sprintf(`{"send_custom_ica_messages":{"messages":[], "timeout_seconds":%d}}`, timeout)

		// Execute the contract:
		err = contract.Execute(ctx, wasmdUser.KeyName(), customMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if channel was closed:
		wasmdChannels, err := relayer.GetChannels(ctx, s.ExecRep, wasmd.Config().ChainID)
		s.Require().NoError(err)
		s.Require().Equal(1, len(wasmdChannels))
		s.Require().Equal(channeltypes.CLOSED.String(), wasmdChannels[0].State)

		simdChannels, err := relayer.GetChannels(ctx, s.ExecRep, simd.Config().ChainID)
		s.Require().NoError(err)
		// sometimes there is a redundant channel for unknown reasons
		s.Require().Greater(len(simdChannels), 0)
		s.Require().Equal(channeltypes.CLOSED.String(), simdChannels[0].State)

		// Check if contract callbacks were executed:
		callbackCounter, err := contract.QueryCallbackCounter(ctx)
		s.Require().NoError(err)
		s.Require().Equal(uint64(2), callbackCounter.Success)
		s.Require().Equal(uint64(1), callbackCounter.Error)
		s.Require().Equal(uint64(1), callbackCounter.Timeout)

		// Check if contract channel state was updated:
		contractChannelState, err := contract.QueryChannelState(ctx)
		s.Require().NoError(err)
		s.Require().Equal(channeltypes.CLOSED.String(), contractChannelState.ChannelStatus)
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
