package types

import "encoding/json"

// QueryResponse is used to represent the response of a query.
// It may contain different types of data, so we need to unmarshal it
type QueryResponse[T any] struct {
	Response json.RawMessage `json:"data"`
}

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

// GetResp unmarshals the response to a T
func (qr QueryResponse[T]) GetResp() (T, error) {
	var resp T
	err := json.Unmarshal(qr.Response, &resp)
	return resp, err
}
