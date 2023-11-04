package types

import (
	"encoding/json"
	"fmt"
)

// newOwnerInstantiateMsg creates a new InstantiateMsg.
func newOwnerInstantiateMsg(admin *string, icaControllerCodeId uint64) string {
	if admin == nil {
		return fmt.Sprintf(`{"ica_controller_code_id":%d}`, icaControllerCodeId)
	} else {
		return fmt.Sprintf(`{"admin":"%s","ica_controller_code_id":%d}`, *admin, icaControllerCodeId)
	}
}

// NewOwnerCreateIcaContractMsg creates a new CreateIcaContractMsg.
func NewOwnerCreateIcaContractMsg(salt *string, coip *ChannelOpenInitOptions) string {
	type CreateIcaContractMsg struct {
		Salt *string `json:"salt,omitempty"`
		// The options to initialize the IBC channel upon contract instantiation.
		// If not specified, the IBC channel is not initialized, and the relayer must.
		ChannelOpenInitOptions *ChannelOpenInitOptions `json:"channel_open_init_options,omitempty"`
	}

	type CreateIcaContractMsgWrapper struct {
		CreateIcaContractMsg CreateIcaContractMsg `json:"create_ica_contract"`
	}

	createIcaContractMsg := CreateIcaContractMsgWrapper{
		CreateIcaContractMsg: CreateIcaContractMsg{
			Salt:                   salt,
			ChannelOpenInitOptions: coip,
		},
	}

	jsonBytes, err := json.Marshal(createIcaContractMsg)
	if err != nil {
		panic(err)
	}

	return string(jsonBytes)
}
