package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/functionx/fx-core/v2/types"
	"github.com/functionx/fx-core/v2/x/ibc/applications/transfer/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseReceiveAndAmountByPacket(t *testing.T) {
	type expect struct {
		address string
		amount  sdk.Int
		fee     sdk.Int
	}
	testCases := []struct {
		name    string
		packet  types.FungibleTokenPacketData
		expPass bool
		err     error
		expect  expect
	}{
		{"no router - expect address is receive", types.FungibleTokenPacketData{Receiver: "fx1yef9232palu3ps25ldjr62ck046rgd8292kc73", Amount: "1", Fee: "0"}, true, nil,
			expect{address: "fx1yef9232palu3ps25ldjr62ck046rgd8292kc73", amount: sdk.NewIntFromUint64(1), fee: sdk.NewIntFromUint64(0)},
		},
		{"no router - expect fee is 0, input 1", types.FungibleTokenPacketData{Receiver: "fx1yef9232palu3ps25ldjr62ck046rgd8292kc73", Amount: "1", Fee: "0"}, true, nil,
			expect{address: "fx1yef9232palu3ps25ldjr62ck046rgd8292kc73", amount: sdk.NewIntFromUint64(1), fee: sdk.NewIntFromUint64(0)},
		},
		{"router - expect address is sender", types.FungibleTokenPacketData{Sender: "fx12qv5llp5mv8m8h5s5nh8tkmxap5267pqd38g7h", Receiver: "fx1yef9232palu3ps25ldjr62ck046rgd8292kc73", Amount: "1", Fee: "0", Router: "erc20"}, true, nil,
			expect{address: "fx12qv5llp5mv8m8h5s5nh8tkmxap5267pqd38g7h", amount: sdk.NewIntFromUint64(1), fee: sdk.NewIntFromUint64(0)},
		},
		{"router - expect fee is 1, input 1", types.FungibleTokenPacketData{Sender: "fx12qv5llp5mv8m8h5s5nh8tkmxap5267pqd38g7h", Receiver: "fx1yef9232palu3ps25ldjr62ck046rgd8292kc73", Amount: "1", Fee: "1", Router: "erc20"}, true, nil,
			expect{address: "fx12qv5llp5mv8m8h5s5nh8tkmxap5267pqd38g7h", amount: sdk.NewIntFromUint64(1), fee: sdk.NewIntFromUint64(1)},
		},
		{"router - expect address is sender, input eip address", types.FungibleTokenPacketData{Sender: "0x50194ffc34DB0fb3De90A4eE75dB66e868AD7820", Receiver: "0x50194ffc34DB0fb3De90A4eE75dB66e868AD7820", Amount: "1", Fee: "1", Router: "erc20"}, true, nil,
			expect{address: "fx12qv5llp5mv8m8h5s5nh8tkmxap5267pqd38g7h", amount: sdk.NewIntFromUint64(1), fee: sdk.NewIntFromUint64(1)},
		},
		{"router - expect address is sender, input eip address", types.FungibleTokenPacketData{Sender: "0x50194ffc34DB0fb3De90A4eE75dB66e868AD7820", Receiver: "0x50194ffc34DB0fb3De90A4eE75dB66e868AD7820", Amount: "1", Fee: "1", Router: "erc20"}, true, nil,
			expect{address: "fx12qv5llp5mv8m8h5s5nh8tkmxap5267pqd38g7h", amount: sdk.NewIntFromUint64(1), fee: sdk.NewIntFromUint64(1)},
		},
		{"error router - expect error, sender eip address is lowercase", types.FungibleTokenPacketData{Sender: "0x50194ffc34db0fb3de90a4ee75db66e868ad7820", Receiver: "0x50194ffc34DB0fb3De90A4eE75dB66e868AD7820", Amount: "1", Fee: "1", Router: "erc20"}, false,
			fmt.Errorf("decoding bech32 failed: invalid character not part of charset: 98"),
			expect{address: "", amount: sdk.Int{}, fee: sdk.Int{}},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualAddress, actualAmount, actualFee, err := parseReceiveAndAmountByPacket(tc.packet)
			if tc.expPass {
				require.NoError(t, err, "valid test case %d failed: %v", i, err)
			} else {
				require.Error(t, err)
				require.EqualValues(t, err.Error(), tc.err.Error())
			}
			require.EqualValues(t, tc.expect.address, actualAddress.String())
			require.EqualValues(t, tc.expect.amount.String(), actualAmount.String())
			require.EqualValues(t, tc.expect.fee.String(), actualFee.String())
		})
	}
}

func TestParsePacketAddress(t *testing.T) {
	testCases := []struct {
		name    string
		address string
		expPass bool
		err     error
		expect  string
	}{
		{"normal fx address", "fx1yef9232palu3ps25ldjr62ck046rgd8292kc73", true, nil, "fx1yef9232palu3ps25ldjr62ck046rgd8292kc73"},
		{"normal cosmos address", "cosmos1yef9232palu3ps25ldjr62ck046rgd82l68dnq", true, nil, "fx1yef9232palu3ps25ldjr62ck046rgd8292kc73"},
		{"normal eip address", "0x2652554541Eff910C154fB643d2b167D743434EA", true, nil, "fx1yef9232palu3ps25ldjr62ck046rgd8292kc73"},

		{"err bech32 addres - kc74", "fx1yef9232palu3ps25ldjr62ck046rgd8292kc74", false, fmt.Errorf("decoding bech32 failed: invalid checksum (expected 92kc73 got 92kc74)"), ""},
		{"err lowercase eip address", "0x2652554541eff910c154fb643d2b167d743434ea", false, fmt.Errorf("decoding bech32 failed: invalid checksum (expected j389ls got 3434ea)"), ""},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualAddress, err := parsePacketAddress(tc.address)
			if tc.expPass {
				require.NoError(t, err, "valid test case %d failed: %v", i, err)
			} else {
				require.Error(t, err)
				require.EqualValues(t, err.Error(), tc.err.Error())
			}
			require.EqualValues(t, tc.expect, actualAddress.String())
		})
	}
}
