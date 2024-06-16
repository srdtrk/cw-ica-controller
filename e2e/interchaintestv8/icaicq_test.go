package main

import (
	"context"

	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"

	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	"github.com/srdtrk/go-codegen/e2esuite/v8/e2esuite"
	"github.com/srdtrk/go-codegen/e2esuite/v8/types/callbackcounter"
	"github.com/srdtrk/go-codegen/e2esuite/v8/types/cwicacontroller"
)

// `TestBankQueries` tests all allowed bank queries in the SendCosmosMsgs message. The following queries are tested:
// - Bank::Balance
// - Bank::AllBalances
// - Bank::AllDenomMetadata
// - Bank::Supply
// TODO: Bank::DenomMetadata
func (s *ContractTestSuite) TestBankQueries() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, cwicacontroller.IbcOrder_OrderUnordered)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser := s.UserA
	simdUser := s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaContractToAddrMap[s.Contract.Address])

	s.Require().True(s.Run("BankQuery_Balance", func() {
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
		s.Require().Equal(int(0), len(callbackCounter.Timeout))

		s.Require().True(s.Run("test unmarshaling ica acknowledgement", func() {
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
		}))

		s.Require().True(s.Run("verify query result", func() {
			s.Require().Nil(callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Error)
			s.Require().NotNil(callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success)
			s.Require().Len(callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses, 1)
			s.Require().Equal(simd.Config().Denom, callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Bank.Balance.Amount.Denom)
			s.Require().Equal(expBalance.String(), string(callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Bank.Balance.Amount.Amount))
		}))
	}))

	s.Require().True(s.Run("Other bank queries", func() {
		allBankQueriesMsg := cwicacontroller.ExecuteMsg{
			SendCosmosMsgs: &cwicacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []cwicacontroller.CosmosMsg_for_Empty{},
				Queries: []cwicacontroller.QueryRequest_for_Empty{
					{
						Bank: &cwicacontroller.QueryRequest_for_Empty_Bank{
							AllBalances: &cwicacontroller.BankQuery_AllBalances{
								Address: simdUser.FormattedAddress(),
							},
						},
					},
					// fail: need to set metadata first
					// {
					// 	Bank: &cwicacontroller.QueryRequest_for_Empty_Bank{
					// 		DenomMetadata: &cwicacontroller.BankQuery_DenomMetadata{
					// 			Denom: simd.Config().Denom,
					// 		},
					// 	},
					// },
					{
						Bank: &cwicacontroller.QueryRequest_for_Empty_Bank{
							AllDenomMetadata: &cwicacontroller.BankQuery_AllDenomMetadata{},
						},
					},
					{
						Bank: &cwicacontroller.QueryRequest_for_Empty_Bank{
							Supply: &cwicacontroller.BankQuery_Supply{
								Denom: simd.Config().Denom,
							},
						},
					},
				},
			},
		}

		expBalance, err := simd.GetBalance(ctx, simdUser.FormattedAddress(), simd.Config().Denom)
		s.Require().NoError(err)
		expSupply, err := e2esuite.GRPCQuery[banktypes.QueryTotalSupplyResponse](ctx, simd, &banktypes.QueryTotalSupplyRequest{})
		s.Require().NoError(err)

		_, err = s.Contract.Execute(ctx, wasmdUser.KeyName(), allBankQueriesMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(2), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))
		s.Require().Equal(int(0), len(callbackCounter.Timeout))

		s.Require().True(s.Run("verify query result", func() {
			s.Require().Nil(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Error)
			s.Require().NotNil(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success)
			s.Require().Len(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses, 3)

			s.Require().Len(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Bank.AllBalances.Amount, 1)
			s.Require().Equal(simd.Config().Denom, callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Bank.AllBalances.Amount[0].Denom)
			s.Require().Equal(expBalance.String(), string(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Bank.AllBalances.Amount[0].Amount))

			s.Require().Len(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[1].Bank.AllDenomMetadata.Metadata, 0)

			s.Require().Equal(simd.Config().Denom, callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[2].Bank.Supply.Amount.Denom)
			s.Require().Less(expSupply.Supply[0].Amount.String(), string(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[2].Bank.Supply.Amount.Amount))
		}))
	}))
}
