package helpers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ed255192 "github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/functionx/fx-core/v2/app"
	fxtypes "github.com/functionx/fx-core/v2/types"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

// SimAppChainID hardcoded chainID for simulation
const (
	SimAppChainID = "sim-app"
)

// DefaultConsensusParams defines the default Tendermint consensus params used
// in GaiaApp testing.
var DefaultConsensusParams = &abci.ConsensusParams{
	Block: &abci.BlockParams{
		MaxBytes: 1048576,
		MaxGas:   -1,
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 1000000,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        100000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
}

type EmptyAppOptions struct{}

func (EmptyAppOptions) Get(string) interface{} { return nil }

func Setup(isCheckTx bool, isShowLog bool) *app.App {
	logger := log.NewNopLogger()
	if isShowLog {
		logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	}

	myApp := app.New(logger, dbm.NewMemDB(),
		nil, true, map[int64]bool{}, os.TempDir(), 5,
		app.MakeEncodingConfig(), EmptyAppOptions{},
	)
	if !isCheckTx {
		// InitChain must be called to stop deliverState from being nil
		stateBytes, err := json.MarshalIndent(DefGenesisState(myApp.AppCodec()), "", " ")
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		myApp.InitChain(abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		})
	}

	return myApp
}

func DefGenesisState(cdc codec.Codec) app.GenesisState {
	genesis := app.NewDefAppGenesisByDenom(fxtypes.DefaultDenom, cdc)
	bankState := new(banktypes.GenesisState)
	cdc.MustUnmarshalJSON(genesis[banktypes.ModuleName], bankState)
	bankState.Balances = append(bankState.Balances, banktypes.Balance{
		Address: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(fxtypes.DefaultDenom, sdk.NewIntFromUint64(4_000).MulRaw(1e18))),
	})
	genesis[banktypes.ModuleName] = cdc.MustMarshalJSON(bankState)
	return genesis
}

func GenerateGenesisValidator(validatorNum int, initCoins sdk.Coins) (valSet *tmtypes.ValidatorSet, genAccs authtypes.GenesisAccounts, balances []banktypes.Balance) {
	if initCoins == nil || initCoins.Len() <= 0 {
		initCoins = sdk.NewCoins(sdk.NewCoin(fxtypes.DefaultDenom, sdk.NewInt(10_000).MulRaw(1e18)))
	}
	validators := make([]*tmtypes.Validator, 0, validatorNum)
	genAccs = make(authtypes.GenesisAccounts, 0)
	balances = make([]banktypes.Balance, 0, validatorNum)
	for i := 0; i < validatorNum; i++ {
		validator := tmtypes.NewValidator(ed255192.GenPrivKey().PubKey(), 1)
		validators = append(validators, validator)

		senderPrivKey := secp256k1.GenPrivKey()
		acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
		genAccs = append(genAccs, acc)

		balance := banktypes.Balance{
			Address: acc.GetAddress().String(),
			Coins:   initCoins,
		}
		balances = append(balances, balance)
	}
	return tmtypes.NewValidatorSet(validators), genAccs, balances
}

// SetupWithGenesisValSet initializes a new SimApp with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit (10^6) in the default token of the simapp from first genesis
// account. A Nop logger is set in SimApp.
func SetupWithGenesisValSet(t *testing.T, valSet *tmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *app.App {
	myApp := Setup(true, false)
	genesisState := DefGenesisState(myApp.AppCodec())

	// set genesis accounts
	var authGenesis authtypes.GenesisState
	myApp.AppCodec().MustUnmarshalJSON(genesisState[authtypes.ModuleName], &authGenesis)
	packAccounts, err := authtypes.PackAccounts(genAccs)
	require.NoError(t, err)
	authGenesis.Accounts = packAccounts
	genesisState[authtypes.ModuleName] = myApp.AppCodec().MustMarshalJSON(&authGenesis)

	// set validators and delegations
	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.NewIntFromUint64(1e18).Mul(sdk.NewInt(100))

	for i, val := range valSet.Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		require.NoError(t, err)
		pkAny, err := codectypes.NewAnyWithValue(pk)
		require.NoError(t, err)
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(genAccs[i].GetAddress()).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdk.OneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			MinSelfDelegation: sdk.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[i].GetAddress(), validator.GetOperator(), sdk.OneDec()))
	}

	var stakingGenesis stakingtypes.GenesisState
	myApp.AppCodec().MustUnmarshalJSON(genesisState[stakingtypes.ModuleName], &stakingGenesis)
	stakingGenesis.Validators = validators
	stakingGenesis.Delegations = delegations
	genesisState[stakingtypes.ModuleName] = myApp.AppCodec().MustMarshalJSON(&stakingGenesis)

	// update balances and total supply
	var bankGenesis banktypes.GenesisState
	myApp.AppCodec().MustUnmarshalJSON(genesisState[banktypes.ModuleName], &bankGenesis)
	for _, b := range balances {
		// add genesis acc tokens and delegated tokens to total supply
		bankGenesis.Supply = bankGenesis.Supply.Add(b.Coins.Add()...)
	}
	for range valSet.Validators {
		bankGenesis.Supply = bankGenesis.Supply.Add(sdk.NewCoin(stakingGenesis.Params.BondDenom, bondAmt))
	}
	bankGenesis.Balances = append(bankGenesis.Balances, balances...)
	// add bonded amount to bonded pool module account
	bankGenesis.Balances = append(bankGenesis.Balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(stakingGenesis.Params.BondDenom, bondAmt.MulRaw(int64(len(validators))))},
	})
	genesisState[banktypes.ModuleName] = myApp.AppCodec().MustMarshalJSON(&bankGenesis)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	myApp.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	// commit genesis changes
	myApp.Commit()
	myApp.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{
		Height:             myApp.LastBlockHeight() + 1,
		AppHash:            myApp.LastCommitID().Hash,
		ValidatorsHash:     valSet.Hash(),
		NextValidatorsHash: valSet.Hash(),
	}})

	return myApp
}

// SetupWithGenesisAccounts initializes a new SimApp with the provided genesis
// accounts and possible balances.
func SetupWithGenesisAccounts(genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *app.App {
	myApp := Setup(true, false)
	genesisState := DefGenesisState(myApp.AppCodec())

	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = myApp.AppCodec().MustMarshalJSON(authGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		totalSupply = totalSupply.Add(b.Coins...)
	}

	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{})
	genesisState[banktypes.ModuleName] = myApp.AppCodec().MustMarshalJSON(bankGenesis)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(err)
	}

	myApp.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	myApp.Commit()
	myApp.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: myApp.LastBlockHeight() + 1}})

	return myApp
}

type GenerateAccountStrategy func(int) []sdk.AccAddress

// createRandomAccounts is a strategy used by addTestAddrs() in order to generated addresses in random order.
func createRandomAccounts(accNum int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, accNum)
	for i := 0; i < accNum; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}

// createIncrementalAccounts is a strategy used by addTestAddrs() in order to generated addresses in ascending order.
func createIncrementalAccounts(accNum int) []sdk.AccAddress {
	var addresses []sdk.AccAddress
	var buffer bytes.Buffer

	// start at 100 so we can make up to 999 test addresses with valid test addresses
	for i := 100; i < (accNum + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6") // base address string

		buffer.WriteString(numString) // adding on final two digits to make addresses unique
		res, _ := sdk.AccAddressFromHex(buffer.String())
		bech := res.String()
		addr, _ := TestAddr(buffer.String(), bech)

		addresses = append(addresses, addr)
		buffer.Reset()
	}

	return addresses
}

// AddTestAddrsFromPubKeys adds the addresses into the SimApp providing only the public keys.
func AddTestAddrsFromPubKeys(myApp *app.App, ctx sdk.Context, pubKeys []cryptotypes.PubKey, accAmt sdk.Int) {
	initCoins := sdk.NewCoins(sdk.NewCoin(myApp.StakingKeeper.BondDenom(ctx), accAmt))

	for _, pk := range pubKeys {
		initAccountWithCoins(myApp, ctx, sdk.AccAddress(pk.Address()), initCoins)
	}
}

// AddTestAddrs constructs and returns accNum amount of accounts with an
// initial balance of accAmt in random order
func AddTestAddrs(myApp *app.App, ctx sdk.Context, accNum int, accAmt sdk.Int) []sdk.AccAddress {
	return addTestAddrs(myApp, ctx, accNum, accAmt, createRandomAccounts)
}

func AddTestAddrsIncremental(myApp *app.App, ctx sdk.Context, accNum int, accAmt sdk.Int) []sdk.AccAddress {
	return addTestAddrs(myApp, ctx, accNum, accAmt, createIncrementalAccounts)
}

func addTestAddrs(myApp *app.App, ctx sdk.Context, accNum int, accAmt sdk.Int, strategy GenerateAccountStrategy) []sdk.AccAddress {
	testAddrs := strategy(accNum)

	initCoins := sdk.NewCoins(sdk.NewCoin(myApp.StakingKeeper.BondDenom(ctx), accAmt))

	for _, addr := range testAddrs {
		initAccountWithCoins(myApp, ctx, addr, initCoins)
	}

	return testAddrs
}

func initAccountWithCoins(myApp *app.App, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	err := myApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	if err != nil {
		panic(err)
	}

	err = myApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins)
	if err != nil {
		panic(err)
	}
}

// ConvertAddrsToValAddrs converts the provided addresses to ValAddress.
func ConvertAddrsToValAddrs(addrs []sdk.AccAddress) []sdk.ValAddress {
	valAddrs := make([]sdk.ValAddress, len(addrs))

	for i, addr := range addrs {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	return valAddrs
}

func TestAddr(addr string, bech string) (sdk.AccAddress, error) {
	res, err := sdk.AccAddressFromHex(addr)
	if err != nil {
		return nil, err
	}
	bechExpected := res.String()
	if bech != bechExpected {
		return nil, fmt.Errorf("bech encoding doesn't match reference")
	}

	bechres, err := sdk.AccAddressFromBech32(bech)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(bechres, res) {
		return nil, err
	}

	return res, nil
}

// CheckBalance checks the balance of an account.
func CheckBalance(t *testing.T, myApp *app.App, addr sdk.AccAddress, balances sdk.Coins) {
	ctxCheck := myApp.BaseApp.NewContext(true, tmproto.Header{})
	require.True(t, balances.IsEqual(myApp.BankKeeper.GetAllBalances(ctxCheck, addr)))
}

// SignCheckDeliver checks a generated signed transaction and simulates a
// block commitment with the given transaction. A test assertion is made using
// the parameter 'expPass' against the result. A corresponding result is
// returned.
func SignCheckDeliver(t *testing.T, txCfg client.TxConfig, app *baseapp.BaseApp, header tmproto.Header, msgs []sdk.Msg,
	chainID string, accNums, accSeqs []uint64, expSimPass, expPass bool, priv ...cryptotypes.PrivKey,
) (sdk.GasInfo, *sdk.Result, error) {

	tx, err := GenTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txCfg,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
		helpers.DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)
	require.NoError(t, err)
	txBytes, err := txCfg.TxEncoder()(tx)
	require.Nil(t, err)

	// Must simulate now as CheckTx doesn't run Msgs anymore
	_, res, err := app.Simulate(txBytes)

	if expSimPass {
		require.NoError(t, err)
		require.NotNil(t, res)
	} else {
		require.Error(t, err)
		require.Nil(t, res)
	}

	// Simulate a sending a transaction and committing a block
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	gInfo, res, err := app.Deliver(txCfg.TxEncoder(), tx)

	if expPass {
		require.NoError(t, err)
		require.NotNil(t, res)
	} else {
		require.Error(t, err)
		require.Nil(t, res)
	}

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	return gInfo, res, err
}

// GenSequenceOfTxs generates a set of signed transactions of messages, such
// that they differ only by having the sequence numbers incremented between
// every transaction.
func GenSequenceOfTxs(txGen client.TxConfig, msgs []sdk.Msg, accNums []uint64, initSeqNums []uint64, numToGenerate int, priv ...cryptotypes.PrivKey) ([]sdk.Tx, error) {
	txs := make([]sdk.Tx, numToGenerate)
	var err error
	for i := 0; i < numToGenerate; i++ {
		txs[i], err = GenTx(
			rand.New(rand.NewSource(time.Now().UnixNano())),
			txGen,
			msgs,
			sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
			helpers.DefaultGenTxGas,
			"",
			accNums,
			initSeqNums,
			priv...,
		)
		if err != nil {
			break
		}
		incrementAllSequenceNumbers(initSeqNums)
	}

	return txs, err
}

func incrementAllSequenceNumbers(initSeqNums []uint64) {
	for i := 0; i < len(initSeqNums); i++ {
		initSeqNums[i]++
	}
}

// GenTx generates a signed mock transaction.
func GenTx(r *rand.Rand, gen client.TxConfig, msgs []sdk.Msg, feeAmt sdk.Coins, gas uint64, chainID string, accNums, accSeqs []uint64, priv ...cryptotypes.PrivKey) (sdk.Tx, error) {
	sigs := make([]signing.SignatureV2, len(priv))

	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))

	signMode := gen.SignModeHandler().DefaultMode()

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range priv {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
			Sequence: accSeqs[i],
		}
	}

	tx := gen.NewTxBuilder()
	err := tx.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = tx.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	tx.SetMemo(memo)
	tx.SetFeeAmount(feeAmt)
	tx.SetGasLimit(gas)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range priv {
		signerData := authsign.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		signBytes, err := gen.SignModeHandler().GetSignBytes(signMode, signerData, tx.GetTx())
		if err != nil {
			panic(err)
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			panic(err)
		}
		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
		err = tx.SetSignatures(sigs...)
		if err != nil {
			panic(err)
		}
	}

	return tx.GetTx(), nil
}

// CreateTestPubKeys returns a total of numPubKeys public keys in ascending order.
func CreateTestPubKeys(numPubKeys int) []cryptotypes.PubKey {
	var publicKeys []cryptotypes.PubKey
	var buffer bytes.Buffer

	// start at 10 to avoid changing 1 to 01, 2 to 02, etc
	for i := 100; i < (numPubKeys + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AF") // base pubkey string
		buffer.WriteString(numString)                                                       // adding on final two digits to make pubkeys unique
		publicKeys = append(publicKeys, NewPubKeyFromHex(buffer.String()))
		buffer.Reset()
	}

	return publicKeys
}

// NewPubKeyFromHex returns a PubKey from a hex string.
func NewPubKeyFromHex(pk string) (res cryptotypes.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	if len(pkBytes) != ed25519.PubKeySize {
		panic(errors.Wrap(errors.ErrInvalidPubKey, "invalid pubkey size"))
	}
	return &ed25519.PubKey{Key: pkBytes}
}
