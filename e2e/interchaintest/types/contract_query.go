package types

type QueryMsg struct {
	GetChannel EmptyObject `json:"get_channel"`
	GetContractState EmptyObject `json:"get_contract_state"`
	GetCallbackCounter EmptyObject `json:"get_callback_counter"`
}

type EmptyObject struct {}

// NewGetChannelQueryMsg creates a new GetChannel QueryMsg.
func NewGetChannelQueryMsg() QueryMsg {
	return QueryMsg{
		GetChannel: EmptyObject{},
	}
}

// NewGetContractStateQueryMsg creates a new GetContractState QueryMsg.
func NewGetContractStateQueryMsg() QueryMsg {
	return QueryMsg{
		GetContractState: EmptyObject{},
	}
}

// NewGetCallbackCounterQueryMsg creates a new GetCallbackCounter QueryMsg.
func NewGetCallbackCounterQueryMsg() QueryMsg {
	return QueryMsg{
		GetCallbackCounter: EmptyObject{},
	}
}
