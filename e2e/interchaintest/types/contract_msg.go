package types

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	codec "github.com/cosmos/cosmos-sdk/codec"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
)

// NewInstantiateMsg creates a new InstantiateMsg.
func NewInstantiateMsg(admin *string) string {
	if admin == nil {
		return `{}`
	} else {
		return fmt.Sprintf(`{"admin":"%s"}`, *admin)
	}
}

// NewSendPredefinedActionMsg creates a new SendPredefinedActionMsg.
func NewSendPredefinedActionMsg(to_address string) string {
	return fmt.Sprintf(`{"send_predefined_action":{"to_address":"%s"}}`, to_address)
}

// NewSendCustomIcaMessagesMsg creates a new SendCustomIcaMessagesMsg.
func NewSendCustomIcaMessagesMsg(cdc codec.BinaryCodec, msgs []proto.Message, encoding string, memo *string, timeout *uint64) string {
	type SendCustomIcaMessagesMsg struct {
		Messages       string  `json:"messages"`
		PacketMemo     *string `json:"packet_memo,omitempty"`
		TimeoutSeconds *uint64 `json:"timeout_seconds,omitempty"`
	}

	type SendCustomIcaMessagesMsgWrapper struct {
		SendCustomIcaMessagesMsg SendCustomIcaMessagesMsg `json:"send_custom_ica_messages"`
	}

	bz, err := icatypes.SerializeCosmosTxWithEncoding(cdc, msgs, encoding)
	if err != nil {
		panic(err)
	}

	messages := base64.StdEncoding.EncodeToString(bz)

	msg := SendCustomIcaMessagesMsgWrapper{
		SendCustomIcaMessagesMsg: SendCustomIcaMessagesMsg{
			Messages:       messages,
			PacketMemo:     memo,
			TimeoutSeconds: timeout,
		},
	}

	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return string(jsonBytes)
}
