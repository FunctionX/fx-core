package keeper_test

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/functionx/fx-core/v7/app"
	"github.com/functionx/fx-core/v7/testutil/helpers"
	fxtypes "github.com/functionx/fx-core/v7/types"
	crosschaintypes "github.com/functionx/fx-core/v7/x/crosschain/types"
	erc20types "github.com/functionx/fx-core/v7/x/erc20/types"
	evmtypes "github.com/functionx/fx-core/v7/x/evm/types"
	"github.com/functionx/fx-core/v7/x/gov/keeper"
	"github.com/functionx/fx-core/v7/x/gov/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app             *app.App
	ctx             sdk.Context
	msgServer       govv1.MsgServer
	legacyMsgServer govv1beta1.MsgServer
	valAddr         []sdk.ValAddress
	addrs           []sdk.AccAddress
	govAcct         string
	MsgServer       types.MsgServer
	queryClient     types.QueryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	valNumber := tmrand.Intn(10) + 1

	valSet, valAccounts, valBalances := helpers.GenerateGenesisValidator(valNumber, sdk.Coins{})
	suite.app = helpers.SetupWithGenesisValSet(suite.T(), valSet, valAccounts, valBalances...)
	suite.ctx = suite.app.NewContext(false, tmproto.Header{
		ChainID:         fxtypes.MainnetChainId,
		Height:          suite.app.LastBlockHeight() + 1,
		ProposerAddress: valSet.Proposer.Address,
	})
	suite.addrs = helpers.AddTestAddrsIncremental(suite.app, suite.ctx, 5, sdk.NewCoins(sdk.NewCoin(fxtypes.DefaultDenom, sdkmath.NewInt(10*1e8).MulRaw(1e18))))
	suite.valAddr = make([]sdk.ValAddress, valNumber)
	for i, addr := range valAccounts {
		suite.valAddr[i] = addr.GetAddress().Bytes()
	}
	suite.govAcct = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	suite.msgServer = keeper.NewMsgServerImpl(govkeeper.NewMsgServerImpl(suite.app.GovKeeper.Keeper), suite.app.GovKeeper)
	suite.legacyMsgServer = govkeeper.NewLegacyMsgServerImpl(suite.govAcct, suite.msgServer)
	suite.MsgServer = keeper.NewMsgServerImpl(govkeeper.NewMsgServerImpl(suite.app.GovKeeper.Keeper), suite.app.GovKeeper)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, app.MakeEncodingConfig().InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, suite.app.GovKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func (suite *KeeperTestSuite) addFundCommunityPool() {
	sender := helpers.GenAccAddress()
	coin := sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: sdkmath.NewInt(5 * 1e8).MulRaw(1e18)}
	helpers.AddTestAddr(suite.app, suite.ctx, sender, sdk.NewCoins(coin))
	err := suite.app.DistrKeeper.FundCommunityPool(suite.ctx, sdk.NewCoins(coin), sender)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) newAddress() sdk.AccAddress {
	address := helpers.GenAccAddress()
	coin := sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: sdkmath.NewInt(50_000).MulRaw(1e18)}
	helpers.AddTestAddr(suite.app, suite.ctx, address, sdk.NewCoins(coin))
	return address
}

func (suite *KeeperTestSuite) TestDeposits() {
	initCoins, testProposalMsg, err := suite.getTextProposal()
	suite.NoError(err)
	proposalResponse, err := suite.msgServer.SubmitProposal(sdk.WrapSDKContext(suite.ctx), testProposalMsg)
	suite.NoError(err)
	addr := suite.newAddress()
	_, found := suite.app.GovKeeper.GetDeposit(suite.ctx, proposalResponse.ProposalId, addr)
	suite.False(found)
	proposal, ok := suite.app.GovKeeper.Keeper.GetProposal(suite.ctx, proposalResponse.ProposalId)
	suite.True(ok)
	suite.Equal(govv1.StatusDepositPeriod, proposal.Status)
	minDeposit := suite.app.GovKeeper.GetParams(suite.ctx).MinDeposit
	suite.True(initCoins.IsAllLT(minDeposit))

	// first deposit
	firstCoins := sdk.Coins{sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: sdkmath.NewInt(1e3).MulRaw(1e18)}}
	votingStarted, err := suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, addr, firstCoins)
	suite.NoError(err)
	suite.False(votingStarted)
	deposit, found := suite.app.GovKeeper.GetDeposit(suite.ctx, proposal.Id, addr)
	suite.True(found)
	suite.Equal(firstCoins.String(), sdk.NewCoins(deposit.Amount...).String())
	suite.Equal(addr.String(), deposit.Depositor)
	proposal, ok = suite.app.GovKeeper.Keeper.GetProposal(suite.ctx, proposalResponse.ProposalId)
	suite.True(ok)
	suite.Equal(govv1.StatusDepositPeriod, proposal.Status)
	suite.Equal(firstCoins, sdk.NewCoins(proposal.TotalDeposit...).Sub(initCoins...))
	suite.True(initCoins.Add(firstCoins...).IsAllLT(minDeposit))

	// second deposit
	secondCoins := sdk.Coins{sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: sdkmath.NewInt(9 * 1e3).MulRaw(1e18)}}
	votingStarted, err = suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, addr, secondCoins)
	suite.NoError(err)
	suite.True(votingStarted)
	deposit, found = suite.app.GovKeeper.GetDeposit(suite.ctx, proposal.Id, addr)
	suite.True(found)
	suite.Equal(firstCoins.Add(secondCoins...).String(), sdk.NewCoins(deposit.Amount...).String())
	suite.Equal(addr.String(), deposit.Depositor)
	proposal, ok = suite.app.GovKeeper.Keeper.GetProposal(suite.ctx, proposalResponse.ProposalId)
	suite.True(ok)
	suite.Equal(govv1.StatusVotingPeriod, proposal.Status)
	suite.True(initCoins.Add(firstCoins...).Add(secondCoins...).IsAllGTE(minDeposit))
}

func (suite *KeeperTestSuite) getTextProposal() (sdk.Coins, *govv1.MsgSubmitProposal, error) {
	initCoins := sdk.Coins{sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: sdkmath.NewInt(1e3).MulRaw(1e18)}}
	content := govv1beta1.NewTextProposal("Test", "description")
	msgExecLegacyContent, err := govv1.NewLegacyContent(content, suite.govAcct)
	suite.NoError(err)
	testProposalMsg, err := govv1.NewMsgSubmitProposal([]sdk.Msg{msgExecLegacyContent}, initCoins, suite.newAddress().String(),
		"", content.GetTitle(), content.GetDescription())
	return initCoins, testProposalMsg, err
}

func (suite *KeeperTestSuite) TestEGFDepositsLessThan1000() {
	suite.addFundCommunityPool()

	egfCoins := sdk.Coins{sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: sdkmath.NewInt(10 * 1e3).MulRaw(1e18)}}

	spendProposal := &distributiontypes.MsgCommunityPoolSpend{Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(), Recipient: helpers.GenAccAddress().String(), Amount: egfCoins}
	minDeposit := suite.app.GovKeeper.EGFProposalMinDeposit(suite.ctx, sdk.MsgTypeURL(&distributiontypes.MsgCommunityPoolSpend{}), egfCoins)
	initCoins := sdk.Coins{sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: sdkmath.NewInt(1 * 1e3).MulRaw(1e18)}}
	suite.True(initCoins.IsEqual(minDeposit))

	communityPoolSpendProposalMsg, err := govv1.NewMsgSubmitProposal([]sdk.Msg{spendProposal}, initCoins, suite.newAddress().String(),
		"", "community Pool Spend Proposal", "description")
	suite.NoError(err)
	proposalResponse, err := suite.msgServer.SubmitProposal(sdk.WrapSDKContext(suite.ctx), communityPoolSpendProposalMsg)
	suite.NoError(err)
	_, found := suite.app.GovKeeper.GetDeposit(suite.ctx, proposalResponse.ProposalId, suite.newAddress())
	suite.False(found)
	proposal, ok := suite.app.GovKeeper.Keeper.GetProposal(suite.ctx, proposalResponse.ProposalId)
	suite.True(ok)
	suite.Equal(govv1.StatusVotingPeriod, proposal.Status)
}

func (suite *KeeperTestSuite) TestEGFDepositsMoreThan1000() {
	suite.addFundCommunityPool()

	thousand := sdkmath.NewInt(1 * 1e3).MulRaw(1e18)
	egfCoins := sdk.Coins{sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: thousand.MulRaw(10).Add(sdkmath.NewInt(10))}}

	initCoins := sdk.Coins{sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: thousand}}
	spendProposal := &distributiontypes.MsgCommunityPoolSpend{Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(), Recipient: helpers.GenAccAddress().String(), Amount: egfCoins}
	minDeposit := suite.app.GovKeeper.EGFProposalMinDeposit(suite.ctx, sdk.MsgTypeURL(&distributiontypes.MsgCommunityPoolSpend{}), egfCoins)

	communityPoolSpendProposalMsg, err := govv1.NewMsgSubmitProposal([]sdk.Msg{spendProposal}, initCoins, suite.newAddress().String(),
		"", "community Pool Spend Proposal", "description")
	suite.NoError(err)
	proposalResponse, err := suite.msgServer.SubmitProposal(sdk.WrapSDKContext(suite.ctx), communityPoolSpendProposalMsg)
	suite.NoError(err)
	_, found := suite.app.GovKeeper.GetDeposit(suite.ctx, proposalResponse.ProposalId, suite.newAddress())
	suite.False(found)
	proposal, ok := suite.app.GovKeeper.Keeper.GetProposal(suite.ctx, proposalResponse.ProposalId)
	suite.True(ok)
	suite.Equal(govv1.StatusDepositPeriod, proposal.Status)
	suite.True(initCoins.IsAllLT(minDeposit))

	depositCoins := sdk.Coins{sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: sdkmath.NewInt(1)}}
	votingStarted, err := suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, suite.newAddress(), depositCoins)
	suite.NoError(err)
	suite.True(votingStarted)
	proposal, ok = suite.app.GovKeeper.Keeper.GetProposal(suite.ctx, proposalResponse.ProposalId)
	suite.True(ok)
	suite.Equal(govv1.StatusVotingPeriod, proposal.Status)
	suite.Equal(sdk.NewCoins(proposal.TotalDeposit...).String(), minDeposit.String())
}

func (suite *KeeperTestSuite) TestEGFDeposits() {
	suite.addFundCommunityPool()

	egfCoins := sdk.Coins{sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: sdkmath.NewInt(150 * 1e3).MulRaw(1e18)}}

	initCoins := sdk.Coins{sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: sdkmath.NewInt(1 * 1e3).MulRaw(1e18)}}
	spendProposal := &distributiontypes.MsgCommunityPoolSpend{Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(), Recipient: helpers.GenAccAddress().String(), Amount: egfCoins}
	minDeposit := suite.app.GovKeeper.EGFProposalMinDeposit(suite.ctx, sdk.MsgTypeURL(&distributiontypes.MsgCommunityPoolSpend{}), egfCoins)

	communityPoolSpendProposalMsg, err := govv1.NewMsgSubmitProposal([]sdk.Msg{spendProposal}, initCoins, suite.newAddress().String(),
		"", "community Pool Spend Proposal", "description")
	suite.NoError(err)
	proposalResponse, err := suite.msgServer.SubmitProposal(sdk.WrapSDKContext(suite.ctx), communityPoolSpendProposalMsg)
	suite.NoError(err)
	_, found := suite.app.GovKeeper.GetDeposit(suite.ctx, proposalResponse.ProposalId, suite.newAddress())
	suite.False(found)
	proposal, ok := suite.app.GovKeeper.Keeper.GetProposal(suite.ctx, proposalResponse.ProposalId)
	suite.True(ok)
	suite.Equal(govv1.StatusDepositPeriod, proposal.Status)
	suite.True(initCoins.IsAllLT(minDeposit))

	// first deposit
	firstCoins := sdk.Coins{sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: sdkmath.NewInt(1 * 1e3).MulRaw(1e18)}}
	addr := suite.newAddress()
	votingStarted, err := suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, addr, firstCoins)
	suite.NoError(err)
	suite.False(votingStarted)
	deposit, found := suite.app.GovKeeper.GetDeposit(suite.ctx, proposal.Id, addr)
	suite.True(found)
	suite.Equal(firstCoins.String(), sdk.NewCoins(deposit.Amount...).String())
	suite.Equal(addr.String(), deposit.Depositor)
	proposal, ok = suite.app.GovKeeper.Keeper.GetProposal(suite.ctx, proposalResponse.ProposalId)
	suite.True(ok)
	suite.Equal(govv1.StatusDepositPeriod, proposal.Status)
	suite.Equal(firstCoins, sdk.NewCoins(proposal.TotalDeposit...).Sub(initCoins...))
	suite.True(initCoins.Add(firstCoins...).IsAllLT(minDeposit))

	// second deposit
	secondCoins := sdk.Coins{sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: sdkmath.NewInt(9 * 1e3).MulRaw(1e18)}}
	votingStarted, err = suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, addr, secondCoins)
	suite.NoError(err)
	suite.False(votingStarted)
	deposit, found = suite.app.GovKeeper.GetDeposit(suite.ctx, proposal.Id, addr)
	suite.True(found)
	suite.Equal(firstCoins.Add(secondCoins...).String(), sdk.NewCoins(deposit.Amount...).String())
	suite.Equal(addr.String(), deposit.Depositor)
	proposal, ok = suite.app.GovKeeper.Keeper.GetProposal(suite.ctx, proposalResponse.ProposalId)
	suite.True(ok)
	suite.Equal(govv1.StatusDepositPeriod, proposal.Status)
	suite.True(initCoins.Add(firstCoins...).Add(secondCoins...).IsAllLT(minDeposit))

	// third deposit
	thirdCoins := sdk.Coins{sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: sdkmath.NewInt(4 * 1e3).MulRaw(1e18)}}
	votingStarted, err = suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, addr, thirdCoins)
	suite.NoError(err)
	suite.True(votingStarted)
	deposit, found = suite.app.GovKeeper.GetDeposit(suite.ctx, proposal.Id, addr)
	suite.True(found)
	suite.Equal(firstCoins.Add(secondCoins...).Add(thirdCoins...).String(), sdk.NewCoins(deposit.Amount...).String())
	suite.Equal(addr.String(), deposit.Depositor)
	proposal, ok = suite.app.GovKeeper.Keeper.GetProposal(suite.ctx, proposalResponse.ProposalId)
	suite.True(ok)
	suite.Equal(govv1.StatusVotingPeriod, proposal.Status)
	suite.True(initCoins.Add(firstCoins...).Add(secondCoins...).Add(thirdCoins...).IsAllGTE(minDeposit))
}

func (suite *KeeperTestSuite) TestUpdateParams() {
	testCases := []struct {
		testName    string
		amount      sdk.Coins
		msg         []sdk.Msg
		result      bool
		expectedErr string
	}{
		{
			testName:    "set Erc20Params",
			amount:      sdk.NewCoins(sdk.NewCoin(fxtypes.DefaultDenom, sdkmath.NewInt(10*1e3).MulRaw(1e18))),
			msg:         []sdk.Msg{&erc20types.MsgUpdateParams{Authority: "0x1", Params: erc20types.DefaultParams()}},
			result:      false,
			expectedErr: "authority",
		},
		{
			testName:    "set CrossChainParam",
			amount:      sdk.NewCoins(sdk.NewCoin(fxtypes.DefaultDenom, sdkmath.NewInt(10*1e3).MulRaw(1e18))),
			msg:         []sdk.Msg{&crosschaintypes.MsgUpdateParams{ChainName: "eth", Authority: suite.govAcct, Params: crosschaintypes.DefaultParams()}},
			result:      true,
			expectedErr: "",
		},
		{
			testName:    "set Erc20Params",
			amount:      sdk.NewCoins(sdk.NewCoin(fxtypes.DefaultDenom, sdkmath.NewInt(10*1e3).MulRaw(1e18))),
			msg:         []sdk.Msg{&erc20types.MsgUpdateParams{Authority: suite.govAcct, Params: erc20types.DefaultParams()}},
			result:      true,
			expectedErr: "",
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.testName), func() {
			proposal, err := suite.app.GovKeeper.SubmitProposal(suite.ctx, tc.msg, "", tc.testName, tc.testName, suite.newAddress())
			if tc.result {
				suite.NoError(err)
				_, err = suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, suite.newAddress(), tc.amount)
				suite.Require().NoError(err)
				proposal, ok := suite.app.GovKeeper.Keeper.GetProposal(suite.ctx, proposal.Id)
				suite.True(ok)
				suite.Equal(govv1.StatusVotingPeriod, proposal.Status)
			} else {
				suite.Error(err, err)
				suite.Require().True(strings.Contains(err.Error(), tc.expectedErr))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetParams() {
	govParams := suite.app.GovKeeper.GetParams(suite.ctx)

	tallyParamsVetoThreshold := govParams.VetoThreshold
	tallyParamsThreshold := govParams.Threshold
	tallyParamsQuorum := govParams.Quorum

	depositParamsMinDeposit := govParams.MinDeposit
	depositParamsMaxDepositPeriod := govParams.MaxDepositPeriod
	votingParamsVotingPeriod := govParams.VotingPeriod

	// unregistered Msg
	params := suite.app.GovKeeper.GetFXParams(suite.ctx, sdk.MsgTypeURL(&erc20types.MsgUpdateParams{}))
	suite.Require().EqualValues(params.MinInitialDeposit.String(), sdk.NewCoin(fxtypes.DefaultDenom, types.DefaultMinInitialDeposit).String())
	suite.Require().EqualValues(params.MinDeposit, depositParamsMinDeposit)
	suite.Require().EqualValues(params.MaxDepositPeriod, depositParamsMaxDepositPeriod)
	suite.Require().EqualValues(params.VotingPeriod, votingParamsVotingPeriod)
	suite.Require().EqualValues(params.VetoThreshold, tallyParamsVetoThreshold)
	suite.Require().EqualValues(params.Threshold, tallyParamsThreshold)
	suite.Require().EqualValues(params.Quorum, tallyParamsQuorum)
	erc20MsgType := []string{
		"/fx.erc20.v1.RegisterCoinProposal",
		"/fx.erc20.v1.RegisterERC20Proposal",
		"/fx.erc20.v1.ToggleTokenConversionProposal",
		"/fx.erc20.v1.UpdateDenomAliasProposal",
		sdk.MsgTypeURL(&erc20types.MsgRegisterCoin{}),
		sdk.MsgTypeURL(&erc20types.MsgRegisterERC20{}),
		sdk.MsgTypeURL(&erc20types.MsgToggleTokenConversion{}),
		sdk.MsgTypeURL(&erc20types.MsgUpdateDenomAlias{}),
	}
	for _, erc20MsgType := range erc20MsgType {
		// registered Msg
		params = suite.app.GovKeeper.GetFXParams(suite.ctx, erc20MsgType)
		suite.Require().NoError(params.ValidateBasic())
		suite.Require().EqualValues(params.MinDeposit, depositParamsMinDeposit)
		suite.Require().EqualValues(params.MinInitialDeposit.String(), sdk.NewCoin(fxtypes.DefaultDenom, types.DefaultMinInitialDeposit).String())
		suite.Require().EqualValues(params.MaxDepositPeriod, depositParamsMaxDepositPeriod)
		suite.Require().EqualValues(params.VotingPeriod, votingParamsVotingPeriod)
		suite.Require().EqualValues(params.VetoThreshold, tallyParamsVetoThreshold)
		suite.Require().EqualValues(params.Threshold, tallyParamsThreshold)
		suite.Require().EqualValues(params.Quorum, types.DefaultErc20Quorum.String())
	}

	params = suite.app.GovKeeper.GetFXParams(suite.ctx, sdk.MsgTypeURL(&evmtypes.MsgCallContract{}))
	suite.Require().NoError(params.ValidateBasic())
	suite.Require().EqualValues(params.MinDeposit, depositParamsMinDeposit)
	suite.Require().EqualValues(params.MinInitialDeposit.String(), sdk.NewCoin(fxtypes.DefaultDenom, types.DefaultMinInitialDeposit).String())
	suite.Require().EqualValues(params.MaxDepositPeriod, depositParamsMaxDepositPeriod)
	suite.Require().EqualValues(params.VotingPeriod.String(), types.DefaultEvmVotingPeriod.String())
	suite.Require().EqualValues(params.VetoThreshold, tallyParamsVetoThreshold)
	suite.Require().EqualValues(params.Threshold, tallyParamsThreshold)
	suite.Require().EqualValues(params.Quorum, types.DefaultEvmQuorum.String())

	params = suite.app.GovKeeper.GetFXParams(suite.ctx, "/cosmos.distribution.v1beta1.CommunityPoolSpendProposal")
	egfParams := suite.app.GovKeeper.GetEGFParams(suite.ctx)
	suite.Require().NoError(params.ValidateBasic())
	suite.Require().NoError(egfParams.ValidateBasic())
	suite.Require().EqualValues(params.MinDeposit, depositParamsMinDeposit)
	suite.Require().EqualValues(params.MinInitialDeposit.String(), sdk.NewCoin(fxtypes.DefaultDenom, types.DefaultMinInitialDeposit).String())
	suite.Require().EqualValues(params.MaxDepositPeriod, depositParamsMaxDepositPeriod)
	suite.Require().EqualValues(params.VotingPeriod.String(), types.DefaultEgfVotingPeriod.String())
	suite.Require().EqualValues(params.VetoThreshold, tallyParamsVetoThreshold)
	suite.Require().EqualValues(params.Threshold, tallyParamsThreshold)
	suite.Require().EqualValues(params.Quorum, tallyParamsQuorum)
	suite.Require().EqualValues(egfParams.EgfDepositThreshold, sdk.NewCoin(fxtypes.DefaultDenom, types.DefaultEgfDepositThreshold))
	suite.Require().EqualValues(egfParams.ClaimRatio, types.DefaultClaimRatio.String())
}

func TestCheckContractAddressIsDisabled(t *testing.T) {
	addr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	methodId, _ := hex.DecodeString("abcdef01")

	tests := []struct {
		name                string
		disabledPrecompiles []string
		addr                common.Address
		methodId            []byte
		expectedError       string
	}{
		{
			name:                "Empty disabled list",
			disabledPrecompiles: []string{},
			addr:                addr,
			methodId:            methodId,
			expectedError:       "",
		},
		{
			name:                "Address is disabled",
			disabledPrecompiles: []string{addr.Hex()},
			addr:                addr,
			methodId:            methodId,
			expectedError:       "precompile address is disabled",
		},
		{
			name:                "Address and method are disabled",
			disabledPrecompiles: []string{fmt.Sprintf("%s/%s", addr.Hex(), "abcdef01")},
			addr:                addr,
			methodId:            methodId,
			expectedError:       "precompile method abcdef01 is disabled",
		},
		{
			name:                "Address and method are not disabled",
			disabledPrecompiles: []string{fmt.Sprintf("%s/%s", addr.Hex(), "12345678")},
			addr:                addr,
			methodId:            methodId,
			expectedError:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keeper.CheckContractAddressIsDisabled(tt.disabledPrecompiles, tt.addr, tt.methodId)
			if tt.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.expectedError)
			}
		})
	}
}
