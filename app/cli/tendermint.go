package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"
	tmcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/consensus"
)

func RollbackStateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback",
		Short: "rollback tendermint state by one height",
		Long: `
A state rollback is performed to recover from an incorrect application state transition,
when Tendermint has persisted an incorrect app hash and is thus unable to make
progress. Rollback overwrites a state at height n with the state at height n - 1.
The application should also roll back to height n - 1. No blocks are removed, so upon
restarting Tendermint the transactions in block n will be re-executed against the
application.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			cfg.EnsureRoot(serverCtx.Config.RootDir)
			fmt.Println(serverCtx.Config.RootDir)
			height, hash, err := tmcmd.RollbackState(serverCtx.Config)
			if err != nil {
				return fmt.Errorf("failed to rollback state: %w", err)
			}

			fmt.Printf("Rolled back state to height %d and hash %X\n", height, hash)
			return nil
		},
	}
}

// ReplayCmd allows replaying of messages from the WAL.
func ReplayCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "replay",
		Short: "Replay messages from WAL",
		Run: func(cmd *cobra.Command, args []string) {
			serverCtx := server.GetServerContextFromCmd(cmd)
			cfg.EnsureRoot(serverCtx.Config.RootDir)

			consensus.RunReplayFile(serverCtx.Config.BaseConfig, serverCtx.Config.Consensus, false)
		},
	}
}

// ReplayConsoleCmd allows replaying of messages from the WAL in a
// console.
func ReplayConsoleCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "replay-console",
		Aliases: []string{"replay_console"},
		Short:   "Replay messages from WAL in a console",
		Run: func(cmd *cobra.Command, args []string) {
			serverCtx := server.GetServerContextFromCmd(cmd)
			cfg.EnsureRoot(serverCtx.Config.RootDir)
			consensus.RunReplayFile(serverCtx.Config.BaseConfig, serverCtx.Config.Consensus, true)
		},
	}
}
