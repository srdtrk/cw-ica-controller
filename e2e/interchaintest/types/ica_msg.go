package types

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	codec "github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
)

// newInstantiateMsg creates a new InstantiateMsg.
func newInstantiateMsg(admin *string) string {
	if admin == nil {
		return `{}`
	} else {
		return fmt.Sprintf(`{"admin":"%s"}`, *admin)
	}
}

type ChannelOpenInitOptions struct {
	// The connection id on this chain.
	ConnectionId string `json:"connection_id"`
	// The counterparty connection id on the counterparty chain.
	CounterpartyConnectionId string `json:"counterparty_connection_id"`
	// The optional counterparty port id.
	CounterpartyPortId *string `json:"counterparty_port_id,omitempty"`
	// The optional tx encoding.
	TxEncoding *string `json:"tx_encoding,omitempty"`
}

// NewInstantiateMsgWithChannelInitOptions creates a new InstantiateMsg with channel init options.
func NewInstantiateMsgWithChannelInitOptions(
	admin *string, connectionId string, counterpartyConnectionId string,
	counterpartyPortId *string, txEncoding *string,
) string {
	type InstantiateMsg struct {
		// The address of the admin of the ICA application.
		// If not specified, the sender is the admin.
		Admin *string `json:"admin,omitempty"`
		// The options to initialize the IBC channel upon contract instantiation.
		// If not specified, the IBC channel is not initialized, and the relayer must.
		ChannelOpenInitOptions *ChannelOpenInitOptions `json:"channel_open_init_options,omitempty"`
	}

	channelOpenInitOptions := ChannelOpenInitOptions{
		ConnectionId:             connectionId,
		CounterpartyConnectionId: counterpartyConnectionId,
		CounterpartyPortId:       counterpartyPortId,
		TxEncoding:               txEncoding,
	}

	instantiateMsg := InstantiateMsg{
		Admin:                  admin,
		ChannelOpenInitOptions: &channelOpenInitOptions,
	}

	jsonBytes, err := json.Marshal(instantiateMsg)
	if err != nil {
		panic(err)
	}

	return string(jsonBytes)
}

func newEmptyCreateChannelMsg() string {
	return `{ "create_channel": {} }`
}

func newCreateChannelMsg(
	connectionId string, counterpartyConnectionId string,
	counterpartyPortId *string, txEncoding *string,
) string {
	type ChannelCreateMsg struct {
		ChannelOpenInitOptions *ChannelOpenInitOptions `json:"channel_open_init_options,omitempty"`
	}

	type ChannelCreateMsgWrapper struct {
		CreateChannelMsg ChannelCreateMsg `json:"create_channel"`
	}

	channelOpenInitOptions := ChannelOpenInitOptions{
		ConnectionId:             connectionId,
		CounterpartyConnectionId: counterpartyConnectionId,
		CounterpartyPortId:       counterpartyPortId,
		TxEncoding:               txEncoding,
	}

	msg := ChannelCreateMsgWrapper{
		CreateChannelMsg: ChannelCreateMsg{
			ChannelOpenInitOptions: &channelOpenInitOptions,
		},
	}

	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return string(jsonBytes)
}

// newSendCustomIcaMessagesMsg creates a new SendCustomIcaMessagesMsg.
func newSendCustomIcaMessagesMsg(cdc codec.BinaryCodec, msgs []proto.Message, encoding string, memo *string, timeout *uint64) string {
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

// newSendCosmosMsgsMsg creates a new SendCosmosMsgsMsg.
func newSendCosmosMsgsMsg(cosmosMsgs []ContractCosmosMsg, memo *string, timeout *uint64) string {
	type SendCosmosMsgsAsIcaTxMsg struct {
		Messages       []ContractCosmosMsg `json:"messages"`
		PacketMemo     *string             `json:"packet_memo,omitempty"`
		TimeoutSeconds *uint64             `json:"timeout_seconds,omitempty"`
	}

	type SendCosmosMsgsAsIcaTxMsgWrapper struct {
		SendCosmosMsgsAsIcaTxMsg SendCosmosMsgsAsIcaTxMsg `json:"send_cosmos_msgs"`
	}

	msg := SendCosmosMsgsAsIcaTxMsgWrapper{
		SendCosmosMsgsAsIcaTxMsg: SendCosmosMsgsAsIcaTxMsg{
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

// newSendCosmosMsgsMsgFromProto creates a new SendCosmosMsgsMsg.
func newSendCosmosMsgsMsgFromProto(msgs []proto.Message, memo *string, timeout *uint64) string {
	type SendCosmosMsgsAsIcaTxMsg struct {
		Messages       []ContractCosmosMsg `json:"messages"`
		PacketMemo     *string             `json:"packet_memo,omitempty"`
		TimeoutSeconds *uint64             `json:"timeout_seconds,omitempty"`
	}

	type SendCosmosMsgsAsIcaTxMsgWrapper struct {
		SendCosmosMsgsAsIcaTxMsg SendCosmosMsgsAsIcaTxMsg `json:"send_cosmos_msgs_as_ica_tx"`
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
		SendCosmosMsgsAsIcaTxMsg: SendCosmosMsgsAsIcaTxMsg{
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
