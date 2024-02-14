package owner

import "github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/icacontroller"

type ContractState struct {
	Admin               string `json:"admin"`
	IcaControllerCodeId uint64 `json:"ica_controller_code_id"`
}

type IcaContractState struct {
	ContractAddr string    `json:"contract_addr"`
	IcaState     *IcaState `json:"ica_state,omitempty"`
}

type IcaState struct {
	IcaId        uint64              `json:"ica_id"`
	IcaAddr      string              `json:"ica_addr"`
	TxEncoding   string              `json:"tx_encoding"`
	ChannelState icacontroller.State `json:"channel_state"`
}
