package main

import (
	"context"
	"encoding/base64"
	"strconv"

	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos/wasm"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	"github.com/srdtrk/go-codegen/e2esuite/v8/chainconfig"
	"github.com/srdtrk/go-codegen/e2esuite/v8/e2esuite"
	"github.com/srdtrk/go-codegen/e2esuite/v8/types/callbackcounter"
	"github.com/srdtrk/go-codegen/e2esuite/v8/types/cwicacontroller"
	"github.com/srdtrk/go-codegen/e2esuite/v8/types/simplecounter"
)

func (s *ContractTestSuite) SetupWasmTestSuite(ctx context.Context) int {
	chainconfig.DefaultChainSpecs = []*interchaintest.ChainSpec{
		chainconfig.DefaultChainSpecs[0],
		{
			ChainConfig: ibc.ChainConfig{
				Type:    "cosmos",
				Name:    "wasmd2",
				ChainID: "wasmd-2",
				Images: []ibc.DockerImage{
					{
						Repository: "cosmwasm/wasmd", // FOR LOCAL IMAGE USE: Docker Image Name
						Version:    "v0.51.0",        // FOR LOCAL IMAGE USE: Docker Image Tag
						UidGid:     "1025:1025",
					},
				},
				Bin:           "wasmd",
				Bech32Prefix:  "wasm",
				Denom:         "stake",
				GasPrices:     "0.00stake",
				GasAdjustment: 1.3,
				// cannot run wasmd commands without wasm encoding
				EncodingConfig: wasm.WasmEncoding(),
				TrustingPeriod: "508h",
				NoHostMount:    false,
			},
		},
	}

	s.SetupSuite(ctx)

	codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/callback_counter.wasm")
	s.Require().Error(err)

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
		},
		SendCallbacksTo: &s.CallbackCounterContract.Address,
	}

	s.Contract, err = cwicacontroller.Instantiate(ctx, s.UserA.KeyName(), codeId, "", s.ChainA, instantiateMsg, "--gas", "500000")
	s.Require().NoError(err)

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	// Query the contract state:
	contractState, err := s.Contract.QueryClient().GetContractState(ctx, &cwicacontroller.QueryMsg_GetContractState{})
	s.Require().NoError(err)

	s.IcaContractToAddrMap[s.Contract.Address] = contractState.IcaInfo.IcaAddress

	// Check the ownership:
	ownershipResponse, err := s.Contract.QueryClient().Ownership(ctx, &cwicacontroller.QueryMsg_Ownership{})
	s.Require().NoError(err)
	s.Require().Equal(s.UserA.FormattedAddress(), *ownershipResponse.Owner)
	s.Require().Nil(ownershipResponse.PendingOwner)
	s.Require().Nil(ownershipResponse.PendingExpiry)

	counterCodeId, err := s.ChainB.StoreContract(ctx, s.UserB.KeyName(), "./testdata/simplecounter.wasm")
	s.Require().NoError(err)

	_, err = s.ChainB.StoreContract(ctx, s.UserB.KeyName(), "./testdata/migratecounter.wasm")
	s.Require().NoError(err)

	counterCodeID, err := strconv.ParseUint(counterCodeId, 10, 64)
	s.Require().NoError(err)

	return int(counterCodeID)
}

func (s *ContractTestSuite) TestSendWasmMsgsProtobufEncoding() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	counterCodeID := s.SetupWasmTestSuite(ctx)
	wasmd, wasmd2 := s.ChainA, s.ChainB
	wasmdUser, wasmd2User := s.UserA, s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaContractToAddrMap[s.Contract.Address])

	var counterContract *simplecounter.Contract
	s.Require().True(s.Run("TestInstantiate", func() {
		icaAddress := s.IcaContractToAddrMap[s.Contract.Address]

		// Instantiate the contract:
		instantiateMsg := cwicacontroller.CosmosMsg_for_Empty{
			Wasm: &cwicacontroller.CosmosMsg_for_Empty_Wasm{
				Instantiate: &cwicacontroller.WasmMsg_Instantiate{
					Admin:  &icaAddress,
					CodeId: counterCodeID,
					Label:  "counter",
					Msg:    cwicacontroller.Binary(toBase64(`{"count": 0}`)),
					Funds:  []cwicacontroller.Coin{},
				},
			},
		}

		// Execute the contract:
		sendCosmosMsgsExecMsg := cwicacontroller.ExecuteMsg{
			SendCosmosMsgs: &cwicacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []cwicacontroller.CosmosMsg_for_Empty{instantiateMsg},
			},
		}
		_, err := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, wasmd2)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(1), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))

		contractByCodeRequest := wasmtypes.QueryContractsByCodeRequest{
			CodeId: uint64(counterCodeID),
		}
		contractByCodeResp, err := e2esuite.GRPCQuery[wasmtypes.QueryContractsByCodeResponse](ctx, wasmd2, &contractByCodeRequest)
		s.Require().NoError(err)
		s.Require().Len(contractByCodeResp.Contracts, 1)

		counterContract, err = simplecounter.NewContract(contractByCodeResp.Contracts[0], strconv.FormatUint(uint64(counterCodeID), 10), wasmd2)
		s.Require().NoError(err)

		// Query the simple counter state:
		counterState, err := counterContract.QueryClient().GetCount(ctx, &simplecounter.QueryMsg_GetCount{})
		s.Require().NoError(err)
		s.Require().Equal(int(0), counterState.Count)
	}))

	var counterContract2 *simplecounter.Contract
	s.Require().True(s.Run("TestExecuteAndInstantiate2AndClearAdminMsg", func() {
		icaAddress := s.IcaContractToAddrMap[s.Contract.Address]

		// Execute the contract:
		executeMsg := cwicacontroller.CosmosMsg_for_Empty{
			Wasm: &cwicacontroller.CosmosMsg_for_Empty_Wasm{
				Execute: &cwicacontroller.WasmMsg_Execute{
					ContractAddr: counterContract.Address,
					Msg:          cwicacontroller.Binary(toBase64(`{"increment": {}}`)),
					Funds:        []cwicacontroller.Coin{},
				},
			},
		}

		clearAdminMsg := cwicacontroller.CosmosMsg_for_Empty{
			Wasm: &cwicacontroller.CosmosMsg_for_Empty_Wasm{
				ClearAdmin: &cwicacontroller.WasmMsg_ClearAdmin{
					ContractAddr: counterContract.Address,
				},
			},
		}

		instantiate2Msg := cwicacontroller.CosmosMsg_for_Empty{
			Wasm: &cwicacontroller.CosmosMsg_for_Empty_Wasm{
				Instantiate2: &cwicacontroller.WasmMsg_Instantiate2{
					Admin:  &icaAddress,
					CodeId: counterCodeID,
					Label:  "counter2",
					Msg:    cwicacontroller.Binary(toBase64(`{"count": 0}`)),
					Funds:  []cwicacontroller.Coin{},
					Salt:   cwicacontroller.Binary(toBase64("salt")),
				},
			},
		}

		// Execute the contract:
		sendCosmosMsgsExecMsg := cwicacontroller.ExecuteMsg{
			SendCosmosMsgs: &cwicacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []cwicacontroller.CosmosMsg_for_Empty{executeMsg, clearAdminMsg, instantiate2Msg},
			},
		}
		_, err := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, wasmd2)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(2), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))

		// Query the simple counter state:
		counterState, err := counterContract.QueryClient().GetCount(ctx, &simplecounter.QueryMsg_GetCount{})
		s.Require().NoError(err)
		s.Require().Equal(int(1), counterState.Count)

		contractInfoRequest := wasmtypes.QueryContractInfoRequest{
			Address: counterContract.Address,
		}
		contractInfoResp, err := e2esuite.GRPCQuery[wasmtypes.QueryContractInfoResponse](ctx, wasmd2, &contractInfoRequest)
		s.Require().NoError(err)

		s.Require().Equal("", contractInfoResp.ContractInfo.Admin)

		contractByCodeRequest := wasmtypes.QueryContractsByCodeRequest{
			CodeId: uint64(counterCodeID),
		}
		contractByCodeResp, err := e2esuite.GRPCQuery[wasmtypes.QueryContractsByCodeResponse](ctx, wasmd2, &contractByCodeRequest)
		s.Require().NoError(err)
		s.Require().Len(contractByCodeResp.Contracts, 2)

		counterContract2, err = simplecounter.NewContract(contractByCodeResp.Contracts[1], strconv.FormatUint(uint64(counterCodeID), 10), wasmd2)
		s.Require().NoError(err)
	}))

	s.Require().True(s.Run("TestMigrateAndUpdateAdmin", func() {
		migrateMsg := cwicacontroller.CosmosMsg_for_Empty{
			Wasm: &cwicacontroller.CosmosMsg_for_Empty_Wasm{
				Migrate: &cwicacontroller.WasmMsg_Migrate{
					ContractAddr: counterContract2.Address,
					NewCodeId:    counterCodeID + 1,
					Msg:          cwicacontroller.Binary(toBase64(`{}`)),
				},
			},
		}

		updateAdminMsg := cwicacontroller.CosmosMsg_for_Empty{
			Wasm: &cwicacontroller.CosmosMsg_for_Empty_Wasm{
				UpdateAdmin: &cwicacontroller.WasmMsg_UpdateAdmin{
					ContractAddr: counterContract2.Address,
					Admin:        wasmd2User.FormattedAddress(),
				},
			},
		}

		// Execute the contract:
		sendCosmosMsgsExecMsg := cwicacontroller.ExecuteMsg{
			SendCosmosMsgs: &cwicacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []cwicacontroller.CosmosMsg_for_Empty{migrateMsg, updateAdminMsg},
			},
		}
		_, err := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, wasmd2)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := s.CallbackCounterContract.QueryClient().GetCallbackCounter(ctx, &callbackcounter.QueryMsg_GetCallbackCounter{})
		s.Require().NoError(err)
		s.Require().Equal(int(3), len(callbackCounter.Success))
		s.Require().Equal(int(0), len(callbackCounter.Error))

		// Query the simple counter state:
		counterState, err := counterContract2.QueryClient().GetCount(ctx, &simplecounter.QueryMsg_GetCount{})
		s.Require().NoError(err)
		s.Require().Equal(int(0), counterState.Count)

		contractInfoResp, err := e2esuite.GRPCQuery[wasmtypes.QueryContractInfoResponse](ctx, wasmd2, &wasmtypes.QueryContractInfoRequest{
			Address: counterContract2.Address,
		})
		s.Require().NoError(err)
		s.Require().Equal(counterCodeID+1, int(contractInfoResp.ContractInfo.CodeID))
		s.Require().Equal(wasmd2User.FormattedAddress(), contractInfoResp.ContractInfo.Admin)
	}))
}

func toBase64(msg string) string {
	return base64.StdEncoding.EncodeToString([]byte(msg))
}
