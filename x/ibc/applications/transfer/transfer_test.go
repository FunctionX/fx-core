package transfer_test

import (
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"testing"

	fxtypes "github.com/functionx/fx-core/v2/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"

	"github.com/functionx/fx-core/v2/x/ibc/applications/transfer/types"
	ibctesting "github.com/functionx/fx-core/v2/x/ibc/testing"
)

var (
	defaultMsgRouter = ""
)

type TransferTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
	chainC *ibctesting.TestChain
}

func (suite *TransferTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 3)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))
	suite.chainC = suite.coordinator.GetChain(ibctesting.GetChainID(3))
}

func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	return path
}

// constructs a send from chainA to chainB on the established channel/connection
// and sends the same coin back from chainB to chainA.
func (suite *TransferTestSuite) TestHandleMsgTransfer() {
	// setup between chainA and chainB
	path := NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	//	originalBalance := suite.chainA.GetSimApp().BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), fxtypes.DefaultDenom)
	timeoutHeight := clienttypes.NewHeight(0, 110)

	amount, ok := sdk.NewIntFromString("9223372036854775808") // 2^63 (one above int64)
	suite.Require().True(ok)
	coinToSendToB := sdk.NewCoin(fxtypes.DefaultDenom, amount)

	// send from chainA to chainB
	msg := types.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coinToSendToB, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, defaultMsgRouter, sdk.NewCoin(coinToSendToB.GetDenom(), sdk.ZeroInt()))
	res, err := suite.chainA.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// relay send
	err = path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	// check that voucher exists on chain B
	voucherDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), fxtypes.DefaultDenom))
	balance := suite.chainB.GetSimApp().BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), voucherDenomTrace.IBCDenom())

	coinSentFromAToB := transfertypes.GetTransferCoin(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, fxtypes.DefaultDenom, amount)
	suite.Require().Equal(coinSentFromAToB, balance)

	// setup between chainB to chainC
	// NOTE:
	// pathBtoC.EndpointA = endpoint on chainB
	// pathBtoC.EndpointB = endpoint on chainC
	pathBtoC := NewTransferPath(suite.chainB, suite.chainC)
	suite.coordinator.Setup(pathBtoC)

	// send from chainB to chainC
	msg = types.NewMsgTransfer(pathBtoC.EndpointA.ChannelConfig.PortID, pathBtoC.EndpointA.ChannelID, coinSentFromAToB, suite.chainB.SenderAccount.GetAddress().String(), suite.chainC.SenderAccount.GetAddress().String(), timeoutHeight, 0, defaultMsgRouter, sdk.NewCoin(coinSentFromAToB.GetDenom(), sdk.ZeroInt()))
	res, err = suite.chainB.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	err = pathBtoC.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	// NOTE: fungible token is prefixed with the full trace in order to verify the packet commitment
	fullDenomPath := transfertypes.GetPrefixedDenom(pathBtoC.EndpointB.ChannelConfig.PortID, pathBtoC.EndpointB.ChannelID, voucherDenomTrace.GetFullDenomPath())

	coinSentFromBToC := sdk.NewCoin(transfertypes.ParseDenomTrace(fullDenomPath).IBCDenom(), amount)
	balance = suite.chainC.GetSimApp().BankKeeper.GetBalance(suite.chainC.GetContext(), suite.chainC.SenderAccount.GetAddress(), coinSentFromBToC.Denom)

	// check that the balance is updated on chainC
	suite.Require().Equal(coinSentFromBToC, balance)

	// check that balance on chain B is empty
	balance = suite.chainB.GetSimApp().BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), coinSentFromBToC.Denom)
	suite.Require().Zero(balance.Amount.Int64())

	// send from chainC back to chainB
	msg = types.NewMsgTransfer(pathBtoC.EndpointB.ChannelConfig.PortID, pathBtoC.EndpointB.ChannelID, coinSentFromBToC, suite.chainC.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, defaultMsgRouter, sdk.NewCoin(coinSentFromBToC.GetDenom(), sdk.ZeroInt()))
	res, err = suite.chainC.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	err = pathBtoC.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	balance = suite.chainB.GetSimApp().BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), coinSentFromAToB.Denom)

	// check that the balance on chainA returned back to the original state
	suite.Require().Equal(coinSentFromAToB, balance)

	// check that module account escrow address is empty
	escrowAddress := transfertypes.GetEscrowAddress(packet.GetDestPort(), packet.GetDestChannel())
	balance = suite.chainB.GetSimApp().BankKeeper.GetBalance(suite.chainB.GetContext(), escrowAddress, fxtypes.DefaultDenom)
	suite.Require().Equal(sdk.NewCoin(fxtypes.DefaultDenom, sdk.ZeroInt()), balance)

	// check that balance on chain B is empty
	balance = suite.chainC.GetSimApp().BankKeeper.GetBalance(suite.chainC.GetContext(), suite.chainC.SenderAccount.GetAddress(), voucherDenomTrace.IBCDenom())
	suite.Require().Zero(balance.Amount.Int64())
}

func TestTransferTestSuite(t *testing.T) {
	suite.Run(t, new(TransferTestSuite))
}
