package crosschain_test

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	tronAddress "github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	bsctypes "github.com/functionx/fx-core/v2/x/bsc/types"
	polygontypes "github.com/functionx/fx-core/v2/x/polygon/types"
	tronkeeper "github.com/functionx/fx-core/v2/x/tron/keeper"
	trontypes "github.com/functionx/fx-core/v2/x/tron/types"

	"github.com/functionx/fx-core/v2/app"
	"github.com/functionx/fx-core/v2/app/helpers"
	"github.com/functionx/fx-core/v2/x/crosschain/keeper"

	fxtypes "github.com/functionx/fx-core/v2/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/functionx/fx-core/v2/x/crosschain"
	"github.com/functionx/fx-core/v2/x/crosschain/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app            *app.App
	ctx            sdk.Context
	oracles        []sdk.AccAddress
	bridgers       []sdk.AccAddress
	externals      []*ecdsa.PrivateKey
	validator      []sdk.ValAddress
	chainName      string
	delegateAmount sdk.Int
}

func TestIntegrationTestSuite(t *testing.T) {
	compile, err := regexp.Compile("^Test")
	require.NoError(t, err)
	for _, moduleName := range []string{bsctypes.ModuleName, polygontypes.ModuleName, trontypes.ModuleName} {
		methodFinder := reflect.TypeOf(new(IntegrationTestSuite))
		for i := 0; i < methodFinder.NumMethod(); i++ {
			method := methodFinder.Method(i)
			if !compile.MatchString(method.Name) {
				continue
			}
			t.Run(method.Name, func(subT *testing.T) {
				mySuite := &IntegrationTestSuite{chainName: moduleName}
				mySuite.SetT(subT)
				mySuite.SetupTest()
				method.Func.Call([]reflect.Value{reflect.ValueOf(mySuite)})
			})
		}
	}
}

func (suite *IntegrationTestSuite) MsgServer() types.MsgServer {
	if suite.chainName == trontypes.ModuleName {
		return tronkeeper.NewMsgServerImpl(suite.app.TronKeeper)
	}
	return keeper.NewMsgServerImpl(suite.Keeper())
}

func (suite *IntegrationTestSuite) Keeper() keeper.Keeper {
	switch suite.chainName {
	case bsctypes.ModuleName:
		return suite.app.BscKeeper
	case polygontypes.ModuleName:
		return suite.app.PolygonKeeper
	case trontypes.ModuleName:
		return suite.app.TronKeeper.Keeper
	default:
		panic("invalid chain name")
	}
}

func (suite *IntegrationTestSuite) SetupTest() {
	valSet, valAccounts, valBalances := helpers.GenerateGenesisValidator(types.MaxOracleSize, sdk.Coins{})
	suite.app = helpers.SetupWithGenesisValSet(suite.T(), valSet, valAccounts, valBalances...)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})
	suite.oracles = helpers.AddTestAddrs(suite.app, suite.ctx, types.MaxOracleSize, sdk.NewInt(300*1e3).MulRaw(1e18))
	suite.bridgers = helpers.AddTestAddrs(suite.app, suite.ctx, types.MaxOracleSize, sdk.NewInt(300*1e3).MulRaw(1e18))
	suite.externals = helpers.CreateMultiEthKey(types.MaxOracleSize)
	suite.delegateAmount = sdk.NewInt(10 * 1e3).MulRaw(1e18)
	for i := 0; i < types.MaxOracleSize; i++ {
		suite.validator = append(suite.validator, valAccounts[i].GetAddress().Bytes())
	}

	proposalOracle := &types.ProposalOracle{}
	for _, oracle := range suite.oracles {
		proposalOracle.Oracles = append(proposalOracle.Oracles, oracle.String())
	}
	suite.Keeper().SetProposalOracle(suite.ctx, proposalOracle)
}

func (suite *IntegrationTestSuite) TestABCIEndBlockDepositClaim() {
	normalMsg := &types.MsgBondedOracle{
		OracleAddress:    suite.oracles[0].String(),
		BridgerAddress:   suite.bridgers[0].String(),
		ExternalAddress:  crypto.PubkeyToAddress(suite.externals[0].PublicKey).Hex(),
		ValidatorAddress: suite.validator[0].String(),
		DelegateAmount:   sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: suite.delegateAmount},
		ChainName:        suite.chainName,
	}
	if trontypes.ModuleName == suite.chainName {
		normalMsg.ExternalAddress = tronAddress.PubkeyToAddress(suite.externals[0].PublicKey).String()
	}
	_, err := suite.MsgServer().BondedOracle(sdk.WrapSDKContext(suite.ctx), normalMsg)
	require.NoError(suite.T(), err)

	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

	suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})

	bridgeToken := "0x3f6795b8ABE0775a88973469909adE1405f7ac09"
	sendToFxSendAddr := "0x3f6795b8ABE0775a88973469909adE1405f7ac09"
	if trontypes.ModuleName == suite.chainName {
		bridgeToken = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
		sendToFxSendAddr = tronAddress.PubkeyToAddress(suite.externals[0].PublicKey).String()
	}
	addBridgeTokenClaim := &types.MsgBridgeTokenClaim{
		EventNonce:     1,
		BlockHeight:    1000,
		TokenContract:  bridgeToken,
		Name:           "Pundix Reward Token",
		Symbol:         "PURES",
		Decimals:       18,
		BridgerAddress: suite.bridgers[0].String(),
		ChannelIbc:     hex.EncodeToString([]byte("transfer/channel-0")),
		ChainName:      suite.chainName,
	}
	_, err = suite.MsgServer().BridgeTokenClaim(sdk.WrapSDKContext(suite.ctx), addBridgeTokenClaim)
	require.NoError(suite.T(), err)

	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
	suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})

	sendToFxClaim := &types.MsgSendToFxClaim{
		EventNonce:     2,
		BlockHeight:    1001,
		TokenContract:  bridgeToken,
		Amount:         sdk.NewInt(1234),
		Sender:         sendToFxSendAddr,
		Receiver:       "fx16wvwsmpp4y4ttgzknyr6kqla877jud6u04lqey",
		TargetIbc:      hex.EncodeToString([]byte("px/transfer/channel-0")),
		BridgerAddress: suite.bridgers[0].String(),
		ChainName:      suite.chainName,
	}
	_, err = suite.MsgServer().SendToFxClaim(sdk.WrapSDKContext(suite.ctx), sendToFxClaim)
	require.NoError(suite.T(), err)

	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
	suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})

	receiveAddr, err := sdk.AccAddressFromBech32(sendToFxClaim.Receiver)
	require.NoError(suite.T(), err)
	allBalances := suite.app.BankKeeper.GetAllBalances(suite.ctx, receiveAddr)
	// transfer/channel-0/bscPURES
	tokenName := fmt.Sprintf("%s%s", suite.chainName, bridgeToken)
	if len(addBridgeTokenClaim.ChannelIbc) > 0 {
		channel, err := hex.DecodeString(addBridgeTokenClaim.ChannelIbc)
		require.NoError(suite.T(), err)
		tokenName = transfertypes.DenomTrace{
			Path:      string(channel),
			BaseDenom: fmt.Sprintf("%s%s", suite.chainName, bridgeToken),
		}.IBCDenom()
	}
	require.EqualValues(suite.T(), fmt.Sprintf("%s%s", sendToFxClaim.Amount.String(), tokenName), allBalances.String())
}

func (suite *IntegrationTestSuite) TestOracleUpdate() {
	for i := 0; i < 10; i++ {
		msgBondedOracle := &types.MsgBondedOracle{
			OracleAddress:    suite.oracles[i].String(),
			BridgerAddress:   suite.bridgers[i].String(),
			ExternalAddress:  crypto.PubkeyToAddress(suite.externals[i].PublicKey).Hex(),
			ValidatorAddress: suite.validator[i].String(),
			DelegateAmount:   sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: suite.delegateAmount},
			ChainName:        suite.chainName,
		}
		if trontypes.ModuleName == suite.chainName {
			msgBondedOracle.ExternalAddress = tronAddress.PubkeyToAddress(suite.externals[i].PublicKey).String()
		}
		require.NoError(suite.T(), msgBondedOracle.ValidateBasic())
		_, err := suite.MsgServer().BondedOracle(sdk.WrapSDKContext(suite.ctx), msgBondedOracle)

		require.NoError(suite.T(), err)
		suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})
		suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
		oracleSets := suite.Keeper().GetOracleSets(suite.ctx)
		require.NotNil(suite.T(), oracleSets)
		require.EqualValues(suite.T(), i+1, len(oracleSets))

		power := suite.Keeper().GetLastTotalPower(suite.ctx)
		expectPower := suite.delegateAmount.Mul(sdk.NewInt(int64(i + 1))).Quo(sdk.DefaultPowerReduction)
		require.True(suite.T(), expectPower.Equal(power))
	}

	bridgeToken := "0x3f6795b8ABE0775a88973469909adE1405f7ac09"
	if trontypes.ModuleName == suite.chainName {
		bridgeToken = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	}

	for i := 0; i < 6; i++ {
		addBridgeTokenClaim := &types.MsgBridgeTokenClaim{
			EventNonce:     1,
			BlockHeight:    1000,
			TokenContract:  bridgeToken,
			Name:           "Pundix Reward Token",
			Symbol:         "PURES",
			Decimals:       18,
			BridgerAddress: suite.bridgers[i].String(),
			ChannelIbc:     hex.EncodeToString([]byte("transfer/channel-0")),
			ChainName:      suite.chainName,
		}
		_, err := suite.MsgServer().BridgeTokenClaim(sdk.WrapSDKContext(suite.ctx), addBridgeTokenClaim)
		require.NoError(suite.T(), err)
		endBlockBeforeAttestation := suite.Keeper().GetAttestation(suite.ctx, addBridgeTokenClaim.EventNonce, addBridgeTokenClaim.ClaimHash())
		require.NotNil(suite.T(), endBlockBeforeAttestation)
		require.False(suite.T(), endBlockBeforeAttestation.Observed)
		require.NotNil(suite.T(), endBlockBeforeAttestation.Votes)
		require.EqualValues(suite.T(), i+1, len(endBlockBeforeAttestation.Votes))

		suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})
		suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
		endBlockAfterAttestation := suite.Keeper().GetAttestation(suite.ctx, addBridgeTokenClaim.EventNonce, addBridgeTokenClaim.ClaimHash())
		require.NotNil(suite.T(), endBlockAfterAttestation)
		require.False(suite.T(), endBlockAfterAttestation.Observed)
	}

	addBridgeTokenClaim := &types.MsgBridgeTokenClaim{
		EventNonce:     1,
		BlockHeight:    1000,
		TokenContract:  bridgeToken,
		Name:           "Pundix Reward Token",
		Symbol:         "PURES",
		Decimals:       18,
		BridgerAddress: suite.bridgers[6].String(),
		ChannelIbc:     hex.EncodeToString([]byte("transfer/channel-0")),
		ChainName:      suite.chainName,
	}
	_, err := suite.MsgServer().BridgeTokenClaim(sdk.WrapSDKContext(suite.ctx), addBridgeTokenClaim)
	require.NoError(suite.T(), err)
	suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
	attestation := suite.Keeper().GetAttestation(suite.ctx, addBridgeTokenClaim.EventNonce, addBridgeTokenClaim.ClaimHash())

	require.NotNil(suite.T(), attestation)

	require.True(suite.T(), attestation.Observed)

	proposalHandler := crosschain.NewChainProposalHandler(suite.app.CrosschainKeeper)

	var newOracleList []string
	for i := 0; i < 7; i++ {
		newOracleList = append(newOracleList, suite.oracles[i].String())
	}
	err = proposalHandler(suite.ctx, &types.UpdateChainOraclesProposal{
		Title:       "proposal 1: try update chain oracle power >= 30%, expect error",
		Description: "",
		Oracles:     newOracleList,
		ChainName:   suite.chainName,
	})
	require.ErrorIs(suite.T(), types.ErrInvalid, err)

	expectTotalPower := suite.delegateAmount.Mul(sdk.NewInt(10)).Quo(sdk.DefaultPowerReduction)
	actualTotalPower := suite.Keeper().GetLastTotalPower(suite.ctx)
	require.True(suite.T(), expectTotalPower.Equal(actualTotalPower))

	expectMaxChangePower := types.AttestationProposalOracleChangePowerThreshold.Mul(expectTotalPower).Quo(sdk.NewInt(100))

	expectDeletePower := suite.delegateAmount.Mul(sdk.NewInt(3)).Quo(sdk.DefaultPowerReduction)
	require.EqualValues(suite.T(), fmt.Sprintf("max change power, maxChangePowerThreshold: %s, deleteTotalPower: %s: %s", expectMaxChangePower.String(), expectDeletePower.String(), types.ErrInvalid), err.Error())

	var newOracleList2 []string
	for i := 0; i < 8; i++ {
		newOracleList2 = append(newOracleList2, suite.oracles[i].String())
	}
	err = proposalHandler(suite.ctx, &types.UpdateChainOraclesProposal{
		Title:       "proposal 2: try update chain oracle power <= 30%, expect success",
		Description: "",
		Oracles:     newOracleList2,
		ChainName:   suite.chainName,
	})
	require.NoError(suite.T(), err)
}

func (suite *IntegrationTestSuite) TestAttestationAfterOracleUpdate() {

	for i := 0; i < 20; i++ {
		msgBondedOracle := &types.MsgBondedOracle{
			OracleAddress:    suite.oracles[i].String(),
			BridgerAddress:   suite.bridgers[i].String(),
			ExternalAddress:  crypto.PubkeyToAddress(suite.externals[i].PublicKey).Hex(),
			ValidatorAddress: suite.validator[i].String(),
			DelegateAmount:   sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: suite.delegateAmount},
			ChainName:        suite.chainName,
		}
		if trontypes.ModuleName == suite.chainName {
			msgBondedOracle.ExternalAddress = tronAddress.PubkeyToAddress(suite.externals[i].PublicKey).String()
		}
		_, err := suite.MsgServer().BondedOracle(sdk.WrapSDKContext(suite.ctx), msgBondedOracle)
		require.NoError(suite.T(), err)
		suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})
		suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
		oracleSets := suite.Keeper().GetOracleSets(suite.ctx)
		require.NotNil(suite.T(), oracleSets)
		require.EqualValues(suite.T(), i+1, len(oracleSets))

		power := suite.Keeper().GetLastTotalPower(suite.ctx)
		expectPower := suite.delegateAmount.Mul(sdk.NewInt(int64(i + 1))).Quo(sdk.DefaultPowerReduction)
		require.True(suite.T(), expectPower.Equal(power))
	}

	bridgeToken := "0x3f6795b8ABE0775a88973469909adE1405f7ac09"
	if trontypes.ModuleName == suite.chainName {
		bridgeToken = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	}

	{
		firstBridgeTokenClaim := &types.MsgBridgeTokenClaim{
			EventNonce:     1,
			BlockHeight:    1000,
			TokenContract:  bridgeToken,
			Name:           "Pundix Reward Token",
			Symbol:         "PURES",
			Decimals:       18,
			BridgerAddress: "",
			ChannelIbc:     hex.EncodeToString([]byte("transfer/channel-0")),
			ChainName:      suite.chainName,
		}

		for i := 0; i < 13; i++ {
			firstBridgeTokenClaim.BridgerAddress = suite.bridgers[i].String()
			_, err := suite.MsgServer().BridgeTokenClaim(sdk.WrapSDKContext(suite.ctx), firstBridgeTokenClaim)
			require.NoError(suite.T(), err)
			endBlockBeforeAttestation := suite.Keeper().GetAttestation(suite.ctx, firstBridgeTokenClaim.EventNonce, firstBridgeTokenClaim.ClaimHash())
			require.NotNil(suite.T(), endBlockBeforeAttestation)
			require.False(suite.T(), endBlockBeforeAttestation.Observed)
			require.NotNil(suite.T(), endBlockBeforeAttestation.Votes)
			require.EqualValues(suite.T(), i+1, len(endBlockBeforeAttestation.Votes))

			suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})
			suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
			endBlockAfterAttestation := suite.Keeper().GetAttestation(suite.ctx, firstBridgeTokenClaim.EventNonce, firstBridgeTokenClaim.ClaimHash())
			require.NotNil(suite.T(), endBlockAfterAttestation)
			require.False(suite.T(), endBlockAfterAttestation.Observed)
		}

		firstBridgeTokenClaim.BridgerAddress = suite.bridgers[13].String()
		_, err := suite.MsgServer().BridgeTokenClaim(sdk.WrapSDKContext(suite.ctx), firstBridgeTokenClaim)
		require.NoError(suite.T(), err)
		suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})
		suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
		attestation := suite.Keeper().GetAttestation(suite.ctx, firstBridgeTokenClaim.EventNonce, firstBridgeTokenClaim.ClaimHash())

		require.NotNil(suite.T(), attestation)

		require.True(suite.T(), attestation.Observed)
	}

	{
		secondBridgeTokenClaim := &types.MsgBridgeTokenClaim{
			EventNonce:     2,
			BlockHeight:    1001,
			TokenContract:  bridgeToken,
			Name:           "Pundix Reward Token2",
			Symbol:         "PURES2",
			Decimals:       18,
			BridgerAddress: "",
			ChannelIbc:     hex.EncodeToString([]byte("transfer/channel-0")),
			ChainName:      suite.chainName,
		}

		for i := 0; i < 6; i++ {
			secondBridgeTokenClaim.BridgerAddress = suite.bridgers[i].String()
			_, err := suite.MsgServer().BridgeTokenClaim(sdk.WrapSDKContext(suite.ctx), secondBridgeTokenClaim)
			require.NoError(suite.T(), err)
			endBlockBeforeAttestation := suite.Keeper().GetAttestation(suite.ctx, secondBridgeTokenClaim.EventNonce, secondBridgeTokenClaim.ClaimHash())
			require.NotNil(suite.T(), endBlockBeforeAttestation)
			require.False(suite.T(), endBlockBeforeAttestation.Observed)
			require.NotNil(suite.T(), endBlockBeforeAttestation.Votes)
			require.EqualValues(suite.T(), i+1, len(endBlockBeforeAttestation.Votes))

			suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})
			suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
			endBlockAfterAttestation := suite.Keeper().GetAttestation(suite.ctx, secondBridgeTokenClaim.EventNonce, secondBridgeTokenClaim.ClaimHash())
			require.NotNil(suite.T(), endBlockAfterAttestation)
			require.False(suite.T(), endBlockAfterAttestation.Observed)
		}

		secondClaimAttestation := suite.Keeper().GetAttestation(suite.ctx, secondBridgeTokenClaim.EventNonce, secondBridgeTokenClaim.ClaimHash())
		require.NotNil(suite.T(), secondClaimAttestation)
		require.False(suite.T(), secondClaimAttestation.Observed)
		require.NotNil(suite.T(), secondClaimAttestation.Votes)
		require.EqualValues(suite.T(), 6, len(secondClaimAttestation.Votes))

		var newOracleList []string
		for i := 0; i < 15; i++ {
			newOracleList = append(newOracleList, suite.oracles[i].String())
		}
		err := crosschain.HandleUpdateChainOraclesProposal(suite.ctx, suite.MsgServer(), &types.UpdateChainOraclesProposal{
			Title:       "proposal 1: try update chain oracle save top 15 oracle, expect success",
			Description: "",
			Oracles:     newOracleList,
			ChainName:   suite.chainName,
		})
		require.NoError(suite.T(), err)
		suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
		suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})

		secondClaimAttestation = suite.Keeper().GetAttestation(suite.ctx, secondBridgeTokenClaim.EventNonce, secondBridgeTokenClaim.ClaimHash())
		require.NotNil(suite.T(), secondClaimAttestation)
		require.False(suite.T(), secondClaimAttestation.Observed)
		require.NotNil(suite.T(), secondClaimAttestation.Votes)
		require.EqualValues(suite.T(), 6, len(secondClaimAttestation.Votes))

		activeOracles := suite.Keeper().GetAllOracles(suite.ctx, true)
		require.NotNil(suite.T(), activeOracles)
		require.EqualValues(suite.T(), 15, len(activeOracles))
		for i := 0; i < 15; i++ {
			require.NotNil(suite.T(), newOracleList[i], activeOracles[i].OracleAddress)
		}

		var newOracleList2 []string
		for i := 0; i < 11; i++ {
			newOracleList2 = append(newOracleList2, suite.oracles[i].String())
		}
		err = crosschain.HandleUpdateChainOraclesProposal(suite.ctx, suite.MsgServer(), &types.UpdateChainOraclesProposal{
			Title:       "proposal 2: try update chain oracle save top 11 oracle, expect success",
			Description: "",
			Oracles:     newOracleList2,
			ChainName:   suite.chainName,
		})
		require.NoError(suite.T(), err)
		suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
		suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})

		secondClaimAttestation = suite.Keeper().GetAttestation(suite.ctx, secondBridgeTokenClaim.EventNonce, secondBridgeTokenClaim.ClaimHash())
		require.NotNil(suite.T(), secondClaimAttestation)
		require.False(suite.T(), secondClaimAttestation.Observed)
		require.NotNil(suite.T(), secondClaimAttestation.Votes)
		require.EqualValues(suite.T(), 6, len(secondClaimAttestation.Votes))

		activeOracles = suite.Keeper().GetAllOracles(suite.ctx, true)
		require.NotNil(suite.T(), activeOracles)
		require.EqualValues(suite.T(), 11, len(activeOracles))
		for i := 0; i < 11; i++ {
			require.NotNil(suite.T(), newOracleList2[i], activeOracles[i].OracleAddress)
		}

		var newOracleList3 []string
		for i := 0; i < 10; i++ {
			newOracleList3 = append(newOracleList3, suite.oracles[i].String())
		}
		err = crosschain.HandleUpdateChainOraclesProposal(suite.ctx, suite.MsgServer(), &types.UpdateChainOraclesProposal{
			Title:       "proposal 3: try update chain oracle save top 10 oracle, expect success",
			Description: "",
			Oracles:     newOracleList3,
			ChainName:   suite.chainName,
		})
		require.NoError(suite.T(), err)
		suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
		suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})

		secondClaimAttestation = suite.Keeper().GetAttestation(suite.ctx, secondBridgeTokenClaim.EventNonce, secondBridgeTokenClaim.ClaimHash())
		require.NotNil(suite.T(), secondClaimAttestation)
		require.False(suite.T(), secondClaimAttestation.Observed)
		require.NotNil(suite.T(), secondClaimAttestation.Votes)
		require.EqualValues(suite.T(), 6, len(secondClaimAttestation.Votes))

		activeOracles = suite.Keeper().GetAllOracles(suite.ctx, true)
		require.NotNil(suite.T(), activeOracles)
		require.EqualValues(suite.T(), 10, len(activeOracles))
		for i := 0; i < 10; i++ {
			require.NotNil(suite.T(), newOracleList3[i], activeOracles[i].OracleAddress)
		}

		secondBridgeTokenClaim.BridgerAddress = suite.bridgers[6].String()
		_, err = suite.MsgServer().BridgeTokenClaim(sdk.WrapSDKContext(suite.ctx), secondBridgeTokenClaim)
		require.NoError(suite.T(), err)

		suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
		suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})

		secondClaimAttestation = suite.Keeper().GetAttestation(suite.ctx, secondBridgeTokenClaim.EventNonce, secondBridgeTokenClaim.ClaimHash())
		require.NotNil(suite.T(), secondClaimAttestation)
		require.True(suite.T(), secondClaimAttestation.Observed)
		require.NotNil(suite.T(), secondClaimAttestation.Votes)
		require.EqualValues(suite.T(), 7, len(secondClaimAttestation.Votes))
	}
}

func (suite *IntegrationTestSuite) TestOracleDelete() {
	for i := 0; i < len(suite.oracles); i++ {
		msgBondedOracle := &types.MsgBondedOracle{
			OracleAddress:    suite.oracles[i].String(),
			BridgerAddress:   suite.bridgers[i].String(),
			ExternalAddress:  crypto.PubkeyToAddress(suite.externals[i].PublicKey).Hex(),
			ValidatorAddress: suite.validator[i].String(),
			DelegateAmount:   sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: suite.delegateAmount},
			ChainName:        suite.chainName,
		}
		if trontypes.ModuleName == suite.chainName {
			msgBondedOracle.ExternalAddress = tronAddress.PubkeyToAddress(suite.externals[i].PublicKey).String()
		}
		require.NoError(suite.T(), msgBondedOracle.ValidateBasic())
		_, err := suite.MsgServer().BondedOracle(sdk.WrapSDKContext(suite.ctx), msgBondedOracle)
		require.NoError(suite.T(), err)
	}
	suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
	allOracles := suite.Keeper().GetAllOracles(suite.ctx, false)
	require.NotNil(suite.T(), allOracles)
	require.EqualValues(suite.T(), len(suite.oracles), len(allOracles))

	oracle := suite.oracles[0]
	bridger := suite.bridgers[0]
	externalAddress := crypto.PubkeyToAddress(suite.externals[0].PublicKey).Hex()
	if trontypes.ModuleName == suite.chainName {
		externalAddress = tronAddress.PubkeyToAddress(suite.externals[0].PublicKey).String()
	}

	oracleAddr, found := suite.Keeper().GetOracleAddressByBridgerKey(suite.ctx, bridger)
	require.True(suite.T(), found)
	require.EqualValues(suite.T(), oracle.String(), oracleAddr.String())

	oracleAddr, found = suite.Keeper().GetOracleByExternalAddress(suite.ctx, externalAddress)
	require.True(suite.T(), found)
	require.EqualValues(suite.T(), oracle.String(), oracleAddr.String())

	oracleData, found := suite.Keeper().GetOracle(suite.ctx, oracle)
	require.True(suite.T(), found)
	require.NotNil(suite.T(), oracleData)
	require.EqualValues(suite.T(), oracle.String(), oracleData.OracleAddress)
	require.EqualValues(suite.T(), bridger.String(), oracleData.BridgerAddress)
	require.EqualValues(suite.T(), externalAddress, oracleData.ExternalAddress)

	require.True(suite.T(), suite.delegateAmount.Equal(oracleData.DelegateAmount))

	var newOracleAddressList []string
	for _, address := range suite.oracles[1:] {
		newOracleAddressList = append(newOracleAddressList, address.String())
	}

	err := crosschain.HandleUpdateChainOraclesProposal(suite.ctx, suite.MsgServer(), &types.UpdateChainOraclesProposal{
		Title:       "proposal 1: try update chain oracle remove first oracle, expect success",
		Description: "",
		Oracles:     newOracleAddressList,
		ChainName:   suite.chainName,
	})
	require.NoError(suite.T(), err)
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
	suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})

	oracleAddr, found = suite.Keeper().GetOracleAddressByBridgerKey(suite.ctx, bridger)
	require.True(suite.T(), found)
	require.Equal(suite.T(), oracleAddr, oracle)

	oracleAddr, found = suite.Keeper().GetOracleByExternalAddress(suite.ctx, externalAddress)
	require.True(suite.T(), found)
	require.Equal(suite.T(), oracleAddr, oracle)

	oracleData, found = suite.Keeper().GetOracle(suite.ctx, oracle)
	require.True(suite.T(), found)
}

func (suite *IntegrationTestSuite) TestOracleSetSlash() {
	for i := 0; i < len(suite.oracles); i++ {
		msgBondedOracle := &types.MsgBondedOracle{
			OracleAddress:    suite.oracles[i].String(),
			BridgerAddress:   suite.bridgers[i].String(),
			ExternalAddress:  crypto.PubkeyToAddress(suite.externals[i].PublicKey).Hex(),
			ValidatorAddress: suite.validator[i].String(),
			DelegateAmount:   sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: suite.delegateAmount},
			ChainName:        suite.chainName,
		}
		if trontypes.ModuleName == suite.chainName {
			msgBondedOracle.ExternalAddress = tronAddress.PubkeyToAddress(suite.externals[i].PublicKey).String()
		}
		require.NoError(suite.T(), msgBondedOracle.ValidateBasic())
		_, err := suite.MsgServer().BondedOracle(sdk.WrapSDKContext(suite.ctx), msgBondedOracle)
		require.NoError(suite.T(), err)
	}
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
	crosschain.EndBlocker(suite.ctx, suite.Keeper())

	allOracles := suite.Keeper().GetAllOracles(suite.ctx, false)
	require.NotNil(suite.T(), allOracles)
	require.Equal(suite.T(), len(suite.oracles), len(allOracles))

	oracleSets := suite.Keeper().GetOracleSets(suite.ctx)
	require.NotNil(suite.T(), oracleSets)
	require.EqualValues(suite.T(), 1, len(oracleSets))

	gravityId := suite.Keeper().GetGravityID(suite.ctx)
	for i := 0; i < len(suite.oracles)-1; i++ {
		externalAddress := crypto.PubkeyToAddress(suite.externals[i].PublicKey).Hex()

		checkpoint, err := oracleSets[0].GetCheckpoint(gravityId)
		require.NoError(suite.T(), err)
		signature, err := types.NewEthereumSignature(checkpoint, suite.externals[i])
		require.NoError(suite.T(), err)

		if trontypes.ModuleName == suite.chainName {
			externalAddress = tronAddress.PubkeyToAddress(suite.externals[i].PublicKey).String()

			checkpoint, err = trontypes.GetCheckpointOracleSet(oracleSets[0], gravityId)
			require.NoError(suite.T(), err)

			signature, err = trontypes.NewTronSignature(checkpoint, suite.externals[i])
			require.NoError(suite.T(), err)
		}
		oracleSetConfirm := &types.MsgOracleSetConfirm{
			Nonce:           oracleSets[0].Nonce,
			BridgerAddress:  suite.bridgers[i].String(),
			ExternalAddress: externalAddress,
			Signature:       hex.EncodeToString(signature),
			ChainName:       suite.chainName,
		}
		require.NoError(suite.T(), oracleSetConfirm.ValidateBasic())
		_, err = suite.MsgServer().OracleSetConfirm(sdk.WrapSDKContext(suite.ctx), oracleSetConfirm)
		require.NoError(suite.T(), err)
	}

	crosschain.EndBlocker(suite.ctx, suite.Keeper())
	oracleSetHeight := int64(oracleSets[0].Height)
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
	suite.app.EndBlock(abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})

	oracle, found := suite.Keeper().GetOracle(suite.ctx, suite.oracles[len(suite.oracles)-1])
	require.True(suite.T(), found)
	require.True(suite.T(), oracle.Online)
	require.Equal(suite.T(), int64(0), oracle.SlashTimes)

	suite.ctx = suite.ctx.WithBlockHeight(oracleSetHeight + int64(suite.Keeper().GetParams(suite.ctx).SignedWindow) + 1)
	crosschain.EndBlocker(suite.ctx, suite.Keeper())

	oracle, found = suite.Keeper().GetOracle(suite.ctx, suite.oracles[len(suite.oracles)-1])
	require.True(suite.T(), found)
	require.False(suite.T(), oracle.Online)
	require.Equal(suite.T(), int64(1), oracle.SlashTimes)
}

func (suite *IntegrationTestSuite) TestSlashOracle() {
	for i := 0; i < len(suite.oracles); i++ {
		msgBondedOracle := &types.MsgBondedOracle{
			OracleAddress:    suite.oracles[i].String(),
			BridgerAddress:   suite.bridgers[i].String(),
			ExternalAddress:  crypto.PubkeyToAddress(suite.externals[i].PublicKey).Hex(),
			ValidatorAddress: suite.validator[i].String(),
			DelegateAmount:   sdk.Coin{Denom: fxtypes.DefaultDenom, Amount: suite.delegateAmount},
			ChainName:        suite.chainName,
		}
		if trontypes.ModuleName == suite.chainName {
			msgBondedOracle.ExternalAddress = tronAddress.PubkeyToAddress(suite.externals[i].PublicKey).String()
		}
		require.NoError(suite.T(), msgBondedOracle.ValidateBasic())
		_, err := suite.MsgServer().BondedOracle(sdk.WrapSDKContext(suite.ctx), msgBondedOracle)
		require.NoError(suite.T(), err)
	}

	params := suite.Keeper().GetParams(suite.ctx)
	suite.Keeper().SetParams(suite.ctx, &params)

	for i := 0; i < len(suite.oracles); i++ {
		oracle, found := suite.Keeper().GetOracle(suite.ctx, suite.oracles[i])
		require.True(suite.T(), found)
		require.True(suite.T(), oracle.Online)
		require.Equal(suite.T(), int64(0), oracle.SlashTimes)

		suite.Keeper().SlashOracle(suite.ctx, oracle.OracleAddress)

		oracle, found = suite.Keeper().GetOracle(suite.ctx, suite.oracles[i])
		require.True(suite.T(), found)
		require.False(suite.T(), oracle.Online)
		require.Equal(suite.T(), int64(1), oracle.SlashTimes)
	}

	// repeat slash test.
	for i := 0; i < len(suite.oracles); i++ {
		oracle, found := suite.Keeper().GetOracle(suite.ctx, suite.oracles[i])
		require.True(suite.T(), found)
		require.False(suite.T(), oracle.Online)
		require.Equal(suite.T(), int64(1), oracle.SlashTimes)

		suite.Keeper().SlashOracle(suite.ctx, oracle.OracleAddress)

		oracle, found = suite.Keeper().GetOracle(suite.ctx, suite.oracles[i])
		require.True(suite.T(), found)
		require.False(suite.T(), oracle.Online)
		require.Equal(suite.T(), int64(1), oracle.SlashTimes)
	}
}
