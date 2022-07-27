package cli

import (
	"encoding/json"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

func GetCmdDenomToIBcDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "denom-convert",
		Short:   "Covert denom to ibc denom",
		Args:    cobra.ExactArgs(1),
		Example: "fxcored query ibc-transfer denom-convert transfer/channel-0/eth0x2170ed0880ac9a755fd29b2688956bd959f933f8",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			denomTrace := transfertypes.ParseDenomTrace(args[0])

			type output struct {
				Prefix   string
				Denom    string
				IBCDenom string
			}

			marshal, err := json.Marshal(output{
				Prefix:   denomTrace.GetPrefix(),
				Denom:    denomTrace.GetBaseDenom(),
				IBCDenom: denomTrace.IBCDenom(),
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintBytes(marshal)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
