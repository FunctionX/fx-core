package v020

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibchost "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	fxtypes "github.com/functionx/fx-core/v2/types"

	bsctypes "github.com/functionx/fx-core/v2/x/bsc/types"
	polygontypes "github.com/functionx/fx-core/v2/x/polygon/types"
	trontypes "github.com/functionx/fx-core/v2/x/tron/types"

	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	erc20types "github.com/functionx/fx-core/v2/x/erc20/types"
	migratetypes "github.com/functionx/fx-core/v2/x/migrate/types"
)

const (
	UpgradeName = "fxv2"

	blockParamsMaxGas = 30_000_000
)

var (
	initGenesis = map[string]bool{
		feemarkettypes.ModuleName: true,
		evmtypes.ModuleName:       true,
		erc20types.ModuleName:     true,
		migratetypes.ModuleName:   true,
		feegrant.ModuleName:       true,
		authz.ModuleName:          true,
	}

	runMigrates = map[string]uint64{
		authtypes.ModuleName:         1,
		banktypes.ModuleName:         1,
		distributiontypes.ModuleName: 1,
		govtypes.ModuleName:          1,
		slashingtypes.ModuleName:     1,
		stakingtypes.ModuleName:      1,
		ibchost.ModuleName:           1,
		bsctypes.ModuleName:          1,
		polygontypes.ModuleName:      1,
		trontypes.ModuleName:         1,
	}

	storeUpgradesDefault = &store.StoreUpgrades{
		Added: []string{
			feemarkettypes.StoreKey,
			evmtypes.StoreKey,
			erc20types.StoreKey,
			migratetypes.StoreKey,
			feegrant.StoreKey,
			authzkeeper.StoreKey,
		},
	}
	storeUpgradesTestnet = &store.StoreUpgrades{
		Added: []string{
			feegrant.StoreKey,
			authzkeeper.StoreKey,
		},
	}
)

func GetStoreUpgrades() *store.StoreUpgrades {
	if fxtypes.ChainId() == fxtypes.TestnetChainId {
		return storeUpgradesTestnet
	}
	return storeUpgradesDefault
}
