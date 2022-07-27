package types_test

import (
	"fmt"
	"testing"

	"github.com/functionx/fx-core/v2/x/ibc/applications/transfer/types"

	_ "github.com/functionx/fx-core/v2/app"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
)

// define constants used for testing
const (
	validPort        = "testportid"
	invalidPort      = "(invalidport1)"
	invalidShortPort = "p"
	// 195 characters
	invalidLongPort = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Duis eros neque, ultricies vel ligula ac, convallis porttitor elit. Maecenas tincidunt turpis elit, vel faucibus nisl pellentesque sodales"

	validChannel        = "testchannel"
	invalidChannel      = "(invalidchannel1)"
	invalidShortChannel = "invalid"
	invalidLongChannel  = "invalidlongchannelinvalidlongchannelinvalidlongchannelinvalidlongchannel"
)

var (
	addr1     = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
	addr2     = sdk.AccAddress("testaddr2").String()
	emptyAddr string

	coin             = sdk.NewCoin("demo", sdk.NewInt(100))
	ibcCoin          = sdk.NewCoin("ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2", sdk.NewInt(100))
	invalidIBCCoin   = sdk.NewCoin("ibc/7F1D3FCF4AE79E1554", sdk.NewInt(100))
	invalidDenomCoin = sdk.Coin{Denom: "0demo", Amount: sdk.NewInt(100)}
	zeroCoin         = sdk.Coin{Denom: "demos", Amount: sdk.NewInt(0)}

	timeoutHeight = clienttypes.NewHeight(0, 10)

	defaultRouter = ""
	defaultFee    = sdk.Coin{Denom: "demo", Amount: sdk.ZeroInt()}
	defaultIbcFee = sdk.NewCoin("ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2", sdk.ZeroInt())
)

// TestMsgTransferRoute tests Route for MsgTransfer
func TestMsgTransferRoute(t *testing.T) {
	msg := types.NewMsgTransfer(validPort, validChannel, coin, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultFee)

	require.Equal(t, types.RouterKey, msg.Route())
}

// TestMsgTransferType tests Type for MsgTransfer
func TestMsgTransferType(t *testing.T) {
	msg := types.NewMsgTransfer(validPort, validChannel, coin, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultFee)

	require.Equal(t, "transfer", msg.Type())
}

func TestMsgTransferGetSignBytes(t *testing.T) {
	msg := types.NewMsgTransfer(validPort, validChannel, coin, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultFee)
	expected := fmt.Sprintf(`{"type":"fxtransfer/MsgTransfer","value":{"fee":{"amount":"0","denom":"demo"},"receiver":"%s","sender":"%s","source_channel":"testchannel","source_port":"testportid","timeout_height":{"revision_height":"10"},"token":{"amount":"100","denom":"demo"}}}`, addr2, addr1)
	require.NotPanics(t, func() {
		res := msg.GetSignBytes()
		require.Equal(t, expected, string(res))
	})
}

// TestMsgTransferValidation tests ValidateBasic for MsgTransfer
func TestMsgTransferValidation(t *testing.T) {
	testCases := []struct {
		name    string
		msg     *types.MsgTransfer
		expPass bool
	}{
		{"valid msg with base denom", types.NewMsgTransfer(validPort, validChannel, coin, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultFee), true},
		{"valid msg with trace hash", types.NewMsgTransfer(validPort, validChannel, ibcCoin, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultIbcFee), true},
		{"invalid ibc denom", types.NewMsgTransfer(validPort, validChannel, invalidIBCCoin, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultFee), false},
		{"too short port id", types.NewMsgTransfer(invalidShortPort, validChannel, coin, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultFee), false},
		{"too long port id", types.NewMsgTransfer(invalidLongPort, validChannel, coin, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultFee), false},
		{"port id contains non-alpha", types.NewMsgTransfer(invalidPort, validChannel, coin, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultFee), false},
		{"too short channel id", types.NewMsgTransfer(validPort, invalidShortChannel, coin, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultFee), false},
		{"too long channel id", types.NewMsgTransfer(validPort, invalidLongChannel, coin, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultFee), false},
		{"channel id contains non-alpha", types.NewMsgTransfer(validPort, invalidChannel, coin, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultFee), false},
		{"invalid denom", types.NewMsgTransfer(validPort, validChannel, invalidDenomCoin, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultFee), false},
		{"zero coin", types.NewMsgTransfer(validPort, validChannel, zeroCoin, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultFee), false},
		{"missing sender address", types.NewMsgTransfer(validPort, validChannel, coin, emptyAddr, addr2, timeoutHeight, 0, defaultRouter, defaultFee), false},
		{"missing recipient address", types.NewMsgTransfer(validPort, validChannel, coin, addr1, "", timeoutHeight, 0, defaultRouter, defaultFee), false},
		{"empty coin", types.NewMsgTransfer(validPort, validChannel, sdk.Coin{}, addr1, addr2, timeoutHeight, 0, defaultRouter, defaultFee), false},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %s", i, tc.name)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

// TestMsgTransferGetSigners tests GetSigners for MsgTransfer
func TestMsgTransferGetSigners(t *testing.T) {
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	msg := types.NewMsgTransfer(validPort, validChannel, coin, addr.String(), addr2, timeoutHeight, 0, defaultRouter, defaultFee)
	res := msg.GetSigners()

	require.Equal(t, []sdk.AccAddress{addr}, res)
}
