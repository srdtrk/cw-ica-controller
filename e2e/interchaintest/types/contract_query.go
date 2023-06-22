package types

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
