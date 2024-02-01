package types

import (
	"context"
	"encoding/json"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

type Contract struct {
	Address string
	CodeID  string
	Chain   *cosmos.CosmosChain
}

// NewContract creates a new Contract instance
func NewContract(address string, codeId string, chain *cosmos.CosmosChain) *Contract {
	return &Contract{
		Address: address,
		CodeID:  codeId,
		Chain:   chain,
	}
}

func (c *Contract) Port() string {
	return "wasm." + c.Address
}

// ExecAnyMsg executes the contract with the given exec message.
func (c *Contract) ExecAnyMsg(ctx context.Context, callerKeyName string, execMsg string, extraExecTxArgs ...string) error {
	_, err := c.Chain.ExecuteContract(ctx, callerKeyName, c.Address, execMsg, extraExecTxArgs...)
	return err
}

// QueryAnyMsg queries the contract with the given query message and returns the response.
func QueryAnyMsg[T any](ctx context.Context, c *Contract, queryMsg any) (*T, error) {
	// QueryResponse is used to represent the response of a query.
	// It may contain different types of data, so we need to unmarshal it
	type QueryResponse struct {
		Response json.RawMessage `json:"data"`
	}

	queryResp := QueryResponse{}
	err := c.Chain.QueryContract(ctx, c.Address, queryMsg, &queryResp)
	if err != nil {
		return nil, err
	}

	var resp T
	err = json.Unmarshal(queryResp.Response, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
