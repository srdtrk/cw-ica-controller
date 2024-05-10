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
// QC is the query client type
type Contract[I, E, Q, QC any] struct {
	Address string
	CodeID  string
	Chain   *cosmos.CosmosChain
	qc      QC
	qcSet   bool
}

// newContract creates a new Contract instance
func newContract[I, E, Q, QC any](address, codeId string, chain *cosmos.CosmosChain) *Contract[I, E, Q, QC] {
	return &Contract[I, E, Q, QC]{
		Address: address,
		CodeID:  codeId,
		Chain:   chain,
	}
}

// Instantiate creates a new contract instance on the chain
// I is the instantiate message type
// E is the execute message type
// Q is the query message type
func Instantiate[I, E, Q, QC any](ctx context.Context, callerKeyName, codeId string, chain *cosmos.CosmosChain, msg I, extraExecTxArgs ...string) (*Contract[I, E, Q, QC], error) {
	contractAddr, err := chain.InstantiateContract(ctx, callerKeyName, codeId, toString(msg), true, extraExecTxArgs...)
	if err != nil {
		return nil, err
	}

	return newContract[I, E, Q, QC](contractAddr, codeId, chain), nil
}

func (c *Contract[I, E, Q, QC]) Port() string {
	return "wasm." + c.Address
}

// QueryClient returns the query client for the contract
func (c *Contract[I, E, Q, QC]) QueryClient() QC {
	if !c.qcSet {
		panic("QueryClient not set")
	}

	return c.qc
}

// SetQueryClient sets the query client for the contract
func (c *Contract[I, E, Q, QC]) SetQueryClient(qc QC) {
	if c.qcSet {
		panic("QueryClient already set")
	}

	c.qc = qc
	c.qcSet = true
}

// Execute executes the contract with the given execute message and returns the transaction response
func (c *Contract[I, E, Q, QC]) Execute(ctx context.Context, callerKeyName string, msg E, extraExecTxArgs ...string) (*sdk.TxResponse, error) {
	return c.Chain.ExecuteContract(ctx, callerKeyName, c.Address, toString(msg), extraExecTxArgs...)
}

// toString converts the message to a string using json
func toString(msg any) string {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return string(bz)
}
