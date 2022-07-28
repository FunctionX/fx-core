package main

import (
	"errors"
	"time"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/store"
	tmtypes "github.com/tendermint/tendermint/types"

	fxtypes "github.com/functionx/fx-core/v2/types"

	"github.com/functionx/fx-core/v2/app/cli"

	sdkserver "github.com/cosmos/cosmos-sdk/server"

	tmcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"

	"github.com/evmos/ethermint/server"

	appCmd "github.com/functionx/fx-core/v2/app/cli"
)

func addTendermintCommands(rootCmd *cobra.Command, defaultNodeHome string, appCreator types.AppCreator, appExport types.AppExporter) {
	tendermintCmd := &cobra.Command{
		Use:   "tendermint",
		Short: "Tendermint subcommands",
	}

	tmRestCmdPreRun := func(cmd *cobra.Command, args []string) {
		serverCtx := sdkserver.GetServerContextFromCmd(cmd)
		globalViper := viper.GetViper()
		for _, s := range serverCtx.Viper.AllKeys() {
			globalViper.Set(s, serverCtx.Viper.Get(s))
		}
	}

	tmcmd.ResetStateCmd.PreRun = tmRestCmdPreRun
	tmcmd.ResetAllCmd.PreRun = tmRestCmdPreRun

	tendermintCmd.AddCommand(
		sdkserver.ShowNodeIDCmd(),
		sdkserver.ShowValidatorCmd(),
		sdkserver.ShowAddressCmd(),
		sdkserver.VersionCmd(),
		appCmd.UnsafeRestPrivValidatorCmd(),
		appCmd.UnsafeResetNodeKeyCmd(),
		appCmd.RollbackStateCmd(),
		appCmd.ReplayCmd(),
		appCmd.ReplayConsoleCmd(),
		tmcmd.ResetAllCmd,
		tmcmd.ResetStateCmd,
		tmcmd.GenValidatorCmd,
		tmcmd.GenNodeKeyCmd,
		//tmcmd.ResetPrivValidatorCmd,
	)

	startCmd := server.StartCmd(appCreator, defaultNodeHome)
	crisis.AddModuleInitFlags(startCmd)
	preRun := startCmd.PreRunE
	startCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if err := preRun(cmd, args); err != nil {
			return err
		}
		serverCtx := sdkserver.GetServerContextFromCmd(cmd)
		genesisDoc, err := tmtypes.GenesisDocFromFile(serverCtx.Config.GenesisFile())
		if err != nil {
			return err
		}
		if err = checkMainnetAndBlock(genesisDoc, serverCtx.Config); err != nil {
			return err
		}
		fxtypes.SetChainId(genesisDoc.ChainID)
		return nil
	}
	rootCmd.AddCommand(
		startCmd,
		tendermintCmd,
		cli.ExportSateCmd(appExport, defaultNodeHome),
		version.NewVersionCommand(),
		sdkserver.NewRollbackCmd(defaultNodeHome),
	)
}

func checkMainnetAndBlock(genesisDoc *tmtypes.GenesisDoc, config *config.Config) error {
	genesisTime, err := time.Parse("2006-01-02T15:04:05Z", "2021-07-05T04:00:00Z")
	if err != nil {
		return err
	}
	blockStoreDB, err := node.DefaultDBProvider(&node.DBContext{ID: "blockstore", Config: config})
	if err != nil {
		return err
	}
	defer blockStoreDB.Close()
	blockStore := store.NewBlockStore(blockStoreDB)
	if genesisDoc.GenesisTime.Equal(genesisTime) && blockStore.Height() <= 0 {
		return errors.New("invalid version: Sync block from scratch please use use fxcored v1.1.x")
	}
	return nil
}
