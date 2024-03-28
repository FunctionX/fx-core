package keeper_test

import (
	"fmt"

	tmrand "github.com/tendermint/tendermint/libs/rand"

	"github.com/functionx/fx-core/v7/testutil/helpers"
	"github.com/functionx/fx-core/v7/x/crosschain/types"
)

func (suite *KeeperTestSuite) TestKeeper_BridgeToken() {
	tokenContract := helpers.GenerateAddress().Hex()
	denom, err := suite.Keeper().SetIbcDenomTrace(suite.ctx, tokenContract, "")
	suite.NoError(err)
	suite.Equal(fmt.Sprintf("%s%s", suite.chainName, tokenContract), denom)

	bridgeTokenTypes := []types.BridgeTokenType{types.BRIDGE_TOKEN_TYPE_ERC20, types.BRIDGE_TOKEN_TYPE_ERC721, types.BRIDGE_TOKEN_TYPE_ERC404}
	randomTokenType := bridgeTokenTypes[tmrand.Intn(len(bridgeTokenTypes))]
	suite.Keeper().AddBridgeTokenWithTokenType(suite.ctx, tokenContract, denom, randomTokenType)

	bridgeToken := &types.BridgeToken{Token: tokenContract, Denom: denom, TokenType: randomTokenType}
	suite.EqualValues(bridgeToken, suite.Keeper().GetBridgeTokenDenom(suite.ctx, tokenContract))

	suite.EqualValues(bridgeToken, suite.Keeper().GetDenomBridgeToken(suite.ctx, denom))

	suite.Keeper().IterateBridgeTokenToDenom(suite.ctx, func(bt *types.BridgeToken) bool {
		suite.Equal(bt.Token, tokenContract)
		suite.Equal(bt.Denom, denom)
		return false
	})
}
