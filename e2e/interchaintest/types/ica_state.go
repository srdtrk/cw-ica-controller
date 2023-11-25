package types

// IcaContractState is used to represent its state in Contract's storage
type IcaContractState struct {
	IcaInfo              IcaContractIcaInfo `json:"ica_info"`
	AllowChannelOpenInit bool               `json:"allow_channel_open_init"`
}

// IcaContractIcaInfo is used to represent the ICA info in the contract's state
type IcaContractIcaInfo struct {
	IcaAddress string `json:"ica_address"`
	ChannelID  string `json:"channel_id"`
}

// ContractCallbackCounter is used to represent the callback counter in the contract's storage
type IcaContractCallbackCounter struct {
	Success uint64 `json:"success"`
	Error   uint64 `json:"error"`
	Timeout uint64 `json:"timeout"`
}

// ContractChannelState is used to represent the channel state in the contract's storage
type IcaContractChannelState struct {
	Channel       CwIbcChannel `json:"channel"`
	ChannelStatus string       `json:"channel_status"`
}

// IsOpen returns true if the channel is open
func (c *IcaContractChannelState) IsOpen() bool {
	return c.ChannelStatus == "STATE_OPEN"
}
