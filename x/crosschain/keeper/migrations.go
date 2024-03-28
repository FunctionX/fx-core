package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/functionx/fx-core/v7/x/crosschain/types"
)

type Migrator struct {
	keeper Keeper
}

func NewMigrator(k Keeper) Migrator {
	return Migrator{
		keeper: k,
	}
}

func (m Migrator) Migrate(ctx sdk.Context) error {
	params := m.keeper.GetParams(ctx)
	params.BridgeCallRefundTimeout = types.DefaultBridgeCallRefundTimeout
	if err := m.keeper.SetParams(ctx, &params); err != nil {
		return err
	}
	addBridgeTokenType(ctx, m.keeper, types.BRIDGE_TOKEN_TYPE_ERC20)
	return nil
}

func addBridgeTokenType(ctx sdk.Context, k Keeper, tokenType types.BridgeTokenType) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DenomToTokenKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		token := string(iterator.Key()[len(types.DenomToTokenKey):])
		store.Set(types.GetTokenTypeToTokenKey(token), sdk.Uint64ToBigEndian(uint64(tokenType)))
	}
}
