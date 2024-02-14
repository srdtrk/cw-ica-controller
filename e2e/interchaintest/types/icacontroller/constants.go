package icacontroller

var (
	// Query request for contract state
	GetContractStateRequest = QueryMsg{GetContractState: &QueryMsg_GetContractState{}}
	// Query request for channel state
	GetChannelRequest = QueryMsg{GetChannel: &QueryMsg_GetChannel{}}
	// Query request for contract ownership
	OwnershipRequest = QueryMsg{Ownership: &QueryMsg_Ownership{}}
)
