package types

// OwnerContractState is used to represent its state in Contract's storage
type OwnerContractState struct {
	Admin               string `json:"admin"`
	IcaControllerCodeID uint64 `json:"ica_controller_code_id"`
}

// OwnerIcaContractState is used to represent its state in Contract's storage
type OwnerIcaContractState struct {
	ContractAddr string         `json:"contract_addr"`
	IcaState     *OwnerIcaState `json:"ica_state"`
}

// OwnerIcaState is the state of the ICA.
type OwnerIcaState struct {
	IcaID        uint32                   `json:"ica_id"`
	IcaAddr      string                   `json:"ica_addr"`
	TxEncoding   string                   `json:"tx_encoding"`
	ChannelState *IcaContractChannelState `json:"channel_state"`
}
