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
		Chain:   chain,
	}

	return NewOwnerContract(contract), nil
}

func (c *OwnerContract) ExecSendPredefinedAction(
	ctx context.Context, callerKeyName string, icaId uint64, toAddress string,
) error {
	msg := newOwnerSendPredefinedActionMsg(icaId, toAddress)
	err := c.ExecAnyMsg(ctx, callerKeyName, msg)
	return err
}

// QueryContractState queries the contract's state
func (c *OwnerContract) QueryContractState(ctx context.Context) (*OwnerContractState, error) {
	queryResp := QueryResponse[OwnerContractState]{}
	err := c.Chain.QueryContract(ctx, c.Address, newOwnerGetContractStateQueryMsg(), &queryResp)
	if err != nil {
		return nil, err
	}

	contractState, err := queryResp.GetResp()
	if err != nil {
		return nil, err
	}

	return &contractState, nil
}

// QueryIcaContractState queries the contract's ica state for a given icaID
func (c *OwnerContract) QueryIcaContractState(ctx context.Context, icaID uint64) (*OwnerIcaContractState, error) {
	queryResp := QueryResponse[OwnerIcaContractState]{}
	err := c.Chain.QueryContract(ctx, c.Address, newOwnerGetIcaContractStateQueryMsg(icaID), &queryResp)
	if err != nil {
		return nil, err
	}

	icaContractState, err := queryResp.GetResp()
	if err != nil {
		return nil, err
	}

	return &icaContractState, nil
}
