package cwicacontroller

import (
	"context"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"

	"github.com/srdtrk/go-codegen/e2esuite/v8/types"
)

// Contract represents a cw-ica-controller contract on a Chain.
type Contract = types.Contract[InstantiateMsg, ExecuteMsg, QueryMsg, QueryClient]

// Instantiate creates a cw-ica-controller new contract instance on the chain.
func Instantiate(ctx context.Context, callerKeyName, codeId, admin string, chain *cosmos.CosmosChain, msg InstantiateMsg, extraExecTxArgs ...string) (*Contract, error) {
	contract, err := types.Instantiate[InstantiateMsg, ExecuteMsg, QueryMsg, QueryClient](ctx, callerKeyName, codeId, admin, chain, msg, extraExecTxArgs...)
	if err != nil {
		return nil, err
	}

	queryClient, err := NewQueryClient(chain.GetHostGRPCAddress(), contract.Address)
	if err != nil {
		return nil, err
	}
	contract.SetQueryClient(queryClient)

	return contract, nil
}
