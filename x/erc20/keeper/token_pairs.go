package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	fxtypes "github.com/functionx/fx-core/v2/types"

	"github.com/functionx/fx-core/v2/x/erc20/types"
)

// GetAllTokenPairs - get all registered token tokenPairs
func (k Keeper) GetAllTokenPairs(ctx sdk.Context) []types.TokenPair {
	tokenPairs := []types.TokenPair{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixTokenPair)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tokenPair types.TokenPair
		k.cdc.MustUnmarshal(iterator.Value(), &tokenPair)

		tokenPairs = append(tokenPairs, tokenPair)
	}

	return tokenPairs
}

// GetTokenPairID returns the pair id from either of the registered tokens.
func (k Keeper) GetTokenPairID(ctx sdk.Context, token string) []byte {
	if common.IsHexAddress(token) {
		addr := common.HexToAddress(token)
		return k.GetERC20Map(ctx, addr)
	}
	return k.GetDenomMap(ctx, token)
}

// GetTokenPair - get registered token pair from the identifier
func (k Keeper) GetTokenPair(ctx sdk.Context, id []byte) (types.TokenPair, bool) {
	if id == nil {
		return types.TokenPair{}, false
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPair)
	var tokenPair types.TokenPair
	bz := store.Get(id)
	if len(bz) == 0 {
		return types.TokenPair{}, false
	}

	k.cdc.MustUnmarshal(bz, &tokenPair)
	return tokenPair, true
}

// SetTokenPair stores a token pair
func (k Keeper) SetTokenPair(ctx sdk.Context, tokenPair types.TokenPair) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPair)
	key := tokenPair.GetID()
	bz := k.cdc.MustMarshal(&tokenPair)
	store.Set(key, bz)
}

// RemoveTokenPair removes a token pair.
func (k Keeper) RemoveTokenPair(ctx sdk.Context, tokenPair types.TokenPair) {
	id := tokenPair.GetID()
	k.DeleteTokenPair(ctx, id)
	k.DeleteERC20Map(ctx, tokenPair.GetERC20Contract())
	k.DeleteDenomMap(ctx, tokenPair.Denom)

	//after support many to one, delete denom with alias
	if ctx.BlockHeight() >= fxtypes.UpgradeExponential1Block() {
		md, found := k.bankKeeper.GetDenomMetaData(ctx, tokenPair.Denom)
		if found && types.IsManyToOneMetadata(md) {
			k.DeleteAliasesDenom(ctx, md.DenomUnits[0].Aliases...)
		}
	}
}

// DeleteTokenPair deletes the token pair for the given id
func (k Keeper) DeleteTokenPair(ctx sdk.Context, id []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPair)
	store.Delete(id)
}

// GetERC20Map returns the token pair id for the given address
func (k Keeper) GetERC20Map(ctx sdk.Context, erc20 common.Address) []byte {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPairByERC20)
	return store.Get(erc20.Bytes())
}

// GetDenomMap returns the token pair id for the given denomination
func (k Keeper) GetDenomMap(ctx sdk.Context, denom string) []byte {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPairByDenom)
	return store.Get([]byte(denom))
}

// SetERC20Map sets the token pair id for the given address
func (k Keeper) SetERC20Map(ctx sdk.Context, erc20 common.Address, id []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPairByERC20)
	store.Set(erc20.Bytes(), id)
}

// DeleteERC20Map deletes the token pair id for the given address
func (k Keeper) DeleteERC20Map(ctx sdk.Context, erc20 common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPairByERC20)
	store.Delete(erc20.Bytes())
}

// SetDenomMap sets the token pair id for the denomination
func (k Keeper) SetDenomMap(ctx sdk.Context, denom string, id []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPairByDenom)
	store.Set([]byte(denom), id)
}

// DeleteDenomMap deletes the token pair id for the given denom
func (k Keeper) DeleteDenomMap(ctx sdk.Context, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPairByDenom)
	store.Delete([]byte(denom))
}

// IsTokenPairRegistered - check if registered token tokenPair is registered
func (k Keeper) IsTokenPairRegistered(ctx sdk.Context, id []byte) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPair)
	return store.Has(id)
}

// IsERC20Registered check if registered ERC20 token is registered
func (k Keeper) IsERC20Registered(ctx sdk.Context, erc20 common.Address) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPairByERC20)
	return store.Has(erc20.Bytes())
}

// IsDenomRegistered check if registered coin denom is registered
func (k Keeper) IsDenomRegistered(ctx sdk.Context, denom string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPairByDenom)
	return store.Has([]byte(denom))
}

// SetAliasesDenom sets the aliases for the denomination
func (k Keeper) SetAliasesDenom(ctx sdk.Context, denom string, aliases ...string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAliasDenom)
	for _, alias := range aliases {
		store.Set([]byte(alias), []byte(denom))
	}
}

// GetAliasDenom returns the denom for the given alias
func (k Keeper) GetAliasDenom(ctx sdk.Context, alias string) []byte {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAliasDenom)
	return store.Get([]byte(alias))
}

// DeleteAliasesDenom deletes the denom-alias for the given alias
func (k Keeper) DeleteAliasesDenom(ctx sdk.Context, aliases ...string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAliasDenom)
	for _, alias := range aliases {
		store.Delete([]byte(alias))
	}
}

// IsAliasDenomRegistered check if registered coin alias is registered
func (k Keeper) IsAliasDenomRegistered(ctx sdk.Context, alias string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAliasDenom)
	return store.Has([]byte(alias))
}
