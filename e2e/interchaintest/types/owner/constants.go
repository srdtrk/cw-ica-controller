package owner

var (
	// Query request for contract state
	GetContractStateRequest = QueryMsg{GetContractState: &QueryMsg_GetContractState{}}
	// Query request for the number of ICA contracts
	GetIcaCountRequest = QueryMsg{GetIcaCount: &QueryMsg_GetIcaCount{}}
)
