package store_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func Benchmark_Iterator(b *testing.B) {
	storeKey := sdk.NewKVStoreKey("test")
	ms := rootmulti.NewStore(dbm.NewMemDB(), log.NewNopLogger())
	ms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, nil)
	assert.NoError(b, ms.LoadLatestVersion())
	store := ms.GetKVStore(storeKey)

	count := 10000
	for i := 0; i < count; i++ {
		key := append([]byte{0x1}, sdk.Uint64ToBigEndian(uint64(i))...)
		store.Set(key, []byte{1, 2, 3})
	}

	b.Run("A", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var data [][]byte
			iter := sdk.KVStorePrefixIterator(store, []byte{0x1})
			for ; iter.Valid(); iter.Next() {
				iter.Value()
				data = append(data, iter.Value())
			}
			assert.Equal(b, count, len(data))
		}
	})

	b.Run("B", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var data [][]byte
			for j := 0; j < count; j++ {
				key := append([]byte{0x1}, sdk.Uint64ToBigEndian(uint64(i))...)
				data = append(data, store.Get(key))
			}
			assert.Equal(b, count, len(data))
		}
	})
}
