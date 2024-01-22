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

// QueryContractState queries the contract's state
func (c *IcaContract) QueryContractState(ctx context.Context) (*icacontroller.ContractState, error) {
	queryResp := QueryResponse[icacontroller.ContractState]{}
	queryReq := icacontroller.QueryMsg{
		GetContractState: &icacontroller.EmptyMsg{},
	}
	err := c.Chain.QueryContract(ctx, c.Address, queryReq, &queryResp)
	if err != nil {
		return nil, err
	}

	contractState, err := queryResp.GetResp()
	if err != nil {
		return nil, err
	}

	return &contractState, nil
}

// QueryChannelState queries the channel state stored in the contract
func (c *IcaContract) QueryChannelState(ctx context.Context) (*icacontroller.ContractChannelState, error) {
	queryResp := QueryResponse[icacontroller.ContractChannelState]{}
	queryReq := icacontroller.QueryMsg{
		GetChannel: &icacontroller.EmptyMsg{},
	}
	err := c.Chain.QueryContract(ctx, c.Address, queryReq, &queryResp)
	if err != nil {
		return nil, err
	}

	channelState, err := queryResp.GetResp()
	if err != nil {
		return nil, err
	}

	return &channelState, nil
}

// QueryCallbackCounter queries the callback counter stored in the contract
func (c *IcaContract) QueryCallbackCounter(ctx context.Context) (*icacontroller.CallbackCounter, error) {
	queryResp := QueryResponse[icacontroller.CallbackCounter]{}
	queryReq := icacontroller.QueryMsg{
		GetCallbackCounter: &icacontroller.EmptyMsg{},
	}
	err := c.Chain.QueryContract(ctx, c.Address, queryReq, &queryResp)
	if err != nil {
		return nil, err
	}

	callbackCounter, err := queryResp.GetResp()
	if err != nil {
		return nil, err
	}

	return &callbackCounter, nil
}

// QueryOwnership queries the owner of the contract
func (c *IcaContract) QueryOwnership(ctx context.Context) (*icacontroller.OwnershipQueryResponse, error) {
	queryResp := QueryResponse[icacontroller.OwnershipQueryResponse]{}
	queryReq := icacontroller.QueryMsg{
		Ownership: &icacontroller.EmptyMsg{},
	}
	err := c.Chain.QueryContract(ctx, c.Address, queryReq, &queryResp)
	if err != nil {
		return nil, err
	}

	ownershipResp, err := queryResp.GetResp()
	if err != nil {
		return nil, err
	}

	return &ownershipResp, nil
}
