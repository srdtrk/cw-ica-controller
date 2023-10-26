package types

import (
	"context"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
)

type Contract struct {
	Address string
	CodeID  string
	chain   *cosmos.CosmosChain
}

func (c *Contract) Port() string {
	return "wasm." + c.Address
}

func (c *Contract) Execute(ctx context.Context, callerKeyName string, execMsg string) error {
	_, err := c.chain.ExecuteContract(ctx, callerKeyName, c.Address, execMsg)
	return err
}
