package types_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos/wasm"

	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types"
)

func TestExecuteMsgs(t *testing.T) {
	const testAddress = "srdtrk"

	t.Parallel()

	// Create deposit message:
	depositMsg := &govtypes.MsgDeposit{
		ProposalId: 1,
		Depositor:  testAddress,
		Amount:     sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10000000))),
	}

	_, err := icatypes.SerializeCosmosTxWithEncoding(wasm.WasmEncoding().Codec, []proto.Message{depositMsg}, icatypes.EncodingProto3JSON)
	require.NoError(t, err)
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
