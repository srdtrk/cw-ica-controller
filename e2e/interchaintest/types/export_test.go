package types

import (
	"github.com/cosmos/gogoproto/proto"

	codec "github.com/cosmos/cosmos-sdk/codec"
)

// NewSendCustomIcaMessagesMsg is a wrapper for newSendCustomIcaMessagesMsg for internal testing
func NewSendCustomIcaMessagesMsg(cdc codec.BinaryCodec, msgs []proto.Message, encoding string, memo *string, timeout *uint64) string {
	return newSendCustomIcaMessagesMsg(cdc, msgs, encoding, memo, timeout)
}
