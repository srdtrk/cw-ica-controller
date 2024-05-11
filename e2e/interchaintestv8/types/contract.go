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
// QC is the query client interface
type Contract[I, E, Q any, QC comparable] struct {
	Address string
	CodeID  string
	Chain   *cosmos.CosmosChain

	qc QC
}

// newContract creates a new Contract instance
func newContract[I, E, Q any, QC comparable](address string, codeId string, chain *cosmos.CosmosChain) *Contract[I, E, Q, QC] {
	return &Contract[I, E, Q, QC]{
		Address: address,
		CodeID:  codeId,
		Chain:   chain,
	}
}

// Instantiate creates a new contract instance on the chain. No admin flag is set if admin is empty.
// I is the instantiate message type
// E is the execute message type
// Q is the query message type
// QC is the query client interface
func Instantiate[I, E, Q any, QC comparable](ctx context.Context, callerKeyName, codeId, admin string, chain *cosmos.CosmosChain, msg I, extraExecTxArgs ...string) (*Contract[I, E, Q, QC], error) {
	isNoAdmin := admin == ""
	if !isNoAdmin {
		extraExecTxArgs = append(extraExecTxArgs, "--admin", admin)
	}

	contractAddr, err := chain.InstantiateContract(ctx, callerKeyName, codeId, toString(msg), isNoAdmin, extraExecTxArgs...)
	if err != nil {
		return nil, err
	}

	return newContract[I, E, Q, QC](contractAddr, codeId, chain), nil
}

// Port returns the port of the contract
func (c *Contract[I, E, Q, QC]) Port() string {
	return "wasm." + c.Address
}

// Execute executes the contract with the given execute message and returns the transaction response
func (c *Contract[I, E, Q, QC]) Execute(ctx context.Context, callerKeyName string, msg E, extraExecTxArgs ...string) (*sdk.TxResponse, error) {
	return c.Chain.ExecuteContract(ctx, callerKeyName, c.Address, toString(msg), extraExecTxArgs...)
}

// QueryClient returns the query client of the contract.
// If the query client is not provided, it panics.
func (c *Contract[I, E, Q, QC]) QueryClient() QC {
	var zeroQc QC
	if c.qc == zeroQc {
		panic("Query client is not provided")
	}

	return c.qc
}

// SetQueryClient sets the query client of the contract.
func (c *Contract[I, E, Q, QC]) SetQueryClient(qc QC) {
	c.qc = qc
}

// Query queries the contract with the given query message
// and unmarshals the response into the given response object.
// This is meant to be used if the query client is not provided.
func (c *Contract[I, E, Q, QC]) Query(ctx context.Context, queryMsg Q, resp any) error {
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

// this line is used by go-codegen # contract/dir

// toString converts the message to a string using json
func toString(msg any) string {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return string(bz)
}
