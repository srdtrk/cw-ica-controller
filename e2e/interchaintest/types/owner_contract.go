package types

import (
	"context"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
)

type OwnerContract struct {
	Contract
}

func NewOwnerContract(contract Contract) *OwnerContract {
	return &OwnerContract{
		Contract: contract,
	}
}

// StoreAndInstantiateNewOwnerContract stores the contract code and instantiates a new contract as the caller.
// Returns a new OwnerContract instance.
func StoreAndInstantiateNewOwnerContract(
	ctx context.Context, chain *cosmos.CosmosChain,
	callerKeyName, fileName string, icaCodeId uint64,
) (*OwnerContract, error) {
	codeId, err := chain.StoreContract(ctx, callerKeyName, fileName)
	if err != nil {
		return nil, err
	}

	contractAddr, err := chain.InstantiateContract(ctx, callerKeyName, codeId, newOwnerInstantiateMsg(nil, icaCodeId), true)
	if err != nil {
		return nil, err
	}

	contract := Contract{
		Address: contractAddr,
		CodeID:  codeId,
		chain:   chain,
	}

	return NewOwnerContract(contract), nil
}
