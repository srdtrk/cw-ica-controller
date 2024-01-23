package callbackcounter

// QueryMsg is the message to query cw-ica-controller
type QueryMsg struct {
	GetCallbackCounter *struct{} `json:"get_callback_counter,omitempty"`
}
