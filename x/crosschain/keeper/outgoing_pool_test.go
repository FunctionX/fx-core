package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/functionx/fx-core/v7/testutil/helpers"
	fxtypes "github.com/functionx/fx-core/v7/types"
	"github.com/functionx/fx-core/v7/x/crosschain/types"
	ethtypes "github.com/functionx/fx-core/v7/x/eth/types"
)

func (suite *KeeperTestSuite) TestKeeper_OutgoingPool() {
	sender := helpers.GenHexAddress().Bytes()
	bridgeToken := helpers.GenHexAddress().Hex()
	denom := types.NewBridgeDenom(suite.chainName, bridgeToken)
	suite.Equal(sdk.NewCoin(denom, sdkmath.ZeroInt()), suite.app.BankKeeper.GetSupply(suite.ctx, denom))

	sendAmount := sdk.NewCoin(denom, sdkmath.NewInt(int64(tmrand.Uint32()*2)))
	err := suite.app.BankKeeper.MintCoins(suite.ctx, suite.chainName, sdk.NewCoins(sendAmount))
	suite.NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, suite.chainName, sender, sdk.NewCoins(sendAmount))
	suite.NoError(err)
	suite.Equal(sendAmount, suite.app.BankKeeper.GetSupply(suite.ctx, denom))

	suite.Keeper().AddBridgeToken(suite.ctx, bridgeToken, denom)

	suite.Equal(suite.app.BankKeeper.GetAllBalances(suite.ctx, sender).AmountOf(denom).String(), sendAmount.Amount.String())
	receiver := helpers.GenHexAddress().Hex()
	amount := sdk.NewCoin(denom, sendAmount.Amount.QuoRaw(2))
	txId, err := suite.Keeper().AddToOutgoingPool(suite.ctx, sender, receiver, amount, amount)
	suite.NoError(err)
	suite.Equal(suite.app.BankKeeper.GetAllBalances(suite.ctx, sender).AmountOf(denom).String(), sdkmath.NewInt(0).String())

	suite.Equal(sdk.NewCoin(denom, sdkmath.ZeroInt()), suite.app.BankKeeper.GetSupply(suite.ctx, denom))

	_, err = suite.Keeper().RemoveFromOutgoingPoolAndRefund(suite.ctx, txId, sender)
	suite.NoError(err)
	suite.Equal(suite.app.BankKeeper.GetAllBalances(suite.ctx, sender).AmountOf(denom).String(), sendAmount.Amount.String())

	suite.Equal(sendAmount, suite.app.BankKeeper.GetSupply(suite.ctx, denom))
}

func (suite *KeeperTestSuite) TestKeeper_OutgoingPool2() {
	sender := helpers.GenHexAddress().Bytes()
	bridgeToken := helpers.GenHexAddress().Hex()
	denom := fxtypes.DefaultDenom
	supply := suite.app.BankKeeper.GetSupply(suite.ctx, denom)

	sendAmount := sdk.NewCoin(denom, sdkmath.NewInt(int64(tmrand.Uint32()*2)))
	err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, suite.chainName, sender, sdk.NewCoins(sendAmount))
	if suite.chainName != ethtypes.ModuleName {
		suite.Error(err)
		return
	}
	suite.NoError(err)

	suite.Keeper().AddBridgeToken(suite.ctx, bridgeToken, denom)

	suite.Equal(suite.app.BankKeeper.GetAllBalances(suite.ctx, sender).AmountOf(denom).String(), sendAmount.Amount.String())
	receiver := helpers.GenHexAddress().Hex()
	amount := sdk.NewCoin(denom, sendAmount.Amount.QuoRaw(2))
	txId, err := suite.Keeper().AddToOutgoingPool(suite.ctx, sender, receiver, amount, amount)
	suite.NoError(err)
	suite.Equal(suite.app.BankKeeper.GetAllBalances(suite.ctx, sender).AmountOf(denom).String(), sdkmath.NewInt(0).String())

	suite.Equal(supply, suite.app.BankKeeper.GetSupply(suite.ctx, denom))

	_, err = suite.Keeper().RemoveFromOutgoingPoolAndRefund(suite.ctx, txId, sender)
	suite.NoError(err)
	suite.Equal(suite.app.BankKeeper.GetAllBalances(suite.ctx, sender).AmountOf(denom).String(), sendAmount.Amount.String())

	suite.Equal(supply, suite.app.BankKeeper.GetSupply(suite.ctx, denom))
}
