package main

import (
	"context"

	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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
func (s *ContractTestSuite) TestBankAndStargateQueries() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, cwicacontroller.IbcOrder_OrderUnordered)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser, simdUser := s.UserA, s.UserB

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

	s.Require().True(s.Run("Other bank queries and stargate query", func() {
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
					{
						Stargate: cwicacontroller.NewStargateQuery_FromProto("/cosmos.auth.v1beta1.Query/ModuleAccountByName", &authtypes.QueryModuleAccountByNameRequest{
							Name: govtypes.ModuleName,
						}),
					},
				},
			},
		}

		expBalance, err := simd.GetBalance(ctx, simdUser.FormattedAddress(), simd.Config().Denom)
		s.Require().NoError(err)
		expSupply, err := e2esuite.GRPCQuery[banktypes.QueryTotalSupplyResponse](ctx, simd, &banktypes.QueryTotalSupplyRequest{})
		s.Require().NoError(err)

		expStargateResp, err := e2esuite.GRPCQuery[authtypes.QueryModuleAccountByNameResponse](ctx, simd, &authtypes.QueryModuleAccountByNameRequest{
			Name: govtypes.ModuleName,
		})
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
			s.Require().Len(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses, 4)

			s.Require().Len(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Bank.AllBalances.Amount, 1)
			s.Require().Equal(simd.Config().Denom, callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Bank.AllBalances.Amount[0].Denom)
			s.Require().Equal(expBalance.String(), string(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Bank.AllBalances.Amount[0].Amount))

			s.Require().Len(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[1].Bank.AllDenomMetadata.Metadata, 0)

			s.Require().Equal(simd.Config().Denom, callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[2].Bank.Supply.Amount.Denom)
			s.Require().Less(expSupply.Supply[0].Amount.String(), string(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[2].Bank.Supply.Amount.Amount))

			s.Require().Equal("/cosmos.auth.v1beta1.Query/ModuleAccountByName", callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[3].Stargate.Path)

			resp := &authtypes.QueryModuleAccountByNameResponse{}
			err = proto.Unmarshal(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[3].Stargate.Data.Unwrap(), resp)
			s.Require().NoError(err)

			s.Require().Equal(expStargateResp, resp)
		}))
	}))
}

// `TestStakingQuery` tests all allowed staking queries in the SendCosmosMsgs message. The following queries are tested:
// - Staking::Delegation
// - Staking::AllDelegations
// - Staking::Validator
// - Staking::AllValidator
// - Staking::BondedDenom
func (s *ContractTestSuite) TestStakingQueries() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, cwicacontroller.IbcOrder_OrderUnordered)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser, _ := s.UserA, s.UserB

	// Fund the ICA address:
	icaAddress := s.IcaContractToAddrMap[s.Contract.Address]
	s.FundAddressChainB(ctx, icaAddress)

	var validator string
	s.Require().True(s.Run("Query after msg", func() {
		var err error
		validator, err = simd.Validators[0].KeyBech32(ctx, "validator", "val")
		s.Require().NoError(err)

		// Stake some tokens through CosmosMsgs:
		stakeAmount := cwicacontroller.Coin{
			Denom:  simd.Config().Denom,
			Amount: "10000000",
		}
		stakeCosmosMsg := cwicacontroller.CosmosMsg_for_Empty{
			Staking: &cwicacontroller.CosmosMsg_for_Empty_Staking{
				Delegate: &cwicacontroller.StakingMsg_Delegate{
					Validator: validator,
					Amount:    stakeAmount,
				},
			},
		}
		// Execute the contract:
		execMsgWithQueries := cwicacontroller.ExecuteMsg{
			SendCosmosMsgs: &cwicacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []cwicacontroller.CosmosMsg_for_Empty{stakeCosmosMsg},
				Queries: []cwicacontroller.QueryRequest_for_Empty{
					{
						Staking: &cwicacontroller.QueryRequest_for_Empty_Staking{
							Delegation: &cwicacontroller.StakingQuery_Delegation{
								Validator: validator,
								Delegator: icaAddress,
							},
						},
					},
					{
						Staking: &cwicacontroller.QueryRequest_for_Empty_Staking{
							AllDelegations: &cwicacontroller.StakingQuery_AllDelegations{
								Delegator: icaAddress,
							},
						},
					},
				},
			},
		}
		_, err = s.Contract.Execute(ctx, wasmdUser.KeyName(), execMsgWithQueries)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(1), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))
		s.Require().Equal(int(0), len(callbackCounter.Timeout))

		s.Require().True(s.Run("verify query result", func() {
			s.Require().Nil(callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Error)
			s.Require().NotNil(callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success)
			s.Require().Len(callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses, 2)

			s.Require().Equal(validator, callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Staking.Delegation.Delegation.Validator)
			s.Require().Equal(icaAddress, callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Staking.Delegation.Delegation.Delegator)
			s.Require().Equal(stakeAmount.Denom, callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Staking.Delegation.Delegation.Amount.Denom)
			s.Require().Equal(string(stakeAmount.Amount), string(callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Staking.Delegation.Delegation.Amount.Amount))

			s.Require().Len(callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[1].Staking.AllDelegations.Delegations, 1)
			s.Require().Equal(validator, callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[1].Staking.AllDelegations.Delegations[0].Validator)
			s.Require().Equal(icaAddress, callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[1].Staking.AllDelegations.Delegations[0].Delegator)
			s.Require().Equal(stakeAmount.Denom, callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[1].Staking.AllDelegations.Delegations[0].Amount.Denom)
			s.Require().Equal(string(stakeAmount.Amount), string(callbackCounter.Success[0].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[1].Staking.AllDelegations.Delegations[0].Amount.Amount))
		}))
	}))

	s.Require().True(s.Run("Other staking queries", func() {
		execMsgWithQueries := cwicacontroller.ExecuteMsg{
			SendCosmosMsgs: &cwicacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []cwicacontroller.CosmosMsg_for_Empty{},
				Queries: []cwicacontroller.QueryRequest_for_Empty{
					{
						Staking: &cwicacontroller.QueryRequest_for_Empty_Staking{
							Validator: &cwicacontroller.StakingQuery_Validator{
								Address: validator,
							},
						},
					},
					{
						Staking: &cwicacontroller.QueryRequest_for_Empty_Staking{
							AllValidators: &cwicacontroller.StakingQuery_AllValidators{},
						},
					},
					{
						Staking: &cwicacontroller.QueryRequest_for_Empty_Staking{
							BondedDenom: &cwicacontroller.StakingQuery_BondedDenom{},
						},
					},
				},
			},
		}

		_, err := s.Contract.Execute(ctx, wasmdUser.KeyName(), execMsgWithQueries)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, simd)
		s.Require().NoError(err)

		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(2), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))
		s.Require().Equal(int(0), len(callbackCounter.Timeout))

		valResp, err := e2esuite.GRPCQuery[stakingtypes.QueryValidatorResponse](ctx, simd, &stakingtypes.QueryValidatorRequest{ValidatorAddr: validator})
		s.Require().NoError(err)

		s.Require().True(s.Run("verify query result", func() {
			s.Require().Nil(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Error)
			s.Require().NotNil(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success)
			s.Require().Len(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses, 3)

			s.Require().Equal(validator, callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Staking.Validator.Validator.Address)
			s.Require().Equal(valResp.Validator.Commission.CommissionRates.Rate.BigInt().String(), string(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Staking.Validator.Validator.Commission))
			s.Require().Equal(valResp.Validator.Commission.CommissionRates.MaxRate.BigInt().String(), string(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Staking.Validator.Validator.MaxCommission))
			s.Require().Equal(valResp.Validator.Commission.CommissionRates.MaxChangeRate.BigInt().String(), string(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Staking.Validator.Validator.MaxChangeRate))

			s.Require().Len(callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[1].Staking.AllValidators.Validators, 2)
			s.Require().Equal(*callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[0].Staking.Validator.Validator, callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[1].Staking.AllValidators.Validators[0])

			s.Require().Equal(simd.Config().Denom, callbackCounter.Success[1].OnAcknowledgementPacketCallback.QueryResult.Success.Responses[2].Staking.BondedDenom.Denom)
		}))
	}))
}
