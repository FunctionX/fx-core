package types

import (
	"math/big"
	"os"
	"sync"
)

// mainnet constant
const (

	/*
		mainnetCrossChainSupportBscBlock            = 1354000
		mainnetCrossChainSupportTronAndPolygonBlock = 2062000
		// gravity prune validator set
		mainnetGravityPruneValsetAndAttestationBlock = 610000
		// gravity not slash no set eth address validator
		mainnetGravityValsetSlashBlock = 1685000
	*/

	MainnetChainId    = "fxcore"
	mainnetEvmChainID = 530

	mainnetSupportDenomManyToOneBlock = 5940000
)

// testnet constant
const (
	TestnetChainId        = "dhobyghaut"
	testnetEvmChainID     = 90001
	testnetIBCRouterBlock = 3433511

	testnetSupportDenomManyToOneBlock = 3918000
	testnetSupportDenomOneToManyBlock = 4140000
)

// SupportDenomManyToOneMsgTypes return msg types
// use method return to prevent accidental modification
func SupportDenomManyToOneMsgTypes() []string {
	return []string{
		"/fx.erc20.v1.MsgConvertDenom",
		"/fx.erc20.v1.UpdateDenomAliasProposal",
	}
}

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
	if TestnetChainId == chainId {
		return big.NewInt(testnetEvmChainID)
	}
	return big.NewInt(mainnetEvmChainID)
}

func IBCRouteBlock() int64 {
	if TestnetChainId == chainId {
		return testnetIBCRouterBlock
	}
	return 0
}

func SetTestingManyToOneBlock(fn func() int64) {
	if os.Getenv("GO_ENV") != "testing" {
		panic("invalid env")
	}
	testingManyToOneBlock = fn
}

var testingManyToOneBlock func() int64

func SupportDenomManyToOneBlock() int64 {
	if os.Getenv("GO_ENV") == "testing" {
		return testingManyToOneBlock()
	}
	if TestnetChainId == chainId {
		return testnetSupportDenomManyToOneBlock
	}
	return mainnetSupportDenomManyToOneBlock
}

func SetTestingSupportDenomOneToManyBlock(fn func() int64) {
	if os.Getenv("GO_ENV") != "testing" {
		panic("invalid env")
	}
	testingSupportDenomOneToManyBlock = fn
}

var testingSupportDenomOneToManyBlock func() int64

func SupportDenomOneToManyBlock() int64 {
	if os.Getenv("GO_ENV") == "testing" {
		return testingSupportDenomOneToManyBlock()
	}
	if TestnetChainId == chainId {
		return testnetSupportDenomOneToManyBlock
	}
	return mainnetSupportDenomManyToOneBlock
}
