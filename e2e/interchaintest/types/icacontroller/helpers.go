package icacontroller

import (
	"encoding/base64"

	"github.com/cosmos/gogoproto/proto"

	codec "github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
)

// NewExecuteMsg_SendCustomIcaMessages_FromProto creates a new ExecuteMsg_SendCustomIcaMessages.
func NewExecuteMsg_SendCustomIcaMessages_FromProto(cdc codec.Codec, msgs []proto.Message, encoding string, memo *string, timeout *uint64) ExecuteMsg {
	bz, err := icatypes.SerializeCosmosTx(cdc, msgs, encoding)
	if err != nil {
		panic(err)
	}

	messages := base64.StdEncoding.EncodeToString(bz)

	return ExecuteMsg{
		SendCustomIcaMessages: &ExecuteMsg_SendCustomIcaMessages{
			Messages:       messages,
			PacketMemo:     memo,
			TimeoutSeconds: timeout,
		},
	}
}

// NewExecuteMsg_SendCosmosMsgs_FromProto creates a new ExecuteMsg_SendCosmosMsgs.
func NewExecuteMsg_SendCosmosMsgs_FromProto(msgs []proto.Message, memo *string, timeout *uint64) ExecuteMsg {
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

	return ExecuteMsg{
		SendCosmosMsgs: &ExecuteMsg_SendCosmosMsgs{
			Messages:       cosmosMsgs,
			PacketMemo:     memo,
			TimeoutSeconds: timeout,
		},
	}
}
