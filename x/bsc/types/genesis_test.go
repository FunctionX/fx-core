package types

import (
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	fxtypes "github.com/functionx/fx-core/v2/types"
	crosschaintypes "github.com/functionx/fx-core/v2/x/crosschain/types"
)

func TestDefaultGenesisState(t *testing.T) {
	tests := []struct {
		name string
		want *crosschaintypes.GenesisState
	}{
		{
			name: "bsc default genesis",
			want: &crosschaintypes.GenesisState{
				Params: crosschaintypes.Params{
					GravityId:                         "fx-bsc-bridge",
					AverageBlockTime:                  5_000,
					ExternalBatchTimeout:              12 * 3600 * 1000,
					AverageExternalBlockTime:          5_000,
					SignedWindow:                      20_000,
					SlashFraction:                     sdk.NewDec(1).Quo(sdk.NewDec(100)),
					OracleSetUpdatePowerChangePercent: sdk.NewDec(1).Quo(sdk.NewDec(10)),
					IbcTransferTimeoutHeight:          20_000,
					DelegateThreshold:                 sdk.NewCoin(fxtypes.DefaultDenom, sdk.NewInt(10_000).MulRaw(1e18)),
					DelegateMultiple:                  10,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultGenesisState(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultGenesisState() = %v, want %v", got, tt.want)
			}
		})
	}
}
