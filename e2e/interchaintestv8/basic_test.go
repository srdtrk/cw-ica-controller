package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	"github.com/srdtrk/go-codegen/e2esuite/v8/e2esuite"
)

// BasicTestSuite is a suite of tests that wraps the TestSuite
// and can provide additional functionality
type BasicTestSuite struct {
	e2esuite.TestSuite
}

// SetupSuite calls the underlying BasicTestSuite's SetupSuite method
func (s *BasicTestSuite) SetupSuite(ctx context.Context) {
	s.TestSuite.SetupSuite(ctx)
}

// TestWithBasicTestSuite is the boilerplate code that allows the test suite to be run
func TestWithBasicTestSuite(t *testing.T) {
	suite.Run(t, new(BasicTestSuite))
}

// TestBasic is an example test function that will be run by the test suite
func (s *BasicTestSuite) TestBasic() {
	ctx := context.Background()

	s.SetupSuite(ctx)

	wasmd1, wasmd2 := s.ChainA, s.ChainB

	// Add your test code here. For example, create a transfer channel between ChainA and ChainB:
	s.Run("CreateTransferChannel", func() {
		err := s.Relayer.CreateChannel(ctx, s.ExecRep, s.PathName, ibc.DefaultChannelOpts())
		s.Require().NoError(err)

		// Wait for the channel to be created
		err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
		s.Require().NoError(err)
	})

	// Test if the handshake was successful
	var (
		wasmd1Channel ibc.ChannelOutput
		wasmd2Channel ibc.ChannelOutput
	)
	s.Run("VerifyTransferChannel", func() {
		wasmd1Channels, err := s.Relayer.GetChannels(ctx, s.ExecRep, wasmd1.Config().ChainID)
		s.Require().NoError(err)
		s.Require().Equal(1, len(wasmd1Channels))

		wasmd1Channel = wasmd1Channels[0]
		s.Require().Equal(transfertypes.PortID, wasmd1Channel.PortID)
		s.Require().Equal(transfertypes.PortID, wasmd1Channel.Counterparty.PortID)
		s.Require().Equal(transfertypes.Version, wasmd1Channel.Version)
		s.Require().Equal(channeltypes.OPEN.String(), wasmd1Channel.State)
		s.Require().Equal(channeltypes.UNORDERED.String(), wasmd1Channel.Ordering)

		wasmd2Channels, err := s.Relayer.GetChannels(ctx, s.ExecRep, wasmd2.Config().ChainID)
		s.Require().NoError(err)
		/*
			The relayer in the test suite sometimes submits multiple `ChannelOpenTry` messages,
			since there is no replay protection for `ChannelOpenTry` messages, there may be
			multiple channels in the state of the counterparty chain. However, only one of them,
			can reach the `OPEN` state since the other will be stuck in the `TRYOPEN` state.
		*/
		s.Require().GreaterOrEqual(len(wasmd2Channels), 1)

		wasmd2Channel = wasmd2Channels[0]
		s.Require().Equal(transfertypes.PortID, wasmd2Channel.PortID)
		s.Require().Equal(transfertypes.PortID, wasmd2Channel.Counterparty.PortID)
		s.Require().Equal(transfertypes.Version, wasmd2Channel.Version)
		s.Require().Equal(channeltypes.OPEN.String(), wasmd2Channel.State)
		s.Require().Equal(channeltypes.UNORDERED.String(), wasmd2Channel.Ordering)
	})

	// Transfer tokens from UserA on ChainA to UserB on ChainB
	s.Run("TransferTokens", func() {
		/*
			Transfer funds to s.UserB on ChainB from s.UserA on ChainA.

			I am using the broadcaster to transfer funds to UserB to demonstrate how to use the
			broadcaster, but you can also use the SendIBCTransfer method from the s.ChainA instance.

			I used 200_000 gas to transfer the funds, but you can use any amount of gas you want.
		*/
		_, err := s.BroadcastMessages(ctx, wasmd1, s.UserA, 200_000, &transfertypes.MsgTransfer{
			SourcePort:       transfertypes.PortID,
			SourceChannel:    wasmd1Channel.ChannelID,
			Token:            sdk.NewInt64Coin(wasmd1.Config().Denom, 100_000),
			Sender:           s.UserA.FormattedAddress(),
			Receiver:         s.UserB.FormattedAddress(),
			TimeoutTimestamp: uint64(time.Now().Add(10 * time.Minute).UnixNano()),
		})
		s.Require().NoError(err)

		s.Require().NoError(testutil.WaitForBlocks(ctx, 5, wasmd1, wasmd2)) // Wait 5 blocks for the packet to be relayed
	})

	// Verify that the tokens were transferred
	s.Run("VerifyTokensTransferred", func() {
		chainBIBCDenom := transfertypes.ParseDenomTrace(
			fmt.Sprintf("%s/%s/%s", wasmd1Channel.PortID, wasmd1Channel.ChannelID, wasmd1.Config().Denom),
		).IBCDenom()

		/*
			Query UserB's balance

			I am using the GRPCQuery to query the new user's balance to demonstrate how to use the GRPCQuery,
			but you can also use the GetBalance method from the s.ChainB instance.
		*/
		balanceResp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, wasmd2, &banktypes.QueryBalanceRequest{
			Address: s.UserB.FormattedAddress(),
			Denom:   chainBIBCDenom,
		})
		s.Require().NoError(err)
		s.Require().NotNil(balanceResp.Balance)

		// Verify the balance
		s.Require().Equal(int64(100_000), balanceResp.Balance.Amount.Int64())
	})
}
