package types

import "encoding/base64"

type ContractCosmosMsg struct {
	Stargate *StargateCosmosMsg `json:"stargate"`
	Bank     *BankCosmosMsg     `json:"bank"`
	IBC      *IbcCosmosMsg      `json:"ibc"`
}

type StargateCosmosMsg struct {
	// Proto Any type URL
	TypeUrl string `json:"type_url"`
	// Base64 encoded bytes
	Value string `json:"value"`
}

type BankCosmosMsg struct {
	Send *BankSendCosmosMsg `json:"send"`
}

type IbcCosmosMsg struct {
	Transfer *IbcTransferCosmosMsg `json:"transfer"`
}

type BankSendCosmosMsg struct {
	ToAddress string `json:"to_address"`
	Amount    []Coin `json:"amount"`
}

type IbcTransferCosmosMsg struct {
	ChannelID string `json:"channel_id"`
	ToAddress string `json:"to_address"`
	Amount    Coin   `json:"amount"`
	// This is optional
	// Timeout string `json:"timeout"`
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// NewCosmosMsgWithStargate creates a new CosmosMsg with a Stargate message
func NewCosmosMsgWithStargate(typeUrl string, value []byte) *ContractCosmosMsg {
	return &ContractCosmosMsg{
		Stargate: &StargateCosmosMsg{
			TypeUrl: typeUrl,
			Value:   base64.StdEncoding.EncodeToString(value),
		},
	}
}

// NewCosmosMsgWithBankSend creates a new CosmosMsg with a Bank message
func NewCosmosMsgWithBankSend(toAddress string, amount []Coin) *ContractCosmosMsg {
	return &ContractCosmosMsg{
		Bank: &BankCosmosMsg{
			Send: &BankSendCosmosMsg{
				ToAddress: toAddress,
				Amount:    amount,
			},
		},
	}
}

// NewCosmosMsgWithIbcTransfer creates a new CosmosMsg with an IBC transfer message
func NewCosmosMsgWithIbcTransfer(channelID string, toAddress string, amount Coin) *ContractCosmosMsg {
	return &ContractCosmosMsg{
		IBC: &IbcCosmosMsg{
			Transfer: &IbcTransferCosmosMsg{
				ChannelID: channelID,
				ToAddress: toAddress,
				Amount:    amount,
			},
		},
	}
}
