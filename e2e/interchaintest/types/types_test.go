package types_test

import (
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"

	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/e2esuite"
)

// This is some boilerplate test code to insert some tests for the types package.
// It is not meant to be executed, but to be used as a way to test some functions when
// debugging developing the types package.
func TestTypes(t *testing.T) {
	const testAddress = "srdtrk"

	t.Parallel()

	// Create deposit message:
	depositMsg := &govv1.MsgDeposit{
		ProposalId: 1,
		Depositor:  testAddress,
		Amount:     sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10000000))),
	}

	_, err := icatypes.SerializeCosmosTx(e2esuite.EncodingConfig().Codec, []proto.Message{depositMsg}, icatypes.EncodingProto3JSON)
	require.NoError(t, err)
}
