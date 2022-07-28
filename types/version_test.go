package types

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetChainId_invalid(t *testing.T) {
	require.Equal(t, ChainId(), MainnetChainId)
	require.Panics(t, func() {
		SetChainId("hello")
	})
}

func TestSetChainId_mainnet(t *testing.T) {
	require.Equal(t, ChainId(), MainnetChainId)

	SetChainId(MainnetChainId)
	require.Equal(t, ChainId(), MainnetChainId)

	SetChainId(TestnetChainId)
	require.Equal(t, ChainId(), MainnetChainId)
}

func TestSetChainId_testnet(t *testing.T) {
	once = sync.Once{}
	require.Equal(t, ChainId(), MainnetChainId)

	SetChainId(TestnetChainId)
	require.Equal(t, ChainId(), TestnetChainId)

	SetChainId(MainnetChainId)
	require.Equal(t, ChainId(), TestnetChainId)
}
