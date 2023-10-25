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

// StoreAndInstantiateNewContract stores the contract code and instantiates a new contract as the caller.
// Returns a new Contract instance.
func StoreAndInstantiateNewContract(
	ctx context.Context, chain *cosmos.CosmosChain,
	callerKeyName, fileName string,
) (*Contract, error) {
	codeId, err := chain.StoreContract(ctx, callerKeyName, fileName)
	if err != nil {
		return nil, err
	}

	contractAddr, err := chain.InstantiateContract(ctx, callerKeyName, codeId, newInstantiateMsg(nil), true)
	if err != nil {
		return nil, err
	}

	return &Contract{
		Address: contractAddr,
		CodeID:  codeId,
		chain:   chain,
	}, nil
}

func (c *Contract) Port() string {
	return "wasm." + c.Address
}

func (c *Contract) Execute(ctx context.Context, callerKeyName string, execMsg string) error {
	_, err := c.chain.ExecuteContract(ctx, callerKeyName, c.Address, execMsg)
	return err
}
