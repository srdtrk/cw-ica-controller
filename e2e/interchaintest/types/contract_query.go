package types

import "encoding/json"

// NewGetChannelQueryMsg creates a new GetChannelQueryMsg.
// This function returns a map[string]interface{} instead of []byte
// because interchaintest uses json.Marshal to convert the map to a string
func NewGetChannelQueryMsg() map[string]interface{} {
	return map[string]interface{}{
		"get_channel": struct{}{},
	}
}

// NewGetContractStateQueryMsg creates a new GetContractStateQueryMsg.
// This function returns a map[string]interface{} instead of []byte
// because interchaintest uses json.Marshal to convert the map to a string
func NewGetContractStateQueryMsg() map[string]interface{} {
	return map[string]interface{}{
		"get_contract_state": struct{}{},
	}
}

// NewGetCallbackCounterQueryMsg creates a new GetCallbackCounterQueryMsg.
// This function returns a map[string]interface{} instead of []byte
// because interchaintest uses json.Marshal to convert the map to a string
func NewGetCallbackCounterQueryMsg() map[string]interface{} {
	return map[string]interface{}{
		"get_callback_counter": struct{}{},
	}
}

// QueryResponse is used to represent the response of a query.
// It may contain different types of data, so we need to unmarshal it
type QueryResponse struct {
	Response json.RawMessage `json:"data"`
}

// GetChannelState unmarshals the response to a ContractChannelState
func (qr QueryResponse) GetChannelState() (ContractChannelState, error) {
	var channelState ContractChannelState
	err := json.Unmarshal(qr.Response, &channelState)
	return channelState, err
}

// GetContractState unmarshals the response to a ContractState
func (qr QueryResponse) GetContractState() (ContractState, error) {
	var contractState ContractState
	err := json.Unmarshal(qr.Response, &contractState)
	return contractState, err
}

// GetCallbackCounter unmarshals the response to a ContractCallbackCounter
func (qr QueryResponse) GetCallbackCounter() (ContractCallbackCounter, error) {
	var callbackCounter ContractCallbackCounter
	err := json.Unmarshal(qr.Response, &callbackCounter)
	return callbackCounter, err
}
