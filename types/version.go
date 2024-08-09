package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
)

// mainnet
const (
	MainnetChainId       = "fxcore"
	mainnetEvmChainID    = 530
	MainnetGenesisHash   = "56629F685970FEC1E35521FC943ACE9AEB2C53448544A0560E4DD5799E1A5593"
	MainnetBlockHeightV2 = 5_713_000
	MainnetBlockHeightV3 = 8_756_000
	MainnetBlockHeightV4 = 10_477_500
	MainnetBlockHeightV5 = 11_601_700
	MainnetBlockHeightV6 = 13_598_000
	MainnetBlockHeightV7 = 16_838_000
)

// testnet
const (
	TestnetChainId        = "dhobyghaut"
	testnetEvmChainID     = 90001
	TestnetGenesisHash    = "06D0A9659E1EC5B0E57E8E2E5F1B1266094808BC9B4081E1A55011FEF4586ACE"
	TestnetBlockHeightV2  = 3_418_880
	TestnetBlockHeightV3  = 6_578_000
	TestnetBlockHeightV4  = 8_088_000
	TestnetBlockHeightV41 = 8_376_000 // v4.1
	TestnetBlockHeightV42 = 8_481_000 // v4.2
	TestnetBlockHeightV5  = 9_773_000
	TestnetBlockHeightV6  = 11_701_000
	TestnetBlockHeightV7  = 12_961_500
	TestnetBlockHeightV71 = 14_369_500 // v7.1
	TestnetBlockHeightV72 = 14_389_000 // v7.2
	TestnetBlockHeightV73 = 14_551_500 // v7.3
	TestnetBlockHeightV74 = 15_614_000 // v7.4
	TestnetBlockHeightV75 = 15_660_500 // v7.5
)

var (
	chainId = MainnetChainId
	once    sync.Once
)

func SetChainId(id string) {
	if id != MainnetChainId && id != TestnetChainId {
		panic("invalid chainId: " + id)
	}
	once.Do(func() {
		chainId = id
	})
}

func ChainId() string {
	return chainId
}

func EIP155ChainID() *big.Int {
	if TestnetChainId == ChainId() {
		return big.NewInt(testnetEvmChainID)
	}
	return big.NewInt(mainnetEvmChainID)
}

func ChainIdWithEIP155() string {
	if TestnetChainId == ChainId() {
		return fmt.Sprintf("%s_%d-1", TestnetChainId, testnetEvmChainID)
	}
	return fmt.Sprintf("%s_%d-1", MainnetChainId, mainnetEvmChainID)
}

// Sha256Hex calculate SHA-256 hash
func Sha256Hex(b []byte) string {
	sha := sha256.New()
	sha.Write(b)
	return strings.ToUpper(hex.EncodeToString(sha.Sum(nil)))
}
