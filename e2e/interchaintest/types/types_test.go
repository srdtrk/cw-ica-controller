package types_test

import (
	"encoding/json"
	"testing"

	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types"
	"github.com/stretchr/testify/require"
)

func TestInstantiateMsg(t *testing.T) {
	t.Parallel()

	msg := types.NewInstantiateMsg(nil)
	require.Equal(t, `{}`, msg)

	admin := "srdtrk"
	msg = types.NewInstantiateMsg(&admin)
	require.Equal(t, `{"admin":"srdtrk"}`, msg)
}

func TestExecuteMsgs(t *testing.T) {
	t.Parallel()

	sendPredefinedActionMsg := types.NewSendPredefinedActionMsg("srdtrk")
	require.Equal(t, `{"send_predefined_action":{"to_address":"srdtrk"}}`, sendPredefinedActionMsg)

	sendCustomIcaMessagesMsg := types.NewSendCustomIcaMessagesMsg(nil, nil, nil, nil)
	require.Equal(t, `{"send_custom_ica_messages":{"messages":[]}}`, sendCustomIcaMessagesMsg)
	memo := "test"
	sendCustomIcaMessagesMsg = types.NewSendCustomIcaMessagesMsg(nil, nil, &memo, nil)
	require.Equal(t, `{"send_custom_ica_messages":{"messages":[],"packet_memo":"test"}}`, sendCustomIcaMessagesMsg)
	timeout := uint64(150)
	sendCustomIcaMessagesMsg = types.NewSendCustomIcaMessagesMsg(nil, nil, nil, &timeout)
	require.Equal(t, `{"send_custom_ica_messages":{"messages":[],"timeout_seconds":150}}`, sendCustomIcaMessagesMsg)
}

func TestQueries(t *testing.T) {
	t.Parallel()

	channelQueryMsg := types.NewGetChannelQueryMsg()
	msg, err := json.Marshal(channelQueryMsg)
	require.NoError(t, err)
	require.Equal(t, `{"get_channel":{}}`, string(msg))

	contractStateQueryMsg := types.NewGetContractStateQueryMsg()
	msg, err = json.Marshal(contractStateQueryMsg)
	require.NoError(t, err)
	require.Equal(t, `{"get_contract_state":{}}`, string(msg))

	callbackCounterQueryMsg := types.NewGetCallbackCounterQueryMsg()
	msg, err = json.Marshal(callbackCounterQueryMsg)
	require.NoError(t, err)
	require.Equal(t, `{"get_callback_counter":{}}`, string(msg))
}
