package app

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	ethermint "github.com/evmos/ethermint/types"
	abci "github.com/tendermint/tendermint/abci/types"

	fxtypes "github.com/functionx/fx-core/v2/types"
	bsctypes "github.com/functionx/fx-core/v2/x/bsc/types"
	crosschaintypes "github.com/functionx/fx-core/v2/x/crosschain/types"
	polygontypes "github.com/functionx/fx-core/v2/x/polygon/types"
	trontypes "github.com/functionx/fx-core/v2/x/tron/types"
)

func init() {
	// set chain id function
	ethermint.SetParseChainIDFunc(func(chainID string) (*big.Int, error) {
		return fxtypes.EIP155ChainID(), nil
	})
	ethermint.SetValidChainIDFunc(func(chainId string) bool {
		return fxtypes.ChainId() == chainId
	})

	crosschaintypes.RegisterValidateBasic(bsctypes.ModuleName, crosschaintypes.EthereumMsgValidate{})
	crosschaintypes.RegisterValidateBasic(polygontypes.ModuleName, crosschaintypes.EthereumMsgValidate{})
	crosschaintypes.RegisterValidateBasic(trontypes.ModuleName, trontypes.TronMsgValidate{})
}

// CosmosApp implements the common methods for a Cosmos SDK-based application
// specific blockchain.
type CosmosApp interface {
	// The assigned name of the app.
	Name() string

	// The application types codec.
	// NOTE: This shoult be sealed before being returned.
	LegacyAmino() *codec.LegacyAmino

	// Application updates every begin block.
	BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock

	// Application updates every end block.
	EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock

	// Application update at chain (i.e app) initialization.
	InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain

	// Loads the app at a given height.
	LoadHeight(height int64) error

	// Exports the state of the application for a genesis file.
	ExportAppStateAndValidators(
		forZeroHeight bool, jailAllowedAddrs []string,
	) (types.ExportedApp, error)

	// All the registered module account addreses.
	ModuleAccountAddrs() map[string]bool

	// Helper for the simulation framework.
	SimulationManager() *module.SimulationManager
}
