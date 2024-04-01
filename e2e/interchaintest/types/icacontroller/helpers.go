package icacontroller

import (
	"encoding/base64"

	"github.com/cosmos/gogoproto/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// NewExecuteMsg_SendCosmosMsgs_FromProto creates a new ExecuteMsg_SendCosmosMsgs.
func NewExecuteMsg_SendCosmosMsgs_FromProto(msgs []proto.Message, memo *string, timeout *int) ExecuteMsg {
	cosmosMsgs := make([]CosmosMsg_for_Empty, len(msgs))

	for i, msg := range msgs {
		protoAny, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}

		cosmosMsgs[i] = CosmosMsg_for_Empty{
			Stargate: &CosmosMsg_for_Empty_Stargate{
				TypeUrl: protoAny.TypeUrl,
				Value:   Binary(base64.StdEncoding.EncodeToString(protoAny.Value)),
			},
		}
	}

	return ExecuteMsg{
		SendCosmosMsgs: &ExecuteMsg_SendCosmosMsgs{
			Messages:       cosmosMsgs,
			PacketMemo:     memo,
			TimeoutSeconds: timeout,
		},
	}
}
