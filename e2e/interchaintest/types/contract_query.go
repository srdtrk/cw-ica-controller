package types

import (
	"encoding/json"
)

// NewGetChannelQueryMsg creates a new GetChannelQueryMsg.
func NewGetChannelQueryMsg() []byte {
	return []byte(`{"get_channel":{}}`)
}

// NewGetContractStateQueryMsg creates a new GetContractStateQueryMsg.
func NewGetContractStateQueryMsg() []byte {
	return []byte(`{"get_contract_state":{}}`)
}

// NewGetCallbackCounterQueryMsg creates a new GetCallbackCounterQueryMsg.
func NewGetCallbackCounterQueryMsg() []byte {
	return []byte(`{"get_callback_counter":{}}`)
}

// UnmarshalGetChannelResp unmarshals the response to a GetChannelQueryMsg.
func UnmarshalGetChannelResp(bz []byte) (ContractChannelState, error) {
	var msg ContractChannelState
	err := json.Unmarshal(bz, &msg)
	return msg, err
}

// UnmarshalGetContractStateResp unmarshals the response to a GetContractStateQueryMsg.
func UnmarshalGetContractStateResp(bz []byte) (ContractState, error) {
	var msg ContractState
	err := json.Unmarshal(bz, &msg)
	return msg, err
}

// UnmarshalGetCallbackCounterResp unmarshals the response to a GetCallbackCounterQueryMsg.
func UnmarshalGetCallbackCounterResp(bz []byte) (ContractCallbackCounter, error) {
	var msg ContractCallbackCounter
	err := json.Unmarshal(bz, &msg)
	return msg, err
}