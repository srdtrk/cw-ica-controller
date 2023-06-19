//go:build tools

// This pattern of a file named tools.go, to import dependent tools,
// comes from the official Go wiki:
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

package main

import (
	_ "github.com/CosmWasm/wasmd/x/wasm/types"
	_ "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/ibc-go/v7/modules/core"
	_ "github.com/strangelove-ventures/interchaintest/v7/testutil"
)