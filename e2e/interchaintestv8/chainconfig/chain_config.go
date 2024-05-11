package chainconfig

import (
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

var DefaultChainSpecs = []*interchaintest.ChainSpec{
	// -- WASMD --
	{
		ChainConfig: ibc.ChainConfig{
			Type:    "cosmos",
			Name:    "wasmd-1",
			ChainID: "wasmd-1",
			Images: []ibc.DockerImage{
				{
					Repository: "cosmwasm/wasmd", // FOR LOCAL IMAGE USE: Docker Image Name
					Version:    "v0.50.0",        // FOR LOCAL IMAGE USE: Docker Image Tag
					UidGid:     "1025:1025",
				},
			},
			Bin:            "wasmd",
			Bech32Prefix:   "wasm",
			Denom:          "stake",
			GasPrices:      "0.00stake",
			GasAdjustment:  1.3,
			EncodingConfig: WasmEncodingConfig(),
			ModifyGenesis:  defaultModifyGenesis(),
			TrustingPeriod: "508h",
			NoHostMount:    false,
		},
	},
	// -- WASMD --
	{
		ChainConfig: ibc.ChainConfig{
			Type:    "cosmos",
			Name:    "wasmd-2",
			ChainID: "wasmd-2",
			Images: []ibc.DockerImage{
				{
					Repository: "cosmwasm/wasmd", // FOR LOCAL IMAGE USE: Docker Image Name
					Version:    "v0.50.0",        // FOR LOCAL IMAGE USE: Docker Image Tag
					UidGid:     "1025:1025",
				},
			},
			Bin:            "wasmd",
			Bech32Prefix:   "wasm",
			Denom:          "stake",
			GasPrices:      "0.00stake",
			GasAdjustment:  1.3,
			EncodingConfig: WasmEncodingConfig(),
			ModifyGenesis:  defaultModifyGenesis(),
			TrustingPeriod: "508h",
			NoHostMount:    false,
		},
	},
}
