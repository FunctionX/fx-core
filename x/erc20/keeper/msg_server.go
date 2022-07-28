package keeper

import (
	"context"
	"math/big"
	"strings"

	gravitytypes "github.com/functionx/fx-core/v2/x/gravity/types"

	fxtypes "github.com/functionx/fx-core/v2/types"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/functionx/fx-core/v2/x/erc20/types"
)

var _ types.MsgServer = &Keeper{}

// ConvertCoin converts Cosmos-native Coins into ERC20 tokens for both
// Cosmos-native and ERC20 TokenPair Owners
func (k Keeper) ConvertCoin(goCtx context.Context, msg *types.MsgConvertCoin) (*types.MsgConvertCoinResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Error checked during msg validation
	receiver := common.HexToAddress(msg.Receiver)
	sender := sdk.MustAccAddressFromBech32(msg.Sender)

	pair, err := k.MintingEnabled(ctx, sender, receiver.Bytes(), msg.Coin.Denom)
	if err != nil {
		return nil, err
	}

	// Remove token pair if contract is suicided
	erc20 := common.HexToAddress(pair.Erc20Address)
	acc := k.evmKeeper.GetAccountWithoutBalance(ctx, erc20)
	if acc == nil || !acc.IsContract() {
		k.DeleteTokenPair(ctx, pair)
		k.Logger(ctx).Debug("deleting selfdestructed token pair from state", "contract", pair.Erc20Address)
		// NOTE: return nil error to persist the changes from the deletion
		return nil, nil
	}

	var res *types.MsgConvertCoinResponse
	var gasMeter sdk.GasMeter
	ctx, gasMeter = k.SupportInfiniteGasMeter(ctx)
	// Check ownership and execute conversion
	switch {
	case pair.IsNativeCoin():
		res, err = k.convertCoinNativeCoin(ctx, pair, msg, receiver, sender) // case 1.1
	case pair.IsNativeERC20():
		res, err = k.convertCoinNativeERC20(ctx, pair, msg, receiver, sender) // case 2.2
	default:
		err = types.ErrUndefinedOwner
	}
	ctx = k.SupportRefundGasWithGasMeter(ctx, gasMeter) //nolint
	return res, err
}

// ConvertERC20 converts ERC20 tokens into Cosmos-native Coins for both
// Cosmos-native and ERC20 TokenPair Owners
func (k Keeper) ConvertERC20(goCtx context.Context, msg *types.MsgConvertERC20) (*types.MsgConvertERC20Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Error checked during msg validation
	receiver := sdk.MustAccAddressFromBech32(msg.Receiver)
	/* sender have two cases
	* 1. cosmos - m/44'/118'/0'/0/0
	* private key ---> cosmos address ---> hex sender
	* hex sender convert from cosmos address
	* private key can authenticate cosmos address in fxcore (verify with cosmos_spec256k1)
	* private key can **NOT** authenticate hex sender address in EVM (verify with eth_spec256k1)
	*
	* 2. eth - m/44'/60'/0'/0/1
	* private key ---> eth public key ---> hex sender ---> cosmos address
	*						 | ---> cosmos public key
	* cosmos address equal to hex sender, generate by eth public key
	* private key can authenticate cosmos address in fxcore (verify with eth_spec256k1)
	* private key can authenticate hex sender address in EVM (verify with eth_spec256k1)
	 */
	sender := common.HexToAddress(msg.Sender)

	pair, err := k.MintingEnabled(ctx, sender.Bytes(), receiver, msg.ContractAddress)
	if err != nil {
		return nil, err
	}

	// Remove token pair if contract is suicided
	erc20 := common.HexToAddress(pair.Erc20Address)
	acc := k.evmKeeper.GetAccountWithoutBalance(ctx, erc20)
	if acc == nil || !acc.IsContract() {
		k.DeleteTokenPair(ctx, pair)
		k.Logger(ctx).Debug("deleting selfdestructed token pair from state", "contract", pair.Erc20Address)
		// NOTE: return nil error to persist the changes from the deletion
		return nil, nil
	}

	var res *types.MsgConvertERC20Response
	var gasMeter sdk.GasMeter
	ctx, gasMeter = k.SupportInfiniteGasMeter(ctx)
	// Check ownership
	switch {
	case pair.IsNativeCoin():
		res, err = k.convertERC20NativeCoin(ctx, pair, msg, receiver, sender) // case 1.2
	case pair.IsNativeERC20():
		res, err = k.convertERC20NativeToken(ctx, pair, msg, receiver, sender) // case 2.1
	default:
		err = types.ErrUndefinedOwner
	}
	ctx = k.SupportRefundGasWithGasMeter(ctx, gasMeter) //nolint
	return res, err
}

// ConvertDenom converts coin into other coin, use for multiple chains in the same currency
func (k Keeper) ConvertDenom(goCtx context.Context, msg *types.MsgConvertDenom) (*types.MsgConvertDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Error checked during msg validation
	sender := sdk.MustAccAddressFromBech32(msg.Sender)
	receiver := sdk.MustAccAddressFromBech32(msg.Receiver)

	var coin sdk.Coin
	var err error
	if len(msg.Target) > 0 {
		// convert one to many
		coin, err = k.convertDenomToMany(ctx, sender, msg.Coin, msg.Target)
	} else {
		coin, err = k.convertDenomToOne(ctx, sender, msg.Coin)
	}
	if err != nil {
		return nil, err
	}

	if err := k.sendCoins(ctx, sender, receiver, sdk.NewCoins(coin)); err != nil {
		return nil, sdkerrors.Wrap(types.ErrConvertDenomSymbolFailed, err.Error())
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"tx", "msg", "convert", "denom", "total"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("denom", msg.Coin.Denom),
				telemetry.NewLabel("target_denom", coin.Denom),
			},
		)
	}()

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeConvertDenom,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
				sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
			),
		},
	)

	return &types.MsgConvertDenomResponse{}, nil
}

// convertCoinNativeCoin handles the Coin conversion flow for a native coin
// token pair:
//  - Escrow Coins on module account (Coins are not burned)
//  - Mint Tokens and send to receiver
//  - Check if token balance increased by amount
func (k Keeper) convertCoinNativeCoin(ctx sdk.Context, pair types.TokenPair, msg *types.MsgConvertCoin, receiver common.Address, sender sdk.AccAddress) (*types.MsgConvertCoinResponse, error) {
	// NOTE: ignore validation from NewCoin constructor
	coins := sdk.Coins{msg.Coin}
	erc20 := fxtypes.GetERC20().ABI
	contract := pair.GetERC20Contract()
	balanceToken, err := k.BalanceOf(ctx, contract, receiver)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEVMCall, "failed to retrieve balance: %s", err.Error())
	}

	// Escrow Coins on module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, coins); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to escrow coins")
	}

	// Mint Tokens and send to receiver
	_, err = k.CallEVM(ctx, erc20, types.ModuleAddress, contract, true, "mint", receiver, msg.Coin.Amount.BigInt())
	if err != nil {
		return nil, err
	}

	if pair.Denom == fxtypes.DefaultDenom {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, contract.Bytes(), coins); err != nil {
			return nil, sdkerrors.Wrap(err, "failed to transfer escrow coins to origin denom")
		}
	}

	// Check expected Receiver balance after transfer execution
	tokens := msg.Coin.Amount.BigInt()
	balanceTokenAfter, err := k.BalanceOf(ctx, contract, receiver)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEVMCall, "failed to retrieve balance: %s", err.Error())
	}
	exp := big.NewInt(0).Add(balanceToken, tokens)

	if r := balanceTokenAfter.Cmp(exp); r != 0 {
		return nil, sdkerrors.Wrapf(
			types.ErrBalanceInvariance,
			"invalid token balance - expected: %v, actual: %v", exp, balanceTokenAfter,
		)
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"tx", "msg", "convert", "coin", "total"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("denom", pair.Denom),
				telemetry.NewLabel("erc20", pair.Erc20Address),
			},
		)
	}()

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeConvertCoin,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
				sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
				sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Coin.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyDenom, msg.Coin.Denom),
				sdk.NewAttribute(types.AttributeKeyTokenAddress, pair.Erc20Address),
			),
		},
	)

	return &types.MsgConvertCoinResponse{}, nil
}

// convertERC20NativeCoin handles the erc20 conversion flow for a native coin token pair:
//  - Burn escrowed tokens
//  - Unescrow coins that have been previously escrowed with ConvertCoin
//  - Check if coin balance increased by amount
//  - Check if token balance decreased by amount
func (k Keeper) convertERC20NativeCoin(ctx sdk.Context, pair types.TokenPair, msg *types.MsgConvertERC20, receiver sdk.AccAddress, sender common.Address) (*types.MsgConvertERC20Response, error) {
	// NOTE: coin fields already validated
	coins := sdk.Coins{sdk.Coin{Denom: pair.Denom, Amount: msg.Amount}}

	erc20 := fxtypes.GetERC20().ABI
	contract := pair.GetERC20Contract()
	balanceCoin := k.bankKeeper.GetBalance(ctx, receiver, pair.Denom)
	balanceToken, err := k.BalanceOf(ctx, contract, sender)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEVMCall, "failed to retrieve balance: %s", err.Error())
	}

	// Burn escrowed tokens
	_, err = k.CallEVM(ctx, erc20, types.ModuleAddress, contract, true, "burn", sender, msg.Amount.BigInt())
	if err != nil {
		return nil, err
	}

	// Transfer origin denom to module
	if pair.Denom == fxtypes.DefaultDenom {
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, contract.Bytes(), types.ModuleName, coins); err != nil {
			return nil, sdkerrors.Wrap(err, "failed to transfer origin denom to module")
		}
	}

	// Unescrow Coins and send to receiver
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, coins); err != nil {
		return nil, err
	}

	// Check expected Receiver balance after transfer execution
	balanceCoinAfter := k.bankKeeper.GetBalance(ctx, receiver, pair.Denom)
	expCoin := balanceCoin.Add(coins[0])
	if ok := balanceCoinAfter.IsEqual(expCoin); !ok {
		return nil, sdkerrors.Wrapf(
			types.ErrBalanceInvariance,
			"invalid coin balance - expected: %v, actual: %v",
			expCoin, balanceCoinAfter,
		)
	}

	// Check expected Sender balance after transfer execution
	tokens := coins[0].Amount.BigInt()
	balanceTokenAfter, err := k.BalanceOf(ctx, contract, sender)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEVMCall, "failed to retrieve balance: %s", err.Error())
	}
	expToken := big.NewInt(0).Sub(balanceToken, tokens)
	if r := balanceTokenAfter.Cmp(expToken); r != 0 {
		return nil, sdkerrors.Wrapf(
			types.ErrBalanceInvariance,
			"invalid token balance - expected: %v, actual: %v",
			expToken, balanceTokenAfter,
		)
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"tx", "msg", "convert", "erc20", "total"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("denom", pair.Denom),
				telemetry.NewLabel("erc20", pair.Erc20Address),
			},
		)
	}()

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeConvertERC20,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
				sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
				sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyDenom, pair.Denom),
				sdk.NewAttribute(types.AttributeKeyTokenAddress, msg.ContractAddress),
			),
		},
	)

	return &types.MsgConvertERC20Response{}, nil
}

// convertERC20NativeToken handles the erc20 conversion flow for a native erc20 token pair:
//  - Escrow tokens on module account (Don't burn as module is not contract owner)
//  - Mint coins on module
//  - Send minted coins to the receiver
//  - Check if coin balance increased by amount
//  - Check if token balance decreased by amount
//  - Check for unexpected `appove` event in logs
func (k Keeper) convertERC20NativeToken(ctx sdk.Context, pair types.TokenPair, msg *types.MsgConvertERC20, receiver sdk.AccAddress, sender common.Address) (*types.MsgConvertERC20Response, error) {
	// NOTE: coin fields already validated
	coins := sdk.Coins{sdk.Coin{Denom: pair.Denom, Amount: msg.Amount}}
	erc20 := fxtypes.GetERC20().ABI
	contract := pair.GetERC20Contract()
	balanceCoin := k.bankKeeper.GetBalance(ctx, receiver, pair.Denom)
	balanceToken, err := k.BalanceOf(ctx, contract, types.ModuleAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEVMCall, "failed to retrieve balance: %s", err.Error())
	}

	// Escrow tokens on module account
	transferData, err := erc20.Pack("transfer", types.ModuleAddress, msg.Amount.BigInt())
	if err != nil {
		return nil, err
	}

	res, err := k.CallEVMWithData(ctx, sender, &contract, transferData, true)
	if err != nil {
		return nil, err
	}

	// Check unpackedRet execution
	var unpackedRet types.ERC20BoolResponse
	if err := erc20.UnpackIntoInterface(&unpackedRet, "transfer", res.Ret); err != nil {
		return nil, err
	}

	if !unpackedRet.Value {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "failed to execute transfer")
	}

	// Check expected escrow balance after transfer execution
	tokens := coins[0].Amount.BigInt()
	balanceTokenAfter, err := k.BalanceOf(ctx, contract, types.ModuleAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEVMCall, "failed to retrieve balance: %s", err.Error())
	}
	expToken := big.NewInt(0).Add(balanceToken, tokens)

	if r := balanceTokenAfter.Cmp(expToken); r != 0 {
		return nil, sdkerrors.Wrapf(
			types.ErrBalanceInvariance,
			"invalid token balance - expected: %v, actual: %v",
			expToken, balanceTokenAfter,
		)
	}

	// Mint coins
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return nil, err
	}

	// Send minted coins to the receiver
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, coins); err != nil {
		return nil, err
	}

	// Check expected Receiver balance after transfer execution
	balanceCoinAfter := k.bankKeeper.GetBalance(ctx, receiver, pair.Denom)
	expCoin := balanceCoin.Add(coins[0])

	if ok := balanceCoinAfter.IsEqual(expCoin); !ok {
		return nil, sdkerrors.Wrapf(
			types.ErrBalanceInvariance,
			"invalid coin balance - expected: %v, actual: %v",
			expCoin, balanceCoinAfter,
		)
	}

	// Check for unexpected `appove` event in logs
	if err = k.monitorApprovalEvent(res); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"tx", "msg", "convert", "erc20", "total"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("coin", pair.Denom),
				telemetry.NewLabel("erc20", pair.Erc20Address),
			},
		)
	}()

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeConvertERC20,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
				sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
				sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyDenom, pair.Denom),
				sdk.NewAttribute(types.AttributeKeyTokenAddress, msg.ContractAddress),
			),
		},
	)

	return &types.MsgConvertERC20Response{}, nil
}

// convertCoinNativeERC20 handles the Coin conversion flow for a native ERC20
// token pair:
//  - Escrow Coins on module account
//  - Unescrow Tokens that have been previously escrowed with ConvertERC20 and send to receiver
//  - Burn escrowed Coins
//  - Check if token balance increased by amount
//  - Check for unexpected `appove` event in logs
func (k Keeper) convertCoinNativeERC20(ctx sdk.Context, pair types.TokenPair, msg *types.MsgConvertCoin, receiver common.Address, sender sdk.AccAddress) (*types.MsgConvertCoinResponse, error) {
	// NOTE: ignore validation from NewCoin constructor
	coins := sdk.Coins{msg.Coin}

	erc20 := fxtypes.GetERC20().ABI
	contract := pair.GetERC20Contract()
	balanceToken, err := k.BalanceOf(ctx, contract, receiver)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEVMCall, "failed to retrieve balance: %s", err.Error())
	}

	// Escrow Coins on module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, coins); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to escrow coins")
	}

	// Unescrow Tokens and send to receiver
	res, err := k.CallEVM(ctx, erc20, types.ModuleAddress, contract, true, "transfer", receiver, msg.Coin.Amount.BigInt())
	if err != nil {
		return nil, err
	}

	// Check unpackedRet execution
	var unpackedRet types.ERC20BoolResponse
	if err := erc20.UnpackIntoInterface(&unpackedRet, "transfer", res.Ret); err != nil {
		return nil, err
	}

	if !unpackedRet.Value {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "failed to execute unescrow tokens from user")
	}

	// Check expected Receiver balance after transfer execution
	tokens := msg.Coin.Amount.BigInt()
	balanceTokenAfter, err := k.BalanceOf(ctx, contract, receiver)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEVMCall, "failed to retrieve balance: %s", err.Error())
	}
	exp := big.NewInt(0).Add(balanceToken, tokens)

	if r := balanceTokenAfter.Cmp(exp); r != 0 {
		return nil, sdkerrors.Wrapf(
			types.ErrBalanceInvariance,
			"invalid token balance - expected: %v, actual: %v", exp, balanceTokenAfter,
		)
	}

	// Burn escrowed Coins
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to burn coins")
	}

	// Check for unexpected `appove` event in logs
	if err = k.monitorApprovalEvent(res); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"tx", "msg", "convert", "coin", "total"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("denom", pair.Denom),
				telemetry.NewLabel("erc20", pair.Erc20Address),
			},
		)
	}()

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeConvertCoin,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
				sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
				sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Coin.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyDenom, msg.Coin.Denom),
				sdk.NewAttribute(types.AttributeKeyTokenAddress, pair.Erc20Address),
			),
		},
	)

	return &types.MsgConvertCoinResponse{}, nil
}

// convertDenomToMany handles the Denom conversion flow for one to many
// token pair:
//  - Escrow Coins on module account
//  - Unescrow Tokens that have been previously escrowed with convertDenomToMany and send to receiver
//  - Burn escrowed Coins
//  - Check if token balance increased by amount
func (k Keeper) convertDenomToMany(ctx sdk.Context, from sdk.AccAddress, coin sdk.Coin, target string) (sdk.Coin, error) {
	//check if denom registered
	if !k.IsDenomRegistered(ctx, coin.Denom) {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidDenom, "denom %s not registered", coin.Denom)
	}
	//check if metadata exist and support many to one
	md, found := k.bankKeeper.GetDenomMetaData(ctx, coin.Denom)
	if !found {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidMetadata, "denom %s not found", coin.Denom)
	}
	if !types.IsManyToOneMetadata(md) {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidMetadata, "denom %s metadata not support", coin.Denom)
	}

	// convert target to denom prefix
	denomPrefix := targetToDenomPrefix(ctx, target)

	aliases := md.DenomUnits[0].Aliases
	targetDenom := ""
	for _, alias := range aliases {
		if strings.HasPrefix(alias, denomPrefix) {
			targetDenom = alias
			break
		}
	}
	if len(targetDenom) == 0 {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidTarget, "target %s denom not exist", target)
	}

	//check if alias not registered
	if !k.IsAliasDenomRegistered(ctx, targetDenom) {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidDenom, "alias %s not registered", targetDenom)
	}

	beforeBalances := k.bankKeeper.GetAllBalances(ctx, from)

	var err error
	targetCoin := sdk.NewCoin(targetDenom, coin.Amount)
	// send symbol denom to module
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, from, types.ModuleName, sdk.NewCoins(coin))
	if err != nil {
		return sdk.Coin{}, sdkerrors.Wrapf(err, "send coin %s to module failed", coin.String())
	}
	// send alias denom to from addr
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, from, sdk.NewCoins(targetCoin))
	if err != nil {
		return sdk.Coin{}, sdkerrors.Wrapf(err, "send coin %s failed", targetCoin.String())
	}
	// burn symbol coin
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(coin))
	if err != nil {
		return sdk.Coin{}, sdkerrors.Wrapf(err, "burn coin %s failed", coin.String())
	}

	// check balances
	afterBalances := k.bankKeeper.GetAllBalances(ctx, from)
	if !beforeBalances.AmountOf(coin.Denom).Equal(afterBalances.AmountOf(coin.Denom).Add(coin.Amount)) ||
		!beforeBalances.AmountOf(targetDenom).Equal(afterBalances.AmountOf(targetDenom).Sub(coin.Amount)) {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrBalanceInvariance,
			"invalid token balance - convert denom %s to %s", coin.Denom, targetDenom)
	}

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeConvertDenomToMany,
				sdk.NewAttribute(types.AttributeKeyFrom, from.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, coin.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyDenom, coin.Denom),
				sdk.NewAttribute(types.AttributeKeyTargetDenom, targetCoin.Denom),
			),
		},
	)

	return targetCoin, nil
}

// convertDenomToOne handles the Denom conversion flow for many to one
// token pair:
//  - Escrow Coins on module account (Coins are not burned)
//  - Mint Tokens and send to from address
//  - Check if token balance increased by amount
func (k Keeper) convertDenomToOne(ctx sdk.Context, from sdk.AccAddress, coin sdk.Coin) (sdk.Coin, error) {
	if k.IsDenomRegistered(ctx, coin.Denom) {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidDenom, "denom %s already registered", coin.Denom)
	}
	aliasDenomBytes := k.GetAliasDenom(ctx, coin.Denom)
	if len(aliasDenomBytes) == 0 {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidDenom, "alias %s not registered", coin.Denom)
	}
	if !k.IsDenomRegistered(ctx, string(aliasDenomBytes)) {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidDenom, "denom %s not registered", string(aliasDenomBytes))
	}
	if ok, err := k.IsManyToOneDenom(ctx, string(aliasDenomBytes)); err != nil || !ok {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidMetadata, "not support with %s", string(aliasDenomBytes))
	}

	beforeBalances := k.bankKeeper.GetAllBalances(ctx, from)

	var err error
	targetCoin := sdk.NewCoin(string(aliasDenomBytes), coin.Amount)
	// send alias denom to module
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, from, types.ModuleName, sdk.NewCoins(coin))
	if err != nil {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrConvertDenomSymbolFailed, "send coin %s to module failed: %v", coin.String(), err)
	}
	//mint symbol denom
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(targetCoin))
	if err != nil {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrConvertDenomSymbolFailed, "mint coin %s failed: %v", targetCoin.String(), err)
	}
	//send symbol denom to from addr
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, from, sdk.NewCoins(targetCoin))
	if err != nil {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrConvertDenomSymbolFailed, "send coin %s failed: %v", targetCoin.String(), err)
	}

	// check balances
	afterBalances := k.bankKeeper.GetAllBalances(ctx, from)
	if !beforeBalances.AmountOf(coin.Denom).Equal(afterBalances.AmountOf(coin.Denom).Add(coin.Amount)) ||
		!beforeBalances.AmountOf(targetCoin.Denom).Equal(afterBalances.AmountOf(targetCoin.Denom).Sub(coin.Amount)) {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrBalanceInvariance,
			"invalid token balance - convert denom %s to %s", coin.Denom, targetCoin.Denom)
	}

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeConvertDenomToOne,
				sdk.NewAttribute(types.AttributeKeyFrom, from.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, coin.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyDenom, coin.Denom),
				sdk.NewAttribute(types.AttributeKeyTargetDenom, targetCoin.Denom),
			),
		},
	)

	return targetCoin, nil
}

func (k Keeper) sendCoins(ctx sdk.Context, from, to sdk.AccAddress, coins sdk.Coins) error {
	if err := k.bankKeeper.IsSendEnabledCoins(ctx, coins...); err != nil {
		return err
	}

	if k.bankKeeper.BlockedAddr(to) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", to.String())
	}

	return k.bankKeeper.SendCoins(ctx, from, to, coins)
}

func (k Keeper) SupportInfiniteGasMeter(ctx sdk.Context) (sdk.Context, sdk.GasMeter) {
	gasMeter := ctx.GasMeter()
	if ctx.BlockHeight() >= fxtypes.SupportDenomOneToManyBlock() {
		ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	}
	return ctx, gasMeter
}

func (k Keeper) SupportRefundGasWithGasMeter(ctx sdk.Context, gasMater sdk.GasMeter) sdk.Context {
	if ctx.BlockHeight() >= fxtypes.SupportDenomOneToManyBlock() {
		ctx = ctx.WithGasMeter(gasMater)
	}
	return ctx
}

func targetToDenomPrefix(ctx sdk.Context, target string) (prefix string) {
	if fxtypes.ChainId() == fxtypes.TestnetChainId && ctx.BlockHeight() < fxtypes.SupportDenomOneToManyBlock() {
		return target
	}
	if target == gravitytypes.ModuleName {
		return gravitytypes.GravityDenomPrefix
	}
	return target
}
