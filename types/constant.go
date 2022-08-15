package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethermint "github.com/evmos/ethermint/types"
)

const (
	Name          = "fxcore"
	AddressPrefix = "fx"

	DefaultDenom = "FX"
	// BaseDenomUnit defines the base denomination unit for Photons.
	// 1 FX = 1x10^{BaseDenomUnit} fx
	BaseDenomUnit = 18

	// DefaultEvmGasPrice is default gas price for evm transactions 500Gwei
	DefaultEvmGasPrice = 500000000000
)

func init() {
	sdk.SetCoinDenomRegex(func() string {
		return `[a-zA-Z][a-zA-Z0-9/-]{1,127}`
	})

	// votingPower = delegateToken / sdk.PowerReduction  --  sdk.TokensToConsensusPower(tokens Int)
	sdk.DefaultPowerReduction = sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(20), nil))

	// set chain id function
	ethermint.SetParseChainIDFunc(func(chainID string) (*big.Int, error) {
		return EIP155ChainID(), nil
	})
	ethermint.SetValidChainIDFunc(func(chainId string) bool {
		return ChainId() == chainId
	})
}

func SetConfig(isCosmosCoinType bool) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(AddressPrefix, AddressPrefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(AddressPrefix+sdk.PrefixValidator+sdk.PrefixOperator, AddressPrefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic)
	config.SetBech32PrefixForConsensusNode(AddressPrefix+sdk.PrefixValidator+sdk.PrefixConsensus, AddressPrefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic)
	if isCosmosCoinType {
		config.SetCoinType(sdk.CoinType)
	} else {
		config.SetCoinType(60)
	}
	config.Seal()

	if err := sdk.RegisterDenom(DefaultDenom, sdk.NewDec(BaseDenomUnit)); err != nil {
		panic(err)
	}
}