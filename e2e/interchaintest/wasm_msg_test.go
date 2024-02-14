package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"

	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos/wasm"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	mysuite "github.com/srdtrk/cw-ica-controller/interchaintest/v2/testsuite"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types"
	callbackcounter "github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/callback-counter"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/icacontroller"
)

type GetCountResponse struct {
	Count int64 `json:"count"`
}

func (s *ContractTestSuite) SetupWasmTestSuite(ctx context.Context, encoding icacontroller.TxEncoding) int {
	wasmChainSpecs := []*interchaintest.ChainSpec{
		chainSpecs[0],
		{
			ChainConfig: ibc.ChainConfig{
				Type:    "cosmos",
				Name:    "wasmd2",
				ChainID: "wasmd-2",
				Images: []ibc.DockerImage{
					{
						Repository: "cosmwasm/wasmd", // FOR LOCAL IMAGE USE: Docker Image Name
						Version:    "v0.50.0",        // FOR LOCAL IMAGE USE: Docker Image Tag
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
	s.SetupSuite(ctx, wasmChainSpecs)

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

	s.Contract.SetIcaAddress(contractState.IcaInfo.IcaAddress)

	s.Require().Equal(s.UserA.FormattedAddress(), *ownershipResponse.Owner)
	s.Require().Nil(ownershipResponse.PendingOwner)
	s.Require().Nil(ownershipResponse.PendingExpiry)

	counterCodeId, err := s.ChainB.StoreContract(ctx, s.UserB.KeyName(), "./test_data/simple_counter.wasm")
	s.Require().NoError(err)

	_, err = s.ChainB.StoreContract(ctx, s.UserB.KeyName(), "./test_data/migrate_counter.wasm")
	s.Require().NoError(err)

	counterCodeID, err := strconv.ParseUint(counterCodeId, 10, 64)
	s.Require().NoError(err)

	return int(counterCodeID)
}

func (s *ContractTestSuite) TestSendWasmMsgsProtobufEncoding() {
	s.SendWasmMsgsTestWithEncoding(icatypes.EncodingProtobuf)
}

// currently, Wasm is only supported with protobuf encoding
func (s *ContractTestSuite) SendWasmMsgsTestWithEncoding(encoding icacontroller.TxEncoding) {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	counterCodeID := s.SetupWasmTestSuite(ctx, encoding)
	wasmd, wasmd2 := s.ChainA, s.ChainB
	wasmdUser, wasmd2User := s.UserA, s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.Contract.IcaAddress)

	var counterContract *types.Contract
	s.Run(fmt.Sprintf("TestInstantiate-%s", encoding), func() {
		// Instantiate the contract:
		instantiateMsg := icacontroller.CosmosMsg_for_Empty{
			Wasm: &icacontroller.CosmosMsg_for_Empty_Wasm{
				Instantiate: &icacontroller.WasmMsg_Instantiate{
					Admin:  &s.Contract.IcaAddress,
					CodeId: counterCodeID,
					Label:  "counter",
					Msg:    icacontroller.Binary(toBase64(`{"count": 0}`)),
					Funds:  []icacontroller.Coin{},
				},
			},
		}

		// Execute the contract:
		sendCosmosMsgsExecMsg := icacontroller.ExecuteMsg{
			SendCosmosMsgs: &icacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []icacontroller.CosmosMsg_for_Empty{instantiateMsg},
			},
		}
		err := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, wasmd2)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(1), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		contractByCodeRequest := wasmtypes.QueryContractsByCodeRequest{
			CodeId: uint64(counterCodeID),
		}
		contractByCodeResp, err := mysuite.GRPCQuery[wasmtypes.QueryContractsByCodeResponse](ctx, wasmd2, &contractByCodeRequest)
		s.Require().NoError(err)
		s.Require().Len(contractByCodeResp.Contracts, 1)

		counterContract = types.NewContract(
			contractByCodeResp.Contracts[0],
			strconv.FormatUint(uint64(counterCodeID), 10),
			wasmd2,
		)

		counterState, err := types.QueryAnyMsg[GetCountResponse](ctx, counterContract, `{"get_count": {}}`)
		s.Require().NoError(err)

		s.Require().Equal(int64(0), counterState.Count)
	})

	var counter2Contract *types.Contract
	s.Run(fmt.Sprintf("TestExecuteAndInstantiate2AndClearAdminMsg-%s", encoding), func() {
		// Execute the contract:
		executeMsg := icacontroller.CosmosMsg_for_Empty{
			Wasm: &icacontroller.CosmosMsg_for_Empty_Wasm{
				Execute: &icacontroller.WasmMsg_Execute{
					ContractAddr: counterContract.Address,
					Msg:          icacontroller.Binary(toBase64(`{"increment": {}}`)),
					Funds:        []icacontroller.Coin{},
				},
			},
		}

		clearAdminMsg := icacontroller.CosmosMsg_for_Empty{
			Wasm: &icacontroller.CosmosMsg_for_Empty_Wasm{
				ClearAdmin: &icacontroller.WasmMsg_ClearAdmin{
					ContractAddr: counterContract.Address,
				},
			},
		}

		instantiate2Msg := icacontroller.CosmosMsg_for_Empty{
			Wasm: &icacontroller.CosmosMsg_for_Empty_Wasm{
				Instantiate2: &icacontroller.WasmMsg_Instantiate2{
					Admin:  &s.Contract.IcaAddress,
					CodeId: counterCodeID,
					Label:  "counter2",
					Msg:    icacontroller.Binary(toBase64(`{"count": 0}`)),
					Funds:  []icacontroller.Coin{},
					Salt:   icacontroller.Binary(toBase64("salt")),
				},
			},
		}

		// Execute the contract:
		sendCosmosMsgsExecMsg := icacontroller.ExecuteMsg{
			SendCosmosMsgs: &icacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []icacontroller.CosmosMsg_for_Empty{executeMsg, clearAdminMsg, instantiate2Msg},
			},
		}
		err := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, wasmd2)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		s.Require().Equal(uint64(2), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		counterState, err := types.QueryAnyMsg[GetCountResponse](ctx, counterContract, `{"get_count": {}}`)
		s.Require().NoError(err)

		s.Require().Equal(int64(1), counterState.Count)

		contractInfoRequest := wasmtypes.QueryContractInfoRequest{
			Address: counterContract.Address,
		}
		contractInfoResp, err := mysuite.GRPCQuery[wasmtypes.QueryContractInfoResponse](ctx, wasmd2, &contractInfoRequest)
		s.Require().NoError(err)

		s.Require().Equal("", contractInfoResp.ContractInfo.Admin)

		contractByCodeRequest := wasmtypes.QueryContractsByCodeRequest{
			CodeId: uint64(counterCodeID),
		}
		contractByCodeResp, err := mysuite.GRPCQuery[wasmtypes.QueryContractsByCodeResponse](ctx, wasmd2, &contractByCodeRequest)
		s.Require().NoError(err)
		s.Require().Len(contractByCodeResp.Contracts, 2)

		counter2Contract = types.NewContract(
			contractByCodeResp.Contracts[1],
			strconv.FormatUint(uint64(counterCodeID), 10),
			wasmd2,
		)
	})

	s.Run(fmt.Sprintf("TestMigrateAndUpdateAdmin-%s", encoding), func() {
		migrateMsg := icacontroller.CosmosMsg_for_Empty{
			Wasm: &icacontroller.CosmosMsg_for_Empty_Wasm{
				Migrate: &icacontroller.WasmMsg_Migrate{
					ContractAddr: counter2Contract.Address,
					NewCodeId:    counterCodeID + 1,
					Msg:          icacontroller.Binary(toBase64(`{}`)),
				},
			},
		}

		updateAdminMsg := icacontroller.CosmosMsg_for_Empty{
			Wasm: &icacontroller.CosmosMsg_for_Empty_Wasm{
				UpdateAdmin: &icacontroller.WasmMsg_UpdateAdmin{
					ContractAddr: counter2Contract.Address,
					Admin:        wasmd2User.FormattedAddress(),
				},
			},
		}

		// Execute the contract:
		sendCosmosMsgsExecMsg := icacontroller.ExecuteMsg{
			SendCosmosMsgs: &icacontroller.ExecuteMsg_SendCosmosMsgs{
				Messages: []icacontroller.CosmosMsg_for_Empty{migrateMsg, updateAdminMsg},
			},
		}
		err := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, wasmd2)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := types.QueryAnyMsg[callbackcounter.CallbackCounter](ctx, s.CallbackCounterContract, callbackcounter.GetCallbackCounterRequest)
		s.Require().NoError(err)

		// s.Require().Equal(uint64(1), callbackCounter.Error)
		s.Require().Equal(uint64(3), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		counterState, err := types.QueryAnyMsg[GetCountResponse](ctx, counter2Contract, `{"get_count": {}}`)
		s.Require().NoError(err)

		s.Require().Equal(int64(0), counterState.Count)

		contractInfoRequest := wasmtypes.QueryContractInfoRequest{
			Address: counter2Contract.Address,
		}
		contractInfoResp, err := mysuite.GRPCQuery[wasmtypes.QueryContractInfoResponse](ctx, wasmd2, &contractInfoRequest)
		s.Require().NoError(err)

		s.Require().Equal(counterCodeID+1, int(contractInfoResp.ContractInfo.CodeID))
		s.Require().Equal(wasmd2User.FormattedAddress(), contractInfoResp.ContractInfo.Admin)
	})
}

func toBase64(msg string) string {
	return base64.StdEncoding.EncodeToString([]byte(msg))
}
