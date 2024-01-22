package icacontroller

var (
	// Query request for contract state
	GetContractStateRequest = QueryMsg{GetContractState: &EmptyMsg{}}
	// Query request for channel state
	GetChannelRequest = QueryMsg{GetChannel: &EmptyMsg{}}
	// Query request for callback counter
	GetCallbackCounterRequest = QueryMsg{GetCallbackCounter: &EmptyMsg{}}
	// Query request for contract ownership
	OwnershipRequest = QueryMsg{Ownership: &EmptyMsg{}}
)
