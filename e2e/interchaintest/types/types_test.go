package types_test

import (
	"encoding/json"
	"testing"

	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types"
	"github.com/stretchr/testify/require"
)

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