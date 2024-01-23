package callbackcounter

// CallbackCounter is used to represent the callback counter in the contract's storage
type CallbackCounter struct {
	Success uint64 `json:"success"`
	Error   uint64 `json:"error"`
	Timeout uint64 `json:"timeout"`
}
