package types

// newOwnerGetContractStateQueryMsg creates a new GetContractStateQueryMsg.
// This function returns a map[string]interface{} instead of []byte
// because interchaintest uses json.Marshal to convert the map to a string
func newOwnerGetContractStateQueryMsg() map[string]interface{} {
	return map[string]interface{}{
		"get_contract_state": struct{}{},
	}
}

// newOwnerGetIcaContractStateQueryMsg creates a new GetIcaContractStateQueryMsg.
// This function returns a map[string]interface{} instead of []byte
// because interchaintest uses json.Marshal to convert the map to a string
func newOwnerGetIcaContractStateQueryMsg(icaID uint64) map[string]interface{} {
	return map[string]interface{}{
		"get_ica_contract_state": map[string]interface{}{
			"ica_id": icaID,
		},
	}
}

// newOwnerGetIcaCountQueryMsg creates a new GetIcaCountQueryMsg.
// This function returns a map[string]interface{} instead of []byte
// because interchaintest uses json.Marshal to convert the map to a string
func newOwnerGetIcaCountQueryMsg() map[string]interface{} {
	return map[string]interface{}{
		"get_ica_count": struct{}{},
	}
}
