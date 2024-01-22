package types

import "encoding/json"

// QueryResponse is used to represent the response of a query.
// It may contain different types of data, so we need to unmarshal it
type QueryResponse[T any] struct {
	Response json.RawMessage `json:"data"`
}

// GetResp unmarshals the response to a T
func (qr QueryResponse[T]) GetResp() (T, error) {
	var resp T
	err := json.Unmarshal(qr.Response, &resp)
	return resp, err
}
