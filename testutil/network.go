package testutil

import (
	"fmt"
	"time"

	"github.com/functionx/fx-core/v2/app/helpers"

	hd2 "github.com/evmos/ethermint/crypto/hd"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/functionx/fx-core/v2/testutil/network"

	"github.com/functionx/fx-core/v2/app"
	fxtypes "github.com/functionx/fx-core/v2/types"
)

func DefNoSupplyGenesisState(cdc codec.Codec) app.GenesisState {
	genesisState := app.NewDefAppGenesisByDenom(fxtypes.DefaultDenom, cdc)
	bankState := banktypes.DefaultGenesisState()
	bankState.DenomMetadata = []banktypes.Metadata{fxtypes.GetFXMetaData(fxtypes.DefaultDenom)}
	genesisState[banktypes.ModuleName] = cdc.MustMarshalJSON(bankState)
	return genesisState
}

// DefaultNetworkConfig returns a sane default configuration suitable for nearly all
// testing requirements.
func DefaultNetworkConfig() network.Config {
	encCfg := app.MakeEncodingConfig()

	config := sdk.GetConfig()
	*config = *sdk.NewConfig()
	config.SetBech32PrefixForAccount(fxtypes.AddressPrefix, fxtypes.AddressPrefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(fxtypes.AddressPrefix+sdk.PrefixValidator+sdk.PrefixOperator, fxtypes.AddressPrefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic)
	config.SetBech32PrefixForConsensusNode(fxtypes.AddressPrefix+sdk.PrefixValidator+sdk.PrefixConsensus, fxtypes.AddressPrefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic)
	config.SetCoinType(118)
	config.Seal()

	return network.Config{
		Codec:             encCfg.Marshaler,
		TxConfig:          encCfg.TxConfig,
		LegacyAmino:       encCfg.Amino,
		InterfaceRegistry: encCfg.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor: func(val network.Validator) servertypes.Application {
			return app.New(
				val.Ctx.Logger, dbm.NewMemDB(), nil, true, make(map[int64]bool), val.Ctx.Config.RootDir, 0,
				encCfg,
				helpers.EmptyAppOptions{},
				baseapp.SetPruning(storetypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
				baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
			)
		},
		GenesisState:    DefNoSupplyGenesisState(encCfg.Marshaler),
		TimeoutCommit:   1 * time.Second,
		ChainID:         fxtypes.MainnetChainId,
		NumValidators:   4,
		BondDenom:       fxtypes.DefaultDenom,
		MinGasPrices:    fmt.Sprintf("4000000000000%s", fxtypes.DefaultDenom),
		AccountTokens:   sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction),
		StakingTokens:   sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
		BondedTokens:    sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),
		PruningStrategy: storetypes.PruningOptionNothing,
		CleanupDir:      true,
		SigningAlgo:     string(hd.Secp256k1Type),
		KeyringOptions: []keyring.Option{
			hd2.EthSecp256k1Option(),
		},
		PrintMnemonic: false,
	}
}
