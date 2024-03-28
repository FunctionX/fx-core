package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/slices"

	"github.com/functionx/fx-core/v7/contract"
	fxtypes "github.com/functionx/fx-core/v7/types"
	"github.com/functionx/fx-core/v7/x/crosschain/types"
)

// AttestationHandler Handle is the entry point for Attestation processing.
//
//gocyclo:ignore
func (k Keeper) AttestationHandler(ctx sdk.Context, externalClaim types.ExternalClaim) error {
	switch claim := externalClaim.(type) {
	case *types.MsgSendToFxClaim:
		bridgeToken := k.GetBridgeTokenDenom(ctx, claim.TokenContract)
		if bridgeToken == nil {
			return errorsmod.Wrap(types.ErrInvalid, "bridge token is not exist")
		}

		coin := sdk.NewCoin(bridgeToken.Denom, claim.Amount)
		receiveAddr, err := sdk.AccAddressFromBech32(claim.Receiver)
		if err != nil {
			return errorsmod.Wrap(types.ErrInvalid, "receiver address")
		}
		isOriginOrConverted := k.erc20Keeper.IsOriginOrConvertedDenom(ctx, bridgeToken.Denom)
		if !isOriginOrConverted {
			// If it is not fxcore originated, mint the coins (aka vouchers)
			if err := k.bankKeeper.MintCoins(ctx, k.moduleName, sdk.NewCoins(coin)); err != nil {
				return errorsmod.Wrapf(err, "mint vouchers coins")
			}
		}
		if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, k.moduleName, receiveAddr, sdk.NewCoins(coin)); err != nil {
			return errorsmod.Wrap(err, "transfer vouchers")
		}

		// convert to base denom
		cacheCtx, commit := ctx.CacheContext()
		targetCoin, err := k.erc20Keeper.ConvertDenomToTarget(cacheCtx, receiveAddr, coin, fxtypes.ParseFxTarget(fxtypes.ERC20Target))
		if err != nil {
			k.Logger(ctx).Info("failed to convert base denom", "error", err)
			return nil
		}
		commit()

		// relay transfer
		if err = k.RelayTransferHandler(ctx, claim.EventNonce, claim.TargetIbc, receiveAddr, targetCoin); err != nil {
			k.Logger(ctx).Info("failed to relay transfer", "error", err)
			return nil
		}

	case *types.MsgBridgeCallClaim:
		assetType, assetData, err := types.UnpackAssetType(claim.Asset)
		if err != nil {
			return errorsmod.Wrap(types.ErrInvalid, "asset")
		}
		switch assetType {
		case contract.AssetERC20:
			return k.BridgeCallERC20Handler(ctx, assetData, claim.MustSender(), claim.MustTo(),
				claim.MustReceiver(), claim.DstChainId, claim.MustMessage(), claim.Value, claim.GasLimit, claim.EventNonce)
		default:
			return errorsmod.Wrap(types.ErrInvalid, "asset not support")
		}

	case *types.MsgSendToExternalClaim:
		k.OutgoingTxBatchExecuted(ctx, claim.TokenContract, claim.BatchNonce)

	case *types.MsgBridgeTokenClaim:
		// Check if it already exists
		isExist := k.HasBridgeToken(ctx, claim.TokenContract)
		if isExist {
			return errorsmod.Wrap(types.ErrInvalid, "bridge token is exist")
		}
		k.Logger(ctx).Info("add bridge token claim", "symbol", claim.Symbol, "token", claim.TokenContract, "channelIbc", claim.ChannelIbc, "tokenType", claim.TokenType)
		if claim.Symbol == types.NativeDenom {
			// Check if denom exists
			metadata, found := k.bankKeeper.GetDenomMetaData(ctx, claim.Symbol)
			if !found {
				return errorsmod.Wrap(
					types.ErrUnknown,
					fmt.Sprintf("denom not found %s", claim.Symbol))
			}

			// Check if attributes of ERC20 match fx denom
			if claim.Name != metadata.Name {
				return errorsmod.Wrap(
					types.ErrInvalid,
					fmt.Sprintf("ERC20 name %s does not match denom display %s", claim.Name, metadata.Description))
			}

			if claim.Symbol != metadata.Symbol {
				return errorsmod.Wrap(
					types.ErrInvalid,
					fmt.Sprintf("ERC20 symbol %s does not match denom display %s", claim.Symbol, metadata.Display))
			}

			if fxtypes.DenomUnit != uint32(claim.Decimals) {
				return errorsmod.Wrap(
					types.ErrInvalid,
					fmt.Sprintf("ERC20 decimals %d does not match denom decimals %d", claim.Decimals, fxtypes.DenomUnit))
			}
			// first to add origin token, update alias first
			if len(metadata.DenomUnits[0].Aliases) == 0 {
				k.AddBridgeTokenWithTokenType(ctx, claim.TokenContract, claim.Symbol, claim.TokenType)
				return nil
			}

			// add more origin token
			denom := NewBridgeDenom(k.moduleName, claim.TokenContract)
			if !slices.Contains(metadata.DenomUnits[0].Aliases, denom) {
				return errorsmod.Wrap(
					types.ErrInvalid,
					fmt.Sprintf("Token %s not exist in metadata", denom))
			}
			k.AddBridgeTokenWithTokenType(ctx, claim.TokenContract, denom, claim.TokenType)
			return nil
		}

		denom, err := k.SetIbcDenomTrace(ctx, claim.TokenContract, claim.ChannelIbc)
		if err != nil {
			return err
		}
		k.AddBridgeTokenWithTokenType(ctx, claim.TokenContract, denom, claim.TokenType)
		k.Logger(ctx).Info("add bridge token success", "symbol", claim.Symbol, "token", claim.TokenContract, "denom", denom, "channelIbc", claim.ChannelIbc)

	case *types.MsgOracleSetUpdatedClaim:
		observedOracleSet := &types.OracleSet{
			Nonce:   claim.OracleSetNonce,
			Members: claim.Members,
		}
		// check the contents of the validator set against the store
		if claim.OracleSetNonce != 0 {
			trustedOracleSet := k.GetOracleSet(ctx, claim.OracleSetNonce)
			if trustedOracleSet == nil {
				ctx.Logger().Error("Received attestation for a oracle set which does not exist in store", "oracleSetNonce", claim.OracleSetNonce, "claim", claim)
				return errorsmod.Wrapf(types.ErrInvalid, "attested oracleSet (%v) does not exist in store", claim.OracleSetNonce)
			}
			// overwrite the height, since it's not part of the claim
			observedOracleSet.Height = trustedOracleSet.Height

			if _, err := trustedOracleSet.Equal(observedOracleSet); err != nil {
				panic(fmt.Sprintf("Potential bridge highjacking: observed oracleSet (%+v) does not match stored oracleSet (%+v)! %s", observedOracleSet, trustedOracleSet, err.Error()))
			}
		}
		k.SetLastObservedOracleSet(ctx, observedOracleSet)

	case *types.MsgRefundTokenClaim:
		k.HandleRefundTokenClaim(ctx, claim)
		return nil
	default:
		return errorsmod.Wrapf(types.ErrInvalid, "event type: %s", claim.GetType())
	}
	return nil
}

func NewBridgeDenom(moduleName string, token string) string {
	return fmt.Sprintf("%s%s", moduleName, token)
}
