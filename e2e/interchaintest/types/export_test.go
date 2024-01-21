package types

import (
	"github.com/cosmos/gogoproto/proto"

	codec "github.com/cosmos/cosmos-sdk/codec"
)

// NewSendCustomIcaMessagesMsg is a wrapper for newSendCustomIcaMessagesMsg for internal testing
func NewSendCustomIcaMessagesMsg(cdc codec.BinaryCodec, msgs []proto.Message, encoding string, memo *string, timeout *uint64) string {
	return newSendCustomIcaMessagesMsg(cdc, msgs, encoding, memo, timeout)
}

// NewGetChannelQueryMsg is a wrapper for newGetChannelQueryMsg for internal testing
func NewGetChannelQueryMsg() map[string]interface{} {
	return newGetCallbackCounterQueryMsg()
}

// NewGetContractStateQueryMsg is a wrapper for newGetContractStateQueryMsg for internal testing
func NewGetContractStateQueryMsg() map[string]interface{} {
	return newGetContractStateQueryMsg()
}

// NewGetCallbackCounterQueryMsg is a wrapper for newGetCallbackCounterQueryMsg for internal testing
func NewGetCallbackCounterQueryMsg() map[string]interface{} {
	return newGetCallbackCounterQueryMsg()
}
