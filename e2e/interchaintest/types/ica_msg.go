package types

import (
	"encoding/base64"
	"encoding/json"

	"github.com/cosmos/gogoproto/proto"

	codec "github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/icacontroller"
)

// newSendCustomIcaMessagesMsg creates a new SendCustomIcaMessagesMsg.
func newSendCustomIcaMessagesMsg(cdc codec.BinaryCodec, msgs []proto.Message, encoding string, memo *string, timeout *uint64) string {
	bz, err := icatypes.SerializeCosmosTxWithEncoding(cdc, msgs, encoding)
	if err != nil {
		panic(err)
	}

	messages := base64.StdEncoding.EncodeToString(bz)

	msg := icacontroller.ExecuteMsg{
		SendCustomIcaMessages: &icacontroller.ExecuteMsg_SendCustomIcaMessages{
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

// newSendCosmosMsgsMsgFromProto creates a new SendCosmosMsgsMsg.
func newSendCosmosMsgsMsgFromProto(msgs []proto.Message, memo *string, timeout *uint64) string {
	type SendCosmosMsgsAsIcaTxMsg struct {
		Messages       []ContractCosmosMsg `json:"messages"`
		PacketMemo     *string             `json:"packet_memo,omitempty"`
		TimeoutSeconds *uint64             `json:"timeout_seconds,omitempty"`
	}

	type SendCosmosMsgsAsIcaTxMsgWrapper struct {
		SendCosmosMsgsMsg SendCosmosMsgsAsIcaTxMsg `json:"send_cosmos_msgs"`
	}

	cosmosMsgs := make([]ContractCosmosMsg, len(msgs))

	for i, msg := range msgs {
		protoAny, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}

		cosmosMsgs[i] = ContractCosmosMsg{
			Stargate: &StargateCosmosMsg{
				TypeUrl: protoAny.TypeUrl,
				Value:   base64.StdEncoding.EncodeToString(protoAny.Value),
			},
		}

		if err != nil {
			panic(err)
		}
	}

	msg := SendCosmosMsgsAsIcaTxMsgWrapper{
		SendCosmosMsgsMsg: SendCosmosMsgsAsIcaTxMsg{
			Messages:       cosmosMsgs,
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
