package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/hashicorp/go-metrics"

	fxtelemetry "github.com/functionx/fx-core/v8/telemetry"
	"github.com/functionx/fx-core/v8/x/crosschain/types"
)

func (k Keeper) AddToOutgoingPool(ctx sdk.Context, sender sdk.AccAddress, receiver string, amount sdk.Coin, fee sdk.Coin) (uint64, error) {
	// get next tx id from keeper
	nextTxID := k.autoIncrementID(ctx, types.KeyLastTxPoolID)
	return nextTxID, k.addToOutgoingPool(ctx, sender, receiver, amount, fee, nextTxID)
}

// AddToOutgoingPool
// - checks a counterpart denominator exists for the given voucher type
// - burns the voucher for transfer amount and fees
// - persists an OutgoingTx
// - adds the TX to the `available` TX pool via a second index
func (k Keeper) addToOutgoingPool(ctx sdk.Context, sender sdk.AccAddress, receiver string, amount sdk.Coin, fee sdk.Coin, txID uint64) error {
	tokenContract, err := k.BaseCoinToBridgeToken(ctx, amount.Add(fee), sender)
	if err != nil {
		return err
	}

	// construct outgoing tx, as part of this process we represent
	// the token as an ERC20 token since it is preparing to go to ETH
	// rather than the denom that is the input to this function.
	outgoing := &types.OutgoingTransferTx{
		Id:          txID,
		Sender:      sender.String(),
		DestAddress: receiver,
		Token:       types.NewERC20Token(amount.Amount, tokenContract),
		Fee:         types.NewERC20Token(fee.Amount, tokenContract),
	}

	if err := k.AddUnbatchedTx(ctx, outgoing); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeSendToExternal,
		sdk.NewAttribute(sdk.AttributeKeyModule, k.moduleName),
		sdk.NewAttribute(types.AttributeKeyOutgoingTxID, fmt.Sprint(txID)),
	))

	if !ctx.IsCheckTx() {
		fxtelemetry.SetGaugeLabelsWithDenom(
			[]string{types.ModuleName, "outgoing_tx_amount"},
			amount.Denom, amount.Amount.BigInt(),
			telemetry.NewLabel("module", k.moduleName),
		)
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "outgoing_tx"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("module", k.moduleName),
				telemetry.NewLabel("denom", amount.Denom),
			},
		)
	}

	return nil
}

// RemoveFromOutgoingPoolAndRefund
// - checks that the provided tx actually exists
// - deletes the unbatched tx from the pool
// - issues the tokens back to the sender
func (k Keeper) RemoveFromOutgoingPoolAndRefund(ctx sdk.Context, txId uint64, sender sdk.AccAddress) (sdk.Coin, error) {
	if ctx.IsZero() || txId < 1 || sender.Empty() {
		return sdk.Coin{}, types.ErrInvalid.Wrapf("arguments")
	}

	// check that we actually have a tx with that id and what it's details are
	tx, err := k.GetUnbatchedTxById(ctx, txId)
	if err != nil {
		return sdk.Coin{}, err
	}
	return k.handleRemoveFromOutgoingPoolAndRefund(ctx, tx, sender)
}

// AddUnbatchedTx creates a new transaction in the pool
func (k Keeper) AddUnbatchedTx(ctx sdk.Context, outgoingTransferTx *types.OutgoingTransferTx) error {
	store := ctx.KVStore(k.storeKey)
	idxKey := types.GetOutgoingTxPoolKey(outgoingTransferTx.Fee, outgoingTransferTx.Id)
	if store.Has(idxKey) {
		return sdkerrors.ErrInvalidRequest.Wrapf("tx with id %d already in pool", outgoingTransferTx.Id)
	}

	store.Set(idxKey, k.cdc.MustMarshal(outgoingTransferTx))
	return nil
}

// pickUnBatchedTx find Tx in pool and remove from "available" second index
func (k Keeper) pickUnBatchedTx(ctx sdk.Context, tokenContract string, maxElements uint, baseFee sdkmath.Int) ([]*types.OutgoingTransferTx, error) {
	var selectedTx []*types.OutgoingTransferTx
	var err error
	k.IterateUnbatchedTransactions(ctx, tokenContract, func(tx *types.OutgoingTransferTx) bool {
		if tx.Fee.Amount.LT(baseFee) {
			return true
		}
		selectedTx = append(selectedTx, tx)
		err = k.removeUnbatchedTx(ctx, tx.Fee, tx.Id)
		oldTx, oldTxErr := k.GetUnbatchedTxByFeeAndId(ctx, tx.Fee, tx.Id)
		if oldTx != nil || oldTxErr == nil {
			panic("picked a duplicate transaction from the pool, duplicates should never exist!")
		}
		return err != nil || uint(len(selectedTx)) == maxElements
	})
	return selectedTx, err
}

// removeUnbatchedTXIndex removes the tx from the pool
func (k Keeper) removeUnbatchedTx(ctx sdk.Context, fee types.ERC20Token, txID uint64) error {
	store := ctx.KVStore(k.storeKey)
	idxKey := types.GetOutgoingTxPoolKey(fee, txID)
	if !store.Has(idxKey) {
		return types.ErrInvalid.Wrapf("tx with id %d not in pool", txID)
	}
	store.Delete(idxKey)
	return nil
}

// GetUnbatchedTxByFeeAndId grabs a tx from the pool given its fee and txID
func (k Keeper) GetUnbatchedTxByFeeAndId(ctx sdk.Context, fee types.ERC20Token, txID uint64) (*types.OutgoingTransferTx, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetOutgoingTxPoolKey(fee, txID))
	if bz == nil {
		return nil, types.ErrInvalid.Wrapf("tx with id %d not in pool", txID)
	}
	var r types.OutgoingTransferTx
	err := k.cdc.Unmarshal(bz, &r)
	return &r, err
}

// GetUnbatchedTxById grabs a tx from the pool given only the txID
// note that due to the way unbatched txs are indexed, the GetUnbatchedTxByFeeAndId method is much faster
func (k Keeper) GetUnbatchedTxById(ctx sdk.Context, txID uint64) (*types.OutgoingTransferTx, error) {
	var r *types.OutgoingTransferTx
	k.IterateUnbatchedTransactions(ctx, "", func(tx *types.OutgoingTransferTx) bool {
		if tx.Id == txID {
			r = tx
			return true
		}
		return false
	})

	if r == nil {
		// We have no return tx, it was either batched or never existed
		return nil, types.ErrInvalid.Wrapf("tx with id %d not in pool", txID)
	}
	return r, nil
}

// GetUnbatchedTransactions used in testing
func (k Keeper) GetUnbatchedTransactions(ctx sdk.Context) []*types.OutgoingTransferTx {
	var txs []*types.OutgoingTransferTx
	k.IterateUnbatchedTransactions(ctx, "", func(tx *types.OutgoingTransferTx) bool {
		txs = append(txs, tx)
		return false
	})
	return txs
}

// IterateUnbatchedTransactions iterates through all unbatched transactions
func (k Keeper) IterateUnbatchedTransactions(ctx sdk.Context, tokenContract string, cb func(tx *types.OutgoingTransferTx) bool) {
	store := ctx.KVStore(k.storeKey)
	prefixKey := types.GetOutgoingTxPoolContractPrefix(tokenContract)
	iter := storetypes.KVStoreReversePrefixIterator(store, prefixKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var transact types.OutgoingTransferTx
		k.cdc.MustUnmarshal(iter.Value(), &transact)
		// cb returns true to stop early
		if cb(&transact) {
			break
		}
	}
}

func (k Keeper) handleRemoveFromOutgoingPoolAndRefund(ctx sdk.Context, tx *types.OutgoingTransferTx, sender sdk.AccAddress) (sdk.Coin, error) {
	txId := tx.Id

	// Check that this user actually sent the transaction, this prevents someone from refunding someone
	// else transaction to themselves.
	txSender := sdk.MustAccAddressFromBech32(tx.Sender)
	if !txSender.Equals(sender) {
		return sdk.Coin{}, types.ErrInvalid.Wrapf("Sender %s did not send Id %d", sender, txId)
	}

	// an inconsistent entry should never enter the store, but this is the ideal place to exploit
	// it such a bug if it did ever occur, so we should double check to be really sure
	if tx.Fee.Contract != tx.Token.Contract {
		return sdk.Coin{}, types.ErrInvalid.Wrapf("Inconsistent tokens to cancel!: %s %s", tx.Fee.Contract, tx.Token.Contract)
	}

	// delete this tx from the pool
	if err := k.removeUnbatchedTx(ctx, tx.Fee, txId); err != nil {
		return sdk.Coin{}, types.ErrInvalid.Wrapf("txId %d not in unbatched index! Must be in a batch!", txId)
	}
	// Make sure the tx was removed
	oldTx, oldTxErr := k.GetUnbatchedTxByFeeAndId(ctx, tx.Fee, tx.Id)
	if oldTx != nil || oldTxErr == nil {
		return sdk.Coin{}, types.ErrInvalid.Wrapf("tx with id %d was not fully removed from the pool, a duplicate must exist", txId)
	}

	return k.handleCancelRefund(ctx, txId, sender, tx.Token.Contract, tx.Token.Amount.Add(tx.Fee.Amount))
}

func (k Keeper) handleCancelRefund(ctx sdk.Context, txId uint64, sender sdk.AccAddress, tokenContract string, refundAmount sdkmath.Int) (sdk.Coin, error) {
	// 1. handler refund and convert to base coin
	baseCoin, err := k.BridgeTokenToBaseCoin(ctx, tokenContract, refundAmount, sender)
	if err != nil {
		return sdk.Coin{}, err
	}

	// 2. handler hook
	if err = k.handleOutgoingTransferRelation(ctx, txId, sender, baseCoin); err != nil {
		return sdk.Coin{}, err
	}

	// 3. emit event
	k.emitCancelEvent(ctx, txId)

	return baseCoin, nil
}

func (k Keeper) handleOutgoingTransferRelation(ctx sdk.Context, txId uint64, sender sdk.AccAddress, targetCoin sdk.Coin) error {
	// if not origin token, refund to contract token
	if !k.erc20Keeper.HasOutgoingTransferRelation(ctx, k.moduleName, txId) {
		return nil
	}
	if err := k.erc20Keeper.HookOutgoingRefund(ctx, k.moduleName, txId, sender, targetCoin); err != nil {
		return err
	}
	return nil
}

func (k Keeper) emitCancelEvent(ctx sdk.Context, txId uint64) {
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeSendToExternalCanceled,
		sdk.NewAttribute(sdk.AttributeKeyModule, k.moduleName),
		sdk.NewAttribute(types.AttributeKeyOutgoingTxID, fmt.Sprint(txId)),
	))
}
