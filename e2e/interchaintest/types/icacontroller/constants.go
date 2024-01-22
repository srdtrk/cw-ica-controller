package icacontroller

var (
	// Query request for contract state
	GetContractStateRequest = QueryMsg{GetContractState: &struct{}{}}
	// Query request for channel state
	GetChannelRequest = QueryMsg{GetChannel: &struct{}{}}
	// Query request for contract ownership
	OwnershipRequest = QueryMsg{Ownership: &struct{}{}}
)
