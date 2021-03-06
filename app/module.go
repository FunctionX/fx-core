package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/x/evm"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/evmos/ethermint/x/feemarket"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	fxtypes "github.com/functionx/fx-core/v2/types"
)

type EvmAppModule struct {
	evm.AppModule
}

// DefaultGenesis returns default genesis state as raw bytes for the evm
// module.
func (EvmAppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genesisState := evmtypes.DefaultGenesisState()
	genesisState.Params.EvmDenom = fxtypes.DefaultDenom
	return cdc.MustMarshalJSON(genesisState)
}

type FeeMarketAppModule struct {
	feemarket.AppModule
}

// DefaultGenesis returns default genesis state as raw bytes for the fee market module.
func (FeeMarketAppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genesisState := feemarkettypes.DefaultGenesisState()
	genesisState.Params.BaseFee = sdk.NewInt(500_000_000_000)
	genesisState.Params.MinGasPrice = sdk.NewDec(500_000_000_000)
	genesisState.Params.MinGasMultiplier = sdk.ZeroDec()
	return cdc.MustMarshalJSON(genesisState)
}
