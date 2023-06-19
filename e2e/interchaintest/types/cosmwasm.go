package types

// CwIbcEndpoint is the endpoint of a channel defined in CosmWasm
type CwIbcEndpoint struct {
	PortID    string `json:"port_id"`
	ChannelID string `json:"channel_id"`
}

// CwIbcChannel is the channel defined in CosmWasm
type CwIbcChannel struct {
	Endpoint             CwIbcEndpoint `json:"endpoint"`
	CounterpartyEndpoint CwIbcEndpoint `json:"counterparty_endpoint"`
	// Order is either "ORDER_UNORDERED" or "ORDER_ORDERED"
	Order        string `json:"order"`
	Version      string `json:"version"`
	ConnectionID string `json:"connection_id"`
}
