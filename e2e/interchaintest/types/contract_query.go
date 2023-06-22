package types

type QueryMsg struct {
	GetChannel EmptyObject `json:"get_channel"`
	GetContractState EmptyObject `json:"get_contract_state"`
	GetCallbackCounter EmptyObject `json:"get_callback_counter"`
}

type EmptyObject struct {}

