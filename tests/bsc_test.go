package tests

import (
	"fmt"
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/functionx/fx-core/v2/app/helpers"

	"github.com/stretchr/testify/suite"

	bsctypes "github.com/functionx/fx-core/v2/x/bsc/types"
)

type BSCTestSuite struct {
	CrosschainTestSuite
}

func TestBSCTestSuite(t *testing.T) {
	suite.Run(t, &BSCTestSuite{
		CrosschainTestSuite: NewCrosschainTestSuite(bsctypes.ModuleName),
	})
}

func (suite *BSCTestSuite) TestCrosschain_BSC() {
	const purseToken = "0xFBBbB4f7B1e5bCb0345c5A5a61584B2547d5D582"
	const purseTokenChannelIBC = "transfer/channel-0"
	purseDenom := transfertypes.DenomTrace{
		Path:      purseTokenChannelIBC,
		BaseDenom: fmt.Sprintf("%s%s", suite.chainName, purseToken),
	}.IBCDenom()

	proposalId := suite.SendUpdateChainOraclesProposal()
	suite.ProposalVote(suite.AdminPrivateKey(), proposalId, govtypes.OptionYes)
	suite.CheckProposal(proposalId, govtypes.StatusPassed)

	suite.BondedOracle()
	suite.SendOracleSetConfirm()

	denom := suite.AddBridgeTokenClaim("PundiX reward token", "PURSE", 18, purseToken, purseTokenChannelIBC)
	suite.Equal(denom, purseDenom)

	suite.SendToFxClaim(purseToken, sdk.NewInt(100).MulRaw(1e18), "")
	suite.SendToFxClaim(purseToken, sdk.NewInt(100).MulRaw(1e18), fmt.Sprintf("px/%s", purseTokenChannelIBC))
	// send tx fee
	suite.Send(suite.AccAddr(), helpers.NewCoin(sdk.NewInt(100).MulRaw(1e18)))
	suite.SendToExternal(5, sdk.NewCoin(purseDenom, sdk.NewInt(10).MulRaw(1e18)))

	suite.SendBatchRequest(5)
	suite.SendConfirmBatch()

	suite.SendToExternalAndCancel(purseToken, purseDenom, sdk.NewInt(10).MulRaw(1e18))
}
