package keeper_test

import (
	"fmt"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/functionx/fx-core/tests"
	"github.com/functionx/fx-core/x/intrarelayer/types"
)

const (
	fip20Name       = "coin"
	fip20Symbol     = "token"
	cosmosTokenName = "coin"
	displayCoinName = "COIN"
	defaultExponent = uint32(18)
	zeroExponent    = uint32(0)
)

func (suite *KeeperTestSuite) setupRegisterFIP20Pair() common.Address {
	suite.SetupTest()
	contractAddr := suite.DeployContract(fip20Name, fip20Symbol, 18)
	suite.Commit()
	_, err := suite.app.IntrarelayerKeeper.RegisterFIP20(suite.ctx, contractAddr)
	suite.Require().NoError(err)
	return contractAddr
}

func (suite *KeeperTestSuite) setupRegisterCoin() (banktypes.Metadata, *types.TokenPair) {
	suite.SetupTest()
	validMetadata := banktypes.Metadata{
		Description: "desc",
		Base:        cosmosTokenName,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    cosmosTokenName,
				Exponent: 0,
			},
			{
				Denom:    displayCoinName,
				Exponent: uint32(18),
			},
		},
		Display: displayCoinName,
	}
	// pair := types.NewTokenPair(contractAddr, cosmosTokenName, true, types.OWNER_MODULE)
	pair, err := suite.app.IntrarelayerKeeper.RegisterCoin(suite.ctx, validMetadata)
	suite.Require().NoError(err)
	suite.Commit()
	return validMetadata, pair
}

func (suite KeeperTestSuite) TestRegisterCoin() {
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"intrarelaying is disabled globally",
			func() {
				params := types.DefaultParams()
				params.EnableIntrarelayer = false
				suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"denom already registered",
			func() {
				regPair := types.NewTokenPair(tests.GenerateAddress(), cosmosTokenName, true, types.OWNER_MODULE)
				suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, regPair.Denom, regPair.GetID())
				suite.Commit()
			},
			false,
		},
		{
			"metadata different that stored",
			func() {
				validMetadata := banktypes.Metadata{
					Description: "desc",
					Base:        cosmosTokenName,
					// NOTE: Denom units MUST be increasing
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    cosmosTokenName,
							Exponent: 0,
						},
						{
							Denom:    displayCoinName,
							Exponent: uint32(1),
						},
						{
							Denom:    "extraDenom",
							Exponent: uint32(2),
						},
					},
					//Name:    "otherName",
					//Symbol:  "token",
					Display: displayCoinName,
				}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, validMetadata)
			},
			false,
		},
		{
			"ok",
			func() {},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()
			validMetadata := banktypes.Metadata{
				Description: "desc",
				Base:        cosmosTokenName,
				// NOTE: Denom units MUST be increasing
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    cosmosTokenName,
						Exponent: 0,
					},
					{
						Denom:    displayCoinName,
						Exponent: uint32(1),
					},
				},
				Display: displayCoinName,
			}
			pair, err := suite.app.IntrarelayerKeeper.RegisterCoin(suite.ctx, validMetadata)
			suite.Commit()
			expPair := &types.TokenPair{
				Fip20Address:  "0x00819E780C6e96c50Ed70eFFf5B73569c15d0bd7",
				Denom:         "coin",
				Enabled:       true,
				ContractOwner: 1,
			}
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(pair, expPair)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite KeeperTestSuite) TestRegisterFIP20() {
	var (
		contractAddr common.Address
		pair         types.TokenPair
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"intrarelaying is disabled globally",
			func() {
				params := types.DefaultParams()
				params.EnableIntrarelayer = false
				suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"token FIP20 already registered",
			func() {
				suite.app.IntrarelayerKeeper.SetFIP20Map(suite.ctx, pair.GetFIP20Contract(), pair.GetID())
			},
			false,
		},
		{
			"denom already registered",
			func() {
				suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, pair.Denom, pair.GetID())
			},
			false,
		},
		{
			"meta data already stored",
			func() {
				suite.app.IntrarelayerKeeper.CreateCoinMetadata(suite.ctx, contractAddr)
			},
			false,
		},
		{
			"ok",
			func() {},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			contractAddr = suite.DeployContract(fip20Name, fip20Symbol, 18)
			suite.Commit()
			coinName := types.CreateDenom(contractAddr.String())
			pair = types.NewTokenPair(contractAddr, coinName, true, types.OWNER_EXTERNAL)

			tc.malleate()

			_, err := suite.app.IntrarelayerKeeper.RegisterFIP20(suite.ctx, contractAddr)
			metadata := suite.app.BankKeeper.GetDenomMetaData(suite.ctx, coinName)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				// Metadata variables
				suite.Require().True(len(metadata.Base) > 0)
				suite.Require().Equal(coinName, metadata.Base)
				//suite.Require().Equal(coinName, metadata.Name)
				suite.Require().Equal(fip20Symbol, metadata.Display)
				//suite.Require().Equal(fip20Symbol, metadata.Symbol)
				// Denom units
				suite.Require().Equal(len(metadata.DenomUnits), 2)
				suite.Require().Equal(coinName, metadata.DenomUnits[0].Denom)
				suite.Require().Equal(uint32(zeroExponent), metadata.DenomUnits[0].Exponent)
				suite.Require().Equal(fip20Symbol, metadata.DenomUnits[1].Denom)
				// Default exponent at contract creation is 18
				suite.Require().Equal(metadata.DenomUnits[1].Exponent, uint32(defaultExponent))
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite KeeperTestSuite) TestToggleRelay() {
	var (
		contractAddr common.Address
		id           []byte
		pair         types.TokenPair
	)

	testCases := []struct {
		name         string
		malleate     func()
		expPass      bool
		relayEnabled bool
	}{
		{
			"token not registered",
			func() {
				contractAddr = suite.DeployContract(fip20Name, fip20Symbol, 18)
				suite.Commit()
				pair = types.NewTokenPair(contractAddr, cosmosTokenName, true, types.OWNER_MODULE)
			},
			false,
			false,
		},
		{
			"token not registered - pair not found",
			func() {
				contractAddr = suite.DeployContract(fip20Name, fip20Symbol, 18)
				suite.Commit()
				pair = types.NewTokenPair(contractAddr, cosmosTokenName, true, types.OWNER_MODULE)
				suite.app.IntrarelayerKeeper.SetFIP20Map(suite.ctx, common.HexToAddress(pair.Fip20Address), pair.GetID())
			},
			false,
			false,
		},
		{
			"disable relay",
			func() {
				contractAddr = suite.setupRegisterFIP20Pair()
				id = suite.app.IntrarelayerKeeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, _ = suite.app.IntrarelayerKeeper.GetTokenPair(suite.ctx, id)
			},
			true,
			false,
		},
		{
			"disable and enable relay",
			func() {
				contractAddr = suite.setupRegisterFIP20Pair()
				id = suite.app.IntrarelayerKeeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, _ = suite.app.IntrarelayerKeeper.GetTokenPair(suite.ctx, id)
				pair, _ = suite.app.IntrarelayerKeeper.ToggleRelay(suite.ctx, contractAddr.String())
			},
			true,
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			var err error
			pair, err = suite.app.IntrarelayerKeeper.ToggleRelay(suite.ctx, contractAddr.String())
			// Request the pair using the GetPairToken func to make sure that is updated on the db
			pair, _ = suite.app.IntrarelayerKeeper.GetTokenPair(suite.ctx, id)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				if tc.relayEnabled {
					suite.Require().True(pair.Enabled)
				} else {
					suite.Require().False(pair.Enabled)
				}
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}
