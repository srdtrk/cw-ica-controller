package types

import (
	"context"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
)

type Contract struct {
	Address string
	CodeID  string
	Chain   *cosmos.CosmosChain
}

// NewContract creates a new Contract instance
func NewContract(address string, codeId string, chain *cosmos.CosmosChain) Contract {
	return Contract{
		Address: address,
		CodeID:  codeId,
		Chain:   chain,
	}
}

func (c *Contract) Port() string {
	return "wasm." + c.Address
}

func (c *Contract) ExecAnyMsg(ctx context.Context, callerKeyName string, execMsg string, extraExecTxArgs ...string) error {
	_, err := c.Chain.ExecuteContract(ctx, callerKeyName, c.Address, execMsg, extraExecTxArgs...)
	return err
}

// InitAnyMsg instantiates a contract with the given instantiateMsg
func (c *Contract) InitAnyMsg(ctx context.Context, callerKeyName string, instantiateMsg string, extraExecTxArgs ...string) (string, error) {
	return c.Chain.InstantiateContract(ctx, callerKeyName, c.CodeID, instantiateMsg, true, extraExecTxArgs...)
}

func QueryAnyMsg[T any](ctx context.Context, c *Contract, queryMsg any) (*T, error) {
	queryResp := QueryResponse[T]{}
	err := c.Chain.QueryContract(ctx, c.Address, queryMsg, &queryResp)
	if err != nil {
		return nil, err
	}

	resp, err := queryResp.GetResp()
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
