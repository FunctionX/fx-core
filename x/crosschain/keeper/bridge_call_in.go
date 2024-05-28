package keeper

import (
	"bytes"
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"

	fxtypes "github.com/functionx/fx-core/v7/types"
	"github.com/functionx/fx-core/v7/x/crosschain/types"
	erc20types "github.com/functionx/fx-core/v7/x/erc20/types"
)

func (k Keeper) BridgeCallHandler(ctx sdk.Context, msg *types.MsgBridgeCallClaim) error {
	tokens := msg.GetTokensAddr()
	erc20Token, err := types.NewERC20Tokens(k.moduleName, tokens, msg.GetAmounts())
	if err != nil {
		return err
	}
	var errCause string
	cacheCtx, commit := ctx.CacheContext()
	sender := msg.GetSenderAddr()
	receiver := msg.GetReceiverAddr()
	eventNonce := msg.EventNonce
	if err = k.BridgeCallTransferAndCallEvm(cacheCtx, sender, receiver, erc20Token, msg.GetToAddr(), msg.MustData(), msg.MustMemo(), msg.Value); err != nil {
		errCause = err.Error()
		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeBridgeCallEvent, sdk.NewAttribute(types.AttributeKeyErrCause, errCause)))
	} else {
		commit()
	}

	if len(errCause) > 0 && len(tokens) > 0 {
		// new outgoing bridge call to refund
		outCall, err := k.AddOutgoingBridgeCall(ctx, receiver, sender.String(), erc20Token, common.Address{}.String(), "", "", eventNonce)
		if err != nil {
			return err
		}
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeBridgeCallRefundOut,
			sdk.NewAttribute(types.AttributeKeyEventNonce, fmt.Sprintf("%d", eventNonce)),
			sdk.NewAttribute(types.AttributeKeyBridgeCallNonce, fmt.Sprintf("%d", outCall.Nonce)),
		))
	}

	if len(errCause) == 0 {
		for i := 0; i < len(erc20Token); i++ {
			bridgeToken := k.GetBridgeTokenDenom(ctx, erc20Token[i].Contract)
			// no need for a double check here, as the bridge token should exist
			k.HandlePendingOutgoingTx(ctx, receiver, eventNonce, bridgeToken)
		}
	}
	return nil
}

func (k Keeper) BridgeCallTransferAndCallEvm(
	ctx sdk.Context,
	sender common.Address,
	receiver sdk.AccAddress,
	tokens []types.ERC20Token,
	to *common.Address,
	data []byte,
	memo []byte,
	value sdkmath.Int,
) error {
	if senderAccount := k.ak.GetAccount(ctx, sender.Bytes()); senderAccount != nil {
		if _, ok := senderAccount.(authtypes.ModuleAccountI); ok {
			return errorsmod.Wrap(types.ErrInvalid, "sender is module account")
		}
	}
	coins, err := k.bridgeCallTransferToSender(ctx, sender.Bytes(), tokens)
	if err != nil {
		return err
	}
	if err = k.bridgeCallTransferToReceiver(ctx, sender.Bytes(), receiver, coins); err != nil {
		return err
	}
	if len(data) > 0 || to != nil {
		callTokens, callAmounts := k.CoinsToBridgeCallTokens(ctx, coins)
		args, err := types.PackBridgeCallback(sender, common.Address(receiver.Bytes()), callTokens, callAmounts, data, memo)
		if err != nil {
			return err
		}
		gasLimit := k.GetParams(ctx).BridgeCallMaxGasLimit
		txResp, err := k.evmKeeper.CallEVM(ctx, k.callbackFrom, to, value.BigInt(), gasLimit, args, true)
		if err != nil {
			return err
		}
		if txResp.Failed() {
			return errorsmod.Wrap(types.ErrInvalid, txResp.VmError)
		}
	}
	return nil
}

func (k Keeper) bridgeCallTransferToSender(ctx sdk.Context, sender sdk.AccAddress, tokens []types.ERC20Token) (sdk.Coins, error) {
	mintCoins := sdk.NewCoins()
	unlockCoins := sdk.NewCoins()
	for i := 0; i < len(tokens); i++ {
		bridgeToken := k.GetBridgeTokenDenom(ctx, tokens[i].Contract)
		if bridgeToken == nil {
			return nil, errorsmod.Wrap(types.ErrInvalid, "bridge token is not exist")
		}
		if !tokens[i].Amount.IsPositive() {
			continue
		}
		coin := sdk.NewCoin(bridgeToken.Denom, tokens[i].Amount)
		isOriginOrConverted := k.erc20Keeper.IsOriginOrConvertedDenom(ctx, bridgeToken.Denom)
		if !isOriginOrConverted {
			mintCoins = mintCoins.Add(coin)
		}
		unlockCoins = unlockCoins.Add(coin)
	}
	if mintCoins.IsAllPositive() {
		if err := k.bankKeeper.MintCoins(ctx, k.moduleName, mintCoins); err != nil {
			return nil, errorsmod.Wrapf(err, "mint vouchers coins")
		}
	}
	if unlockCoins.IsAllPositive() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, k.moduleName, sender, unlockCoins); err != nil {
			return nil, errorsmod.Wrap(err, "transfer vouchers")
		}
	}

	targetCoins := sdk.NewCoins()
	for _, coin := range unlockCoins {
		targetCoin, err := k.erc20Keeper.ConvertDenomToTarget(ctx, sender, coin, fxtypes.ParseFxTarget(fxtypes.ERC20Target))
		if err != nil {
			return nil, errorsmod.Wrap(err, "convert to target coin")
		}
		targetCoins = targetCoins.Add(targetCoin)
	}
	return targetCoins, nil
}

func (k Keeper) bridgeCallTransferToReceiver(ctx sdk.Context, sender sdk.AccAddress, receiver []byte, coins sdk.Coins) error {
	for _, coin := range coins {
		if coin.Denom == fxtypes.DefaultDenom {
			if bytes.Equal(sender, receiver) {
				continue
			}
			if err := k.bankKeeper.SendCoins(ctx, sender, receiver, sdk.NewCoins(coin)); err != nil {
				return err
			}
			continue
		}
		if _, err := k.erc20Keeper.ConvertCoin(sdk.WrapSDKContext(ctx), &erc20types.MsgConvertCoin{
			Coin:     coin,
			Receiver: common.BytesToAddress(receiver).String(),
			Sender:   sender.String(),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) CoinsToBridgeCallTokens(ctx sdk.Context, coins sdk.Coins) ([]common.Address, []*big.Int) {
	_tokens := make([]common.Address, 0, len(coins))
	_amounts := make([]*big.Int, 0, len(coins))
	for _, coin := range coins {
		_amounts = append(_amounts, coin.Amount.BigInt())
		if coin.Denom == fxtypes.DefaultDenom {
			_tokens = append(_tokens, common.Address{})
			continue
		}
		// bridgeCallTransferToReceiver().ConvertCoin hava already checked.
		pair, _ := k.erc20Keeper.GetTokenPair(ctx, coin.Denom)
		_tokens = append(_tokens, common.HexToAddress(pair.Erc20Address))
	}
	return _tokens, _amounts
}
