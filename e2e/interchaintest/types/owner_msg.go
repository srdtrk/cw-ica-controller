package types

import (
	"encoding/json"
	"fmt"

	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/icacontroller"
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
func NewOwnerCreateIcaContractMsg(salt *string, coip *icacontroller.ChannelOpenInitOptions) string {
	type CreateIcaContractMsg struct {
		Salt *string `json:"salt,omitempty"`
		// The options to initialize the IBC channel upon contract instantiation.
		// If not specified, the IBC channel is not initialized, and the relayer must.
		ChannelOpenInitOptions *icacontroller.ChannelOpenInitOptions `json:"channel_open_init_options,omitempty"`
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

// newOwnerSendPredefinedActionMsg creates a new SendPredefinedActionMsg.
func newOwnerSendPredefinedActionMsg(icaId uint64, toAddress string) string {
	return fmt.Sprintf(`{"send_predefined_action":{"ica_id":%d,"to_address":"%s"}}`, icaId, toAddress)
}
