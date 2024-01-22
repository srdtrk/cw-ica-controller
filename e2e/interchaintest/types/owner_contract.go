package types

import (
	"context"

	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/owner"
)

type OwnerContract struct {
	Contract
}

func NewOwnerContract(contract Contract) *OwnerContract {
	return &OwnerContract{Contract: contract}
}

func (c *OwnerContract) Execute(ctx context.Context, callerKeyName string, msg owner.ExecuteMsg, extraExecTxArgs ...string) error {
	return c.Contract.ExecAnyMsg(ctx, callerKeyName, msg.ToString(), extraExecTxArgs...)
}
