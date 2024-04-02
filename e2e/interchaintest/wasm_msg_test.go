package main

import (
	"context"
	"encoding/base64"
	"strconv"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos/wasm"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/chainconfig"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/e2esuite"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types"
	callbackcounter "github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/callback-counter"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/icacontroller"
	simplecounter "github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/simple-counter"
)

func (s *ContractTestSuite) SetupWasmTestSuite(ctx context.Context) int {
	wasmChainSpecs := []*interchaintest.ChainSpec{
		chainconfig.DefaultChainSpecs[0],
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

	s.CallbackCounterContract, err = types.Instantiate[callbackcounter.InstantiateMsg, callbackcounter.ExecuteMsg, callbackcounter.QueryMsg](ctx, s.UserA.KeyName(), codeId, s.ChainA, callbackcounter.InstantiateMsg{})
	s.Require().NoError(err)

	codeId, err = s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	// Instantiate the contract with channel:
	instantiateMsg := icacontroller.InstantiateMsg{
		Owner: nil,
		ChannelOpenInitOptions: icacontroller.ChannelOpenInitOptions{
			ConnectionId:             s.ChainAConnID,
			CounterpartyConnectionId: s.ChainBConnID,
			CounterpartyPortId:       nil,
		},
		SendCallbacksTo: &s.CallbackCounterContract.Address,
	}

	s.Contract, err = types.Instantiate[icacontroller.InstantiateMsg, icacontroller.ExecuteMsg, icacontroller.QueryMsg](ctx, s.UserA.KeyName(), codeId, s.ChainA, instantiateMsg, "--gas", "500000")
	s.Require().NoError(err)

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	// Query the contract state:
	contractState := &icacontroller.State_2{}
	err = s.Contract.Query(ctx, icacontroller.GetContractStateRequest, contractState)
	s.Require().NoError(err)

	// Check the ownership:
	ownershipResponse := &icacontroller.Ownership_for_String{}
	err = s.Contract.Query(ctx, icacontroller.OwnershipRequest, ownershipResponse)
	s.Require().NoError(err)

	s.IcaContractToAddrMap[s.Contract.Address] = contractState.IcaInfo.IcaAddress

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
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	counterCodeID := s.SetupWasmTestSuite(ctx)
	wasmd, wasmd2 := s.ChainA, s.ChainB
	wasmdUser, wasmd2User := s.UserA, s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaContractToAddrMap[s.Contract.Address])

	var counterContract *types.Contract[simplecounter.InstantiateMsg, simplecounter.ExecuteMsg, simplecounter.QueryMsg]
	s.Run("TestInstantiate", func() {
		icaAddress := s.IcaContractToAddrMap[s.Contract.Address]

		// Instantiate the contract:
		instantiateMsg := icacontroller.CosmosMsg_for_Empty{
			Wasm: &icacontroller.CosmosMsg_for_Empty_Wasm{
				Instantiate: &icacontroller.WasmMsg_Instantiate{
					Admin:  &icaAddress,
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
		_, err := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, wasmd2)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter := &callbackcounter.CallbackCounter{}
		err = s.CallbackCounterContract.Query(ctx, callbackcounter.GetCallbackCounterRequest, callbackCounter)
		s.Require().NoError(err)

		s.Require().Equal(int(1), callbackCounter.Success)
		s.Require().Equal(int(0), callbackCounter.Error)

		contractByCodeRequest := wasmtypes.QueryContractsByCodeRequest{
			CodeId: uint64(counterCodeID),
		}
		contractByCodeResp, err := e2esuite.GRPCQuery[wasmtypes.QueryContractsByCodeResponse](ctx, wasmd2, &contractByCodeRequest)
		s.Require().NoError(err)
		s.Require().Len(contractByCodeResp.Contracts, 1)

		counterContract = &types.Contract[simplecounter.InstantiateMsg, simplecounter.ExecuteMsg, simplecounter.QueryMsg]{
			Address: contractByCodeResp.Contracts[0],
			CodeID:  strconv.FormatUint(uint64(counterCodeID), 10),
			Chain:   wasmd2,
		}

		// Query the simple counter state:
		counterState := &simplecounter.GetCountResponse{}
		err = counterContract.Query(ctx, simplecounter.GetCountRequest, counterState)
		s.Require().NoError(err)

		s.Require().Equal(int(0), counterState.Count)
	})

	var counterContract2 *types.Contract[simplecounter.InstantiateMsg, simplecounter.ExecuteMsg, simplecounter.QueryMsg]
	s.Run("TestExecuteAndInstantiate2AndClearAdminMsg", func() {
		icaAddress := s.IcaContractToAddrMap[s.Contract.Address]

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
					Admin:  &icaAddress,
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
		_, err := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, wasmd2)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter := &callbackcounter.CallbackCounter{}
		err = s.CallbackCounterContract.Query(ctx, callbackcounter.GetCallbackCounterRequest, callbackCounter)
		s.Require().NoError(err)

		s.Require().Equal(int(2), callbackCounter.Success)
		s.Require().Equal(int(0), callbackCounter.Error)

		// Query the simple counter state:
		counterState := &simplecounter.GetCountResponse{}
		err = counterContract.Query(ctx, simplecounter.GetCountRequest, counterState)
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

		counterContract2 = &types.Contract[simplecounter.InstantiateMsg, simplecounter.ExecuteMsg, simplecounter.QueryMsg]{
			Address: contractByCodeResp.Contracts[1],
			CodeID:  strconv.FormatUint(uint64(counterCodeID), 10),
			Chain:   wasmd2,
		}
	})

	s.Run("TestMigrateAndUpdateAdmin", func() {
		migrateMsg := icacontroller.CosmosMsg_for_Empty{
			Wasm: &icacontroller.CosmosMsg_for_Empty_Wasm{
				Migrate: &icacontroller.WasmMsg_Migrate{
					ContractAddr: counterContract2.Address,
					NewCodeId:    counterCodeID + 1,
					Msg:          icacontroller.Binary(toBase64(`{}`)),
				},
			},
		}

		updateAdminMsg := icacontroller.CosmosMsg_for_Empty{
			Wasm: &icacontroller.CosmosMsg_for_Empty_Wasm{
				UpdateAdmin: &icacontroller.WasmMsg_UpdateAdmin{
					ContractAddr: counterContract2.Address,
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
		_, err := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendCosmosMsgsExecMsg)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, wasmd2)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter := &callbackcounter.CallbackCounter{}
		err = s.CallbackCounterContract.Query(ctx, callbackcounter.GetCallbackCounterRequest, callbackCounter)
		s.Require().NoError(err)

		s.Require().Equal(int(3), callbackCounter.Success)
		s.Require().Equal(int(0), callbackCounter.Error)

		// Query the simple counter state:
		counterState := &simplecounter.GetCountResponse{}
		err = counterContract2.Query(ctx, simplecounter.GetCountRequest, counterState)
		s.Require().NoError(err)

		s.Require().Equal(int(0), counterState.Count)

		contractInfoRequest := wasmtypes.QueryContractInfoRequest{
			Address: counterContract2.Address,
		}
		contractInfoResp, err := e2esuite.GRPCQuery[wasmtypes.QueryContractInfoResponse](ctx, wasmd2, &contractInfoRequest)
		s.Require().NoError(err)

		s.Require().Equal(counterCodeID+1, int(contractInfoResp.ContractInfo.CodeID))
		s.Require().Equal(wasmd2User.FormattedAddress(), contractInfoResp.ContractInfo.Admin)
	})
}

func toBase64(msg string) string {
	return base64.StdEncoding.EncodeToString([]byte(msg))
}
