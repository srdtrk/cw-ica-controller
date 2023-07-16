package types

import (
	"context"

	"github.com/cosmos/gogoproto/proto"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
)

type Contract struct {
	Address string
	CodeID  string
	chain   *cosmos.CosmosChain
}

// StoreAndInstantiateNewContract stores the contract code and instantiates a new contract as the caller.
// Returns a new Contract instance.
func StoreAndInstantiateNewContract(
	ctx context.Context, chain *cosmos.CosmosChain,
	callerKeyName, fileName string,
) (*Contract, error) {
	codeId, err := chain.StoreContract(ctx, callerKeyName, fileName)
	if err != nil {
		return nil, err
	}

	contractAddr, err := chain.InstantiateContract(ctx, callerKeyName, codeId, NewInstantiateMsg(nil), true)
	if err != nil {
		return nil, err
	}

	return &Contract{
		Address: contractAddr,
		CodeID:  codeId,
		chain:   chain,
	}, nil
}

func (c *Contract) Port() string {
	return "wasm." + c.Address
}

// ExecCustomMessages invokes the contract's `CustomIcaMessages` message as the caller
func (c *Contract) ExecCustomIcaMessages(
	ctx context.Context, callerKeyName string,
	messages []proto.Message, encoding string,
	memo *string, timeout *uint64,
) error {
	customMsg := NewSendCustomIcaMessagesMsg(c.chain.Config().EncodingConfig.Codec, messages, encoding, memo, timeout)
	err := c.Execute(ctx, callerKeyName, customMsg)
	return err
}

// ExecPredefinedAction executes the contract's predefined action message as the caller
func (c *Contract) ExecPredefinedAction(ctx context.Context, callerKeyName string, toAddress string) error {
	err := c.Execute(ctx, callerKeyName, NewSendPredefinedActionMsg(toAddress))
	return err
}

func (c *Contract) Execute(ctx context.Context, callerKeyName string, execMsg string) error {
	_, err := c.chain.ExecuteContract(ctx, callerKeyName, c.Address, execMsg)
	return err
}

// QueryContractState queries the contract's state
func (c *Contract) QueryContractState(ctx context.Context) (*ContractState, error) {
	queryResp := QueryResponse{}
	err := c.chain.QueryContract(ctx, c.Address, NewGetContractStateQueryMsg(), &queryResp)
	if err != nil {
		return nil, err
	}

	contractState, err := queryResp.GetContractState()
	if err != nil {
		return nil, err
	}

	return &contractState, nil
}

// QueryChannelState queries the channel state stored in the contract
func (c *Contract) QueryChannelState(ctx context.Context) (*ContractChannelState, error) {
	queryResp := QueryResponse{}
	err := c.chain.QueryContract(ctx, c.Address, NewGetChannelQueryMsg(), &queryResp)
	if err != nil {
		return nil, err
	}

	channelState, err := queryResp.GetChannelState()
	if err != nil {
		return nil, err
	}

	return &channelState, nil
}

// QueryCallbackCounter queries the callback counter stored in the contract
func (c *Contract) QueryCallbackCounter(ctx context.Context) (*ContractCallbackCounter, error) {
	queryResp := QueryResponse{}
	err := c.chain.QueryContract(ctx, c.Address, NewGetCallbackCounterQueryMsg(), &queryResp)
	if err != nil {
		return nil, err
	}

	callbackCounter, err := queryResp.GetCallbackCounter()
	if err != nil {
		return nil, err
	}

	return &callbackCounter, nil
}
