package cli

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/node"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
)

func QueryGasPricesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gas-prices",
		Short: "Query node gas prices",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := node.NewServiceClient(clientCtx)
			res, err := queryClient.Config(context.Background(), &node.ConfigRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func QueryStoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store <store name> <hex key>",
		Short: "Query for a blockchain store",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			hexKey, err := hex.DecodeString(args[1])
			if err != nil {
				return err
			}
			data, height, err := clientCtx.QueryStore(hexKey, args[0])
			if err != nil {
				return err
			}
			raw, err := json.Marshal(map[string]interface{}{
				"block_height": height,
				"data":         hex.EncodeToString(data),
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintRaw(raw)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func QueryValidatorByConsAddr() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator [validator-consAddr]",
		Short: "Query details about an individual validator cons address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			consAddr, err := sdk.ConsAddressFromBech32(args[0])
			if err != nil {
				consAddr, err = hex.DecodeString(args[0])
				if err != nil {
					return errors.New("expected hex or bech32 address")
				}
			}
			opAddr, _, err := clientCtx.QueryStore(types.GetValidatorByConsAddrKey(consAddr), types.StoreKey)
			if err != nil {
				return err
			}
			if opAddr == nil {
				return fmt.Errorf("not found validator by consAddress: %s", consAddr.String())
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Validator(context.Background(), &types.QueryValidatorRequest{
				ValidatorAddr: sdk.ValAddress(opAddr).String(),
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(&res.Validator)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
