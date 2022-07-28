package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"sync"
	"testing"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/functionx/fx-core/v2/testutil"
	"github.com/functionx/fx-core/v2/testutil/network"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	hd2 "github.com/evmos/ethermint/crypto/hd"
	"github.com/gogo/protobuf/proto"

	"github.com/functionx/fx-core/v2/client/jsonrpc"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"

	"github.com/functionx/fx-core/v2/app/helpers"
	"github.com/functionx/fx-core/v2/client/grpc"
	fxtypes "github.com/functionx/fx-core/v2/types"
)

type TestClient interface {
	AppVersion() (string, error)
	GetChainId() (chain string, err error)
	GetBlockHeight() (int64, error)
	GetMintDenom() (denom string, err error)
	GetGasPrices() (sdk.Coins, error)
	GetAddressPrefix() (prefix string, err error)
	QueryAccount(address string) (authtypes.AccountI, error)
	QueryBalance(address string, denom string) (sdk.Coin, error)
	QueryBalances(address string) (sdk.Coins, error)
	QuerySupply() (sdk.Coins, error)
	BuildTx(privKey cryptotypes.PrivKey, msgs []sdk.Msg) (*tx.TxRaw, error)
	EstimatingGas(raw *tx.TxRaw) (*sdk.GasInfo, error)
	BroadcastTx(txRaw *tx.TxRaw, mode ...tx.BroadcastMode) (*sdk.TxResponse, error)
	TxByHash(txHash string) (*sdk.TxResponse, error)
}

type IntegrationTestSuite struct {
	suite.Suite

	network *network.Network
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.T().Log("setting up integration test suite")

	cfg := testutil.DefaultNetworkConfig()
	cfg.NumValidators = 1
	cfg.Mnemonics = append(cfg.Mnemonics, helpers.NewMnemonic())

	baseDir, err := ioutil.TempDir(suite.T().TempDir(), cfg.ChainID)
	suite.Require().NoError(err)
	suite.network, err = network.New(suite.T(), baseDir, cfg)
	suite.Require().NoError(err)

	_, err = suite.network.WaitForHeight(1)
	suite.Require().NoError(err)

	suite.FirstValidatorTransferTo(1, sdk.NewInt(1_000).MulRaw(1e18))
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	suite.T().Log("tearing down integration test suite")

	// This is important and must be called to ensure other tests can create
	// a network!
	suite.network.Cleanup()
}

func (suite *IntegrationTestSuite) GetFirstValidator() *network.Validator {
	return suite.network.Validators[0]
}

func (suite *IntegrationTestSuite) GetFirstValidatorPrivKey() cryptotypes.PrivKey {
	return suite.GetPrivKeyByIndex(hd.Secp256k1Type, 0)
}

func (suite *IntegrationTestSuite) GetPrivKeyByIndex(algo hd.PubKeyType, index uint32) cryptotypes.PrivKey {
	privKey, err := helpers.PrivKeyFromMnemonic(suite.network.Config.Mnemonics[0], algo, 0, index)
	suite.NoError(err)
	return privKey
}

func (suite *IntegrationTestSuite) GetClients() []TestClient {
	validator := suite.GetFirstValidator()
	suite.True(validator.AppConfig.GRPC.Enable)
	grpcClient, err := grpc.NewClient(fmt.Sprintf("http://%s", validator.AppConfig.GRPC.Address))
	suite.NoError(err)
	return []TestClient{
		grpcClient,
		jsonrpc.NewNodeRPC(jsonrpc.NewFastClient(validator.RPCAddress)),
	}
}

func (suite *IntegrationTestSuite) FirstValidatorTransferTo(index uint32, amount sdk.Int) {
	validator := suite.GetFirstValidator()
	suite.True(validator.AppConfig.GRPC.Enable)
	grpcClient, err := grpc.NewClient(fmt.Sprintf("http://%s", validator.AppConfig.GRPC.Address))
	suite.NoError(err)
	valKey := suite.GetFirstValidatorPrivKey()
	nextValKey := suite.GetPrivKeyByIndex(hd.Secp256k1Type, index)
	txRaw, err := grpcClient.BuildTxV2(valKey,
		[]sdk.Msg{
			banktypes.NewMsgSend(
				valKey.PubKey().Address().Bytes(),
				nextValKey.PubKey().Address().Bytes(),
				sdk.NewCoins(sdk.NewCoin(fxtypes.DefaultDenom, amount)),
			),
		},
		500000,
		"",
		0,
	)
	suite.NoError(err)
	txResponse, err := grpcClient.BroadcastTx(txRaw)
	suite.NoError(err)
	suite.Equal(uint32(0), txResponse.Code)
}

func (suite *IntegrationTestSuite) TestClient_Tx() {
	privKey := suite.GetPrivKeyByIndex(hd.Secp256k1Type, 1)

	clients := suite.GetClients()
	for i := 0; i < len(clients); i++ {
		cli := clients[i]
		toAddress := sdk.AccAddress(helpers.NewPriKey().PubKey().Address())
		txRaw, err := cli.BuildTx(privKey, []sdk.Msg{
			banktypes.NewMsgSend(
				privKey.PubKey().Address().Bytes(),
				toAddress,
				sdk.NewCoins(sdk.NewCoin(fxtypes.DefaultDenom, sdk.NewInt(1))),
			)},
		)
		suite.NoError(err)

		gas, err := cli.EstimatingGas(txRaw)
		suite.NoError(err)
		suite.True(gas.GasUsed < 90000)
		suite.Equal(uint64(0), gas.GasWanted)

		txResponse, err := cli.BroadcastTx(txRaw)
		suite.NoError(err)
		suite.Equal(uint32(0), txResponse.Code)

		err = suite.network.WaitForNextBlock()
		suite.NoError(err)

		txRes, err := cli.TxByHash(txResponse.TxHash)
		suite.NoError(err)
		txRes.Tx = nil
		txRes.Timestamp = ""
		suite.Equal(txResponse, txRes)

		account, err := cli.QueryAccount(toAddress.String())
		suite.NoError(err)
		suite.Equal(authtypes.NewBaseAccount(toAddress, nil, uint64(12+i), 0), account)
	}

	ethPrivKey := suite.GetPrivKeyByIndex(hd2.EthSecp256k1Type, 0)

	ethAddress := sdk.AccAddress(ethPrivKey.PubKey().Address().Bytes())

	for i := 0; i < len(clients); i++ {
		cli := clients[i]
		txRaw, err := cli.BuildTx(privKey, []sdk.Msg{
			banktypes.NewMsgSend(
				privKey.PubKey().Address().Bytes(),
				ethAddress,
				sdk.NewCoins(sdk.NewCoin(fxtypes.DefaultDenom, sdk.NewInt(10).MulRaw(1e18))),
			)},
		)
		suite.NoError(err)

		gas, err := cli.EstimatingGas(txRaw)
		suite.NoError(err)
		suite.True(gas.GasUsed < 90000)
		suite.Equal(uint64(0), gas.GasWanted)

		txResponse, err := cli.BroadcastTx(txRaw)
		suite.NoError(err)
		suite.Equal(uint32(0), txResponse.Code)

		err = suite.network.WaitForNextBlock()
		suite.NoError(err)

		account, err := cli.QueryAccount(ethAddress.String())
		suite.NoError(err)
		suite.Equal(authtypes.NewBaseAccount(ethAddress, nil, 14, 0), account)
	}

	for i := 0; i < len(clients); i++ {
		cli := clients[i]
		toAddress := sdk.AccAddress(helpers.NewPriKey().PubKey().Address())
		txRaw, err := cli.BuildTx(ethPrivKey, []sdk.Msg{
			banktypes.NewMsgSend(
				ethPrivKey.PubKey().Address().Bytes(),
				toAddress,
				sdk.NewCoins(sdk.NewCoin(fxtypes.DefaultDenom, sdk.NewInt(1))),
			)},
		)
		suite.NoError(err)

		gas, err := cli.EstimatingGas(txRaw)
		suite.NoError(err)
		suite.True(gas.GasUsed < 90000)
		suite.Equal(uint64(0), gas.GasWanted)

		txResponse, err := cli.BroadcastTx(txRaw)
		suite.NoError(err)
		suite.Equal(uint32(0), txResponse.Code)

		err = suite.network.WaitForNextBlock()
		suite.NoError(err)

		account, err := cli.QueryAccount(ethAddress.String())
		suite.NoError(err)
		baseAccount, ok := account.(*authtypes.BaseAccount)
		suite.True(ok)
		if baseAccount.PubKey.TypeUrl != "" {
			pubAny, err := types.NewAnyWithValue(ethPrivKey.PubKey())
			suite.NoError(err)
			suite.Equal("/"+proto.MessageName(&ethsecp256k1.PubKey{}), baseAccount.PubKey.TypeUrl)
			suite.Equal(pubAny, baseAccount.PubKey)
		}
		suite.Equal(uint64(i+1), account.GetSequence())
	}
}

func (suite *IntegrationTestSuite) TestQueryBlockHeight() {
	clients := suite.GetClients()
	for i := 0; i < len(clients); i++ {
		height, err := clients[i].GetBlockHeight()
		suite.NoError(err)
		suite.True(height >= int64(10))
	}
}

func (suite *IntegrationTestSuite) TestQuerySupply() {
	clients := suite.GetClients()
	for i := 0; i < len(clients); i++ {
		supply, err := clients[i].QuerySupply()
		suite.NoError(err)
		nodeCoin := sdk.Coin{
			Denom:  "node0token",
			Amount: sdk.NewInt(100_000).MulRaw(1e18),
		}
		suite.Equal(supply.AmountOf(nodeCoin.Denom), nodeCoin.Amount)
		suite.True(supply.AmountOf(fxtypes.DefaultDenom).GTE(sdk.NewInt(50_000).MulRaw(1e18)))
	}
}

func (suite *IntegrationTestSuite) TestClient_Query() {
	tests := []struct {
		funcName string
		params   []interface{}
		wantRes  []interface{}
	}{
		{
			funcName: "GetChainId",
			params:   []interface{}{},
			wantRes:  []interface{}{fxtypes.MainnetChainId, nil},
		},
		{
			funcName: "GetMintDenom",
			params:   []interface{}{},
			wantRes:  []interface{}{fxtypes.DefaultDenom, nil},
		},
		{
			funcName: "GetAddressPrefix",
			params:   []interface{}{},
			wantRes:  []interface{}{fxtypes.AddressPrefix, nil},
		},
		{
			funcName: "AppVersion",
			params:   []interface{}{},
			wantRes:  []interface{}{"", nil},
		},
		{
			funcName: "GetGasPrices",
			params:   []interface{}{},
			wantRes: []interface{}{
				sdk.Coins{
					sdk.Coin{
						Denom:  fxtypes.DefaultDenom,
						Amount: sdk.NewInt(4).MulRaw(1e12),
					},
				},
				nil,
			},
		},
		{
			funcName: "QueryAccount",
			params:   []interface{}{suite.GetFirstValidator().Address.String()},
			wantRes: []interface{}{authtypes.NewBaseAccount(
				suite.GetFirstValidator().Address,
				suite.GetFirstValidator().PubKey,
				0,
				0,
			),
				nil},
		},
		{
			funcName: "QueryBalance",
			params:   []interface{}{suite.GetFirstValidator().Address.String(), fxtypes.DefaultDenom},
			wantRes: []interface{}{
				sdk.Coin{
					Denom:  fxtypes.DefaultDenom,
					Amount: sdk.NewInt(38998).MulRaw(1e18),
				},
				nil,
			},
		},
		{
			funcName: "QueryBalances",
			params:   []interface{}{suite.GetFirstValidator().Address.String()},
			wantRes: []interface{}{
				sdk.Coins{
					sdk.Coin{
						Denom:  fxtypes.DefaultDenom,
						Amount: sdk.NewInt(38998).MulRaw(1e18),
					},
					sdk.Coin{
						Denom:  "node0token",
						Amount: sdk.NewInt(100_000).MulRaw(1e18),
					},
				},
				nil,
			},
		},
	}
	clients := suite.GetClients()
	for _, tt := range tests {
		suite.Run(tt.funcName, func() {
			for i := 0; i < len(clients); i++ {
				typeOf := reflect.TypeOf(clients[i])
				method, is := typeOf.MethodByName(tt.funcName)
				suite.True(is)
				params := make([]reflect.Value, len(tt.params)+1)
				params[0] = reflect.ValueOf(clients[i])
				for i := 1; i < len(params); i++ {
					params[i] = reflect.ValueOf(tt.params[i-1])
				}
				results := method.Func.Call(params)
				for i := 0; i < len(results); i++ {
					suite.EqualValues(
						fmt.Sprintf("%v", tt.wantRes[i]),
						fmt.Sprintf("%v", results[i]),
					)
				}
			}
		})
	}
}

func (suite *IntegrationTestSuite) TestTmClient() {
	validator := suite.GetFirstValidator()
	tmRPC := validator.RPCClient
	callTmRPC := func(funcName string, params []interface{}) []reflect.Value {
		typeOf := reflect.TypeOf(tmRPC)
		method, is := typeOf.MethodByName(funcName)
		suite.True(is)
		callParams := make([]reflect.Value, len(params))
		for i, param := range params {
			callParams[i] = reflect.ValueOf(param)
		}
		callParams = append([]reflect.Value{reflect.ValueOf(tmRPC), reflect.ValueOf(context.Background())}, callParams...)
		results := method.Func.Call(callParams)
		return results
	}

	nodeRPC := jsonrpc.NewNodeRPC(jsonrpc.NewFastClient(validator.RPCAddress))
	callNodeRPC := func(funcName string, params []interface{}) []reflect.Value {
		typeOf := reflect.TypeOf(nodeRPC)
		method, is := typeOf.MethodByName(funcName)
		suite.True(is)
		callParams := make([]reflect.Value, len(params))
		for i, param := range params {
			callParams[i] = reflect.Indirect(reflect.ValueOf(param))
		}
		callParams = append([]reflect.Value{reflect.ValueOf(nodeRPC)}, callParams...)
		results := method.Func.Call(callParams)
		return results
	}

	var height = int64(1)
	var limit = 1
	tests := []struct {
		funcName string
		params   []interface{}
		wantRes  []interface{}
	}{
		//ABCIClient
		{
			funcName: "ABCIInfo",
			params:   []interface{}{},
		},
		//HistoryClient
		{
			funcName: "Genesis",
			params:   []interface{}{},
		},
		{
			funcName: "BlockchainInfo",
			params:   []interface{}{int64(1), int64(1)},
		},
		//StatusClient
		//{
		//	funcName:   "Status",
		//	params: []interface{}{},
		//},
		//NetworkClient
		{
			funcName: "NetInfo",
			params:   []interface{}{},
		},
		{
			funcName: "DumpConsensusState",
			params:   []interface{}{},
		},
		{
			funcName: "ConsensusState",
			params:   []interface{}{},
		},
		{
			funcName: "ConsensusParams",
			params:   []interface{}{&height},
		},
		{
			funcName: "Health",
			params:   []interface{}{},
		},
		//MempoolClient
		{
			funcName: "UnconfirmedTxs",
			params:   []interface{}{&limit},
		},
		{
			funcName: "NumUnconfirmedTxs",
			params:   []interface{}{},
		},
	}
	for _, tt := range tests {
		suite.Run(tt.funcName, func() {
			wg := sync.WaitGroup{}
			wg.Add(1)
			resultChan := make(chan []reflect.Value, 2)
			go func() {
				defer wg.Done()
				resultChan <- callTmRPC(tt.funcName, tt.params)
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				resultChan <- callNodeRPC(tt.funcName, tt.params)
			}()
			wg.Wait()

			result1 := <-resultChan
			result2 := <-resultChan
			suite.Equal(len(result1), len(result2))
			for i := 0; i < len(result1); i++ {
				if i != 0 && result1[i].IsNil() && result2[i].IsNil() {
					continue
				}
				if result1[i].IsNil() || result2[i].IsNil() {
					suite.T().Log("warn", result1[i], result2[i])
					continue
				}
				data1, err1 := json.Marshal(reflect.Indirect(result1[i]).Interface())
				suite.NoError(err1)
				data2, err2 := json.Marshal(reflect.Indirect(result2[i]).Interface())
				suite.NoError(err2)
				suite.JSONEq(string(data1), string(data2))
			}
			close(resultChan)
		})
	}
}

func (suite *IntegrationTestSuite) TestJsonRPC_GetStakeValidators() {
	validator := suite.GetFirstValidator()
	nodeRPC := jsonrpc.NewNodeRPC(jsonrpc.NewFastClient(validator.RPCAddress))
	validators, err := nodeRPC.GetStakeValidators(stakingtypes.Bonded)
	suite.Require().NoError(err)
	suite.Require().Len(validators, 1)
}
