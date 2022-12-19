package tests

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	crosschaintypes "github.com/functionx/fx-core/v3/x/crosschain/types"
	erc20types "github.com/functionx/fx-core/v3/x/erc20/types"
	trontypes "github.com/functionx/fx-core/v3/x/tron/types"
)

func (suite *IntegrationTest) ERC20Test() {
	suite.Send(suite.erc20.AccAddress(), suite.NewCoin(sdk.NewInt(10_100).MulRaw(1e18)))

	var aliases []string
	for _, chain := range suite.crosschain {
		denoms := chain.GetBridgeDenoms()
		aliases = append(aliases, denoms...)
	}
	suite.erc20.metadata.DenomUnits[0].Aliases = aliases
	suite.T().Log(suite.erc20.metadata.DenomUnits[0].Aliases)
	proposalId := suite.erc20.RegisterCoinProposal(suite.erc20.metadata)
	suite.NoError(suite.network.WaitForNextBlock())
	suite.CheckProposal(proposalId, govtypes.StatusPassed)
	suite.erc20.CheckRegisterCoin(suite.erc20.metadata.Base)

	denom := suite.erc20.metadata.Base
	tokenPair := suite.erc20.TokenPair(denom)
	suite.Equal(tokenPair.Denom, denom)
	suite.Equal(tokenPair.Enabled, true)
	suite.Equal(tokenPair.ContractOwner, erc20types.OWNER_MODULE)
	suite.T().Log("token pair", tokenPair.String())

	for i, chain := range suite.crosschain {
		bridgeDenom := chain.GetBridgeDenoms()[0]
		tokenContract := chain.GetBridgeToken(bridgeDenom)
		chain.SendToFxClaim(tokenContract, sdk.NewInt(200), "module/evm")
		balance := suite.erc20.BalanceOf(tokenPair.GetERC20Contract(), chain.HexAddress())
		suite.Equal(balance, big.NewInt(200))

		suite.erc20.TransferERC20(chain.privKey, tokenPair.GetERC20Contract(), suite.erc20.HexAddress(), big.NewInt(100))
		balance = suite.erc20.BalanceOf(tokenPair.GetERC20Contract(), suite.erc20.HexAddress())
		suite.Equal(balance, big.NewInt(100))

		receive := suite.erc20.HexAddress().String()
		if chain.chainName == trontypes.ModuleName {
			receive = trontypes.AddressFromHex(receive)
		}
		suite.erc20.TransferCrossChain(suite.erc20.privKey, tokenPair.GetERC20Contract(), receive,
			big.NewInt(50), big.NewInt(50), fmt.Sprintf("chain/%s", chain.chainName))

		resp, err := chain.CrosschainQuery().GetPendingSendToExternal(suite.ctx,
			&crosschaintypes.QueryPendingSendToExternalRequest{
				ChainName:     chain.chainName,
				SenderAddress: suite.erc20.AccAddress().String(),
			})
		suite.NoError(err)
		suite.Equal(1, len(resp.UnbatchedTransfers))
		suite.Equal(int64(50), resp.UnbatchedTransfers[0].Token.Amount.Int64())
		suite.Equal(int64(50), resp.UnbatchedTransfers[0].Fee.Amount.Int64())
		suite.Equal(suite.erc20.AccAddress().String(), resp.UnbatchedTransfers[0].Sender)
		if chain.chainName == trontypes.ModuleName {
			suite.Equal(trontypes.AddressFromHex(suite.erc20.HexAddress().String()), resp.UnbatchedTransfers[0].DestAddress)
		} else {
			suite.Equal(suite.erc20.HexAddress().String(), resp.UnbatchedTransfers[0].DestAddress)
		}

		// convert
		suite.CheckBalance(suite.erc20.AccAddress(), sdk.NewCoin(bridgeDenom, sdk.NewInt(50)))
		suite.erc20.ConvertDenom(suite.erc20.privKey, suite.erc20.AccAddress(), sdk.NewCoin(bridgeDenom, sdk.NewInt(50)), "")
		suite.CheckBalance(suite.erc20.AccAddress(), sdk.NewCoin(denom, sdk.NewInt(50)))
		suite.CheckBalance(suite.erc20.AccAddress(), sdk.NewCoin(bridgeDenom, sdk.ZeroInt()))

		suite.erc20.ConvertDenom(suite.erc20.privKey, suite.erc20.AccAddress(), sdk.NewCoin(denom, sdk.NewInt(50)), chain.chainName)
		suite.CheckBalance(suite.erc20.AccAddress(), sdk.NewCoin(bridgeDenom, sdk.NewInt(50)))

		// todo why can't delete the last alias
		if i < len(suite.crosschain)-1 {
			// remove proposal
			proposalId := suite.erc20.UpdateDenomAliasProposal(denom, bridgeDenom)
			suite.NoError(suite.network.WaitForNextBlock())
			suite.CheckProposal(proposalId, govtypes.StatusPassed)

			// check remove
			response, err := suite.erc20.ERC20Query().DenomAliases(suite.ctx, &erc20types.QueryDenomAliasesRequest{Denom: denom})
			suite.NoError(err)
			suite.Equal(len(suite.crosschain)-i-1, len(response.Aliases))

			_, err = suite.erc20.ERC20Query().AliasDenom(suite.ctx, &erc20types.QueryAliasDenomRequest{Alias: bridgeDenom})
			suite.Error(err)
		}
	}

	proposalId = suite.erc20.ToggleTokenConversionProposal(denom)
	suite.NoError(suite.network.WaitForNextBlock())
	suite.CheckProposal(proposalId, govtypes.StatusPassed)

	suite.False(suite.erc20.TokenPair(denom).Enabled)
}
