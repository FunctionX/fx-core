package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(AddressPrefix, AddressPrefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(AddressPrefix+sdk.PrefixValidator+sdk.PrefixOperator, AddressPrefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic)
	config.SetBech32PrefixForConsensusNode(AddressPrefix+sdk.PrefixValidator+sdk.PrefixConsensus, AddressPrefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic)
	config.SetCoinType(60)
	config.Seal()

	// votingPower = delegateToken / sdk.PowerReduction  --  sdk.TokensToConsensusPower(tokens Int)
	sdk.DefaultPowerReduction = sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(20), nil))

	if err := sdk.RegisterDenom(DefaultDenom, sdk.NewDec(18)); err != nil {
		panic(err)
	}
}
