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

// NewContract creates a new Contract instance
func NewContract(address string, codeId string, chain *cosmos.CosmosChain) Contract {
	return Contract{
		Address: address,
		CodeID:  codeId,
		chain:   chain,
	}
}

func (c *Contract) Port() string {
	return "wasm." + c.Address
}

func (c *Contract) ExecAnyMsg(ctx context.Context, callerKeyName string, execMsg string, extraExecTxArgs ...string) error {
	_, err := c.chain.ExecuteContract(ctx, callerKeyName, c.Address, execMsg, extraExecTxArgs...)
	return err
}

func QueryContract[T any](ctx context.Context, chain *cosmos.CosmosChain, contractAddr string, queryMsg string) (*T, error) {
	queryResp := QueryResponse[T]{}
	err := chain.QueryContract(ctx, contractAddr, queryMsg, &queryResp)
	if err != nil {
		return nil, err
	}

	resp, err := queryResp.GetResp()
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
