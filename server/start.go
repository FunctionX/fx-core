package server

import (
	"errors"
	"fmt"
	"os"
	"time"

	tmcfg "github.com/cometbft/cometbft/config"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/store"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	serverlog "github.com/cosmos/cosmos-sdk/server/log"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	ethermintserver "github.com/evmos/ethermint/server"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	fxtypes "github.com/functionx/fx-core/v7/types"
)

// StartCmd runs the service passed in, either stand-alone or in-process with
// CometBFT.
func StartCmd(appCreator types.AppCreator, defaultNodeHome string) *cobra.Command {
	startCmd := ethermintserver.StartCmd(ethermintserver.NewDefaultStartOptions(appCreator, defaultNodeHome))
	startCmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		serverCtx := server.GetServerContextFromCmd(cmd)

		if _, err := server.GetPruningOptionsFromFlags(serverCtx.Viper); err != nil {
			return err
		}

		genDocFile := serverCtx.Config.GenesisFile()
		genesisBytes, err := os.ReadFile(genDocFile)
		if err != nil {
			return fmt.Errorf("couldn't read GenesisDoc file: %w", err)
		}
		expectGenesisHash := serverCtx.Viper.GetString("genesis_hash")
		actualGenesisHash := fxtypes.Sha256Hex(genesisBytes)
		if len(expectGenesisHash) != 0 && fxtypes.Sha256Hex(genesisBytes) != expectGenesisHash {
			return fmt.Errorf("--genesis_hash=%s does not match %s hash: %s", expectGenesisHash, genDocFile, actualGenesisHash)
		}
		genesisDoc, err := tmtypes.GenesisDocFromJSON(genesisBytes)
		if err != nil {
			return fmt.Errorf("error reading GenesisDoc at %s: %w", genDocFile, err)
		}
		if err = checkMainnetAndBlock(genesisDoc, serverCtx.Config); err != nil {
			return err
		}
		fxtypes.SetChainId(genesisDoc.ChainID)

		filterLogTypes := serverCtx.Viper.GetStringSlice(FlagLogFilter)
		zeroLog := serverCtx.Logger.(serverlog.CometLoggerWrapper).Logger.Impl().(*zerolog.Logger)
		serverCtx.Logger = NewFxZeroLogWrapper(zeroLog, filterLogTypes)

		clientCtx := client.GetClientContextFromCmd(cmd)
		if len(clientCtx.ChainID) == 0 {
			clientCtx.ChainID = fxtypes.ChainId()
		}
		if len(clientCtx.HomeDir) == 0 {
			clientCtx.HomeDir = serverCtx.Config.RootDir
		}
		if err = client.SetCmdClientContext(cmd, clientCtx); err != nil {
			return err
		}
		serverCtx.Viper.Set(flags.FlagChainID, fxtypes.ChainIdWithEIP155())
		if err = server.SetCmdServerContext(cmd, serverCtx); err != nil {
			return err
		}
		return nil
	}
	startCmd.Flags().StringSlice(FlagLogFilter, nil, `The logging filter can discard custom log type (ABCIQuery)`)
	crisis.AddModuleInitFlags(startCmd)
	return startCmd
}

func checkMainnetAndBlock(genesisDoc *tmtypes.GenesisDoc, config *tmcfg.Config) error {
	if genesisDoc.InitialHeight > 1 || genesisDoc.ChainID != fxtypes.MainnetChainId || config.StateSync.Enable {
		return nil
	}
	genesisTime, err := time.Parse("2006-01-02T15:04:05Z", "2021-07-05T04:00:00Z")
	if err != nil {
		return err
	}
	blockStoreDB, err := node.DefaultDBProvider(&node.DBContext{ID: "blockstore", Config: config})
	if err != nil {
		return err
	}
	defer func() {
		_ = blockStoreDB.Close()
	}()
	blockStore := store.NewBlockStore(blockStoreDB)
	if genesisDoc.GenesisTime.Equal(genesisTime) {
		genesisBytes, _ := tmjson.Marshal(genesisDoc)
		if fxtypes.Sha256Hex(genesisBytes) != fxtypes.MainnetGenesisHash {
			return nil
		}
		if blockStore.Height() < fxtypes.MainnetBlockHeightV2 {
			return errors.New("invalid version: The current block height is less than the fxv2 upgrade height(5_713_000), " +
				"sync block from scratch please use use fxcored v1.x.x")
		}
		if blockStore.Height() < fxtypes.MainnetBlockHeightV3 {
			return errors.New("invalid version: The current block height is less than the fxv3 upgrade height(8_756_000)," +
				" please use the v2.x.x version to synchronize the block or download the latest snapshot")
		}
		if blockStore.Height() < fxtypes.MainnetBlockHeightV4 {
			return errors.New("invalid version: The current block height is less than the v4.2.0 upgrade height(10_477_500)," +
				" please use the v3.x.x version to synchronize the block or download the latest snapshot")
		}
		if blockStore.Height() < fxtypes.MainnetBlockHeightV5 {
			return errors.New("invalid version: The current block height is less than the v5.0.0 upgrade height(11_601_700)," +
				" please use the v4.x.x version to synchronize the block or download the latest snapshot")
		}
		if blockStore.Height() < fxtypes.MainnetBlockHeightV6 {
			return errors.New("invalid version: The current block height is less than the v6.0.0 upgrade height(13_598_000)," +
				" please use the v5.x.x version to synchronize the block or download the latest snapshot")
		}
		if blockStore.Height() < fxtypes.MainnetBlockHeightV7 {
			return errors.New("invalid version: The current block height is less than the v7.5.0 upgrade height(16_838_000)," +
				" please use the v6.x.x version to synchronize the block or download the latest snapshot")
		}
		return errors.New("invalid version: The current version is not released, please use the corresponding version")
	}
	return nil
}
