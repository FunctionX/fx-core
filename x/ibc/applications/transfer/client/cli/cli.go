package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	transfercli "github.com/cosmos/ibc-go/v3/modules/apps/transfer/client/cli"
)

// GetQueryCmd returns the query commands for IBC connections
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        "fx-ibc-transfer",
		Short:                      "IBC fungible token transfer query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	queryCmd.AddCommand(
		transfercli.GetCmdQueryDenomTrace(),
		transfercli.GetCmdQueryDenomTraces(),
		transfercli.GetCmdParams(),
		GetCmdDenomToIBcDenom(),
		transfercli.GetCmdQueryEscrowAddress(),
		transfercli.GetCmdQueryDenomHash(),
	)

	return queryCmd
}

// NewTxCmd returns the transaction commands for IBC fungible token transfer
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "fx-ibc-transfer",
		Short:                      "IBC fungible token transfer transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewTransferTxCmd(),
	)

	return txCmd
}
