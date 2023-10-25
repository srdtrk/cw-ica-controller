package types

import "encoding/json"

// newGetChannelQueryMsg creates a new GetChannelQueryMsg.
// This function returns a map[string]interface{} instead of []byte
// because interchaintest uses json.Marshal to convert the map to a string
func newGetChannelQueryMsg() map[string]interface{} {
	return map[string]interface{}{
		"get_channel": struct{}{},
	}
}

// newGetContractStateQueryMsg creates a new GetContractStateQueryMsg.
// This function returns a map[string]interface{} instead of []byte
// because interchaintest uses json.Marshal to convert the map to a string
func newGetContractStateQueryMsg() map[string]interface{} {
	return map[string]interface{}{
		"get_contract_state": struct{}{},
	}
}

// newGetCallbackCounterQueryMsg creates a new GetCallbackCounterQueryMsg.
// This function returns a map[string]interface{} instead of []byte
// because interchaintest uses json.Marshal to convert the map to a string
func newGetCallbackCounterQueryMsg() map[string]interface{} {
	return map[string]interface{}{
		"get_callback_counter": struct{}{},
	}
}

// getChannelState unmarshals the response to a ContractChannelState
func (qr QueryResponse) getChannelState() (IcaContractChannelState, error) {
	var channelState IcaContractChannelState
	err := json.Unmarshal(qr.Response, &channelState)
	return channelState, err
}

// getContractState unmarshals the response to a ContractState
func (qr QueryResponse) getContractState() (IcaContractState, error) {
	var contractState IcaContractState
	err := json.Unmarshal(qr.Response, &contractState)
	return contractState, err
}

// getCallbackCounter unmarshals the response to a ContractCallbackCounter
func (qr QueryResponse) getCallbackCounter() (IcaContractCallbackCounter, error) {
	var callbackCounter IcaContractCallbackCounter
	err := json.Unmarshal(qr.Response, &callbackCounter)
	return callbackCounter, err
}
