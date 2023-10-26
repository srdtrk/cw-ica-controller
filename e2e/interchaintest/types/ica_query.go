package types

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
