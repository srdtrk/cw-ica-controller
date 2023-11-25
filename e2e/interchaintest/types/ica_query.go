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

// newOwnershipQueryMsg creates a new OwnershipQueryMsg.
// This function returns a map[string]interface{} instead of []byte
// because interchaintest uses json.Marshal to convert the map to a string
func newOwnershipQueryMsg() map[string]interface{} {
	return map[string]interface{}{
		"ownership": struct{}{},
	}
}

// OwnershipQueryResponse is the response type for the OwnershipQueryMsg
type OwnershipQueryResponse struct {
	// The current owner of the contract.
	// This contract must have an owner.
	Owner string `json:"owner"`
	// The pending owner of the contract if one exists.
	PendingOwner *string `json:"pending_owner"`
	// The height at which the pending owner offer expires.
	// Not sure how to represent this, so we'll just use a raw message
	PendingExpiry *json.RawMessage `json:"pending_expiry"`
}
