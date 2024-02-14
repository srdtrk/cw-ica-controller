package types

import (
	"context"
	"encoding/json"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"

	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/icacontroller"
)

type IcaContract struct {
	Contract
	IcaAddress string
}

func NewIcaContract(contract Contract) *IcaContract {
	return &IcaContract{Contract: contract, IcaAddress: ""}
}

func (c *IcaContract) SetIcaAddress(icaAddress string) {
	c.IcaAddress = icaAddress
}

func (c *IcaContract) Execute(ctx context.Context, callerKeyName string, msg icacontroller.ExecuteMsg, extraExecTxArgs ...string) error {
	return c.Contract.ExecAnyMsg(ctx, callerKeyName, toString(msg), extraExecTxArgs...)
}

func (c *IcaContract) Instantiate(ctx context.Context, callerKeyName string, chain *cosmos.CosmosChain, codeId string, msg icacontroller.InstantiateMsg, extraExecTxArgs ...string) error {
	contractAddr, err := chain.InstantiateContract(ctx, callerKeyName, codeId, toString(msg), true, extraExecTxArgs...)
	if err != nil {
		return err
	}

	c.Address = contractAddr
	c.CodeID = codeId
	c.Chain = chain
	return nil
}

// toString converts the message to a string using json
func toString(msg interface{}) string {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return string(bz)
}
