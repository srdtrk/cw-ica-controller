package e2esuite

import (
	"context"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	grouptypes "github.com/cosmos/cosmos-sdk/x/group"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	proposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	feetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	v7migrations "github.com/cosmos/ibc-go/v8/modules/core/02-client/migrations/v7"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	solomachine "github.com/cosmos/ibc-go/v8/modules/light-clients/06-solomachine"
	ibctmtypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	localhost "github.com/cosmos/ibc-go/v8/modules/light-clients/09-localhost"
	simappparams "github.com/cosmos/ibc-go/v8/testing/simapp/params"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
)

// FundAddressChainA sends funds to the given address on chain A.
// The amount sent is 1,000,000,000 of the chain's denom.
func (s *TestSuite) FundAddressChainA(ctx context.Context, address string) {
	s.fundAddress(ctx, s.ChainA, s.UserA.KeyName(), address)
}

// FundAddressChainB sends funds to the given address on chain B.
// The amount sent is 1,000,000,000 of the chain's denom.
func (s *TestSuite) FundAddressChainB(ctx context.Context, address string) {
	s.fundAddress(ctx, s.ChainB, s.UserB.KeyName(), address)
}

// GetModuleAddress returns the address of the given module on the given chain.
// Added because interchaintest's method doesn't work.
func (s *TestSuite) GetModuleAddress(ctx context.Context, chain *cosmos.CosmosChain, moduleName string) string {
	modAccResp, err := GRPCQuery[authtypes.QueryModuleAccountByNameResponse](
		ctx, chain, &authtypes.QueryModuleAccountByNameRequest{Name: moduleName},
	)
	s.Require().NoError(err)

	cfg := chain.Config().EncodingConfig
	var account sdk.AccountI
	err = cfg.InterfaceRegistry.UnpackAny(modAccResp.Account, &account)
	s.Require().NoError(err)

	govAccount, ok := account.(authtypes.ModuleAccountI)
	s.Require().True(ok)
	s.Require().NotEmpty(govAccount.GetAddress().String())

	return govAccount.GetAddress().String()
}

// fundAddress sends funds to the given address on the given chain
func (s *TestSuite) fundAddress(ctx context.Context, chain *cosmos.CosmosChain, keyName, address string) {
	err := chain.SendFunds(ctx, keyName, ibc.WalletAmount{
		Address: address,
		Denom:   chain.Config().Denom,
		Amount:  math.NewInt(1_000_000_000),
	})
	s.Require().NoError(err)

	// wait for 2 blocks for the funds to be received
	err = testutil.WaitForBlocks(ctx, 2, chain)
	s.Require().NoError(err)
}

// EncodingConfig returns the global E2E encoding config.
// It includes CosmosSDK, IBC, and Wasm messages
func EncodingConfig() *sdktestutil.TestEncodingConfig {
	_, cfg := codecAndEncodingConfig()
	return &sdktestutil.TestEncodingConfig{
		InterfaceRegistry: cfg.InterfaceRegistry,
		Codec:             cfg.Codec,
		TxConfig:          cfg.TxConfig,
		Amino:             cfg.Amino,
	}
}

func codecAndEncodingConfig() (*codec.ProtoCodec, simappparams.EncodingConfig) {
	cfg := simappparams.MakeTestEncodingConfig()

	// ibc types
	icacontrollertypes.RegisterInterfaces(cfg.InterfaceRegistry)
	icahosttypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	solomachine.RegisterInterfaces(cfg.InterfaceRegistry)
	v7migrations.RegisterInterfaces(cfg.InterfaceRegistry)
	transfertypes.RegisterInterfaces(cfg.InterfaceRegistry)
	clienttypes.RegisterInterfaces(cfg.InterfaceRegistry)
	channeltypes.RegisterInterfaces(cfg.InterfaceRegistry)
	connectiontypes.RegisterInterfaces(cfg.InterfaceRegistry)
	ibctmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	localhost.RegisterInterfaces(cfg.InterfaceRegistry)

	// sdk types
	upgradetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	banktypes.RegisterInterfaces(cfg.InterfaceRegistry)
	govv1beta1.RegisterInterfaces(cfg.InterfaceRegistry)
	govv1.RegisterInterfaces(cfg.InterfaceRegistry)
	authtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	cryptocodec.RegisterInterfaces(cfg.InterfaceRegistry)
	grouptypes.RegisterInterfaces(cfg.InterfaceRegistry)
	proposaltypes.RegisterInterfaces(cfg.InterfaceRegistry)
	authz.RegisterInterfaces(cfg.InterfaceRegistry)
	txtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	stakingtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	minttypes.RegisterInterfaces(cfg.InterfaceRegistry)
	distrtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	slashingtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	consensustypes.RegisterInterfaces(cfg.InterfaceRegistry)

	// custom module types
	wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)

	cdc := codec.NewProtoCodec(cfg.InterfaceRegistry)
	return cdc, cfg
}
