package app_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/functionx/fx-core/v7/app"
	v7 "github.com/functionx/fx-core/v7/app/upgrades/v7"
	"github.com/functionx/fx-core/v7/testutil/helpers"
	fxtypes "github.com/functionx/fx-core/v7/types"
	"github.com/functionx/fx-core/v7/x/crosschain/types"
	ethtypes "github.com/functionx/fx-core/v7/x/eth/types"
	layer2types "github.com/functionx/fx-core/v7/x/layer2/types"
)

func Test_TestnetUpgrade(t *testing.T) {
	helpers.SkipTest(t, "Skipping local test: ", t.Name())

	fxtypes.SetConfig(true)
	fxtypes.SetChainId(fxtypes.TestnetChainId) // only for testnet

	testCases := []struct {
		name                  string
		fromVersion           int
		toVersion             int
		LocalStoreBlockHeight uint64
		plan                  upgradetypes.Plan
	}{
		{
			name:        "upgrade v7",
			fromVersion: 6,
			toVersion:   7,
			plan: upgradetypes.Plan{
				Name: v7.Upgrade.UpgradeName,
				Info: "local test upgrade v7",
			},
		},
	}

	db, err := dbm.NewDB("application", dbm.GoLevelDBBackend, filepath.Join(os.Getenv("HOME"), "tmp/data"))
	require.NoError(t, err)

	makeEncodingConfig := app.MakeEncodingConfig()
	myApp := app.New(log.NewFilter(log.NewTMLogger(os.Stdout), log.AllowAll()),
		db, nil, false, map[int64]bool{}, fxtypes.GetDefaultNodeHome(), 0,
		makeEncodingConfig, app.EmptyAppOptions{})
	myApp.SetStoreLoader(upgradetypes.UpgradeStoreLoader(myApp.LastBlockHeight()+1, v7.Upgrade.StoreUpgrades()))
	err = myApp.LoadLatestVersion()
	require.NoError(t, err)

	ctx := newContext(t, myApp)
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.plan.Height = ctx.BlockHeight()
			myApp.UpgradeKeeper.ApplyUpgrade(ctx, testCase.plan)

			myApp.EthKeeper.IterateAttestations(ctx, func(attestation *types.Attestation) bool {
				if _, err = types.UnpackAttestationClaim(makeEncodingConfig.Codec, attestation); err != nil {
					t.Log(ethtypes.ModuleName, attestation.Claim.TypeUrl, attestation.Observed, attestation.Height)
					require.True(t,
						attestation.Claim.TypeUrl == sdk.MsgTypeURL(&types.MsgBridgeCallClaim{}) ||
							attestation.Claim.TypeUrl == sdk.MsgTypeURL(&types.MsgBridgeCallResultClaim{}))
				}
				return false
			})

			myApp.Layer2Keeper.IterateAttestations(ctx, func(attestation *types.Attestation) bool {
				if _, err = types.UnpackAttestationClaim(makeEncodingConfig.Codec, attestation); err != nil {
					t.Log(layer2types.ModuleName, attestation.Claim.TypeUrl, attestation.Observed, attestation.Height)
				}
				return false
			})
		})
	}
}

func newContext(t *testing.T, myApp *app.App) sdk.Context {
	chainId := fxtypes.MainnetChainId
	if os.Getenv("CHAIN_ID") == fxtypes.TestnetChainId {
		chainId = fxtypes.TestnetChainId
	}
	ctx := myApp.NewUncachedContext(false, tmproto.Header{
		ChainID: chainId, Height: myApp.LastBlockHeight(),
	})
	// set the first validator to proposer
	validators := myApp.StakingKeeper.GetAllValidators(ctx)
	assert.True(t, len(validators) > 0)
	var pubKey cryptotypes.PubKey
	assert.NoError(t, myApp.AppCodec().UnpackAny(validators[0].ConsensusPubkey, &pubKey))
	ctx = ctx.WithProposer(pubKey.Address().Bytes())
	return ctx
}
