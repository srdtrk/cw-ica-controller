package owner

import (
	"encoding/json"

	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/icacontroller"
)

type InstantiateMsg struct {
	// The admin address. If not specified, the sender is the admin.
	Admin               *string `json:"admin,omitempty"`
	IcaControllerCodeId uint64  `json:"ica_controller_code_id"`
}

type ExecuteMsg struct {
	CreateIcaContract    *ExecuteMsg_CreateIcaContract    `json:"create_ica_contract,omitempty"`
	SendPredefinedAction *ExecuteMsg_SendPredefinedAction `json:"send_predefined_action,omitempty"`
}

type ExecuteMsg_CreateIcaContract struct {
	Salt                   *string                              `json:"salt,omitempty"`
	ChannelOpenInitOptions icacontroller.ChannelOpenInitOptions `json:"channel_open_init_options,omitempty"`
}

type ExecuteMsg_SendPredefinedAction struct {
	IcaId     uint64 `json:"ica_id"`
	ToAddress string `json:"to_address"`
}

type QueryMsg struct {
	// GetContractState returns the contract state
	GetContractState    *struct{}                     `json:"get_contract_state,omitempty"`
	GetIcaContractState *QueryMsg_GetIcaContractState `json:"get_ica_contract_state,omitempty"`
	GetIcaCount         *struct{}                     `json:"get_ica_count,omitempty"`
}

type QueryMsg_GetIcaContractState struct {
	IcaId uint64 `json:"ica_id"`
}

// ToString returns a string representation of the message
func (m *InstantiateMsg) ToString() string {
	return toString(m)
}

// ToString returns a string representation of the message
func (m *ExecuteMsg) ToString() string {
	return toString(m)
}

// ToString returns a string representation of the message
func (m *QueryMsg) ToString() string {
	return toString(m)
}

func toString(v any) string {
	jsonBz, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(jsonBz)
}
