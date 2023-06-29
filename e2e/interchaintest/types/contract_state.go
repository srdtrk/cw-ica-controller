package types

// ContractState is used to represent its state in Contract's storage
type ContractState struct {
	Admin   string          `json:"admin"`
	IcaInfo ContractIcaInfo `json:"ica_info"`
}

// ContractIcaInfo is used to represent the ICA info in the contract's state
type ContractIcaInfo struct {
	IcaAddress string `json:"ica_address"`
	ChannelID  string `json:"channel_id"`
}

// ContractCallbackCounter is used to represent the callback counter in the contract's storage
type ContractCallbackCounter struct {
	Success uint64 `json:"success"`
	Error   uint64 `json:"error"`
	Timeout uint64 `json:"timeout"`
}

// ContractChannelState is used to represent the channel state in the contract's storage
type ContractChannelState struct {
	Channel       CwIbcChannel `json:"channel"`
	ChannelStatus string       `json:"channel_status"`
}

// IsOpen returns true if the channel is open
func (c *ContractChannelState) IsOpen() bool {
	return c.ChannelStatus == "STATE_OPEN"
}
