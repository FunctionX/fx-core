package precompile_test

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"

	"github.com/functionx/fx-core/v8/contract"
	testcontract "github.com/functionx/fx-core/v8/tests/contract"
	"github.com/functionx/fx-core/v8/testutil/helpers"
	fxtypes "github.com/functionx/fx-core/v8/types"
	bsctypes "github.com/functionx/fx-core/v8/x/bsc/types"
	crosschainkeeper "github.com/functionx/fx-core/v8/x/crosschain/keeper"
	"github.com/functionx/fx-core/v8/x/crosschain/precompile"
	crosschaintypes "github.com/functionx/fx-core/v8/x/crosschain/types"
	"github.com/functionx/fx-core/v8/x/erc20/types"
	ethtypes "github.com/functionx/fx-core/v8/x/eth/types"
)

func TestCrossChainABI(t *testing.T) {
	crossChain := precompile.NewCrossChainMethod(nil)

	require.Equal(t, 6, len(crossChain.Method.Inputs))
	require.Equal(t, 1, len(crossChain.Method.Outputs))

	require.Equal(t, 8, len(crossChain.Event.Inputs))
}

//nolint:gocyclo
func (suite *PrecompileTestSuite) TestCrossChain() {
	// todo fix this test
	suite.T().SkipNow()
	testCases := []struct {
		name     string
		malleate func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string)
		error    func(args []string) string
		result   bool
	}{
		{
			name: "ok - address",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)

				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), crosschaintypes.GetAddress(), randMint)

				moduleName := md.RandModule()

				fee := big.NewInt(1)
				amount := big.NewInt(0).Sub(randMint, fee)
				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					amount,
					fee,
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, nil
			},
			result: true,
		},
		{
			name: "ok - address - no fee",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)

				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), crosschaintypes.GetAddress(), randMint)

				moduleName := md.RandModule()

				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, nil
			},
			result: true,
		},
		{
			name: "ok - address - origin token",
			malleate: func(_ *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				suite.MintToken(signer.AccAddress(), sdk.NewCoin(fxtypes.DefaultDenom, sdkmath.NewIntFromBigInt(randMint)))

				balance := suite.App.BankKeeper.GetBalance(suite.Ctx, signer.AccAddress(), fxtypes.DefaultDenom)
				suite.Require().Equal(randMint.String(), balance.Amount.BigInt().String())
				moduleName := ethtypes.ModuleName

				suite.AddFXBridgeToken(helpers.GenHexAddress().String())

				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					common.Address{},
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				pair, found := suite.App.Erc20Keeper.GetTokenPair(suite.Ctx, fxtypes.DefaultDenom)
				suite.Require().True(found)

				return data, &pair, randMint, moduleName, nil
			},
			result: true,
		},
		{
			name: "ok - address - origin erc20 token",
			malleate: func(_ *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				moduleName := ethtypes.ModuleName
				denomAddr := helpers.GenHexAddress().String()
				alias := crosschaintypes.NewBridgeDenom(moduleName, denomAddr)

				suite.AddBridgeToken(moduleName, denomAddr)

				token, err := suite.DeployContract(signer.Address())
				suite.Require().NoError(err)

				suite.MintERC20Token(signer, token, signer.Address(), randMint)
				balOf := suite.BalanceOf(token, signer.Address())
				suite.Require().Equal(randMint.String(), balOf.String())

				pair, err := suite.App.Erc20Keeper.RegisterNativeERC20(suite.Ctx, token, alias)
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, token, crosschaintypes.GetAddress(), randMint)

				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, nil
			},
			result: true,
		},
		{
			name: "ok - address - wrapper origin token",
			malleate: func(_ *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				pair, found := suite.App.Erc20Keeper.GetTokenPair(suite.Ctx, fxtypes.DefaultDenom)
				suite.Require().True(found)

				moduleName := ethtypes.ModuleName
				suite.AddFXBridgeToken(helpers.GenHexAddress().String())

				coin := sdk.NewCoin(fxtypes.DefaultDenom, sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)

				balance := suite.App.BankKeeper.GetBalance(suite.Ctx, signer.AccAddress(), fxtypes.DefaultDenom)
				suite.Require().Equal(randMint.String(), balance.Amount.BigInt().String())

				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), crosschaintypes.GetAddress(), randMint)

				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, &pair, big.NewInt(0), moduleName, nil
			},
			result: true,
		},
		// todo: fix this test case
		// {
		//	name: "ok - ibc token",
		//	malleate: func(_ *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
		//		tokenAddress := helpers.GenHexAddress()
		//		bridgeDenom := suite.AddBridgeToken(bsctypes.ModuleName, tokenAddress.Hex())
		//
		//		symbol := helpers.NewRandSymbol()
		//		ibcMD := banktypes.Metadata{
		//			Description: "The cross chain token of the Function X",
		//			DenomUnits: []*banktypes.DenomUnit{
		//				{
		//					Denom:    bridgeDenom,
		//					Exponent: 0,
		//				},
		//				{
		//					Denom:    symbol,
		//					Exponent: 18,
		//				},
		//			},
		//			Base:    bridgeDenom,
		//			Display: bridgeDenom,
		//			Name:    fmt.Sprintf("%s Token", symbol),
		//			Symbol:  symbol,
		//		}
		//		pair, err := suite.App.Erc20Keeper.RegisterNativeCoin(suite.Ctx, ibcMD)
		//		suite.Require().NoError(err)
		//
		//		coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
		//		suite.MintToken(signer.AccAddress(), coin)
		//		_, err = suite.App.Erc20Keeper.ConvertCoin(suite.Ctx,
		//			&types.MsgConvertCoin{Coin: coin, Receiver: signer.Address().Hex(), Sender: signer.AccAddress().String()})
		//		suite.Require().NoError(err)
		//
		//		suite.ERC20Approve(signer, pair.GetERC20Contract(), crosschaintypes.GetAddress(), randMint)
		//
		//		data, err := crosschaintypes.GetABI().Pack(
		//			"crossChain",
		//			pair.GetERC20Contract(),
		//			helpers.GenExternalAddr(bsctypes.ModuleName),
		//			randMint,
		//			big.NewInt(0),
		//			fxtypes.MustStrToByte32(bsctypes.ModuleName),
		//			"",
		//		)
		//		suite.Require().NoError(err)
		//
		//		return data, pair, big.NewInt(0), bsctypes.ModuleName, nil
		//	},
		//	result: true,
		// },
		{
			name: "ok - multiple chain transfer ibc token to outside",
			malleate: func(_ *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				tokenAddress := helpers.GenHexAddress()
				// add to bsc chain
				bridgeDenom := suite.AddBridgeToken(bsctypes.ModuleName, tokenAddress.Hex())
				// add to eth chain
				ethBridgeToken := crosschaintypes.NewBridgeDenom(ethtypes.ModuleName, tokenAddress.Hex())
				suite.AddBridgeToken(ethtypes.ModuleName, tokenAddress.Hex())

				symbol := helpers.NewRandSymbol()
				ibcMD := banktypes.Metadata{
					Description: "The cross chain token of the Function X",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    bridgeDenom,
							Exponent: 0,
							Aliases:  []string{ethBridgeToken},
						},
						{
							Denom:    symbol,
							Exponent: 18,
						},
					},
					Base:    bridgeDenom,
					Display: bridgeDenom,
					Name:    fmt.Sprintf("%s Token", symbol),
					Symbol:  symbol,
				}
				pair, err := suite.App.Erc20Keeper.RegisterNativeCoin(suite.Ctx, ibcMD)
				suite.Require().NoError(err)

				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)

				_, err = suite.App.Erc20Keeper.ConvertCoin(suite.Ctx,
					&types.MsgConvertCoin{Coin: coin, Receiver: signer.Address().Hex(), Sender: signer.AccAddress().String()})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), crosschaintypes.GetAddress(), randMint)

				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(bsctypes.ModuleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(bsctypes.ModuleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), bsctypes.ModuleName, nil
			},
			result: true,
		},
		{
			name: "ok - multiple chain transfer bridge token to outside",
			malleate: func(_ *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				tokenAddress := helpers.GenHexAddress()
				// add to bsc chain
				bridgeDenom := suite.AddBridgeToken(bsctypes.ModuleName, tokenAddress.Hex())
				// add to eth chain
				ethBridgeToken := crosschaintypes.NewBridgeDenom(ethtypes.ModuleName, tokenAddress.Hex())
				suite.AddBridgeToken(ethtypes.ModuleName, tokenAddress.Hex())

				symbol := helpers.NewRandSymbol()
				ibcMD := banktypes.Metadata{
					Description: "The cross chain token of the Function X",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    bridgeDenom,
							Exponent: 0,
							Aliases:  []string{ethBridgeToken},
						},
						{
							Denom:    symbol,
							Exponent: 18,
						},
					},
					Base:    bridgeDenom,
					Display: bridgeDenom,
					Name:    fmt.Sprintf("%s Token", symbol),
					Symbol:  symbol,
				}
				pair, err := suite.App.Erc20Keeper.RegisterNativeCoin(suite.Ctx, ibcMD)
				suite.Require().NoError(err)

				coin := sdk.NewCoin(ethBridgeToken, sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				suite.AddTokenToModule(types.ModuleName, sdk.NewCoins(sdk.NewCoin(ibcMD.Base, sdkmath.NewIntFromBigInt(randMint))))

				targetCoin, err := suite.App.Erc20Keeper.ConvertDenomToTarget(suite.Ctx, signer.AccAddress(), coin, fxtypes.ParseFxTarget(types.ModuleName))
				suite.Require().NoError(err)

				_, err = suite.App.Erc20Keeper.ConvertCoin(suite.Ctx,
					&types.MsgConvertCoin{Coin: targetCoin, Receiver: signer.Address().Hex(), Sender: signer.AccAddress().String()})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), crosschaintypes.GetAddress(), randMint)

				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(ethtypes.ModuleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(ethtypes.ModuleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), ethtypes.ModuleName, nil
			},
			result: true,
		},
		{
			name: "ok - multiple FX transfer outside",
			malleate: func(_ *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				pair, found := suite.App.Erc20Keeper.GetTokenPair(suite.Ctx, fxtypes.DefaultDenom)
				suite.Require().True(found)

				md, found := suite.App.BankKeeper.GetDenomMetaData(suite.Ctx, fxtypes.DefaultDenom)
				suite.Require().True(found)

				tokenContract := helpers.GenExternalAddr(bsctypes.ModuleName)
				newAlias := crosschaintypes.NewBridgeDenom(bsctypes.ModuleName, tokenContract)
				suite.AddBridgeToken(bsctypes.ModuleName, tokenContract)
				update, err := suite.App.Erc20Keeper.UpdateDenomAliases(suite.Ctx, md.Base, newAlias)
				suite.Require().NoError(err)
				suite.Require().True(update)

				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				_, err = suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), crosschaintypes.GetAddress(), randMint)
				moduleName := ethtypes.ModuleName

				fee := big.NewInt(1)
				amount := big.NewInt(0).Sub(randMint, fee)
				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					amount,
					fee,
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, &pair, big.NewInt(0), moduleName, nil
			},
			result: true,
		},
		{
			name: "ok - multiple FX transfer outside other chain",
			malleate: func(_ *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				pair, found := suite.App.Erc20Keeper.GetTokenPair(suite.Ctx, fxtypes.DefaultDenom)
				suite.Require().True(found)

				md, found := suite.App.BankKeeper.GetDenomMetaData(suite.Ctx, fxtypes.DefaultDenom)
				suite.Require().True(found)

				tokenContract := helpers.GenExternalAddr(bsctypes.ModuleName)
				newAlias := crosschaintypes.NewBridgeDenom(bsctypes.ModuleName, tokenContract)
				suite.AddBridgeToken(bsctypes.ModuleName, tokenContract)
				update, err := suite.App.Erc20Keeper.UpdateDenomAliases(suite.Ctx, md.Base, newAlias)
				suite.Require().NoError(err)
				suite.Require().True(update)

				coin := sdk.NewCoin(newAlias, sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				suite.AddTokenToModule(types.ModuleName, sdk.NewCoins(sdk.NewCoin(md.Base, sdkmath.NewIntFromBigInt(randMint))))

				targetCoin, err := suite.App.Erc20Keeper.ConvertDenomToTarget(suite.Ctx, signer.AccAddress(), coin, fxtypes.ParseFxTarget(types.ModuleName))
				suite.Require().NoError(err)

				_, err = suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     targetCoin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), crosschaintypes.GetAddress(), randMint)
				moduleName := bsctypes.ModuleName

				fee := big.NewInt(1)
				amount := big.NewInt(0).Sub(randMint, fee)
				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					amount,
					fee,
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, &pair, big.NewInt(0), moduleName, nil
			},
			result: true,
		},
		{
			name: "failed - msg.value not equal",
			malleate: func(pair *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				suite.MintToken(signer.AccAddress(), sdk.NewCoin(fxtypes.DefaultDenom, sdkmath.NewIntFromBigInt(randMint)))

				moduleName := ethtypes.ModuleName
				suite.AddFXBridgeToken(helpers.GenHexAddress().String())
				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					common.Address{},
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(1),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, randMint, moduleName, nil
			},
			error: func(args []string) string {
				return "amount + fee not equal msg.value"
			},
			result: false,
		},
		{
			name: "failed - token pair not found",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), crosschaintypes.GetAddress(), randMint)

				moduleName := md.RandModule()
				token := helpers.GenHexAddress()
				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					token,
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, []string{token.String()}
			},
			error: func(args []string) string {
				return fmt.Sprintf("token pair not found: %s", args[0])
			},
			result: false,
		},
		{
			name: "failed - not approve",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				moduleName := md.RandModule()
				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, nil
			},
			error: func(args []string) string {
				return "call transferFrom: execution reverted"
			},
			result: false,
		},

		{
			name: "contract - ok",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), suite.crosschain, randMint)
				allowance := suite.ERC20Allowance(pair.GetERC20Contract(), signer.Address(), suite.crosschain)
				suite.Require().Equal(randMint.String(), allowance.String())

				moduleName := md.RandModule()
				fee := big.NewInt(1)
				amount := big.NewInt(0).Sub(randMint, fee)
				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					amount,
					fee,
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, nil
			},
			result: true,
		},
		{
			name: "contract - ok - no fee",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), suite.crosschain, randMint)
				allowance := suite.ERC20Allowance(pair.GetERC20Contract(), signer.Address(), suite.crosschain)
				suite.Require().Equal(randMint.String(), allowance.String())

				moduleName := md.RandModule()
				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, nil
			},
			result: true,
		},
		{
			name: "contract - ok - origin token",
			malleate: func(_ *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				suite.MintToken(signer.AccAddress(), sdk.NewCoin(fxtypes.DefaultDenom, sdkmath.NewIntFromBigInt(randMint)))

				moduleName := ethtypes.ModuleName
				suite.AddFXBridgeToken(helpers.GenHexAddress().String())

				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					common.Address{},
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				pair, found := suite.App.Erc20Keeper.GetTokenPair(suite.Ctx, fxtypes.DefaultDenom)
				suite.Require().True(found)

				return data, &pair, randMint, moduleName, nil
			},
			result: true,
		},
		{
			name: "contract - ok - address - wrapper origin token",
			malleate: func(_ *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				pair, found := suite.App.Erc20Keeper.GetTokenPair(suite.Ctx, fxtypes.DefaultDenom)
				suite.Require().True(found)

				moduleName := ethtypes.ModuleName
				suite.AddFXBridgeToken(helpers.GenHexAddress().String())

				coin := sdk.NewCoin(fxtypes.DefaultDenom, sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)

				balance := suite.App.BankKeeper.GetBalance(suite.Ctx, signer.AccAddress(), fxtypes.DefaultDenom)
				suite.Require().Equal(randMint.String(), balance.Amount.BigInt().String())

				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), suite.crosschain, randMint)

				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, &pair, big.NewInt(0), moduleName, nil
			},
			result: true,
		},
		// todo: fix this case
		// {
		//	name: "contract - ok - ibc token",
		//	malleate: func(_ *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
		//		tokenAddress := helpers.GenHexAddress()
		//		bridgeDenom := suite.AddBridgeToken(bsctypes.ModuleName, tokenAddress.Hex())
		//
		//		symbol := helpers.NewRandSymbol()
		//		ibcMD := banktypes.Metadata{
		//			Description: "The cross chain token of the Function X",
		//			DenomUnits: []*banktypes.DenomUnit{
		//				{
		//					Denom:    bridgeDenom,
		//					Exponent: 0,
		//				},
		//				{
		//					Denom:    symbol,
		//					Exponent: 18,
		//				},
		//			},
		//			Base:    bridgeDenom,
		//			Display: bridgeDenom,
		//			Name:    fmt.Sprintf("%s Token", symbol),
		//			Symbol:  symbol,
		//		}
		//		pair, err := suite.App.Erc20Keeper.RegisterNativeCoin(suite.Ctx, ibcMD)
		//		suite.Require().NoError(err)
		//
		//		coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
		//		suite.MintToken(signer.AccAddress(), coin)
		//		_, err = suite.App.Erc20Keeper.ConvertCoin(suite.Ctx,
		//			&types.MsgConvertCoin{Coin: coin, Receiver: signer.Address().Hex(), Sender: signer.AccAddress().String()})
		//		suite.Require().NoError(err)
		//
		//		suite.ERC20Approve(signer, pair.GetERC20Contract(), suite.crosschain, randMint)
		//
		//		data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
		//			"crossChain",
		//			pair.GetERC20Contract(),
		//			helpers.GenExternalAddr(bsctypes.ModuleName),
		//			randMint,
		//			big.NewInt(0),
		//			fxtypes.MustStrToByte32(bsctypes.ModuleName),
		//			"",
		//		)
		//		suite.Require().NoError(err)
		//
		//		return data, pair, big.NewInt(0), bsctypes.ModuleName, nil
		//	},
		//	result: true,
		// },
		{
			name: "contract - failed - msg.value not equal amount",
			malleate: func(pair *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				suite.MintToken(signer.AccAddress(), sdk.NewCoin(fxtypes.DefaultDenom, sdkmath.NewIntFromBigInt(randMint)))

				moduleName := ethtypes.ModuleName
				suite.AddFXBridgeToken(helpers.GenHexAddress().String())
				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					common.Address{},
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(1),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, randMint, moduleName, nil
			},
			error: func(args []string) string {
				return "msg.value not equal amount + fee"
			},
			result: false,
		},
		{
			name: "contract - failed - token pair not found",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				erc20Token, err := suite.DeployContract(signer.Address())
				suite.Require().NoError(err)
				suite.MintERC20Token(signer, erc20Token, signer.Address(), randMint)
				suite.ERC20Approve(signer, erc20Token, suite.crosschain, randMint)

				moduleName := md.RandModule()
				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					erc20Token,
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, []string{erc20Token.String()}
			},
			error: func(args []string) string {
				return fmt.Sprintf("token pair not found: %s", args[0])
			},
			result: false,
		},
		{
			name: "contract - failed - not approve",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				moduleName := md.RandModule()
				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, nil
			},
			error: func(args []string) string {
				return "transfer amount exceeds allowance"
			},
			result: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			signer := suite.RandSigner()
			// token pair
			md := suite.GenerateCrossChainDenoms()
			pair, err := suite.App.Erc20Keeper.RegisterNativeCoin(suite.Ctx, md.GetMetadata())
			suite.Require().NoError(err)
			randMint := big.NewInt(int64(tmrand.Uint32() + 10))
			suite.MintLockNativeTokenToModule(md.GetMetadata(), sdkmath.NewIntFromBigInt(randMint))

			chainBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, signer.AccAddress())
			suite.Require().True(chainBalances.IsZero(), chainBalances.String())
			balance := suite.BalanceOf(pair.GetERC20Contract(), signer.Address())
			suite.Require().True(balance.Cmp(big.NewInt(0)) == 0, balance.String())

			packData, newPair, value, moduleName, errArgs := tc.malleate(pair, md, signer, randMint)

			contractAddr := crosschaintypes.GetAddress()
			addrQuery := signer.Address()
			if strings.HasPrefix(tc.name, "contract") {
				contractAddr = suite.crosschain
				addrQuery = suite.crosschain
			}

			queryServer := crosschainkeeper.NewQueryServerImpl(suite.CrossChainKeepers()[moduleName])
			resp, err := queryServer.GetPendingSendToExternal(suite.Ctx,
				&crosschaintypes.QueryPendingSendToExternalRequest{
					ChainName:     moduleName,
					SenderAddress: sdk.AccAddress(addrQuery.Bytes()).String(),
				})
			suite.Require().NoError(err)
			suite.Require().Equal(0, len(resp.UnbatchedTransfers))
			suite.Require().Equal(0, len(resp.TransfersInBatches))

			totalBefore, err := suite.App.BankKeeper.TotalSupply(suite.Ctx, &banktypes.QueryTotalSupplyRequest{})
			suite.Require().NoError(err)

			res := suite.EthereumTx(signer, contractAddr, value, packData)

			if tc.result {
				suite.Require().False(res.Failed(), res.VmError)
				// signer balance zero
				chainBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, sdk.AccAddress(addrQuery.Bytes()))
				suite.Require().True(chainBalances.IsZero(), chainBalances.String())
				balance := suite.BalanceOf(newPair.GetERC20Contract(), addrQuery)
				suite.Require().True(balance.Cmp(big.NewInt(0)) == 0, balance.String())

				manyToOne := make(map[string]bool)
				suite.App.BankKeeper.IterateAllDenomMetaData(suite.Ctx, func(md banktypes.Metadata) bool {
					if len(md.DenomUnits) > 0 && len(md.DenomUnits[0].Aliases) > 0 {
						manyToOne[md.Base] = true
					}
					return false
				})
				totalAfter, err := suite.App.BankKeeper.TotalSupply(suite.Ctx, &banktypes.QueryTotalSupplyRequest{})
				suite.Require().NoError(err)

				newMD, found := suite.App.BankKeeper.GetDenomMetaData(suite.Ctx, newPair.GetDenom())
				suite.Require().True(found)

				for _, coin := range totalBefore.Supply {
					if manyToOne[coin.Denom] {
						continue
					}
					expect := totalAfter.Supply.AmountOf(coin.Denom)

					has := false
					if len(newMD.DenomUnits) > 0 && len(newMD.DenomUnits[0].Aliases) > 0 {
						for _, alias := range newMD.DenomUnits[0].Aliases {
							if strings.HasPrefix(alias, moduleName) &&
								alias == coin.GetDenom() && !suite.ConvertOneToManyToken(newMD) {
								has = true
								break
							}
						}
					}
					if has || strings.HasPrefix(coin.GetDenom(), "ibc/") {
						expect = expect.Add(sdkmath.NewIntFromBigInt(randMint))
					}

					if suite.ConvertOneToManyToken(newMD) &&
						slices.Contains(newMD.DenomUnits[0].Aliases, coin.Denom) {
						coin.Amount = coin.Amount.Sub(sdkmath.NewIntFromBigInt(randMint))
					}

					suite.Require().Equal(coin.Amount.String(), expect.String(), coin.Denom, randMint.String())
				}

				// pending send to external
				resp, err := queryServer.GetPendingSendToExternal(suite.Ctx,
					&crosschaintypes.QueryPendingSendToExternalRequest{
						ChainName:     moduleName,
						SenderAddress: sdk.AccAddress(addrQuery.Bytes()).String(),
					})
				suite.Require().NoError(err)
				suite.Require().Equal(1, len(resp.UnbatchedTransfers))
				suite.Require().Equal(0, len(resp.TransfersInBatches))
				suite.Require().Equal(sdk.AccAddress(addrQuery.Bytes()).String(), resp.UnbatchedTransfers[0].Sender)
				// NOTE: fee + amount == randMint
				suite.Require().Equal(randMint.String(), resp.UnbatchedTransfers[0].Fee.Amount.Add(resp.UnbatchedTransfers[0].Token.Amount).BigInt().String())

				if !strings.EqualFold(resp.UnbatchedTransfers[0].Token.Contract, strings.TrimPrefix(Metadata{metadata: newMD}.GetDenom(moduleName), moduleName)) {
					tokenContract, found := suite.CrossChainKeepers()[moduleName].GetContractByBridgeDenom(suite.Ctx, newPair.Denom)
					suite.Require().True(found)
					suite.Require().Equal(resp.UnbatchedTransfers[0].Token.Contract, tokenContract, moduleName)
				}

				for _, log := range res.Logs {
					event := crosschaintypes.GetABI().Events["IncreaseBridgeFee"]
					if log.Topics[0] == event.ID.String() {
						suite.Require().Equal(3, len(log.Topics))
						suite.Require().Equal(log.Address, crosschaintypes.GetAddress().String())
						suite.Require().Equal(log.Topics[1], addrQuery.Hash().String())

						unpack, err := event.Inputs.NonIndexed().Unpack(log.Data)
						suite.Require().NoError(err)
						denom := unpack[0].(string)

						if moduleName == ethtypes.ModuleName && value.Cmp(big.NewInt(0)) == 1 {
							suite.Require().Equal(log.Topics[2], common.Address{}.Hash().String())
							suite.Require().Equal(fxtypes.DefaultDenom, denom)
						} else {
							suite.Require().Equal(log.Topics[2], newPair.GetERC20Contract().Hash().String())
							suite.Require().Equal(denom, newPair.GetDenom())
						}

						amount := unpack[2].(*big.Int)
						fee := unpack[3].(*big.Int)
						suite.Require().Equal(randMint.String(), big.NewInt(0).Add(amount, fee).String())
						target := unpack[4].([32]byte)
						suite.Require().Equal(moduleName, fxtypes.Byte32ToString(target))
						memo := unpack[5].(string)
						suite.Require().Equal("", memo)
					}
				}

				if value.Cmp(big.NewInt(0)) == 0 {
					relation := suite.App.Erc20Keeper.HasOutgoingTransferRelation(suite.Ctx, moduleName, resp.UnbatchedTransfers[0].Id)
					suite.Require().True(relation)
				}
			} else {
				suite.Error(res, errors.New(tc.error(errArgs)))
			}
		})
	}
}

func (suite *PrecompileTestSuite) TestCrossChainExternal() {
	testCases := []struct {
		name     string
		malleate func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string)
		error    func(args []string) string
		result   bool
	}{
		{
			name: "ok - address",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				suite.MintERC20Token(signer, pair.GetERC20Contract(), suite.App.Erc20Keeper.ModuleAddress(), randMint)

				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), crosschaintypes.GetAddress(), randMint)

				moduleName := md.RandModule()

				fee := big.NewInt(1)
				amount := big.NewInt(0).Sub(randMint, fee)
				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					amount,
					fee,
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, nil
			},
			result: true,
		},
		{
			name: "ok - address - no fee",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				suite.MintERC20Token(signer, pair.GetERC20Contract(), suite.App.Erc20Keeper.ModuleAddress(), randMint)

				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), crosschaintypes.GetAddress(), randMint)

				moduleName := md.RandModule()

				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, nil
			},
			result: true,
		},

		{
			name: "failed - token pair not found",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				suite.MintERC20Token(signer, pair.GetERC20Contract(), suite.App.Erc20Keeper.ModuleAddress(), randMint)
				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), crosschaintypes.GetAddress(), randMint)

				moduleName := md.RandModule()
				token := helpers.GenHexAddress()
				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					token,
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, []string{token.String()}
			},
			error: func(args []string) string {
				return fmt.Sprintf("token pair not found: %s", args[0])
			},
			result: false,
		},
		{
			name: "failed - not approve",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				suite.MintERC20Token(signer, pair.GetERC20Contract(), suite.App.Erc20Keeper.ModuleAddress(), randMint)
				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				moduleName := md.RandModule()
				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, nil
			},
			error: func(args []string) string {
				return "call transferFrom: execution reverted"
			},
			result: false,
		},

		{
			name: "contract - ok",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				suite.MintERC20Token(signer, pair.GetERC20Contract(), suite.App.Erc20Keeper.ModuleAddress(), randMint)
				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), suite.crosschain, randMint)
				allowance := suite.ERC20Allowance(pair.GetERC20Contract(), signer.Address(), suite.crosschain)
				suite.Require().Equal(randMint.String(), allowance.String())

				moduleName := md.RandModule()
				fee := big.NewInt(1)
				amount := big.NewInt(0).Sub(randMint, fee)
				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					amount,
					fee,
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, nil
			},
			result: true,
		},
		{
			name: "contract - ok - no fee",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				suite.MintERC20Token(signer, pair.GetERC20Contract(), suite.App.Erc20Keeper.ModuleAddress(), randMint)
				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), suite.crosschain, randMint)
				allowance := suite.ERC20Allowance(pair.GetERC20Contract(), signer.Address(), suite.crosschain)
				suite.Require().Equal(randMint.String(), allowance.String())

				moduleName := md.RandModule()
				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, nil
			},
			result: true,
		},

		{
			name: "contract - failed - token pair not found",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				erc20Token, err := suite.DeployContract(signer.Address())
				suite.Require().NoError(err)
				suite.MintERC20Token(signer, erc20Token, signer.Address(), randMint)
				suite.ERC20Approve(signer, erc20Token, suite.crosschain, randMint)

				moduleName := md.RandModule()
				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					erc20Token,
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, []string{erc20Token.String()}
			},
			error: func(args []string) string {
				return fmt.Sprintf("token pair not found: %s", args[0])
			},
			result: false,
		},
		{
			name: "contract - failed - not approve",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *types.TokenPair, *big.Int, string, []string) {
				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				suite.MintERC20Token(signer, pair.GetERC20Contract(), suite.App.Erc20Keeper.ModuleAddress(), randMint)
				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx, &types.MsgConvertCoin{
					Coin:     coin,
					Receiver: signer.Address().Hex(),
					Sender:   signer.AccAddress().String(),
				})
				suite.Require().NoError(err)

				moduleName := md.RandModule()
				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					pair.GetERC20Contract(),
					helpers.GenExternalAddr(moduleName),
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(moduleName),
					"",
				)
				suite.Require().NoError(err)

				return data, pair, big.NewInt(0), moduleName, nil
			},
			error: func(args []string) string {
				return "transfer amount exceeds allowance"
			},
			result: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			signer := suite.RandSigner()
			// token pair
			md := suite.GenerateCrossChainDenoms()

			// deploy fip20 external
			fip20External, err := suite.App.Erc20Keeper.DeployUpgradableToken(suite.Ctx, signer.Address(), "Test token", "TEST", 18)
			suite.Require().NoError(err)
			// token pair
			pair, err := suite.App.Erc20Keeper.RegisterNativeERC20(suite.Ctx, fip20External, md.GetMetadata().DenomUnits[0].Aliases...)
			suite.Require().NoError(err)

			randMint := big.NewInt(int64(tmrand.Uint32() + 10))
			suite.MintLockNativeTokenToModule(md.GetMetadata(), sdkmath.NewIntFromBigInt(randMint))

			chainBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, signer.AccAddress())
			suite.Require().True(chainBalances.IsZero(), chainBalances.String())
			balance := suite.BalanceOf(pair.GetERC20Contract(), signer.Address())
			suite.Require().True(balance.Cmp(big.NewInt(0)) == 0, balance.String())

			packData, newPair, value, moduleName, errArgs := tc.malleate(pair, md, signer, randMint)

			contractAddr := crosschaintypes.GetAddress()
			addrQuery := signer.Address()
			if strings.HasPrefix(tc.name, "contract") {
				contractAddr = suite.crosschain
				addrQuery = suite.crosschain
			}

			queryServer := crosschainkeeper.NewQueryServerImpl(suite.CrossChainKeepers()[moduleName])
			resp, err := queryServer.GetPendingSendToExternal(suite.Ctx,
				&crosschaintypes.QueryPendingSendToExternalRequest{
					ChainName:     moduleName,
					SenderAddress: sdk.AccAddress(addrQuery.Bytes()).String(),
				})
			suite.Require().NoError(err)
			suite.Require().Equal(0, len(resp.UnbatchedTransfers))
			suite.Require().Equal(0, len(resp.TransfersInBatches))

			totalBefore, err := suite.App.BankKeeper.TotalSupply(suite.Ctx, &banktypes.QueryTotalSupplyRequest{})
			suite.Require().NoError(err)

			res := suite.EthereumTx(signer, contractAddr, value, packData)

			if tc.result {
				suite.Require().False(res.Failed(), res.VmError)
				// signer balance zero
				chainBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, sdk.AccAddress(addrQuery.Bytes()))
				suite.Require().True(chainBalances.IsZero(), chainBalances.String())
				balance := suite.BalanceOf(newPair.GetERC20Contract(), addrQuery)
				suite.Require().True(balance.Cmp(big.NewInt(0)) == 0, balance.String())

				manyToOne := make(map[string]bool)
				suite.App.BankKeeper.IterateAllDenomMetaData(suite.Ctx, func(md banktypes.Metadata) bool {
					if len(md.DenomUnits) > 0 && len(md.DenomUnits[0].Aliases) > 0 {
						manyToOne[md.Base] = true
					}
					return false
				})
				totalAfter, err := suite.App.BankKeeper.TotalSupply(suite.Ctx, &banktypes.QueryTotalSupplyRequest{})
				suite.Require().NoError(err)

				for _, coin := range totalAfter.Supply {
					if manyToOne[coin.Denom] {
						continue
					}
					expect := totalBefore.Supply.AmountOf(coin.Denom)
					if coin.Denom == md.GetDenom(moduleName) {
						suite.Require().Equal(coin.Amount.String(), expect.Add(sdkmath.NewIntFromBigInt(randMint)).String(), coin.Denom)
					} else {
						suite.Require().Equal(coin.Amount.String(), expect.String(), coin.Denom)
					}
				}

				// pending send to external
				resp, err := queryServer.GetPendingSendToExternal(suite.Ctx,
					&crosschaintypes.QueryPendingSendToExternalRequest{
						ChainName:     moduleName,
						SenderAddress: sdk.AccAddress(addrQuery.Bytes()).String(),
					})
				suite.Require().NoError(err)
				suite.Require().Equal(1, len(resp.UnbatchedTransfers))
				suite.Require().Equal(0, len(resp.TransfersInBatches))
				suite.Require().Equal(sdk.AccAddress(addrQuery.Bytes()).String(), resp.UnbatchedTransfers[0].Sender)
				// NOTE: fee + amount == randMint
				suite.Require().Equal(randMint.String(), resp.UnbatchedTransfers[0].Fee.Amount.Add(resp.UnbatchedTransfers[0].Token.Amount).BigInt().String())

				if !strings.EqualFold(resp.UnbatchedTransfers[0].Token.Contract, strings.TrimPrefix(md.GetDenom(moduleName), moduleName)) {
					tokenContract, found := suite.CrossChainKeepers()[moduleName].GetContractByBridgeDenom(suite.Ctx, newPair.Denom)
					suite.Require().True(found)
					suite.Require().Equal(resp.UnbatchedTransfers[0].Token.Contract, tokenContract, moduleName)
				}

				for _, log := range res.Logs {
					event := crosschaintypes.GetABI().Events["IncreaseBridgeFee"]
					if log.Topics[0] == event.ID.String() {
						suite.Require().Equal(3, len(log.Topics))
						suite.Require().Equal(log.Address, crosschaintypes.GetAddress().String())
						suite.Require().Equal(log.Topics[1], addrQuery.Hash().String())

						unpack, err := event.Inputs.NonIndexed().Unpack(log.Data)
						suite.Require().NoError(err)
						denom := unpack[0].(string)

						if moduleName == ethtypes.ModuleName && value.Cmp(big.NewInt(0)) == 1 {
							suite.Require().Equal(log.Topics[2], common.Address{}.Hash().String())
							suite.Require().Equal(fxtypes.DefaultDenom, denom)
						} else {
							suite.Require().Equal(log.Topics[2], newPair.GetERC20Contract().Hash().String())
							suite.Require().Equal(denom, newPair.GetDenom())
						}

						amount := unpack[2].(*big.Int)
						fee := unpack[3].(*big.Int)
						suite.Require().Equal(randMint.String(), big.NewInt(0).Add(amount, fee).String())
						target := unpack[4].([32]byte)
						suite.Require().Equal(moduleName, fxtypes.Byte32ToString(target))
						memo := unpack[5].(string)
						suite.Require().Equal("", memo)
					}
				}
			} else {
				suite.Error(res, errors.New(tc.error(errArgs)))
			}
		})
	}
}

//nolint:gocyclo
func (suite *PrecompileTestSuite) TestCrossChainIBC() {
	testCases := []struct {
		name     string
		malleate func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *big.Int, string, string, []string)
		error    func(args []string) string
		result   bool
	}{
		{
			name: "ok - origin token",
			malleate: func(_ *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *big.Int, string, string, []string) {
				suite.MintToken(signer.AccAddress(), sdk.NewCoin(fxtypes.DefaultDenom, sdkmath.NewIntFromBigInt(randMint)))

				sourcePort, sourceChannel := suite.RandTransferChannel()

				prefix, recipient := suite.RandPrefixAndAddress()
				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					common.Address{},
					recipient,
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(fmt.Sprintf("%s/%s/%s", prefix, sourcePort, sourceChannel)),
					"ibc memo",
				)
				suite.Require().NoError(err)

				return data, randMint, sourcePort, sourceChannel, nil
			},
			result: true,
		},
		{
			name: "failed - not zero fee",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *big.Int, string, string, []string) {
				sourcePort, sourceChannel := suite.RandTransferChannel()

				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx,
					&types.MsgConvertCoin{Coin: coin, Receiver: signer.Address().Hex(), Sender: signer.AccAddress().String()})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), crosschaintypes.GetAddress(), randMint)

				prefix, recipient := suite.RandPrefixAndAddress()

				fee := big.NewInt(1)
				amount := big.NewInt(0).Sub(randMint, fee)
				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					pair.GetERC20Contract(),
					recipient,
					amount,
					fee,
					fxtypes.MustStrToByte32(fmt.Sprintf("%s/%s/%s", prefix, sourcePort, sourceChannel)),
					"ibc memo",
				)
				suite.Require().NoError(err)

				return data, big.NewInt(0), sourcePort, sourceChannel, []string{sdk.Coin{Amount: sdkmath.NewIntFromBigInt(fee), Denom: md.metadata.Base}.String()}
			},
			error: func(args []string) string {
				return fmt.Sprintf("ibc transfer fee must be zero: %s", args[0])
			},
			result: false,
		},
		{
			name: "failed - not zero fee - ibc denom",
			malleate: func(_ *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *big.Int, string, string, []string) {
				sourcePort, sourceChannel := suite.RandTransferChannel()
				tokenAddress := helpers.GenHexAddress()
				denom, err := suite.CrossChainKeepers()[bsctypes.ModuleName].SetIbcDenomTrace(suite.Ctx,
					tokenAddress.Hex(), hex.EncodeToString([]byte(fmt.Sprintf("%s/%s", sourcePort, sourceChannel))))
				suite.Require().NoError(err)
				suite.AddBridgeToken(bsctypes.ModuleName, tokenAddress.Hex())

				symbol := helpers.NewRandSymbol()
				ibcMD := banktypes.Metadata{
					Description: "The cross chain token of the Function X",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
						},
						{
							Denom:    symbol,
							Exponent: 18,
						},
					},
					Base:    denom,
					Display: denom,
					Name:    fmt.Sprintf("%s Token", symbol),
					Symbol:  symbol,
				}
				pair, err := suite.App.Erc20Keeper.RegisterNativeCoin(suite.Ctx, ibcMD)
				suite.Require().NoError(err)

				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				_, err = suite.App.Erc20Keeper.ConvertCoin(suite.Ctx,
					&types.MsgConvertCoin{Coin: coin, Receiver: signer.Address().Hex(), Sender: signer.AccAddress().String()})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), crosschaintypes.GetAddress(), randMint)

				prefix, recipient := suite.RandPrefixAndAddress()

				fee := big.NewInt(1)
				amount := big.NewInt(0).Sub(randMint, fee)
				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					pair.GetERC20Contract(),
					recipient,
					amount,
					fee,
					fxtypes.MustStrToByte32(fmt.Sprintf("%s/%s/%s", prefix, sourcePort, sourceChannel)),
					"ibc memo",
				)
				suite.Require().NoError(err)

				return data, big.NewInt(0), sourcePort, sourceChannel, []string{sdk.Coin{Amount: sdkmath.NewIntFromBigInt(fee), Denom: denom}.String()}
			},
			error: func(args []string) string {
				return fmt.Sprintf("ibc transfer fee must be zero: %s", args[0])
			},
			result: false,
		},
		{
			name: "failed - not zero fee - origin token",
			malleate: func(_ *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *big.Int, string, string, []string) {
				suite.MintToken(signer.AccAddress(), sdk.NewCoin(fxtypes.DefaultDenom, sdkmath.NewIntFromBigInt(randMint)))

				sourcePort, sourceChannel := suite.RandTransferChannel()

				prefix, recipient := suite.RandPrefixAndAddress()
				fee := big.NewInt(1)
				amount := big.NewInt(0).Sub(randMint, fee)
				data, err := crosschaintypes.GetABI().Pack(
					"crossChain",
					common.Address{},
					recipient,
					amount,
					fee,
					fxtypes.MustStrToByte32(fmt.Sprintf("%s/%s/%s", prefix, sourcePort, sourceChannel)),
					"ibc memo",
				)
				suite.Require().NoError(err)

				return data, randMint, sourcePort, sourceChannel, []string{sdk.Coin{Amount: sdkmath.NewIntFromBigInt(fee), Denom: fxtypes.DefaultDenom}.String()}
			},
			error: func(args []string) string {
				return fmt.Sprintf("ibc transfer fee must be zero: %s", args[0])
			},
			result: false,
		},
		{
			name: "contract - ok - origin token",
			malleate: func(_ *types.TokenPair, _ Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *big.Int, string, string, []string) {
				suite.MintToken(signer.AccAddress(), sdk.NewCoin(fxtypes.DefaultDenom, sdkmath.NewIntFromBigInt(randMint)))

				sourcePort, sourceChannel := suite.RandTransferChannel()

				prefix, recipient := suite.RandPrefixAndAddress()
				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					common.Address{},
					recipient,
					randMint,
					big.NewInt(0),
					fxtypes.MustStrToByte32(fmt.Sprintf("%s/%s/%s", prefix, sourcePort, sourceChannel)),
					"ibc memo",
				)
				suite.Require().NoError(err)

				return data, randMint, sourcePort, sourceChannel, nil
			},
			result: true,
		},
		{
			name: "contract - failed - not zero fee",
			malleate: func(pair *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *big.Int, string, string, []string) {
				sourcePort, sourceChannel := suite.RandTransferChannel()

				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				_, err := suite.App.Erc20Keeper.ConvertCoin(suite.Ctx,
					&types.MsgConvertCoin{Coin: coin, Receiver: signer.Address().Hex(), Sender: signer.AccAddress().String()})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), suite.crosschain, randMint)

				prefix, recipient := suite.RandPrefixAndAddress()
				fee := big.NewInt(1)
				amount := big.NewInt(0).Sub(randMint, fee)
				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					pair.GetERC20Contract(),
					recipient,
					amount,
					fee,
					fxtypes.MustStrToByte32(fmt.Sprintf("%s/%s/%s", prefix, sourcePort, sourceChannel)),
					"ibc memo",
				)
				suite.Require().NoError(err)

				return data, big.NewInt(0), sourcePort, sourceChannel, []string{sdk.Coin{Amount: sdkmath.NewIntFromBigInt(fee), Denom: md.metadata.Base}.String()}
			},
			error: func(args []string) string {
				return fmt.Sprintf("ibc transfer fee must be zero: %s", args[0])
			},
			result: false,
		},
		{
			name: "contract - failed - not zero fee - ibc denom",
			malleate: func(_ *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *big.Int, string, string, []string) {
				sourcePort, sourceChannel := suite.RandTransferChannel()
				tokenAddress := helpers.GenHexAddress()
				denom, err := suite.CrossChainKeepers()[bsctypes.ModuleName].SetIbcDenomTrace(suite.Ctx,
					tokenAddress.Hex(), hex.EncodeToString([]byte(fmt.Sprintf("%s/%s", sourcePort, sourceChannel))))
				suite.Require().NoError(err)
				suite.AddBridgeToken(bsctypes.ModuleName, tokenAddress.Hex())

				symbol := helpers.NewRandSymbol()
				ibcMD := banktypes.Metadata{
					Description: "The cross chain token of the Function X",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
						},
						{
							Denom:    symbol,
							Exponent: 18,
						},
					},
					Base:    denom,
					Display: denom,
					Name:    fmt.Sprintf("%s Token", symbol),
					Symbol:  symbol,
				}
				pair, err := suite.App.Erc20Keeper.RegisterNativeCoin(suite.Ctx, ibcMD)
				suite.Require().NoError(err)

				coin := sdk.NewCoin(pair.GetDenom(), sdkmath.NewIntFromBigInt(randMint))
				suite.MintToken(signer.AccAddress(), coin)
				_, err = suite.App.Erc20Keeper.ConvertCoin(suite.Ctx,
					&types.MsgConvertCoin{Coin: coin, Receiver: signer.Address().Hex(), Sender: signer.AccAddress().String()})
				suite.Require().NoError(err)

				suite.ERC20Approve(signer, pair.GetERC20Contract(), suite.crosschain, randMint)

				prefix, recipient := suite.RandPrefixAndAddress()

				fee := big.NewInt(1)
				amount := big.NewInt(0).Sub(randMint, fee)
				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					pair.GetERC20Contract(),
					recipient,
					amount,
					fee,
					fxtypes.MustStrToByte32(fmt.Sprintf("%s/%s/%s", prefix, sourcePort, sourceChannel)),
					"ibc memo",
				)
				suite.Require().NoError(err)

				return data, big.NewInt(0), sourcePort, sourceChannel, []string{sdk.Coin{Amount: sdkmath.NewIntFromBigInt(fee), Denom: denom}.String()}
			},
			error: func(args []string) string {
				return fmt.Sprintf("ibc transfer fee must be zero: %s", args[0])
			},
			result: false,
		},
		{
			name: "contract - failed - not zero fee - origin token",
			malleate: func(_ *types.TokenPair, md Metadata, signer *helpers.Signer, randMint *big.Int) ([]byte, *big.Int, string, string, []string) {
				suite.MintToken(signer.AccAddress(), sdk.NewCoin(fxtypes.DefaultDenom, sdkmath.NewIntFromBigInt(randMint)))

				sourcePort, sourceChannel := suite.RandTransferChannel()

				prefix, recipient := suite.RandPrefixAndAddress()
				fee := big.NewInt(1)
				amount := big.NewInt(0).Sub(randMint, fee)
				data, err := contract.MustABIJson(testcontract.CrossChainTestMetaData.ABI).Pack(
					"crossChain",
					common.Address{},
					recipient,
					amount,
					fee,
					fxtypes.MustStrToByte32(fmt.Sprintf("%s/%s/%s", prefix, sourcePort, sourceChannel)),
					"ibc memo",
				)
				suite.Require().NoError(err)

				return data, randMint, sourcePort, sourceChannel, []string{sdk.Coin{Amount: sdkmath.NewIntFromBigInt(fee), Denom: fxtypes.DefaultDenom}.String()}
			},
			error: func(args []string) string {
				return fmt.Sprintf("ibc transfer fee must be zero: %s", args[0])
			},
			result: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			signer := suite.RandSigner()
			// token pair
			md := suite.GenerateCrossChainDenoms()
			pair, err := suite.App.Erc20Keeper.RegisterNativeCoin(suite.Ctx, md.GetMetadata())
			suite.Require().NoError(err)
			randMint := big.NewInt(int64(tmrand.Uint32() + 10))
			suite.MintLockNativeTokenToModule(md.GetMetadata(), sdkmath.NewIntFromBigInt(randMint))

			chainBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, signer.AccAddress())
			suite.Require().True(chainBalances.IsZero(), chainBalances.String())
			balance := suite.BalanceOf(pair.GetERC20Contract(), signer.Address())
			suite.Require().True(balance.Cmp(big.NewInt(0)) == 0, balance.String())

			packData, value, portId, channelId, errArgs := tc.malleate(pair, md, signer, randMint)

			crosschainContract := crosschaintypes.GetAddress()
			addrQuery := signer.Address()
			if strings.HasPrefix(tc.name, "contract") {
				crosschainContract = suite.crosschain
				addrQuery = suite.crosschain
			}

			commitments := suite.App.IBCKeeper.ChannelKeeper.GetAllPacketCommitmentsAtChannel(suite.Ctx, portId, channelId)
			ibcTxs := make(map[string]bool, len(commitments))
			for _, commitment := range commitments {
				ibcTxs[fmt.Sprintf("%s/%s/%d", commitment.PortId, commitment.ChannelId, commitment.Sequence)] = true
			}

			totalBefore, err := suite.App.BankKeeper.TotalSupply(suite.Ctx, &banktypes.QueryTotalSupplyRequest{})
			suite.Require().NoError(err)

			res := suite.EthereumTx(signer, crosschainContract, value, packData)

			if tc.result {
				suite.Require().False(res.Failed(), res.VmError)

				chainBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, sdk.AccAddress(addrQuery.Bytes()))
				suite.Require().True(chainBalances.IsZero(), chainBalances.String())
				balance := suite.BalanceOf(pair.GetERC20Contract(), addrQuery)
				suite.Require().True(balance.Cmp(big.NewInt(0)) == 0, balance.String())

				manyToOne := make(map[string]bool)
				suite.App.BankKeeper.IterateAllDenomMetaData(suite.Ctx, func(md banktypes.Metadata) bool {
					if len(md.DenomUnits) > 0 && len(md.DenomUnits[0].Aliases) > 0 {
						manyToOne[md.Base] = true
					}
					return false
				})
				totalAfter, err := suite.App.BankKeeper.TotalSupply(suite.Ctx, &banktypes.QueryTotalSupplyRequest{})
				suite.Require().NoError(err)

				for _, coin := range totalBefore.Supply {
					if manyToOne[coin.Denom] {
						continue
					}
					expect := totalAfter.Supply.AmountOf(coin.Denom)
					if strings.HasPrefix(coin.GetDenom(), "ibc/") {
						expect = expect.Add(sdkmath.NewIntFromBigInt(randMint))
					}
					suite.Require().Equal(coin.Amount.String(), expect.String(), coin.Denom)
				}

				for _, event := range suite.Ctx.EventManager().Events() {
					if event.Type != ibcchanneltypes.EventTypeSendPacket {
						continue
					}
					var eventPortId, eventChannelId string
					var sequence string
					var data []byte

					for _, attr := range event.Attributes {
						attrKey, attrValue := attr.Key, attr.Value
						if attrKey == ibcchanneltypes.AttributeKeyDataHex {
							data, err = hex.DecodeString(attrValue)
							suite.Require().NoError(err)
						}
						if attrKey == ibcchanneltypes.AttributeKeySequence {
							sequence = attrValue
						}
						if attrKey == ibcchanneltypes.AttributeKeySrcPort {
							eventPortId = attrValue
						}
						if attrKey == ibcchanneltypes.AttributeKeySrcChannel {
							eventChannelId = attrValue
						}
					}
					if eventPortId != portId || eventChannelId != channelId {
						continue
					}
					txKey := fmt.Sprintf("%s/%s/%s", portId, channelId, sequence)
					if ibcTxs[txKey] {
						continue
					}
					var packet ibctransfertypes.FungibleTokenPacketData
					err = suite.App.LegacyAmino().UnmarshalJSON(data, &packet)
					suite.Require().NoError(err)
					suite.Require().Equal(sdk.AccAddress(addrQuery.Bytes()).String(), packet.Sender)
					suite.Require().Equal(randMint.String(), packet.Amount)
				}
			} else {
				suite.Error(res, errors.New(tc.error(errArgs)))
			}
		})
	}
}
