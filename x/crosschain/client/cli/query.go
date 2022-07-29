package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/ethereum/go-ethereum/crypto"

	fxtypes "github.com/functionx/fx-core/v2/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	abcitype "github.com/tendermint/tendermint/abci/types"

	"github.com/functionx/fx-core/v2/x/crosschain/types"
)

const (
	flagChannelIBC = "channelIBC"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands from cross chain",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand([]*cobra.Command{
		// query module params
		CmdGetParams(),

		// query Oracle
		CmdGetOracles(),
		CmdGetOracleReward(),
		CmdGetChainOracles(),
		CmdGetOracleByAddr(),
		CmdGetOracleByBridgerAddr(),
		CmdGetOracleByExternalAddr(),

		// query oracle set
		CmdGetCurrentOracleSet(),
		CmdGetOracleSetRequest(),

		// need oracle consensus sign
		// 1. oracle set change confirm
		CmdGetLastOracleSetRequests(),
		CmdGetPendingOracleSetRequest(),
		CmdGetOracleSetConfirm(),
		CmdGetOracleSetConfirms(),

		// 2. request batch confirm
		CmdGetPendingOutgoingTXBatchRequest(),
		CmdBatchConfirm(),
		CmdBatchConfirms(),

		// send to external
		CmdBatchRequestByNonce(),
		CmdGetPendingSendToExternal(),
		CmdOutgoingTxBatches(),
		CmdGetBatchFees(),

		CmdGetLastObservedBlockHeight(),
		CmdProjectedBatchTimeoutHeight(),

		// denom <-> external token
		CmdGetDenomToExternalToken(),
		CmdGetExternalTokenToDenom(),
		CmdGetBridgeTokens(),

		// 1. oracle event nonce
		CmdGetOracleEventNonce(),
		// 2. event nonce block height
		CmdGetOracleEventBlockHeight(),

		// help cmd.
		CmdCovertBridgeToken(),
	}...)

	for _, command := range cmd.Commands() {
		flags.AddQueryFlagsToCmd(command)
	}
	return cmd
}

func CmdGetParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params [chain-name]",
		Short: "Query the current parameters information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)
			chainName := args[0]
			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{ChainName: chainName})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(&res.Params)
		},
	}

	return cmd
}

func CmdGetOracles() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracles [chain-name]",
		Short: "Query oracles",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Oracles(cmd.Context(), &types.QueryOraclesRequest{
				ChainName: args[0],
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

func CmdGetChainOracles() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain-oracles [chain-name]",
		Short: "Query active oracles address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			req := abcitype.RequestQuery{
				Data: types.ProposalOracleKey,
				Path: fmt.Sprintf("/store/%s/key", args[0]),
			}
			abciResp, err := clientCtx.QueryABCI(req)
			if err != nil {
				return err
			}

			var proposalOracle types.ProposalOracle
			if err := clientCtx.LegacyAmino.Unmarshal(abciResp.Value, &proposalOracle); err != nil {
				return err
			}
			return clientCtx.PrintProto(&proposalOracle)
		},
	}

	return cmd
}

func CmdGetOracleReward() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reward [chain-name] [oracle-address]",
		Short: "Query oracle reward",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			oracleAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			data := append(oracleAddr.Bytes(), []byte(args[0])...)
			queryClient := distrtypes.NewQueryClient(clientCtx)
			rewards, err := queryClient.DelegationTotalRewards(cmd.Context(), &distrtypes.QueryDelegationTotalRewardsRequest{
				DelegatorAddress: sdk.AccAddress(crypto.Keccak256(data)[12:]).String(),
			})
			if err != nil {
				return err
			}
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(rewards)
		},
	}

	return cmd
}

func CmdGetOracleByAddr() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle-by-addr [chain-name] [oracle-address]",
		Short: "Query Oracle for a given oracle address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			oracleAddress, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			req := &types.QueryOracleByAddrRequest{
				OracleAddress: oracleAddress.String(),
				ChainName:     args[0],
			}
			res, err := queryClient.GetOracleByAddr(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

func CmdGetOracleByBridgerAddr() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle-by-bridger [chain-name] [bridger-address]",
		Short: "Query Oracle for a given bridger address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			bridgerAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			req := &types.QueryOracleByBridgerAddrRequest{
				BridgerAddress: bridgerAddr.String(),
				ChainName:      args[0],
			}
			res, err := queryClient.GetOracleByBridgerAddr(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

func CmdGetOracleByExternalAddr() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle-key-by-external-addr [chain-name] [external-address]",
		Short: "Query Oracle by external address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			externalAddress := args[1]
			if strings.HasPrefix(externalAddress, "0x") {
				if !gethCommon.IsHexAddress(externalAddress) {
					return fmt.Errorf("target address is invalid!address: [%s]", externalAddress)
				}
				externalAddress = gethCommon.HexToAddress(externalAddress).Hex()
			}

			res, err := queryClient.GetOracleByExternalAddr(cmd.Context(), &types.QueryOracleByExternalAddrRequest{
				ExternalAddress: externalAddress,
				ChainName:       args[0],
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdGetCurrentOracleSet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current-oracle-set [chain-name]",
		Short: "Query current oracle-set",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.CurrentOracleSet(cmd.Context(), &types.QueryCurrentOracleSetRequest{
				ChainName: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

func CmdGetOracleSetRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle-set-request [chain-name] [nonce]",
		Short: "Query requested oracle-set with a particular nonce",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			var nonce uint64
			if len(args) == 1 {
				queryAbciResp, err := clientCtx.QueryABCI(abcitype.RequestQuery{
					Path: fmt.Sprintf("store/%s/key", args[0]),
					Data: types.LatestOracleSetNonce,
				})
				if err != nil {
					return err
				}
				nonce = sdk.BigEndianToUint64(queryAbciResp.Value)
				if err = clientCtx.PrintString(fmt.Sprintf("latest oracle-set nonce:[%d]\n", nonce)); err != nil {
					return err
				}
			} else {
				var err error
				nonce, err = strconv.ParseUint(args[1], 10, 64)
				if err != nil {
					return err
				}
			}
			res, err := queryClient.OracleSetRequest(cmd.Context(), &types.QueryOracleSetRequestRequest{
				ChainName: args[0],
				Nonce:     nonce,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdGetLastOracleSetRequests() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last-oracle-set-requests [chain-name]",
		Short: "Query last oracle set requests",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryLastOracleSetRequestsRequest{
				ChainName: args[0],
			}

			res, err := queryClient.LastOracleSetRequests(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

func CmdGetPendingOracleSetRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-oracle-set-request [chain-name] [bridger]",
		Short: "Query the latest oracle-set request which has not been signed by a particular oracle bridger",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			bridgerAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			req := &types.QueryLastPendingOracleSetRequestByAddrRequest{
				BridgerAddress: bridgerAddr.String(),
				ChainName:      args[0],
			}

			res, err := queryClient.LastPendingOracleSetRequestByAddr(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

func CmdGetOracleSetConfirm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle-set-confirm [chain-name] [nonce] [bridger-address]",
		Short: "Query oracle-set confirmation with a particular nonce from a particular oracle bridger",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			nonce, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}
			bridgerAddr, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return err
			}
			res, err := queryClient.OracleSetConfirm(cmd.Context(), &types.QueryOracleSetConfirmRequest{
				Nonce:          nonce,
				BridgerAddress: bridgerAddr.String(),
				ChainName:      args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdGetOracleSetConfirms() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle-set-confirms [chain-name] [nonce]",
		Short: "Query oracle-set confirmations with a particular nonce",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			nonce, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}
			res, err := queryClient.OracleSetConfirmsByNonce(cmd.Context(), &types.QueryOracleSetConfirmsByNonceRequest{
				Nonce:     nonce,
				ChainName: args[0],
			})
			if err != nil {
				return err
			}
			if err = clientCtx.PrintString(fmt.Sprintf("confirm size:[%d]\n", len(res.Confirms))); err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdGetPendingOutgoingTXBatchRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-batch-request [chain-name] [bridger-address]",
		Short: "Query the latest outgoing TX batch request which has not been signed by a particular oracle bridger address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			bridgerAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			res, err := queryClient.LastPendingBatchRequestByAddr(cmd.Context(), &types.QueryLastPendingBatchRequestByAddrRequest{
				BridgerAddress: bridgerAddr.String(),
				ChainName:      args[0],
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdBatchConfirm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-confirm [chain-name] [token-contract] [nonce] [bridger-address]",
		Short: "Query outgoing tx batches confirm by oracle bridger address",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			tokenContract := args[1]
			if strings.HasPrefix(tokenContract, "0x") {
				if !gethCommon.IsHexAddress(tokenContract) {
					return fmt.Errorf("contract address is invalid!address:[%s]", tokenContract)
				}
				tokenContract = gethCommon.HexToAddress(tokenContract).Hex()
			}

			nonce, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}
			bridgerAddr, err := sdk.AccAddressFromBech32(args[3])
			if err != nil {
				return err
			}
			res, err := queryClient.BatchConfirm(cmd.Context(), &types.QueryBatchConfirmRequest{
				ChainName:      args[0],
				TokenContract:  tokenContract,
				Nonce:          nonce,
				BridgerAddress: bridgerAddr.String(),
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdBatchConfirms() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-confirms [chain-name] [token-contract] [nonce]",
		Short: "Query outgoing tx batches confirms",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			tokenContract := args[1]
			if strings.HasPrefix(tokenContract, "0x") {
				if !gethCommon.IsHexAddress(tokenContract) {
					return fmt.Errorf("contract address is invalid!address:[%s]", tokenContract)
				}
				tokenContract = gethCommon.HexToAddress(tokenContract).Hex()
			}
			nonce, err := strconv.Atoi(args[2])
			if err != nil {
				return err
			}
			res, err := queryClient.BatchConfirms(cmd.Context(), &types.QueryBatchConfirmsRequest{
				TokenContract: tokenContract,
				Nonce:         uint64(nonce),
				ChainName:     args[0],
			})
			if err != nil {
				return err
			}
			if err = clientCtx.PrintString(fmt.Sprintf("confirm size:[%d]\n", len(res.Confirms))); err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdBatchRequestByNonce() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-request [chain-name] [token-contract] [nonce]",
		Short: "Query outgoing tx batches",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			nonce, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.BatchRequestByNonce(cmd.Context(), &types.QueryBatchRequestByNonceRequest{
				ChainName:     args[0],
				TokenContract: args[1],
				Nonce:         nonce,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdGetPendingSendToExternal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-send-to-external [chain-name] [address]",
		Short: "Query pending send to external txs",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)
			chainName := args[0]
			if err := types.ValidateModuleName(chainName); err != nil {
				return err
			}
			fxAddr := args[1]
			if _, err := sdk.AccAddressFromBech32(fxAddr); err != nil {
				return nil
			}
			res, err := queryClient.GetPendingSendToExternal(cmd.Context(), &types.QueryPendingSendToExternalRequest{
				ChainName:     chainName,
				SenderAddress: fxAddr,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdOutgoingTxBatches() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outgoing-tx-batches [chain-name]",
		Short: "Query outgoing tx batches",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.OutgoingTxBatches(cmd.Context(), &types.QueryOutgoingTxBatchesRequest{
				ChainName: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdGetBatchFees() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-fees [chain-name]",
		Short: "Query a list of send to external transaction fees to be processed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.BatchFees(cmd.Context(), &types.QueryBatchFeeRequest{
				ChainName: args[0],
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

func CmdGetLastObservedBlockHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last-observed-block-height [chain-name]",
		Short: "Query last observed block height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.LastObservedBlockHeight(cmd.Context(), &types.QueryLastObservedBlockHeightRequest{
				ChainName: args[0],
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdProjectedBatchTimeoutHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projected-batch-timeout-height [chain-name]",
		Short: "Query projected batch timeout height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ProjectedBatchTimeoutHeight(cmd.Context(), &types.QueryProjectedBatchTimeoutHeightRequest{
				ChainName: args[0],
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdGetDenomToExternalToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token [chain-name] [denom]",
		Short: "Query contract address from denom",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DenomToToken(cmd.Context(), &types.QueryDenomToTokenRequest{
				ChainName: args[0],
				Denom:     args[1],
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

func CmdGetExternalTokenToDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "denom [chain-name] [token-contract]",
		Short: "Query denom from contract address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			tokenAddress := args[1]
			if strings.HasPrefix(tokenAddress, "0x") {
				if !gethCommon.IsHexAddress(tokenAddress) {
					return fmt.Errorf("invalid contract address:%s", tokenAddress)
				}
				tokenAddress = gethCommon.HexToAddress(tokenAddress).Hex()
			}

			res, err := queryClient.TokenToDenom(cmd.Context(), &types.QueryTokenToDenomRequest{
				ChainName: args[0],
				Token:     tokenAddress,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

func CmdGetBridgeTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridge-tokens [chain-name]",
		Short: "Query bridge token list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.BridgeTokens(cmd.Context(), &types.QueryBridgeTokensRequest{
				ChainName: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

func CmdGetOracleEventNonce() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event-nonce [chain-name] [bridger-address]",
		Short: "Query last event nonce by oracle bridger address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			bridgerAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}
			res, err := queryClient.LastEventNonceByAddr(cmd.Context(), &types.QueryLastEventNonceByAddrRequest{
				ChainName:      args[0],
				BridgerAddress: bridgerAddr.String(),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

func CmdGetOracleEventBlockHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event-block-height [chain-name] [bridger-address]",
		Short: "Query last event block height by oracle bridger address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			bridgerAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}
			res, err := queryClient.LastEventBlockHeightByAddr(cmd.Context(), &types.QueryLastEventBlockHeightByAddrRequest{
				BridgerAddress: bridgerAddr.String(),
				ChainName:      args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdCovertBridgeToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "covert-bridge-token [chain-name] [token-contract]",
		Short:   "Covert bridge claim token name",
		Example: "fxcored q crosschain covert-bridge-token bsc 0x3f6795b8ABE0775a88973469909adE1405f7ac09 --channelIBC=transfer/channel-0",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			chainName := args[0]
			if err := types.ValidateModuleName(chainName); err != nil {
				return err
			}
			tokenContract := args[1]
			if err := fxtypes.ValidateEthereumAddress(tokenContract); err != nil {
				return err
			}
			channelIbc, err := cmd.Flags().GetString(flagChannelIBC)
			if err != nil {
				return err
			}
			coinName := fmt.Sprintf("%s%s", chainName, tokenContract)
			if len(channelIbc) > 0 {
				coinName = transfertypes.DenomTrace{
					Path:      channelIbc,
					BaseDenom: coinName,
				}.IBCDenom()
			}
			type output struct {
				ChainName     string
				TokenContract string
				ChannelIbc    string
				CoinName      string
			}
			indent, err := json.MarshalIndent(output{
				ChainName:     chainName,
				TokenContract: tokenContract,
				ChannelIbc:    channelIbc,
				CoinName:      coinName,
			}, "", "  ")
			if err != nil {
				return err
			}
			return clientCtx.PrintBytes(indent)
		},
	}
	cmd.Flags().String(flagChannelIBC, "", "bridge bind channel: transfer/channel-0")
	return cmd
}
