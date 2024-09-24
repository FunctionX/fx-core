package helpers

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	tenderminttypes "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/stretchr/testify/suite"

	"github.com/functionx/fx-core/v8/app"
)

type BaseSuite struct {
	suite.Suite
	MintValNumber int
	ValSet        *tmtypes.ValidatorSet
	ValAddr       []sdk.ValAddress
	App           *app.App
	Ctx           sdk.Context
}

func (s *BaseSuite) SetupTest() {
	valNumber := s.MintValNumber
	if valNumber <= 0 {
		valNumber = tmrand.Intn(10) + 1
	}
	valSet, valAccounts, valBalances := GenerateGenesisValidator(valNumber, sdk.Coins{})
	s.ValSet = valSet
	for _, account := range valAccounts {
		s.ValAddr = append(s.ValAddr, account.GetAddress().Bytes())
	}
	s.App = SetupWithGenesisValSet(s.T(), valSet, valAccounts, valBalances...)
	s.Ctx = s.App.GetContextForFinalizeBlock(nil)
	s.Ctx = s.Ctx.WithProposer(s.ValSet.Proposer.Address.Bytes())
}

func (s *BaseSuite) AddTestSigner() *Signer {
	signer := NewSigner(NewEthPrivKey())
	s.MintToken(signer.AccAddress(), NewStakingCoin(100, 18))
	return signer
}

func (s *BaseSuite) Commit(block ...int64) sdk.Context {
	ctx := s.Ctx
	lastBlockHeight := s.Ctx.BlockHeight()
	nextHeight := lastBlockHeight + 1
	if len(block) > 0 {
		nextHeight = lastBlockHeight + block[0]
	}
	commitInfo := abci.CommitInfo{
		Round: 1,
	}

	for _, val := range s.ValSet.Validators {
		pk, err := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
		s.NoError(err)
		commitInfo.Votes = append(commitInfo.Votes, abci.VoteInfo{
			Validator: abci.Validator{
				Address: pk.Address(),
				Power:   val.VotingPower,
			},
			BlockIdFlag: tenderminttypes.BlockIDFlagCommit,
		})

		signingInfo := slashingtypes.NewValidatorSigningInfo(
			sdk.ConsAddress(pk.Address()),
			ctx.BlockHeight(),
			0,
			time.Unix(0, 0),
			false,
			0,
		)
		s.NoError(s.App.SlashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(pk.Address()), signingInfo))
	}

	for i := lastBlockHeight; i < nextHeight; i++ {
		// 1. try to finalize the block + commit finalizeBlockState
		if _, err := s.App.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height:            i,
			Time:              tmtime.Now(),
			ProposerAddress:   s.Ctx.BlockHeader().ProposerAddress,
			DecidedLastCommit: commitInfo,
		}); err != nil {
			panic(err)
		}

		// 2. commit lastCommitInfo
		if _, err := s.App.Commit(); err != nil {
			panic(err)
		}

		// 3. prepare to process new blocks (myApp.GetContextForFinalizeBlock(nil))
		if _, err := s.App.ProcessProposal(&abci.RequestProcessProposal{
			Height:             i + 1,
			Time:               tmtime.Now(),
			ProposerAddress:    s.Ctx.BlockHeader().ProposerAddress,
			ProposedLastCommit: commitInfo,
		}); err != nil {
			panic(err)
		}

		// 4. get new ctx for finalizeBlockState
		ctx = s.App.GetContextForFinalizeBlock(nil)
	}
	s.Ctx = ctx
	return ctx
}

func (s *BaseSuite) AddTestSigners(accNum int, coin sdk.Coin) []*Signer {
	signers := make([]*Signer, accNum)
	for i := 0; i < accNum; i++ {
		signer := NewSigner(NewEthPrivKey())
		s.MintToken(signer.AccAddress(), coin)
	}
	return signers
}

func (s *BaseSuite) MintToken(address sdk.AccAddress, amount sdk.Coin) {
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(amount))
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, address.Bytes(), sdk.NewCoins(amount))
	s.Require().NoError(err)
}
