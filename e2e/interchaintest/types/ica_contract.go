package types

import (
	"context"

	"github.com/cosmos/gogoproto/proto"

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

// StoreAndInstantiateNewIcaContract stores the contract code and instantiates a new contract as the caller.
// Returns a new Contract instance.
func StoreAndInstantiateNewIcaContract(
	ctx context.Context, chain *cosmos.CosmosChain,
	callerKeyName, fileName string,
) (*IcaContract, error) {
	codeId, err := chain.StoreContract(ctx, callerKeyName, fileName)
	if err != nil {
		return nil, err
	}

	contractAddr, err := chain.InstantiateContract(ctx, callerKeyName, codeId, "{}", true)
	if err != nil {
		return nil, err
	}

	contract := Contract{
		Address: contractAddr,
		CodeID:  codeId,
		Chain:   chain,
	}

	return NewIcaContract(contract), nil
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

// ExecCustomMessages invokes the contract's `CustomIcaMessages` message as the caller
func (c *IcaContract) ExecCustomIcaMessages(
	ctx context.Context, callerKeyName string,
	messages []proto.Message, encoding string,
	memo *string, timeout *uint64,
) error {
	customMsg := newSendCustomIcaMessagesMsg(c.Chain.Config().EncodingConfig.Codec, messages, encoding, memo, timeout)
	err := c.ExecAnyMsg(ctx, callerKeyName, customMsg)
	return err
}

// ExecSendCosmosMsgs invokes the contract's `SendCosmosMsgsAsIcaTx` message as the caller
// This version takes a slice of proto messages instead of ContractCosmosMsgs to make it easier to use
// the Stargate Cosmos Message type
func (c *IcaContract) ExecSendStargateMsgs(
	ctx context.Context, callerKeyName string,
	msgs []proto.Message, memo *string, timeout *uint64,
) error {
	cosmosMsg := newSendCosmosMsgsMsgFromProto(msgs, memo, timeout)
	err := c.ExecAnyMsg(ctx, callerKeyName, cosmosMsg)
	return err
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
