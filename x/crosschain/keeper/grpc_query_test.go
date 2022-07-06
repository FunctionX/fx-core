package keeper_test

import (
	"encoding/hex"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/functionx/fx-core/app"
	"github.com/functionx/fx-core/app/helpers"
	"github.com/functionx/fx-core/x/crosschain/keeper"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/functionx/fx-core/x/crosschain/types"
)

type CrossChainGrpcTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	chainName string
	oracles   []sdk.AccAddress
	bridgers  []sdk.AccAddress

	msgServer   types.MsgServer
	queryClient types.QueryClient
}

func TestCrossChainGrpcTestSuite(t *testing.T) {
	suite.Run(t, new(CrossChainGrpcTestSuite))
}

func (suite *CrossChainGrpcTestSuite) SetupTest() {
	valSet, valAccounts, valBalances := helpers.GenerateGenesisValidator(types.MaxOracleSize, sdk.Coins{})
	suite.app = helpers.SetupWithGenesisValSet(suite.T(), valSet, valAccounts, valBalances...)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.CrosschainKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	suite.chainName = "bsc"
	suite.oracles = helpers.AddTestAddrs(suite.app, suite.ctx, types.MaxOracleSize, sdk.NewInt(300*1e3).MulRaw(1e18))
	suite.bridgers = helpers.AddTestAddrs(suite.app, suite.ctx, types.MaxOracleSize, sdk.NewInt(300*1e3).MulRaw(1e18))
	suite.msgServer = keeper.NewMsgServerImpl(suite.app.BscKeeper)

}

func (suite *CrossChainGrpcTestSuite) Keeper() keeper.Keeper {
	return suite.app.BscKeeper
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_CurrentOracleSet() {
	testCases := []struct {
		name          string
		malleate      func() *types.QueryCurrentOracleSetResponse
		expectedError error
		expPass       bool
	}{
		{
			name: "no oracle set",
			malleate: func() *types.QueryCurrentOracleSetResponse {
				return &types.QueryCurrentOracleSetResponse{OracleSet: types.NewOracleSet(1, 0, nil)}
			},
			expPass: true,
		},
		{
			name: "query current oracle set",
			malleate: func() *types.QueryCurrentOracleSetResponse {
				newOracleSet := &types.OracleSet{
					Members: make([]types.BridgeValidator, 0),
				}
				for i := 0; i < 6; i++ {
					key, _ := ethsecp256k1.GenerateKey()
					externalAcc := common.BytesToAddress(key.PubKey().Address())
					delegateAmount := sdk.DefaultPowerReduction.Mul(sdk.NewInt(100))
					if i == 5 {
						delegateAmount = sdk.ZeroInt()
					}
					suite.Keeper().SetOracle(suite.ctx, types.Oracle{
						OracleAddress:   suite.oracles[i].String(),
						BridgerAddress:  suite.bridgers[i].String(),
						ExternalAddress: externalAcc.String(),
						DelegateAmount:  delegateAmount,
						Online:          true,
						StartHeight:     int64(10 + i),
					})
					if i != 5 {
						newOracleSet.Members = append(newOracleSet.Members, types.BridgeValidator{
							Power:           858993459,
							ExternalAddress: externalAcc.String(),
						})
					}
				}
				suite.ctx = suite.ctx.WithBlockHeight(100)
				newOracleSet.Height = 100
				suite.Keeper().SetLatestOracleSetNonce(suite.ctx, 10)
				newOracleSet.Nonce = 11
				return &types.QueryCurrentOracleSetResponse{OracleSet: newOracleSet}
			},
			expPass: true,
		},
	}
	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			response := testCase.malleate()
			res, err := suite.Keeper().CurrentOracleSet(sdk.WrapSDKContext(suite.ctx), nil)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().ElementsMatch(response.OracleSet.Members, res.OracleSet.Members)
				suite.Require().Equal(response.OracleSet.Nonce, res.OracleSet.Nonce)
				suite.Require().Equal(response.OracleSet.Height, res.OracleSet.Height)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, testCase.expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_OracleSetRequest() {
	var (
		request       *types.QueryOracleSetRequestRequest
		response      *types.QueryCurrentOracleSetResponse
		expectedError error
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "oracle set nonce does not exist",
			malleate: func() {
				request = &types.QueryOracleSetRequestRequest{
					Nonce: 1,
				}
				response = &types.QueryCurrentOracleSetResponse{OracleSet: nil}
			},
			expPass: true,
		},
		{
			name: "oracle set nonce is zero",
			malleate: func() {
				request = &types.QueryOracleSetRequestRequest{
					Nonce: 0,
				}
				expectedError = sdkerrors.Wrapf(types.ErrUnknown, "nonce")
			},
			expPass: false,
		},
		{
			name: "normal oracle set",
			malleate: func() {
				members := []types.BridgeValidator{
					{
						Power:           10000,
						ExternalAddress: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
					},
				}
				request = &types.QueryOracleSetRequestRequest{
					Nonce: 3,
				}
				suite.Keeper().StoreOracleSet(suite.ctx, &types.OracleSet{
					Nonce:   3,
					Members: members,
					Height:  100,
				})
				response = &types.QueryCurrentOracleSetResponse{
					OracleSet: types.NewOracleSet(3, 100, members),
				}
			},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().OracleSetRequest(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response.OracleSet, res.OracleSet)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_OracleSetConfirm() {
	var (
		request       *types.QueryOracleSetConfirmRequest
		response      *types.QueryOracleSetConfirmResponse
		expectedError error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "oracle set bridger address error",
			malleate: func() {
				request = &types.QueryOracleSetConfirmRequest{
					ChainName:      suite.chainName,
					BridgerAddress: "fx1",
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "bridger address")
			},
			expPass: false,
		},
		{
			name: "oracle set nonce error",
			malleate: func() {
				request = &types.QueryOracleSetConfirmRequest{
					ChainName:      suite.chainName,
					BridgerAddress: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
					Nonce:          0,
				}
				expectedError = sdkerrors.Wrap(types.ErrUnknown, "nonce")
			},
			expPass: false,
		},
		{
			name: "oracle set bridger address does not exist",
			malleate: func() {
				request = &types.QueryOracleSetConfirmRequest{
					ChainName:      suite.chainName,
					BridgerAddress: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
					Nonce:          3,
				}
				expectedError = types.ErrNoFoundOracle
			},
			expPass: false,
		},
		{
			"oracle set normal",
			func() {
				request = &types.QueryOracleSetConfirmRequest{
					ChainName:      suite.chainName,
					BridgerAddress: suite.bridgers[0].String(),
					Nonce:          3,
				}
				suite.Keeper().SetOracleByBridger(suite.ctx, suite.oracles[0], suite.bridgers[0])
				suite.Keeper().SetOracleSetConfirm(suite.ctx, suite.oracles[0], &types.MsgOracleSetConfirm{
					Nonce:          3,
					BridgerAddress: suite.bridgers[0].String(),
					ChainName:      suite.chainName,
				})
				response = &types.QueryOracleSetConfirmResponse{
					Confirm: &types.MsgOracleSetConfirm{
						Nonce:          3,
						BridgerAddress: suite.bridgers[0].String(),
						ChainName:      suite.chainName,
					},
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().OracleSetConfirm(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_OracleSetConfirmsByNonce() {
	var (
		request       *types.QueryOracleSetConfirmsByNonceRequest
		response      *types.QueryOracleSetConfirmsByNonceResponse
		expectedError error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "query nonce is zero",
			malleate: func() {
				request = &types.QueryOracleSetConfirmsByNonceRequest{
					ChainName: suite.chainName,
					Nonce:     0,
				}
				expectedError = sdkerrors.Wrapf(types.ErrUnknown, "nonce")
			},
			expPass: false,
		},
		{
			name: "query nonce does not exist",
			malleate: func() {
				request = &types.QueryOracleSetConfirmsByNonceRequest{
					ChainName: suite.chainName,
					Nonce:     5,
				}
				response = &types.QueryOracleSetConfirmsByNonceResponse{}
			},
			expPass: true,
		},
		{
			name: "query nonce normal",
			malleate: func() {
				suite.Keeper().SetOracleByBridger(suite.ctx, suite.oracles[0], suite.bridgers[0])
				suite.Keeper().SetOracleSetConfirm(suite.ctx, suite.oracles[0], &types.MsgOracleSetConfirm{
					Nonce:          3,
					BridgerAddress: suite.bridgers[0].String(),
					ChainName:      suite.chainName,
				})
				request = &types.QueryOracleSetConfirmsByNonceRequest{
					ChainName: suite.chainName,
					Nonce:     3,
				}
				response = &types.QueryOracleSetConfirmsByNonceResponse{Confirms: []*types.MsgOracleSetConfirm{
					{
						Nonce:          3,
						BridgerAddress: suite.bridgers[0].String(),
						ChainName:      suite.chainName,
					},
				}}
			},
			expPass: true,
		},
	}
	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().OracleSetConfirmsByNonce(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_LastOracleSetRequest() {
	testCases := []struct {
		name          string
		malleate      func() *types.QueryLastOracleSetRequestsResponse
		expectedError error
		expPass       bool
	}{
		{
			name: "query params",
			malleate: func() *types.QueryLastOracleSetRequestsResponse {
				oracleSetList := make([]*types.OracleSet, 0)
				for i := 0; i < 10; i++ {
					key, _ := ethsecp256k1.GenerateKey()
					newOracleSet := &types.OracleSet{
						Nonce: uint64(i),
						Members: []types.BridgeValidator{
							{
								Power:           100000,
								ExternalAddress: common.BytesToAddress(key.PubKey().Address().Bytes()).String(),
							},
						},
						Height: uint64((i + 1) * 33),
					}
					suite.Keeper().StoreOracleSet(suite.ctx, newOracleSet)
					oracleSetList = append(oracleSetList, newOracleSet)
				}
				return &types.QueryLastOracleSetRequestsResponse{
					OracleSets: oracleSetList[len(oracleSetList)-5:],
				}
			},
			expPass: true,
		},
	}
	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			response := testCase.malleate()
			res, err := suite.Keeper().LastOracleSetRequests(sdk.WrapSDKContext(suite.ctx), nil)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().ElementsMatch(response.OracleSets, res.OracleSets)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, testCase.expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_LastPendingOracleSetRequestByAddr() {
	var (
		request       *types.QueryLastPendingOracleSetRequestByAddrRequest
		response      *types.QueryLastPendingOracleSetRequestByAddrResponse
		expectedError error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "query oracla set address error",
			malleate: func() {
				request = &types.QueryLastPendingOracleSetRequestByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: "fx1",
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "bridger address")
			},
			expPass: false,
		},
		{
			name: "not found oracle address by bridger",
			malleate: func() {
				request = &types.QueryLastPendingOracleSetRequestByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
				}
				expectedError = types.ErrNoFoundOracle
			},
			expPass: false,
		},
		{
			name: "not found oracle by oracle address",
			malleate: func() {
				suite.Keeper().SetOracleByBridger(suite.ctx, suite.oracles[0], suite.bridgers[0])
				request = &types.QueryLastPendingOracleSetRequestByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: suite.bridgers[0].String(),
				}
				expectedError = types.ErrNoFoundOracle
			},
			expPass: false,
		},
		{
			name: "not found oracle by oracle address",
			malleate: func() {
				key, err := ethsecp256k1.GenerateKey()
				suite.Require().NoError(err)
				externalAcc := common.BytesToAddress(key.PubKey().Address().Bytes())
				suite.ctx = suite.ctx.WithBlockHeight(100)

				oracleSet := &types.OracleSet{
					Nonce: 3,
					Members: []types.BridgeValidator{
						{
							Power:           10000,
							ExternalAddress: externalAcc.String(),
						},
					},
					Height: 100,
				}

				suite.Keeper().SetOracleByBridger(suite.ctx, suite.oracles[0], suite.bridgers[0])

				suite.Keeper().SetOracle(suite.ctx, types.Oracle{
					OracleAddress:   suite.oracles[0].String(),
					BridgerAddress:  suite.bridgers[0].String(),
					ExternalAddress: externalAcc.String(),
					StartHeight:     0,
				})
				suite.Keeper().StoreOracleSet(suite.ctx, oracleSet)
				request = &types.QueryLastPendingOracleSetRequestByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: suite.bridgers[0].String(),
				}

				response = &types.QueryLastPendingOracleSetRequestByAddrResponse{
					OracleSets: []*types.OracleSet{oracleSet},
				}
			},
			expPass: true,
		},
	}
	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().LastPendingOracleSetRequestByAddr(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_BatchFees() {
	var (
		request       *types.QueryBatchFeeRequest
		response      *types.QueryBatchFeeResponse
		expectedError error
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "batch fee BaseFee is negative",
			malleate: func() {
				request = &types.QueryBatchFeeRequest{
					ChainName: suite.chainName,
					MinBatchFees: []types.MinBatchFee{
						{
							TokenContract: suite.bridgers[0].String(),
							BaseFee:       sdk.NewInt(-1),
						},
					},
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "base fee")
			},
			expPass: false,
		},
		{
			name: "batch fee normal",
			malleate: func() {
				externalKey, _ := ethsecp256k1.GenerateKey()
				externalAcc := common.BytesToAddress(externalKey.PubKey().Address())
				token := crypto.CreateAddress(common.BytesToAddress(externalKey.PubKey().Address()), 0).String()
				err := suite.app.BscKeeper.AttestationHandler(suite.ctx, types.Attestation{}, &types.MsgBridgeTokenClaim{
					TokenContract:  token,
					BridgerAddress: suite.bridgers[0].String(),
					ChannelIbc:     hex.EncodeToString([]byte("transfer/channel-0")),
					ChainName:      suite.chainName,
				})
				suite.Require().NoError(err)
				denom := suite.app.BscKeeper.GetBridgeTokenDenom(suite.ctx, token)
				initBalances := sdk.NewIntFromUint64(1e18).Mul(sdk.NewInt(20000))
				err = suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(denom.Denom, initBalances)))
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, suite.bridgers[0], sdk.NewCoins(sdk.NewCoin(denom.Denom, initBalances)))
				suite.Require().NoError(err)
				minBatchFee := []types.MinBatchFee{
					{
						TokenContract: denom.Token,
						BaseFee:       sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e6), big.NewInt(100))),
					},
				}
				for i := uint64(1); i <= 3; i++ {
					_, err := suite.app.BscKeeper.AddToOutgoingPool(
						suite.ctx,
						suite.bridgers[0],
						externalAcc.String(),
						sdk.NewCoin(denom.Denom, sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e6), big.NewInt(100)))),
						sdk.NewCoin(denom.Denom, sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e6), big.NewInt(100)))))
					suite.Require().NoError(err)
				}
				for i := uint64(1); i <= 2; i++ {
					_, err := suite.app.BscKeeper.AddToOutgoingPool(
						suite.ctx,
						suite.bridgers[0],
						externalAcc.String(),
						sdk.NewCoin(denom.Denom, sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e6), big.NewInt(100)))),
						sdk.NewCoin(denom.Denom, sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e6), big.NewInt(10)))))
					suite.Require().NoError(err)
				}
				request = &types.QueryBatchFeeRequest{
					ChainName:    suite.chainName,
					MinBatchFees: minBatchFee,
				}
				response = &types.QueryBatchFeeResponse{BatchFees: []*types.BatchFees{
					{
						TokenContract: denom.Token,
						TotalFees:     sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e6), big.NewInt(300))),
						TotalTxs:      3,
						TotalAmount:   sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e6), big.NewInt(300))),
					},
				}}
			},
			expPass: true,
		},
		{
			name: "batch fee mul normal",
			malleate: func() {
				bridgeTokenList := make([]*types.BridgeToken, 0)

				externalKey, _ := ethsecp256k1.GenerateKey()
				externalAcc := common.BytesToAddress(externalKey.PubKey().Address())

				for i := 0; i < 2; i++ {
					token := crypto.CreateAddress(common.BytesToAddress(externalKey.PubKey().Address()), uint64(i)).String()
					err := suite.app.BscKeeper.AttestationHandler(suite.ctx, types.Attestation{}, &types.MsgBridgeTokenClaim{
						TokenContract:  token,
						BridgerAddress: suite.bridgers[0].String(),
						ChannelIbc:     hex.EncodeToString([]byte("transfer/channel-0")),
						ChainName:      suite.chainName,
					})
					suite.Require().NoError(err)
					denom := suite.app.BscKeeper.GetBridgeTokenDenom(suite.ctx, token)
					bridgeTokenList = append(bridgeTokenList, denom)
					initBalances := sdk.NewIntFromUint64(1e18).Mul(sdk.NewInt(20000))
					err = suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(denom.Denom, initBalances)))
					suite.Require().NoError(err)
					err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, suite.bridgers[0], sdk.NewCoins(sdk.NewCoin(denom.Denom, initBalances)))
					suite.Require().NoError(err)
				}
				minBatchFee := []types.MinBatchFee{
					{
						TokenContract: bridgeTokenList[0].Token,
						BaseFee:       sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(100), big.NewInt(1e6))),
					},
					{
						TokenContract: bridgeTokenList[1].Token,
						BaseFee:       sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))),
					},
				}
				for i := uint64(1); i <= 2; i++ {
					_, err := suite.app.BscKeeper.AddToOutgoingPool(
						suite.ctx,
						suite.bridgers[0],
						externalAcc.String(),
						sdk.NewCoin(bridgeTokenList[0].Denom, sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e6), big.NewInt(100)))),
						sdk.NewCoin(bridgeTokenList[0].Denom, sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e6), big.NewInt(10)))))
					suite.Require().NoError(err)
				}
				_, err := suite.app.BscKeeper.AddToOutgoingPool(
					suite.ctx,
					suite.bridgers[0],
					externalAcc.String(),
					sdk.NewCoin(bridgeTokenList[0].Denom, sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e6), big.NewInt(100)))),
					sdk.NewCoin(bridgeTokenList[0].Denom, sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e6), big.NewInt(100)))))
				suite.Require().NoError(err)

				for i := uint64(1); i <= 3; i++ {
					_, err := suite.app.BscKeeper.AddToOutgoingPool(
						suite.ctx,
						suite.bridgers[0],
						externalAcc.String(),
						sdk.NewCoin(bridgeTokenList[1].Denom, sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e18), big.NewInt(100)))),
						sdk.NewCoin(bridgeTokenList[1].Denom, sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e18), big.NewInt(100)))))
					suite.Require().NoError(err)
				}
				request = &types.QueryBatchFeeRequest{
					ChainName:    suite.chainName,
					MinBatchFees: minBatchFee,
				}
				response = &types.QueryBatchFeeResponse{BatchFees: []*types.BatchFees{
					{
						TokenContract: bridgeTokenList[0].Token,
						TotalFees:     sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e6), big.NewInt(100))),
						TotalTxs:      1,
						TotalAmount:   sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e6), big.NewInt(100))),
					},
					{
						TokenContract: bridgeTokenList[1].Token,
						TotalFees:     sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e18), big.NewInt(300))),
						TotalTxs:      3,
						TotalAmount:   sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e18), big.NewInt(300))),
					},
				}}
			},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().BatchFees(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().ElementsMatch(response.BatchFees, res.BatchFees)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_LastPendingBatchRequestByAddr() {
	var (
		request       *types.QueryLastPendingBatchRequestByAddrRequest
		response      *types.QueryLastPendingBatchRequestByAddrResponse
		expectedError error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "bridger address error",
			malleate: func() {
				request = &types.QueryLastPendingBatchRequestByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: "fx1",
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "bridger address")
			},
			expPass: false,
		},
		{
			name: "not found oracle by bridger",
			malleate: func() {
				request = &types.QueryLastPendingBatchRequestByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
				}
				expectedError = types.ErrNoFoundOracle
			},
			expPass: false,
		},
		{
			name: "not found oracle",
			malleate: func() {
				suite.Keeper().SetOracleByBridger(suite.ctx, suite.oracles[0], suite.bridgers[0])
				request = &types.QueryLastPendingBatchRequestByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: suite.bridgers[0].String(),
				}
				expectedError = types.ErrNoFoundOracle
			},
			expPass: false,
		},
		{
			name: "normal test",
			malleate: func() {
				externalKey, err := ethsecp256k1.GenerateKey()
				suite.Require().NoError(err)
				externalAcc := common.BytesToAddress(externalKey.PubKey().Address().Bytes())
				externalToken := crypto.CreateAddress(common.BytesToAddress(externalKey.PubKey().Address().Bytes()), 0)

				suite.Keeper().SetOracleByBridger(suite.ctx, suite.oracles[0], suite.bridgers[0])
				suite.Keeper().SetOracle(suite.ctx, types.Oracle{
					OracleAddress:   suite.oracles[0].String(),
					BridgerAddress:  suite.bridgers[0].String(),
					ExternalAddress: externalAcc.String(),
					StartHeight:     10,
				})
				request = &types.QueryLastPendingBatchRequestByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: suite.bridgers[0].String(),
				}
				suite.ctx = suite.ctx.WithBlockHeight(100)
				err = suite.Keeper().StoreBatch(suite.ctx, &types.OutgoingTxBatch{
					BatchNonce:   3,
					BatchTimeout: 10000,
					Transactions: []*types.OutgoingTransferTx{
						{
							Id:          0,
							Sender:      sdk.AccAddress(externalKey.PubKey().Address()).String(),
							DestAddress: externalAcc.String(),
							Token:       types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), externalToken.String()),
							Fee:         types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), externalToken.String()),
						},
					},
					TokenContract: externalToken.String(),
					FeeReceive:    externalAcc.String(),
				})
				suite.Require().NoError(err)
				response = &types.QueryLastPendingBatchRequestByAddrResponse{Batch: &types.OutgoingTxBatch{
					BatchNonce:   3,
					BatchTimeout: 10000,
					Transactions: []*types.OutgoingTransferTx{
						{
							Id:          0,
							Sender:      sdk.AccAddress(externalKey.PubKey().Address()).String(),
							DestAddress: externalAcc.String(),
							Token:       types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), externalToken.String()),
							Fee:         types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), externalToken.String()),
						},
					},
					TokenContract: externalToken.String(),
					Block:         100,
					FeeReceive:    externalAcc.String(),
				}}
			},
			expPass: true,
		},
		{
			name: "test batch confirm tx",
			malleate: func() {
				externalKey, err := ethsecp256k1.GenerateKey()
				suite.Require().NoError(err)
				externalAcc := common.BytesToAddress(externalKey.PubKey().Address().Bytes())
				externalToken := crypto.CreateAddress(common.BytesToAddress(externalKey.PubKey().Address().Bytes()), 0)
				suite.Keeper().SetOracleByBridger(suite.ctx, suite.oracles[0], suite.bridgers[0])
				suite.Keeper().SetOracle(suite.ctx, types.Oracle{
					OracleAddress:   suite.oracles[0].String(),
					BridgerAddress:  suite.bridgers[0].String(),
					ExternalAddress: externalAcc.String(),
					StartHeight:     10,
				})
				request = &types.QueryLastPendingBatchRequestByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: suite.bridgers[0].String(),
				}
				suite.ctx = suite.ctx.WithBlockHeight(100)
				err = suite.Keeper().StoreBatch(suite.ctx, &types.OutgoingTxBatch{
					BatchNonce:   3,
					BatchTimeout: 10000,
					Transactions: []*types.OutgoingTransferTx{
						{
							Id:          0,
							Sender:      sdk.AccAddress(externalKey.PubKey().Address()).String(),
							DestAddress: externalAcc.String(),
							Token:       types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), externalToken.String()),
							Fee:         types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), externalToken.String()),
						},
					},
					TokenContract: externalToken.String(),
					FeeReceive:    externalAcc.String(),
				})
				suite.Require().NoError(err)
				suite.Keeper().SetBatchConfirm(suite.ctx, suite.oracles[0], &types.MsgConfirmBatch{
					Nonce:           3,
					TokenContract:   externalToken.String(),
					BridgerAddress:  suite.bridgers[0].String(),
					ExternalAddress: externalAcc.String(),
					Signature:       "0x1",
					ChainName:       suite.chainName,
				})
				response = &types.QueryLastPendingBatchRequestByAddrResponse{}
			},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().LastPendingBatchRequestByAddr(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response.Batch, res.Batch)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_OutgoingTxBatches() {
	testCases := []struct {
		name          string
		malleate      func() *types.QueryOutgoingTxBatchesResponse
		expectedError error
		expPass       bool
	}{
		{
			name: "query outgoing tx batches",
			malleate: func() *types.QueryOutgoingTxBatchesResponse {
				newBatchList := make([]*types.OutgoingTxBatch, 0)
				for i := 0; i < 10; i++ {
					suite.ctx = suite.ctx.WithBlockHeight(int64(i + 3))
					token := crypto.PubkeyToAddress(genEthKey(1)[0].PublicKey)
					newOutgoingTx := &types.OutgoingTxBatch{
						BatchNonce:   uint64(i + 3),
						BatchTimeout: uint64(1000),
						Transactions: []*types.OutgoingTransferTx{
							{
								Id:          uint64(i),
								Sender:      sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
								DestAddress: crypto.PubkeyToAddress(genEthKey(1)[0].PublicKey).String(),
								Token:       types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), token.String()),
								Fee:         types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), token.String()),
							},
						},
						TokenContract: token.String(),
						Block:         uint64(i + 3),
						FeeReceive:    crypto.PubkeyToAddress(genEthKey(1)[0].PublicKey).String(),
					}
					err := suite.Keeper().StoreBatch(suite.ctx, newOutgoingTx)
					suite.Require().NoError(err)
					newBatchList = append(newBatchList, newOutgoingTx)
				}
				return &types.QueryOutgoingTxBatchesResponse{Batches: newBatchList}
			},
			expPass: true,
		},
		{
			name: "query outgoing tx batches more than 100",
			malleate: func() *types.QueryOutgoingTxBatchesResponse {
				for i := 1; i < 110; i++ {
					suite.ctx = suite.ctx.WithBlockHeight(int64(i))
					token := crypto.PubkeyToAddress(genEthKey(1)[0].PublicKey)
					newOutgoingTx := &types.OutgoingTxBatch{
						BatchNonce:   uint64(i),
						BatchTimeout: uint64(1000 + i),
						Transactions: []*types.OutgoingTransferTx{
							{
								Id:          uint64(i),
								Sender:      sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
								DestAddress: crypto.PubkeyToAddress(genEthKey(1)[0].PublicKey).String(),
								Token:       types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), token.String()),
								Fee:         types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), token.String()),
							},
						},
						TokenContract: token.String(),
						Block:         uint64(i),
						FeeReceive:    crypto.PubkeyToAddress(genEthKey(1)[0].PublicKey).String(),
					}
					err := suite.Keeper().StoreBatch(suite.ctx, newOutgoingTx)
					suite.Require().NoError(err)
				}
				return &types.QueryOutgoingTxBatchesResponse{}
			},
			expPass: true,
		},
	}
	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			response := testCase.malleate()
			res, err := suite.Keeper().OutgoingTxBatches(sdk.WrapSDKContext(suite.ctx), nil)
			suite.Require().NoError(err)
			if testCase.expPass {
				suite.Require().True(len(res.Batches) <= 100)
				if len(res.Batches) < 100 {
					suite.Require().ElementsMatch(response.Batches, res.Batches)
				}
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, testCase.expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_BatchRequestByNonce() {
	var (
		request       *types.QueryBatchRequestByNonceRequest
		response      *types.QueryBatchRequestByNonceResponse
		expectedError error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "query token contract error",
			malleate: func() {
				request = &types.QueryBatchRequestByNonceRequest{
					ChainName:     suite.chainName,
					TokenContract: "0x1",
					Nonce:         3,
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "token contract address")
			},
			expPass: false,
		},
		{
			name: "query token contract error",
			malleate: func() {
				key, _ := ethsecp256k1.GenerateKey()
				request = &types.QueryBatchRequestByNonceRequest{
					ChainName:     suite.chainName,
					TokenContract: crypto.CreateAddress(common.BytesToAddress(key.PubKey().Address().Bytes()), 0).String(),
					Nonce:         0,
				}
				expectedError = sdkerrors.Wrap(types.ErrUnknown, "nonce")
			},
			expPass: false,
		},
		{
			name: "query does not exist tx batch",
			malleate: func() {
				key, _ := ethsecp256k1.GenerateKey()
				request = &types.QueryBatchRequestByNonceRequest{
					ChainName:     suite.chainName,
					TokenContract: crypto.CreateAddress(common.BytesToAddress(key.PubKey().Address().Bytes()), 0).String(),
					Nonce:         3,
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "can not find tx batch")
			},
			expPass: false,
		},
		{
			name: "query tx batch normal",
			malleate: func() {
				key, _ := ethsecp256k1.GenerateKey()
				token := crypto.CreateAddress(common.BytesToAddress(key.PubKey().Address().Bytes()), 0)

				newBatch := &types.OutgoingTxBatch{
					BatchNonce:   3,
					BatchTimeout: 10000,
					Transactions: []*types.OutgoingTransferTx{
						{
							Id:    0,
							Token: types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), token.String()),
							Fee:   types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), token.String()),
						},
					},
					TokenContract: token.String(),
					Block:         100,
				}
				err := suite.Keeper().StoreBatch(suite.ctx, newBatch)
				suite.Require().NoError(err)
				request = &types.QueryBatchRequestByNonceRequest{
					ChainName:     suite.chainName,
					TokenContract: token.String(),
					Nonce:         3,
				}
				response = &types.QueryBatchRequestByNonceResponse{Batch: newBatch}
			},
			expPass: true,
		},
	}
	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().BatchRequestByNonce(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_BatchConfirm() {
	var (
		request       *types.QueryBatchConfirmRequest
		response      *types.QueryBatchConfirmResponse
		expectedError error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "bridger address error",
			malleate: func() {
				request = &types.QueryBatchConfirmRequest{
					ChainName:      suite.chainName,
					BridgerAddress: "fx1",
					Nonce:          3,
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "bridger address")
			},
			expPass: false,
		},
		{
			name: "query nonce error",
			malleate: func() {
				request = &types.QueryBatchConfirmRequest{
					ChainName:      suite.chainName,
					BridgerAddress: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
					Nonce:          0,
				}
				expectedError = sdkerrors.Wrap(types.ErrUnknown, "nonce")
			},
			expPass: false,
		},
		{
			name: "query oracle not found",
			malleate: func() {
				request = &types.QueryBatchConfirmRequest{
					ChainName:      suite.chainName,
					BridgerAddress: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
					Nonce:          3,
				}
				expectedError = types.ErrNoFoundOracle
			},
			expPass: false,
		},
		{
			name: "query batch confirm normal",
			malleate: func() {
				suite.Keeper().SetOracleByBridger(suite.ctx, suite.oracles[0], suite.bridgers[0])

				suite.Keeper().SetBatchConfirm(suite.ctx, suite.oracles[0], &types.MsgConfirmBatch{
					Nonce:          3,
					BridgerAddress: suite.bridgers[0].String(),
					ChainName:      suite.chainName,
				})
				request = &types.QueryBatchConfirmRequest{
					ChainName:      suite.chainName,
					BridgerAddress: suite.bridgers[0].String(),
					Nonce:          3,
				}
				response = &types.QueryBatchConfirmResponse{Confirm: &types.MsgConfirmBatch{
					Nonce:          3,
					BridgerAddress: suite.bridgers[0].String(),
					ChainName:      suite.chainName,
				}}
			},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().BatchConfirm(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_BatchConfirms() {
	var (
		request       *types.QueryBatchConfirmsRequest
		response      *types.QueryBatchConfirmsResponse
		expectedError error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "query token address error",
			malleate: func() {
				request = &types.QueryBatchConfirmsRequest{
					ChainName:     suite.chainName,
					TokenContract: "0x11",
					Nonce:         3,
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "token contract address")
			},
			expPass: false,
		},
		{
			name: "query nonce error",
			malleate: func() {
				key, _ := ethsecp256k1.GenerateKey()
				token := crypto.CreateAddress(common.BytesToAddress(key.PubKey().Address()), 0)

				request = &types.QueryBatchConfirmsRequest{
					ChainName:     suite.chainName,
					TokenContract: token.String(),
					Nonce:         0,
				}
				expectedError = sdkerrors.Wrap(types.ErrUnknown, "nonce")
			},
			expPass: false,
		},
		{
			name: "batch confirms normal",
			malleate: func() {
				key, _ := ethsecp256k1.GenerateKey()
				token := crypto.CreateAddress(common.BytesToAddress(key.PubKey().Address()), 0)
				confirms := make([]*types.MsgConfirmBatch, 0)

				for i := 0; i < 3; i++ {
					newMsg := &types.MsgConfirmBatch{
						Nonce:          3,
						TokenContract:  token.String(),
						BridgerAddress: suite.bridgers[i].String(),
						ChainName:      suite.chainName,
					}
					suite.Keeper().SetBatchConfirm(suite.ctx, suite.oracles[i], newMsg)
					confirms = append(confirms, newMsg)
				}

				request = &types.QueryBatchConfirmsRequest{
					ChainName:     suite.chainName,
					TokenContract: token.String(),
					Nonce:         3,
				}
				response = &types.QueryBatchConfirmsResponse{Confirms: confirms}
			},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().BatchConfirms(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().ElementsMatch(response.Confirms, res.Confirms)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_LastEventNonceByAddr() {
	var (
		request       *types.QueryLastEventNonceByAddrRequest
		response      *types.QueryLastEventNonceByAddrResponse
		expectedError error
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "query bridger address ",
			malleate: func() {
				request = &types.QueryLastEventNonceByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: "fx1",
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "bridger address")
			},
			expPass: false,
		},
		{
			name: "query not found oracle by bridger",
			malleate: func() {
				request = &types.QueryLastEventNonceByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
				}
				expectedError = types.ErrNoFoundOracle
			},
			expPass: false,
		},
		{
			name: "query last event nonce from lastObservedEventNonce",
			malleate: func() {
				suite.Keeper().SetOracleByBridger(suite.ctx, suite.oracles[0], suite.bridgers[0])
				suite.Keeper().SetLastObservedEventNonce(suite.ctx, 5)

				request = &types.QueryLastEventNonceByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: suite.bridgers[0].String(),
				}
				response = &types.QueryLastEventNonceByAddrResponse{EventNonce: 4}
			},
			expPass: true,
		},
		{
			name: "query last event nonce not found",
			malleate: func() {
				suite.Keeper().SetOracleByBridger(suite.ctx, suite.oracles[0], suite.bridgers[0])
				request = &types.QueryLastEventNonceByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: suite.bridgers[0].String(),
				}
				response = &types.QueryLastEventNonceByAddrResponse{EventNonce: 0}
			},
			expPass: true,
		},
		{
			name: "query last event nonce normal",
			malleate: func() {
				suite.Keeper().SetOracleByBridger(suite.ctx, suite.oracles[0], suite.bridgers[0])
				suite.Keeper().SetLastEventNonceByOracle(suite.ctx, suite.oracles[0], 3)

				request = &types.QueryLastEventNonceByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: suite.bridgers[0].String(),
				}
				response = &types.QueryLastEventNonceByAddrResponse{EventNonce: 3}
			},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().LastEventNonceByAddr(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_DenomToToken() {
	var (
		request       *types.QueryDenomToTokenRequest
		response      *types.QueryDenomToTokenResponse
		expectedError error
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"denom is nil",
			func() {
				request = &types.QueryDenomToTokenRequest{
					ChainName: suite.chainName,
				}
				expectedError = sdkerrors.Wrap(types.ErrUnknown, "denom")
			},
			false,
		},
		{
			"bridgetoken not exist",
			func() {
				request = &types.QueryDenomToTokenRequest{
					ChainName: suite.chainName,
					Denom:     "bsc0xfbbbb4f7b1e5bcb0345c5a5a61584b2547d5d582",
				}
				expectedError = sdkerrors.Wrap(types.ErrEmpty, "bridge token is not exist")
			},
			false,
		},
		{
			"bridge token and ChannelIbc is exist and true",
			func() {
				key, _ := ethsecp256k1.GenerateKey()
				token := common.BytesToAddress(key.PubKey().Address()).String()

				err := suite.app.BscKeeper.AttestationHandler(suite.ctx, types.Attestation{}, &types.MsgBridgeTokenClaim{
					TokenContract: token,
					ChannelIbc:    hex.EncodeToString([]byte("transfer/channel-0")),
					Symbol:        "fxcoin",
				})
				suite.Require().NoError(err)
				denom := suite.app.BscKeeper.GetBridgeTokenDenom(suite.ctx, token)
				request = &types.QueryDenomToTokenRequest{
					ChainName: suite.chainName,
					Denom:     denom.Denom,
				}
				response = &types.QueryDenomToTokenResponse{
					Token:      token,
					ChannelIbc: "transfer/channel-0",
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().DenomToToken(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_TokenToDenom() {

	var (
		request       *types.QueryTokenToDenomRequest
		response      *types.QueryTokenToDenomResponse
		expectedError error
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"token address is error or null",
			func() {
				request = &types.QueryTokenToDenomRequest{
					ChainName: suite.chainName,
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "token address")
			},
			false,
		},
		{
			"bridge token is not exist",
			func() {
				key, _ := ethsecp256k1.GenerateKey()
				request = &types.QueryTokenToDenomRequest{
					ChainName: suite.chainName,
					Token:     common.BytesToAddress(key.PubKey().Address()).String(),
				}
				expectedError = sdkerrors.Wrap(types.ErrEmpty, "bridge token is not exist")
			},
			false,
		},
		{
			"token normal",
			func() {
				key, _ := ethsecp256k1.GenerateKey()
				token := common.BytesToAddress(key.PubKey().Address()).String()
				err := suite.app.BscKeeper.AttestationHandler(suite.ctx, types.Attestation{}, &types.MsgBridgeTokenClaim{
					TokenContract: token,
					ChannelIbc:    hex.EncodeToString([]byte("transfer/channel-0")),
					Symbol:        "fxcoin",
				})
				suite.Require().NoError(err)
				request = &types.QueryTokenToDenomRequest{
					ChainName: suite.chainName,
					Token:     token,
				}
				denom := suite.app.BscKeeper.GetBridgeTokenDenom(suite.ctx, token)
				response = &types.QueryTokenToDenomResponse{
					Denom:      denom.Denom,
					ChannelIbc: "transfer/channel-0",
				}
				expectedError = sdkerrors.Wrap(types.ErrEmpty, "bridge token is not exist")
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().TokenToDenom(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_GetOracleByAddr() {
	var (
		request       *types.QueryOracleByAddrRequest
		response      *types.QueryOracleResponse
		expectedError error
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "query oracle address error",
			malleate: func() {
				request = &types.QueryOracleByAddrRequest{
					ChainName:     suite.chainName,
					OracleAddress: "fx1",
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "oracle address")
			},
			expPass: false,
		},
		{
			name: "query oracle does not exist",
			malleate: func() {
				request = &types.QueryOracleByAddrRequest{
					ChainName:     suite.chainName,
					OracleAddress: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
				}
				expectedError = types.ErrNoFoundOracle
			},
			expPass: false,
		},
		{
			name: "query oracle normal",
			malleate: func() {
				key, err := ethsecp256k1.GenerateKey()
				suite.Require().NoError(err)
				externalAcc := common.BytesToAddress(key.PubKey().Address().Bytes())
				suite.ctx = suite.ctx.WithBlockHeight(100)
				newOracle := types.Oracle{
					OracleAddress:   suite.oracles[0].String(),
					BridgerAddress:  suite.bridgers[0].String(),
					ExternalAddress: externalAcc.String(),
					DelegateAmount:  sdk.NewIntFromBigInt(big.NewInt(10000)),
					StartHeight:     0,
				}
				suite.Keeper().SetOracle(suite.ctx, newOracle)
				request = &types.QueryOracleByAddrRequest{
					ChainName:     suite.chainName,
					OracleAddress: suite.oracles[0].String(),
				}
				response = &types.QueryOracleResponse{Oracle: &newOracle}
			},
			expPass: true,
		},
	}
	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().GetOracleByAddr(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_GetOracleByBridgerAddr() {
	var (
		request       *types.QueryOracleByBridgerAddrRequest
		response      *types.QueryOracleResponse
		expectedError error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "query bridger address error",
			malleate: func() {
				request = &types.QueryOracleByBridgerAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: "fx1",
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "bridger address")
			},

			expPass: false,
		},
		{
			name: "query oracle by bridger address does not exist",
			malleate: func() {
				request = &types.QueryOracleByBridgerAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
				}
				expectedError = types.ErrNoFoundOracle
			},
			expPass: false,
		},
		{
			name: "query oracle by oracle address does not exist",
			malleate: func() {
				suite.Keeper().SetOracleByBridger(suite.ctx, suite.oracles[0], suite.bridgers[0])
				request = &types.QueryOracleByBridgerAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: suite.bridgers[0].String(),
				}
				expectedError = types.ErrNoFoundOracle
			},
			expPass: false,
		},
		{
			name: "query oracle by oracle address normal",
			malleate: func() {
				suite.Keeper().SetOracleByBridger(suite.ctx, suite.oracles[0], suite.bridgers[0])
				key, err := ethsecp256k1.GenerateKey()
				suite.Require().NoError(err)
				externalAcc := common.BytesToAddress(key.PubKey().Address().Bytes())
				suite.ctx = suite.ctx.WithBlockHeight(100)

				newOracle := types.Oracle{
					OracleAddress:   suite.oracles[0].String(),
					BridgerAddress:  suite.bridgers[0].String(),
					ExternalAddress: externalAcc.String(),
					DelegateAmount:  sdk.NewIntFromBigInt(big.NewInt(10000)),
					StartHeight:     0,
				}

				suite.Keeper().SetOracle(suite.ctx, newOracle)
				request = &types.QueryOracleByBridgerAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: suite.bridgers[0].String(),
				}
				response = &types.QueryOracleResponse{Oracle: &newOracle}
			},
			expPass: true,
		},
	}
	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().GetOracleByBridgerAddr(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_GetOracleByExternalAddr() {
	var (
		request       *types.QueryOracleByExternalAddrRequest
		response      *types.QueryOracleResponse
		expectedError error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "query external address error",
			malleate: func() {
				request = &types.QueryOracleByExternalAddrRequest{
					ChainName:       suite.chainName,
					ExternalAddress: "0x123",
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "external address")
			},
			expPass: false,
		},
		{
			name: "query oracle by external address does not exist",
			malleate: func() {
				key, err := ethsecp256k1.GenerateKey()
				suite.Require().NoError(err)
				externalAcc := common.BytesToAddress(key.PubKey().Address().Bytes())
				request = &types.QueryOracleByExternalAddrRequest{
					ChainName:       suite.chainName,
					ExternalAddress: externalAcc.String(),
				}
				expectedError = types.ErrNoFoundOracle
			},
			expPass: false,
		},
		{
			name: "query oracle does not exist",
			malleate: func() {
				key, err := ethsecp256k1.GenerateKey()
				suite.Require().NoError(err)
				externalAcc := common.BytesToAddress(key.PubKey().Address().Bytes())
				suite.Keeper().SetExternalAddressForOracle(suite.ctx, suite.oracles[0], externalAcc.String())
				request = &types.QueryOracleByExternalAddrRequest{
					ChainName:       suite.chainName,
					ExternalAddress: externalAcc.String(),
				}
				expectedError = types.ErrNoFoundOracle
			},
			expPass: false,
		},
		{
			name: "query oracle normal",
			malleate: func() {
				key, err := ethsecp256k1.GenerateKey()
				suite.Require().NoError(err)
				externalAcc := common.BytesToAddress(key.PubKey().Address().Bytes())
				suite.Keeper().SetExternalAddressForOracle(suite.ctx, suite.oracles[0], externalAcc.String())
				newOracle := types.Oracle{
					OracleAddress:   suite.oracles[0].String(),
					BridgerAddress:  suite.bridgers[0].String(),
					ExternalAddress: externalAcc.String(),
					DelegateAmount:  sdk.NewIntFromBigInt(big.NewInt(10000)),
					StartHeight:     0,
				}

				suite.Keeper().SetOracle(suite.ctx, newOracle)
				suite.ctx = suite.ctx.WithBlockHeight(100)
				request = &types.QueryOracleByExternalAddrRequest{
					ChainName:       suite.chainName,
					ExternalAddress: externalAcc.String(),
				}
				response = &types.QueryOracleResponse{Oracle: &newOracle}
			},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().GetOracleByExternalAddr(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_GetPendingSendToExternal() {
	var (
		request       *types.QueryPendingSendToExternalRequest
		response      *types.QueryPendingSendToExternalResponse
		expectedError error
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"sender address error",
			func() {
				request = &types.QueryPendingSendToExternalRequest{
					ChainName:     suite.chainName,
					SenderAddress: "fx1",
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "sender address")
			},
			false,
		},
		{
			"sender outgoing transfer tx not exist",
			func() {
				externalKey, _ := ethsecp256k1.GenerateKey()
				request = &types.QueryPendingSendToExternalRequest{
					ChainName:     suite.chainName,
					SenderAddress: common.BytesToAddress(externalKey.PubKey().Address()).String(),
				}

			},
			false,
		},
		{
			name: "sender pending send to external in batches",
			malleate: func() {
				externalKey, _ := ethsecp256k1.GenerateKey()
				externalAcc := common.BytesToAddress(externalKey.PubKey().Address())
				externalToken := crypto.CreateAddress(common.BytesToAddress(externalKey.PubKey().Address()), 0)
				err := suite.Keeper().StoreBatch(suite.ctx, &types.OutgoingTxBatch{
					Transactions: []*types.OutgoingTransferTx{
						{
							Id:          0,
							Sender:      sdk.AccAddress(externalKey.PubKey().Address()).String(),
							DestAddress: externalAcc.String(),
							Token:       types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), externalToken.String()),
							Fee:         types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), externalToken.String()),
						},
					},
				})
				suite.Require().NoError(err)
				request = &types.QueryPendingSendToExternalRequest{
					ChainName:     suite.chainName,
					SenderAddress: sdk.AccAddress(externalKey.PubKey().Address()).String(),
				}
				response = &types.QueryPendingSendToExternalResponse{
					TransfersInBatches: []*types.OutgoingTransferTx{
						{
							Id:          0,
							Sender:      sdk.AccAddress(externalKey.PubKey().Address()).String(),
							DestAddress: externalAcc.String(),
							Token:       types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), externalToken.String()),
							Fee:         types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), externalToken.String()),
						},
					},
					UnbatchedTransfers: []*types.OutgoingTransferTx{},
				}
			},
			expPass: true,
		},
		{
			name: "pending send to external in batch and unbatched",
			malleate: func() {
				externalKey, _ := ethsecp256k1.GenerateKey()
				externalAcc := common.BytesToAddress(externalKey.PubKey().Address())
				token := crypto.CreateAddress(common.BytesToAddress(externalKey.PubKey().Address()), 0)
				bridgeAcc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
				err := suite.app.BscKeeper.AttestationHandler(suite.ctx, types.Attestation{}, &types.MsgBridgeTokenClaim{
					TokenContract:  token.String(),
					BridgerAddress: suite.bridgers[0].String(),
					ChannelIbc:     hex.EncodeToString([]byte("transfer/channel-0")),
					ChainName:      suite.chainName,
				})
				suite.Require().NoError(err)
				denom := suite.app.BscKeeper.GetBridgeTokenDenom(suite.ctx, token.String())
				initBalances := sdk.NewIntFromUint64(1e18).Mul(sdk.NewInt(20000))
				err = suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(denom.Denom, initBalances)))
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, bridgeAcc, sdk.NewCoins(sdk.NewCoin(denom.Denom, initBalances)))
				suite.Require().NoError(err)
				pool, err := suite.app.BscKeeper.AddToOutgoingPool(
					suite.ctx,
					bridgeAcc,
					externalAcc.String(),
					sdk.NewCoin(denom.Denom, sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e18), big.NewInt(100)))),
					sdk.NewCoin(denom.Denom, sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e18), big.NewInt(100)))),
				)
				suite.Require().NoError(err)
				suite.Require().Equal(pool, uint64(1))
				bridgeToken := suite.app.BscKeeper.GetDenomByBridgeToken(suite.ctx, denom.Denom)
				suite.Require().Equal(bridgeToken.Denom, denom.Denom)
				suite.Require().Equal(bridgeToken.Token, denom.Token)
				bridgeTokenFee := types.NewERC20Token(sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e18), big.NewInt(100))), bridgeToken.Token)

				err = suite.Keeper().StoreBatch(suite.ctx, &types.OutgoingTxBatch{
					Transactions: []*types.OutgoingTransferTx{
						{
							Id:          0,
							Sender:      bridgeAcc.String(),
							DestAddress: externalAcc.String(),
							Token:       types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), token.String()),
							Fee:         types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), token.String()),
						},
					},
				})
				suite.Require().NoError(err)
				request = &types.QueryPendingSendToExternalRequest{
					ChainName:     suite.chainName,
					SenderAddress: bridgeAcc.String(),
				}
				response = &types.QueryPendingSendToExternalResponse{
					TransfersInBatches: []*types.OutgoingTransferTx{
						{
							Id:          0,
							Sender:      bridgeAcc.String(),
							DestAddress: externalAcc.String(),
							Token:       types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), token.String()),
							Fee:         types.NewERC20Token(sdk.NewIntFromBigInt(big.NewInt(1e18)), token.String()),
						},
					},
					UnbatchedTransfers: []*types.OutgoingTransferTx{
						{
							Id:          1,
							Sender:      bridgeAcc.String(),
							DestAddress: externalAcc.String(),
							Token:       types.NewERC20Token(sdk.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(1e18), big.NewInt(100))), denom.Token),
							Fee:         bridgeTokenFee,
						},
					},
				}
			},
			expPass: true,
		},
	}
	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().GetPendingSendToExternal(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().ElementsMatch(response.TransfersInBatches, res.TransfersInBatches)
				suite.Require().ElementsMatch(response.UnbatchedTransfers, res.UnbatchedTransfers)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_LastEventBlockHeightByAddr() {
	var (
		request       *types.QueryLastEventBlockHeightByAddrRequest
		response      *types.QueryLastEventBlockHeightByAddrResponse
		expectedError error
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"BridgerAddress is error",
			func() {
				request = &types.QueryLastEventBlockHeightByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: "fx1",
				}
				expectedError = sdkerrors.Wrap(types.ErrInvalid, "bridger address")
			},
			false,
		},
		{
			"BridgerAddress exist oracle is nil",
			func() {
				request = &types.QueryLastEventBlockHeightByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: suite.bridgers[0].String(),
				}
				expectedError = types.ErrNoFoundOracle
			},
			false,
		},
		{
			"BridgerAddress exist oracle is not nil",
			func() {
				request = &types.QueryLastEventBlockHeightByAddrRequest{
					ChainName:      suite.chainName,
					BridgerAddress: suite.bridgers[0].String(),
				}
				suite.ctx = suite.ctx.WithBlockHeight(100)
				suite.Keeper().SetOracleByBridger(suite.ctx, suite.oracles[0], suite.bridgers[0])

				suite.Keeper().SetOracle(suite.ctx, types.Oracle{
					OracleAddress:  suite.oracles[0].String(),
					BridgerAddress: suite.bridgers[0].String(),
					StartHeight:    100,
					Online:         true,
				})
				_, err := suite.msgServer.BridgeTokenClaim(sdk.WrapSDKContext(suite.ctx), &types.MsgBridgeTokenClaim{
					EventNonce:     1,
					BlockHeight:    100,
					TokenContract:  crypto.PubkeyToAddress(genEthKey(1)[0].PublicKey).String(),
					Name:           "test token",
					Symbol:         "tt",
					Decimals:       18,
					BridgerAddress: suite.bridgers[0].String(),
					ChainName:      suite.chainName,
				})
				suite.Require().NoError(err)
				response = &types.QueryLastEventBlockHeightByAddrResponse{
					BlockHeight: uint64(100),
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().LastEventBlockHeightByAddr(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_LastObservedBlockHeight() {
	var (
		request       *types.QueryLastObservedBlockHeightRequest
		response      *types.QueryLastObservedBlockHeightResponse
		expectedError error
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"ExternalBlockHeight not exist",
			func() {
				request = &types.QueryLastObservedBlockHeightRequest{
					ChainName: suite.chainName,
				}
				response = &types.QueryLastObservedBlockHeightResponse{
					ExternalBlockHeight: 0,
					BlockHeight:         0,
				}
			},
			true,
		},
		{
			"ExternalBlockHeight exist",
			func() {
				suite.ctx = suite.ctx.WithBlockHeight(100)
				suite.Keeper().SetLastObservedBlockHeight(suite.ctx, uint64(30))

				request = &types.QueryLastObservedBlockHeightRequest{
					ChainName: suite.chainName,
				}

				response = &types.QueryLastObservedBlockHeightResponse{
					ExternalBlockHeight: uint64(30),
					BlockHeight:         uint64(100),
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().LastObservedBlockHeight(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_Oracles() {
	var (
		request       *types.QueryOraclesRequest
		response      *types.QueryOraclesResponse
		expectedError error
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"Oracles exist and online is false",
			func() {
				externalKey, err := ethsecp256k1.GenerateKey()
				suite.Require().NoError(err)
				externalAcc := common.BytesToAddress(externalKey.PubKey().Address())
				suite.Keeper().SetOracle(suite.ctx, types.Oracle{
					OracleAddress:   suite.oracles[0].String(),
					BridgerAddress:  suite.bridgers[0].String(),
					ExternalAddress: externalAcc.String(),
					DelegateAmount:  sdk.ZeroInt(),
					StartHeight:     10,
					Online:          false,
				})
				request = &types.QueryOraclesRequest{
					ChainName: suite.chainName,
				}
				response = &types.QueryOraclesResponse{
					Oracles: []types.Oracle{{
						OracleAddress:   suite.oracles[0].String(),
						BridgerAddress:  suite.bridgers[0].String(),
						ExternalAddress: externalAcc.String(),
						DelegateAmount:  sdk.ZeroInt(),
						StartHeight:     10,
						Online:          false,
					},
					},
				}
			},
			true,
		},
		{
			"Oracles  exist and online is true",
			func() {
				externalKey, err := ethsecp256k1.GenerateKey()
				suite.Require().NoError(err)
				externalAcc := common.BytesToAddress(externalKey.PubKey().Address().Bytes())
				for i := 1; i < 4; i++ {
					online := true
					if i == 2 {
						online = false
					}
					suite.Keeper().SetOracle(suite.ctx, types.Oracle{
						OracleAddress:   suite.oracles[i].String(),
						BridgerAddress:  suite.bridgers[i].String(),
						ExternalAddress: externalAcc.String(),
						DelegateAmount:  sdk.ZeroInt(),
						StartHeight:     int64(i),
						Online:          online,
					})
				}
				request = &types.QueryOraclesRequest{
					ChainName: suite.chainName,
				}
				response = &types.QueryOraclesResponse{
					Oracles: []types.Oracle{
						{
							OracleAddress:   suite.oracles[1].String(),
							BridgerAddress:  suite.bridgers[1].String(),
							ExternalAddress: externalAcc.String(),
							DelegateAmount:  sdk.ZeroInt(),
							StartHeight:     int64(1),
							Online:          true,
						},
						{
							OracleAddress:   suite.oracles[2].String(),
							BridgerAddress:  suite.bridgers[2].String(),
							ExternalAddress: externalAcc.String(),
							DelegateAmount:  sdk.ZeroInt(),
							StartHeight:     int64(2),
							Online:          false,
						},
						{
							OracleAddress:   suite.oracles[3].String(),
							BridgerAddress:  suite.bridgers[3].String(),
							ExternalAddress: externalAcc.String(),
							DelegateAmount:  sdk.ZeroInt(),
							StartHeight:     int64(3),
							Online:          true,
						},
					},
				}

			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().Oracles(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().ElementsMatch(response.Oracles, res.Oracles)

			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_ProjectedBatchTimeoutHeight() {
	var (
		request       *types.QueryProjectedBatchTimeoutHeightRequest
		response      *types.QueryProjectedBatchTimeoutHeightResponse
		expectedError error
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"ExternalBlockHeight is 0",
			func() {
				request = &types.QueryProjectedBatchTimeoutHeightRequest{
					ChainName: suite.chainName,
				}
				suite.Require().Equal(uint64(0), suite.Keeper().GetLastObservedBlockHeight(suite.ctx).ExternalBlockHeight)
				suite.Require().Equal(uint64(0), suite.Keeper().GetLastObservedBlockHeight(suite.ctx).BlockHeight)
				response = &types.QueryProjectedBatchTimeoutHeightResponse{
					TimeoutHeight: 0,
				}
			},
			true,
		},
		{
			name: "ProjectedBatchTimeoutHeight exist",
			malleate: func() {
				suite.ctx = suite.ctx.WithBlockHeight(100)
				suite.Keeper().SetLastObservedBlockHeight(suite.ctx, 99)
				heights := suite.Keeper().GetLastObservedBlockHeight(suite.ctx)
				suite.Assert().Equal(uint64(99), heights.ExternalBlockHeight)
				suite.Assert().Equal(uint64(100), heights.BlockHeight)
				params := suite.Keeper().GetParams(suite.ctx)
				suite.ctx = suite.ctx.WithBlockHeight(1000)
				projectedMillis := (1000 - heights.BlockHeight) * params.AverageBlockTime
				projectedCurrentEthereumHeight := (projectedMillis / params.AverageExternalBlockTime) + heights.ExternalBlockHeight
				blocksToAdd := params.ExternalBatchTimeout / params.AverageExternalBlockTime
				request = &types.QueryProjectedBatchTimeoutHeightRequest{
					ChainName: suite.chainName,
				}
				response = &types.QueryProjectedBatchTimeoutHeightResponse{
					TimeoutHeight: projectedCurrentEthereumHeight + blocksToAdd,
				}
			},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			testCase.malleate()
			res, err := suite.Keeper().ProjectedBatchTimeoutHeight(sdk.WrapSDKContext(suite.ctx), request)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, expectedError)
			}
		})
	}
}

func (suite *CrossChainGrpcTestSuite) TestKeeper_BridgeTokens() {
	testCases := []struct {
		name          string
		malleate      func() *types.QueryBridgeTokensResponse
		expectedError error
		expPass       bool
	}{
		{
			name: "query bridge tokens",
			malleate: func() *types.QueryBridgeTokensResponse {

				newBridgeTokens := make([]*types.BridgeToken, 3)

				for i := 0; i < 3; i++ {
					key, _ := ethsecp256k1.GenerateKey()
					channelIbc := ""
					if i == 2 {
						channelIbc = "transfer/channel-0"
					}
					err := suite.app.BscKeeper.AttestationHandler(suite.ctx, types.Attestation{}, &types.MsgBridgeTokenClaim{
						TokenContract:  common.BytesToAddress(key.PubKey().Address()).String(),
						BridgerAddress: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
						ChannelIbc:     hex.EncodeToString([]byte(channelIbc)),
					})

					suite.Require().NoError(err)
					bridgeTokenFromToken := suite.Keeper().GetBridgeTokenDenom(suite.ctx, common.BytesToAddress(key.PubKey().Address()).String())
					suite.Require().Equal(bridgeTokenFromToken.Token, common.BytesToAddress(key.PubKey().Address()).String())

					bridgeTokenFromDenom := suite.Keeper().GetDenomByBridgeToken(suite.ctx, bridgeTokenFromToken.Denom)
					suite.Require().Equal(bridgeTokenFromDenom.Token, common.BytesToAddress(key.PubKey().Address()).String())
					suite.Require().Equal(bridgeTokenFromDenom.Denom, bridgeTokenFromToken.Denom)

					newBridgeTokens[i] = &types.BridgeToken{
						Token:      common.BytesToAddress(key.PubKey().Address()).String(),
						Denom:      bridgeTokenFromToken.Denom,
						ChannelIbc: channelIbc,
					}
				}
				return &types.QueryBridgeTokensResponse{BridgeTokens: newBridgeTokens}
			},
			expPass: true,
		},
	}
	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.SetupTest()
			response := testCase.malleate()
			res, err := suite.Keeper().BridgeTokens(sdk.WrapSDKContext(suite.ctx), nil)
			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().ElementsMatch(response.BridgeTokens, res.BridgeTokens)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, testCase.expectedError)
			}
		})
	}
}
