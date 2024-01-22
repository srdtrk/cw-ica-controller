package types

import (
	"context"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"

	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/icacontroller"
)

type IcaContract struct {
	Contract
	IcaAddress string
}

func NewIcaContract(contract Contract) *IcaContract {
	return &IcaContract{
		Contract:   contract,
		IcaAddress: "",
	}
}

func (c *IcaContract) SetIcaAddress(icaAddress string) {
	c.IcaAddress = icaAddress
}

func (c *IcaContract) Execute(ctx context.Context, callerKeyName string, msg icacontroller.ExecuteMsg, extraExecTxArgs ...string) error {
	return c.Contract.ExecAnyMsg(ctx, callerKeyName, msg.ToString(), extraExecTxArgs...)
}

func (c *IcaContract) Instantiate(ctx context.Context, callerKeyName string, chain *cosmos.CosmosChain, codeId string, msg icacontroller.InstantiateMsg, extraExecTxArgs ...string) error {
	contractAddr, err := chain.InstantiateContract(ctx, callerKeyName, codeId, msg.ToString(), true, extraExecTxArgs...)
	if err != nil {
		return err
	}

	c.Address = contractAddr
	c.CodeID = codeId
	c.Chain = chain
	return nil
}
