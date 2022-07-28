package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func Test_Address(t *testing.T) {
	accAddress := sdk.AccAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	assert.Equal(t, "fx1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc59jattd", accAddress.String())

	bech32, err := sdk.AccAddressFromBech32(accAddress.String())
	assert.NoError(t, err)
	assert.Equal(t, bech32, accAddress)
}

func Test_NormalizeCoin(t *testing.T) {
	denom, err := sdk.GetBaseDenom()
	assert.NoError(t, err)
	assert.Equal(t, DefaultDenom, denom)
	coin := sdk.NewCoin(DefaultDenom, sdk.NewInt(1e18))
	assert.Equal(t, coin, sdk.NormalizeCoin(coin))
}
