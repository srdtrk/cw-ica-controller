package types

import (
	"context"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

// Contract represents a smart contract on a Chain.
// I is the instantiate message type
// E is the execute message type
// Q is the query message type
type Contract[I, E, Q any] struct {
	Address string
	CodeID  string
	Chain   *cosmos.CosmosChain
}

// newContract creates a new Contract instance
func newContract[I, E, Q any](address string, codeId string, chain *cosmos.CosmosChain) *Contract[I, E, Q] {
	return &Contract[I, E, Q]{
		Address: address,
		CodeID:  codeId,
		Chain:   chain,
	}
}

// Instantiate creates a new contract instance on the chain
// I is the instantiate message type
// E is the execute message type
// Q is the query message type
func Instantiate[I, E, Q any](ctx context.Context, callerKeyName string, codeId string, chain *cosmos.CosmosChain, msg I, extraExecTxArgs ...string) (*Contract[I, E, Q], error) {
	contractAddr, err := chain.InstantiateContract(ctx, callerKeyName, codeId, toString(msg), true, extraExecTxArgs...)
	if err != nil {
		return nil, err
	}

	return newContract[I, E, Q](contractAddr, codeId, chain), nil
}

func (c *Contract[I, E, Q]) Port() string {
	return "wasm." + c.Address
}

// Execute executes the contract with the given execute message and returns the transaction response
func (c *Contract[I, E, Q]) Execute(ctx context.Context, callerKeyName string, msg E, extraExecTxArgs ...string) (*sdk.TxResponse, error) {
	return c.Chain.ExecuteContract(ctx, callerKeyName, c.Address, toString(msg), extraExecTxArgs...)
}

// Query queries the contract with the given query message
// and unmarshals the response into the given response object
func (c *Contract[I, E, Q]) Query(ctx context.Context, queryMsg Q, resp any) error {
	// queryResponse is used to represent the response of a query.
	// It may contain different types of data, so we need to unmarshal it
	type queryResponse struct {
		Response json.RawMessage `json:"data"`
	}

	queryResp := queryResponse{}
	err := c.Chain.QueryContract(ctx, c.Address, queryMsg, &queryResp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(queryResp.Response, resp)
	if err != nil {
		return err
	}

	return nil
}

// toString converts the message to a string using json
func toString(msg any) string {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return string(bz)
}
