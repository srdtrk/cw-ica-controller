package types

import (
	"encoding/base64"
	"encoding/json"

	codec "github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewSendCustomIcaMessagesMsg(cdc codec.BinaryCodec, msgs []sdk.Msg, memo *string, timeout *uint64) ([]byte, error) {
	type SendCustomIcaMessagesMsg struct {
		Messages       []string `json:"messages"`
		PacketMemo     *string  `json:"packet_memo,omitempty"`
		TimeoutSeconds *uint64  `json:"timeout_seconds,omitempty"`
	}

	type SendCustomIcaMessagesMsgWrapper struct {
		SendCustomIcaMessagesMsg SendCustomIcaMessagesMsg `json:"send_custom_ica_messages"`
	}

	messages := make([]string, len(msgs))

	for _, msg := range msgs {
		bz, err := cdc.(*codec.ProtoCodec).MarshalJSON(msg)
		if err != nil {
			return nil, err
		}
		b64 := base64.StdEncoding.EncodeToString(bz)
		messages = append(messages, b64)
	}

	msg := SendCustomIcaMessagesMsgWrapper{
		SendCustomIcaMessagesMsg: SendCustomIcaMessagesMsg{
			Messages:       messages,
			PacketMemo:     memo,
			TimeoutSeconds: timeout,
		},
	}

	return json.Marshal(msg)
}
