package keeper

import (
	"fmt"
	"math"
	"sort"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/functionx/fx-core/v2/x/crosschain/types"
)

/////////////////////////////
//   ORACLE SET REQUESTS   //
/////////////////////////////

// GetCurrentOracleSet gets powers from the store and normalizes them
// into an integer percentage with a resolution of uint32 Max meaning
// a given validators 'gravity power' is computed as
// Cosmos power / total cosmos power = x / uint32 Max
// where x is the voting power on the gravity contract. This allows us
// to only use integer division which produces a known rounding error
// from truncation equal to the ratio of the validators
// Cosmos power / total cosmos power ratio, leaving us at uint32 Max - 1
// total voting power. This is an acceptable rounding error since floating
// point may cause consensus problems if different floating point unit
// implementations are involved.
func (k Keeper) GetCurrentOracleSet(ctx sdk.Context) *types.OracleSet {
	allOracles := k.GetAllOracles(ctx, true)
	var bridgeValidators []types.BridgeValidator
	var totalPower uint64

	for _, oracle := range allOracles {
		power := oracle.GetPower()
		if power.LTE(sdk.ZeroInt()) {
			continue
		}
		totalPower += power.Uint64()
		bridgeValidators = append(bridgeValidators, types.BridgeValidator{
			Power:           power.Uint64(),
			ExternalAddress: oracle.ExternalAddress,
		})
	}
	// normalize power values
	for i := range bridgeValidators {
		bridgeValidators[i].Power = sdk.NewUint(bridgeValidators[i].Power).MulUint64(math.MaxUint32).QuoUint64(totalPower).Uint64()
	}

	oracleSetNonce := k.GetLatestOracleSetNonce(ctx) + 1
	return types.NewOracleSet(oracleSetNonce, uint64(ctx.BlockHeight()), bridgeValidators)
}

// AddOracleSetRequest returns a new instance of the Gravity BridgeValidatorSet
func (k Keeper) AddOracleSetRequest(ctx sdk.Context, currentOracleSet *types.OracleSet) {
	// if currentOracleSet member is empty, not store OracleSet.
	if len(currentOracleSet.Members) <= 0 {
		return
	}
	k.StoreOracleSet(ctx, currentOracleSet)

	k.CommonSetOracleTotalPower(ctx)

	gravityId := k.GetGravityID(ctx)
	checkpoint, err := currentOracleSet.GetCheckpoint(gravityId)
	if err != nil {
		panic(err)
	}
	k.SetPastExternalSignatureCheckpoint(ctx, checkpoint)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeOracleSetUpdate,
		sdk.NewAttribute(sdk.AttributeKeyModule, k.moduleName),
		sdk.NewAttribute(types.AttributeKeyOracleSetNonce, fmt.Sprint(currentOracleSet.Nonce)),
		sdk.NewAttribute(types.AttributeKeyOracleSetLen, fmt.Sprint(len(currentOracleSet.Members))),
	))
}

// StoreOracleSet is for storing a oracle set at a given height
func (k Keeper) StoreOracleSet(ctx sdk.Context, oracleSet *types.OracleSet) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetOracleSetKey(oracleSet.Nonce), k.cdc.MustMarshal(oracleSet))
	k.SetLatestOracleSetNonce(ctx, oracleSet.Nonce)
}

// HasOracleSetRequest returns true if a oracleSet defined by a nonce exists
func (k Keeper) HasOracleSetRequest(ctx sdk.Context, nonce uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetOracleSetKey(nonce))
}

// DeleteOracleSet deletes the oracleSet at a given nonce from state
func (k Keeper) DeleteOracleSet(ctx sdk.Context, nonce uint64) {
	ctx.KVStore(k.storeKey).Delete(types.GetOracleSetKey(nonce))
}

// SetLatestOracleSetNonce sets the latest oracleSet nonce
func (k Keeper) SetLatestOracleSetNonce(ctx sdk.Context, nonce uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LatestOracleSetNonce, sdk.Uint64ToBigEndian(nonce))
}

// GetLatestOracleSetNonce returns the latest oracleSet nonce
func (k Keeper) GetLatestOracleSetNonce(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	data := store.Get(types.LatestOracleSetNonce)
	if len(data) == 0 {
		return 0
	}
	return sdk.BigEndianToUint64(data)
}

// GetOracleSet returns a oracleSet by nonce
func (k Keeper) GetOracleSet(ctx sdk.Context, nonce uint64) *types.OracleSet {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetOracleSetKey(nonce))
	if bz == nil {
		return nil
	}
	var oracleSet types.OracleSet
	k.cdc.MustUnmarshal(bz, &oracleSet)
	return &oracleSet
}

// IterateOracleSets returns all oracleSetRequests
func (k Keeper) IterateOracleSets(ctx sdk.Context, cb func(key []byte, val *types.OracleSet) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.OracleSetRequestKey)
	iter := prefixStore.ReverseIterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var oracleSet types.OracleSet
		k.cdc.MustUnmarshal(iter.Value(), &oracleSet)
		// cb returns true to stop early
		if cb(iter.Key(), &oracleSet) {
			break
		}
	}
}

// GetOracleSets returns all the oracle sets in state
func (k Keeper) GetOracleSets(ctx sdk.Context) (out []*types.OracleSet) {
	k.IterateOracleSets(ctx, func(_ []byte, val *types.OracleSet) bool {
		out = append(out, val)
		return false
	})
	sort.Sort(types.OracleSets(out))
	return
}

// GetLatestOracleSet returns the latest oracle set in state
func (k Keeper) GetLatestOracleSet(ctx sdk.Context) *types.OracleSet {
	latestOracleSetNonce := k.GetLatestOracleSetNonce(ctx)
	return k.GetOracleSet(ctx, latestOracleSetNonce)
}

// SetLastSlashedOracleSetNonce sets the latest slashed oracleSet nonce
func (k Keeper) SetLastSlashedOracleSetNonce(ctx sdk.Context, nonce uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastSlashedOracleSetNonce, sdk.Uint64ToBigEndian(nonce))
}

// GetLastSlashedOracleSetNonce returns the latest slashed oracleSet nonce
func (k Keeper) GetLastSlashedOracleSetNonce(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	data := store.Get(types.LastSlashedOracleSetNonce)
	if len(data) == 0 {
		return 0
	}
	return sdk.BigEndianToUint64(data)
}

/////////////////////////////
//   ORACLE SET CONFIRMS   //
/////////////////////////////

// GetUnSlashedOracleSets returns all the unSlashed oracle sets in state
func (k Keeper) GetUnSlashedOracleSets(ctx sdk.Context, maxHeight uint64) (oracleSets types.OracleSets) {
	lastSlashedOracleSetNonce := k.GetLastSlashedOracleSetNonce(ctx)
	k.IterateOracleSetBySlashedOracleSetNonce(ctx, lastSlashedOracleSetNonce, maxHeight, func(_ []byte, oracleSet *types.OracleSet) bool {
		if oracleSet.Nonce > lastSlashedOracleSetNonce && maxHeight > oracleSet.Height {
			oracleSets = append(oracleSets, oracleSet)
		}
		return false
	})
	sort.Sort(oracleSets)
	return
}

// IterateOracleSetBySlashedOracleSetNonce iterates through all oracleSet by last slashed oracleSet nonce in ASC order
func (k Keeper) IterateOracleSetBySlashedOracleSetNonce(ctx sdk.Context, lastSlashedOracleSetNonce uint64, maxHeight uint64, cb func([]byte, *types.OracleSet) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.OracleSetRequestKey)
	iter := prefixStore.Iterator(sdk.Uint64ToBigEndian(lastSlashedOracleSetNonce), sdk.Uint64ToBigEndian(maxHeight))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var oracleSet types.OracleSet
		k.cdc.MustUnmarshal(iter.Value(), &oracleSet)
		// cb returns true to stop early
		if cb(iter.Key(), &oracleSet) {
			break
		}
	}
}

// GetOracleSetConfirm returns a oracleSet confirmation by a nonce and external address
func (k Keeper) GetOracleSetConfirm(ctx sdk.Context, nonce uint64, oracleAddr sdk.AccAddress) *types.MsgOracleSetConfirm {
	store := ctx.KVStore(k.storeKey)
	entity := store.Get(types.GetOracleSetConfirmKey(nonce, oracleAddr))
	if entity == nil {
		return nil
	}
	confirm := types.MsgOracleSetConfirm{}
	k.cdc.MustUnmarshal(entity, &confirm)
	return &confirm
}

// SetOracleSetConfirm sets a oracleSet confirmation
func (k Keeper) SetOracleSetConfirm(ctx sdk.Context, oracleAddr sdk.AccAddress, oracleSetConfirm *types.MsgOracleSetConfirm) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetOracleSetConfirmKey(oracleSetConfirm.Nonce, oracleAddr)
	store.Set(key, k.cdc.MustMarshal(oracleSetConfirm))
}

// GetOracleSetConfirms returns all oracle set confirmations by nonce
func (k Keeper) GetOracleSetConfirms(ctx sdk.Context, nonce uint64) (confirms []*types.MsgOracleSetConfirm) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.OracleSetConfirmKey)
	start, end := prefixRange(sdk.Uint64ToBigEndian(nonce))
	iterator := prefixStore.Iterator(start, end)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		confirm := types.MsgOracleSetConfirm{}
		k.cdc.MustUnmarshal(iterator.Value(), &confirm)
		confirms = append(confirms, &confirm)
	}

	return confirms
}

// IterateOracleSetConfirmByNonce iterates through all oracleSet confirms by nonce in ASC order
// MARK finish-batches: this is where the key is iterated in the old (presumed working) code
func (k Keeper) IterateOracleSetConfirmByNonce(ctx sdk.Context, nonce uint64, cb func([]byte, types.MsgOracleSetConfirm) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.OracleSetConfirmKey)
	iter := prefixStore.Iterator(prefixRange(sdk.Uint64ToBigEndian(nonce)))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		confirm := types.MsgOracleSetConfirm{}
		k.cdc.MustUnmarshal(iter.Value(), &confirm)
		// cb returns true to stop early
		if cb(iter.Key(), confirm) {
			break
		}
	}
}

// GetLastObservedOracleSet retrieves the last observed oracle set from the store
// WARNING: This value is not an up to date oracle set on Ethereum, it is a oracle set
// that AT ONE POINT was the one in the bridge on Ethereum. If you assume that it's up
// to date you may break the bridge
func (k Keeper) GetLastObservedOracleSet(ctx sdk.Context) *types.OracleSet {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LastObservedOracleSetKey)

	if len(bytes) == 0 {
		return nil
	}
	oracleSet := types.OracleSet{}
	k.cdc.MustUnmarshal(bytes, &oracleSet)
	return &oracleSet
}

// SetLastObservedOracleSet updates the last observed oracle set in the store
func (k Keeper) SetLastObservedOracleSet(ctx sdk.Context, oracleSet types.OracleSet) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastObservedOracleSetKey, k.cdc.MustMarshal(&oracleSet))
}
