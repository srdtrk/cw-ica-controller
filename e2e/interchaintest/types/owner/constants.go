package owner

var (
	// Query request for contract state
	GetContractStateRequest = QueryMsg{GetContractState: &struct{}{}}
	// Query request for the number of ICA contracts
	GetIcaCountRequest = QueryMsg{GetIcaCount: &struct{}{}}
)
