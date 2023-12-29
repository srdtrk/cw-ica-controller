package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos/wasm"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"

	mysuite "github.com/srdtrk/cw-ica-controller/interchaintest/v2/testsuite"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types"
)

func (s *ContractTestSuite) SetupWasmTestSuite(ctx context.Context, encoding string) uint64 {
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
						Version:    "v0.45.0",        // FOR LOCAL IMAGE USE: Docker Image Tag
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

	codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	// Instantiate the contract with channel:
	instantiateMsg := types.NewInstantiateMsgWithChannelInitOptions(nil, s.ChainAConnID, s.ChainBConnID, nil, &encoding)

	contractAddr, err := s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), codeId, instantiateMsg, true, "--gas", "500000")
	s.Require().NoError(err)

	s.Contract = types.NewIcaContract(types.NewContract(contractAddr, codeId, s.ChainA))

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	contractState, err := s.Contract.QueryContractState(ctx)
	s.Require().NoError(err)

	ownershipResponse, err := s.Contract.QueryOwnership(ctx)
	s.Require().NoError(err)

	s.IcaAddress = contractState.IcaInfo.IcaAddress
	s.Contract.SetIcaAddress(s.IcaAddress)

	s.Require().Equal(s.UserA.FormattedAddress(), ownershipResponse.Owner)
	s.Require().Nil(ownershipResponse.PendingOwner)
	s.Require().Nil(ownershipResponse.PendingExpiry)

	counterCodeId, err := s.ChainB.StoreContract(ctx, s.UserB.KeyName(), "./test_data/simple_counter.wasm")
	s.Require().NoError(err)

	_, err = s.ChainB.StoreContract(ctx, s.UserB.KeyName(), "./test_data/migrate_counter.wasm")
	s.Require().NoError(err)

	counterCodeID, err := strconv.ParseUint(counterCodeId, 10, 64)
	s.Require().NoError(err)

	return counterCodeID
}

func (s *ContractTestSuite) TestSendWasmMsgsProto3JsonEncoding() {
	s.SendWasmMsgsTestWithEncoding(icatypes.EncodingProto3JSON)
}

func (s *ContractTestSuite) TestSendWasmMsgsProtobufEncoding() {
	s.SendWasmMsgsTestWithEncoding(icatypes.EncodingProtobuf)
}

func (s *ContractTestSuite) SendWasmMsgsTestWithEncoding(encoding string) {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	counterCodeID := s.SetupWasmTestSuite(ctx, encoding)
	wasmd, wasmd2 := s.ChainA, s.ChainB
	wasmdUser, wasmd2User := s.UserA, s.UserB

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaAddress)

	var counterAddress string
	s.Run(fmt.Sprintf("TestInstantiate-%s", encoding), func() {
		// Instantiate the contract:
		instantiateMsg := types.ContractCosmosMsg{
			Wasm: &types.WasmCosmosMsg{
				Instantiate: &types.WasmInstantiateCosmosMsg{
					Admin:  s.IcaAddress,
					CodeID: counterCodeID,
					Label:  "counter",
					Msg:    toBase64(`{"count": 0}`),
					Funds:  []types.Coin{},
				},
			},
		}

		// Execute the contract:
		err := s.Contract.ExecSendCosmosMsgs(ctx, wasmdUser.KeyName(), []types.ContractCosmosMsg{instantiateMsg}, nil, nil)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, wasmd2)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := s.Contract.QueryCallbackCounter(ctx)
		s.Require().NoError(err)

		s.Require().Equal(uint64(1), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		contractByCodeQuerier := mysuite.NewGRPCQuerier[wasmtypes.QueryContractsByCodeResponse](s.T(), wasmd2, "/cosmwasm.wasm.v1.Query/ContractsByCode")

		contractByCodeRequest := wasmtypes.QueryContractsByCodeRequest{
			CodeId: counterCodeID,
		}
		contractByCodeResp, err := contractByCodeQuerier.GRPCQuery(ctx, &contractByCodeRequest)
		s.Require().NoError(err)
		s.Require().Len(contractByCodeResp.Contracts, 1)

		counterAddress = contractByCodeResp.Contracts[0]

		counterState, err := types.QueryContract[types.GetCountResponse](ctx, wasmd2, counterAddress, `{"get_count": {}}`)
		s.Require().NoError(err)

		s.Require().Equal(int64(0), counterState.Count)
	})

	var counter2Address string
	s.Run(fmt.Sprintf("TestExecuteAndInstantiate2AndClearAdminMsg-%s", encoding), func() {
		// Execute the contract:
		executeMsg := types.ContractCosmosMsg{
			Wasm: &types.WasmCosmosMsg{
				Execute: &types.WasmExecuteCosmosMsg{
					ContractAddr: counterAddress,
					Msg:          toBase64(`{"increment": {}}`),
					Funds:        []types.Coin{},
				},
			},
		}

		clearAdminMsg := types.ContractCosmosMsg{
			Wasm: &types.WasmCosmosMsg{
				ClearAdmin: &types.WasmClearAdminCosmosMsg{
					ContractAddr: counterAddress,
				},
			},
		}

		instantiate2Msg := types.ContractCosmosMsg{
			Wasm: &types.WasmCosmosMsg{
				Instantiate2: &types.WasmInstantiate2CosmosMsg{
					Admin:  s.IcaAddress,
					CodeID: counterCodeID,
					Label:  "counter2",
					Msg:    toBase64(`{"count": 0}`),
					Funds:  []types.Coin{},
					Salt:   toBase64("salt"),
				},
			},
		}

		// Execute the contract:
		err := s.Contract.ExecSendCosmosMsgs(ctx, wasmdUser.KeyName(), []types.ContractCosmosMsg{executeMsg, clearAdminMsg, instantiate2Msg}, nil, nil)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, wasmd2)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := s.Contract.QueryCallbackCounter(ctx)
		s.Require().NoError(err)

		s.Require().Equal(uint64(2), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		counterState, err := types.QueryContract[types.GetCountResponse](ctx, wasmd2, counterAddress, `{"get_count": {}}`)
		s.Require().NoError(err)

		s.Require().Equal(int64(1), counterState.Count)

		contractInfoQuerier := mysuite.NewGRPCQuerier[wasmtypes.QueryContractInfoResponse](s.T(), wasmd2, "/cosmwasm.wasm.v1.Query/ContractInfo")

		contractInfoRequest := wasmtypes.QueryContractInfoRequest{
			Address: counterAddress,
		}
		contractInfoResp, err := contractInfoQuerier.GRPCQuery(ctx, &contractInfoRequest)
		s.Require().NoError(err)

		s.Require().Equal("", contractInfoResp.ContractInfo.Admin)

		contractByCodeQuerier := mysuite.NewGRPCQuerier[wasmtypes.QueryContractsByCodeResponse](s.T(), wasmd2, "/cosmwasm.wasm.v1.Query/ContractsByCode")

		contractByCodeRequest := wasmtypes.QueryContractsByCodeRequest{
			CodeId: counterCodeID,
		}
		contractByCodeResp, err := contractByCodeQuerier.GRPCQuery(ctx, &contractByCodeRequest)
		s.Require().NoError(err)
		s.Require().Len(contractByCodeResp.Contracts, 2)

		counter2Address = contractByCodeResp.Contracts[1]
	})

	s.Run(fmt.Sprintf("TestMigrateAndUpdateAdmin-%s", encoding), func() {
		migrateMsg := types.ContractCosmosMsg{
			Wasm: &types.WasmCosmosMsg{
				Migrate: &types.WasmMigrateCosmosMsg{
					ContractAddr: counter2Address,
					NewCodeID:    counterCodeID + 1,
					Msg:          toBase64(`{}`),
				},
			},
		}

		updateAdminMsg := types.ContractCosmosMsg{
			Wasm: &types.WasmCosmosMsg{
				UpdateAdmin: &types.WasmUpdateAdminCosmosMsg{
					ContractAddr: counter2Address,
					Admin:        wasmd2User.FormattedAddress(),
				},
			},
		}

		// Execute the contract:
		err := s.Contract.ExecSendCosmosMsgs(ctx, wasmdUser.KeyName(), []types.ContractCosmosMsg{migrateMsg, updateAdminMsg}, nil, nil)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 5, wasmd, wasmd2)
		s.Require().NoError(err)

		// Check if contract callbacks were executed:
		callbackCounter, err := s.Contract.QueryCallbackCounter(ctx)
		s.Require().NoError(err)

		// s.Require().Equal(uint64(1), callbackCounter.Error)
		s.Require().Equal(uint64(3), callbackCounter.Success)
		s.Require().Equal(uint64(0), callbackCounter.Error)

		counterState, err := types.QueryContract[types.GetCountResponse](ctx, wasmd2, counter2Address, `{"get_count": {}}`)
		s.Require().NoError(err)

		s.Require().Equal(int64(0), counterState.Count)

		contractInfoQuerier := mysuite.NewGRPCQuerier[wasmtypes.QueryContractInfoResponse](s.T(), wasmd2, "/cosmwasm.wasm.v1.Query/ContractInfo")

		contractInfoRequest := wasmtypes.QueryContractInfoRequest{
			Address: counter2Address,
		}
		contractInfoResp, err := contractInfoQuerier.GRPCQuery(ctx, &contractInfoRequest)
		s.Require().NoError(err)

		s.Require().Equal(counterCodeID+1, contractInfoResp.ContractInfo.CodeID)
		s.Require().Equal(wasmd2User.FormattedAddress(), contractInfoResp.ContractInfo.Admin)
	})
}

func toBase64(msg string) string {
	return base64.StdEncoding.EncodeToString([]byte(msg))
}
