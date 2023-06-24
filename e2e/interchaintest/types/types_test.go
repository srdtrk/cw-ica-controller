package types_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types"

	sdkmath "cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos/wasm"

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
	const testAddress = "srdtrk"

	t.Parallel()

	// Basic tests:
	sendPredefinedActionMsg := types.NewSendPredefinedActionMsg(testAddress)
	require.Equal(t, `{"send_predefined_action":{"to_address":"srdtrk"}}`, sendPredefinedActionMsg)

	sendCustomIcaMessagesMsg := types.NewSendCustomIcaMessagesMsg(nil, nil, nil, nil)
	require.Equal(t, `{"send_custom_ica_messages":{"messages":[]}}`, sendCustomIcaMessagesMsg)
	memo := "test"
	sendCustomIcaMessagesMsg = types.NewSendCustomIcaMessagesMsg(nil, nil, &memo, nil)
	require.Equal(t, `{"send_custom_ica_messages":{"messages":[],"packet_memo":"test"}}`, sendCustomIcaMessagesMsg)
	timeout := uint64(150)
	sendCustomIcaMessagesMsg = types.NewSendCustomIcaMessagesMsg(nil, nil, nil, &timeout)
	require.Equal(t, `{"send_custom_ica_messages":{"messages":[],"timeout_seconds":150}}`, sendCustomIcaMessagesMsg)

	// Test with custom messages:

	type SendCustomIcaMessagesMsg struct {
		Messages       []string `json:"messages"`
		PacketMemo     *string  `json:"packet_memo,omitempty"`
		TimeoutSeconds *uint64  `json:"timeout_seconds,omitempty"`
	}

	type SendCustomIcaMessagesMsgWrapper struct {
		SendCustomIcaMessagesMsg SendCustomIcaMessagesMsg `json:"send_custom_ica_messages"`
	}

	testProposal := &govtypes.TextProposal{
		Title:       "IBC Gov Proposal",
		Description: "tokens for all!",
	}
	protoAny, err := codectypes.NewAnyWithValue(testProposal)
	require.NoError(t, err)
	proposalMsg := &govtypes.MsgSubmitProposal{
		Content:        protoAny,
		InitialDeposit: sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(5000))),
		Proposer:       testAddress,
	}

	// Create deposit message:
	depositMsg := &govtypes.MsgDeposit{
		ProposalId: 1,
		Depositor:  testAddress,
		Amount:     sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10000000))),
	}

	customMsg := types.NewSendCustomIcaMessagesMsg(wasm.WasmEncoding().Codec, []sdk.Msg{proposalMsg, depositMsg}, nil, nil)
	unmarshaledCustomMsg := SendCustomIcaMessagesMsgWrapper{}
	err = json.Unmarshal([]byte(customMsg), &unmarshaledCustomMsg)
	require.NoError(t, err)
	base64Msgs := unmarshaledCustomMsg.SendCustomIcaMessagesMsg.Messages
	require.Len(t, base64Msgs, 2)
	stringMsgs := make([]string, len(base64Msgs))
	for i, base64Msg := range base64Msgs {
		msg, err := base64.StdEncoding.DecodeString(base64Msg)
		require.NoError(t, err)
		stringMsgs[i] = string(msg)
	}

	expectedMsg1 := `{"@type":"/cosmos.gov.v1beta1.MsgSubmitProposal","content":{"@type":"/cosmos.gov.v1beta1.TextProposal","title":"IBC Gov Proposal","description":"tokens for all!"},"initial_deposit":[{"denom":"stake","amount":"5000"}],"proposer":"srdtrk"}`
	expectedMsg2 := `{"@type":"/cosmos.gov.v1beta1.MsgDeposit","proposal_id":"1","depositor":"srdtrk","amount":[{"denom":"stake","amount":"10000000"}]}`
	expectedMsgs := []string{expectedMsg1, expectedMsg2}

	require.Equal(t, expectedMsgs, stringMsgs)
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
