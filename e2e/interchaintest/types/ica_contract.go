package types

import (
	"context"

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

func (c *IcaContract) Instantiate(ctx context.Context, callerKeyName string, codeId string, msg icacontroller.InstantiateMsg, extraExecTxArgs ...string) error {
	c.CodeID = codeId

	contractAddr, err := c.Contract.InitAnyMsg(ctx, callerKeyName, msg.ToString(), extraExecTxArgs...)
	if err != nil {
		return err
	}

	c.Address = contractAddr
	return nil
}
