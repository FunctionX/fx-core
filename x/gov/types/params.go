package types

import (
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/gogoproto/proto"

	fxtypes "github.com/functionx/fx-core/v7/types"
	erc20types "github.com/functionx/fx-core/v7/x/erc20/types"
	evmtypes "github.com/functionx/fx-core/v7/x/evm/types"
)

var (
	DefaultMinInitialDeposit   = sdkmath.NewInt(1000).Mul(sdkmath.NewInt(1e18))
	DefaultEgfDepositThreshold = sdkmath.NewInt(10_000).Mul(sdkmath.NewInt(1e18))
	DefaultClaimRatio          = sdk.NewDecWithPrec(1, 1)  // 10%
	DefaultErc20Quorum         = sdk.NewDecWithPrec(25, 2) // 25%
	DefaultEvmQuorum           = sdk.NewDecWithPrec(25, 2) // 25%
	DefaultEgfVotingPeriod     = time.Hour * 24 * 14       // Default egf period for deposits & voting  14 days
	DefaultEvmVotingPeriod     = time.Hour * 24 * 2        // Default evm period for deposits & voting  2 days

	// FxBaseParamsKeyPrefix is the key to query all base params
	FxBaseParamsKeyPrefix = []byte("0x90")
	// FxEGFParamsKey is the key to query all EGF params
	FxEGFParamsKey    = []byte("0x91")
	FxSwitchParamsKey = []byte{0x92}
)

func NewParam(msgType string, minDeposit []sdk.Coin, minInitialDeposit sdk.Coin, votingPeriod *time.Duration,
	quorum string, maxDepositPeriod *time.Duration, threshold, vetoThreshold, minInitialDepositRatio string,
	burnVoteQuorum, burnProposalDepositPrevote, burnVoteVeto bool,
) *Params {
	return &Params{
		MsgType:                    msgType,
		MinDeposit:                 minDeposit,
		MinInitialDeposit:          minInitialDeposit,
		VotingPeriod:               votingPeriod,
		Quorum:                     quorum,
		MaxDepositPeriod:           maxDepositPeriod,
		Threshold:                  threshold,
		VetoThreshold:              vetoThreshold,
		MinInitialDepositRatio:     minInitialDepositRatio,
		BurnVoteQuorum:             burnVoteQuorum,
		BurnProposalDepositPrevote: burnProposalDepositPrevote,
		BurnVoteVeto:               burnVoteVeto,
	}
}

func NewEGFParam(egfDepositThreshold sdk.Coin, claimRatio string) *EGFParams {
	return &EGFParams{
		EgfDepositThreshold: egfDepositThreshold,
		ClaimRatio:          claimRatio,
	}
}

func DefaultParams() *Params {
	p := govv1.DefaultParams()
	return NewParam(sdk.MsgTypeURL(&evmtypes.MsgCallContract{}),
		p.GetMinDeposit(),
		sdk.NewCoin(fxtypes.DefaultDenom, DefaultMinInitialDeposit),
		&DefaultEvmVotingPeriod,
		p.Quorum,
		p.MaxDepositPeriod,
		p.Threshold,
		p.VetoThreshold,
		p.MinInitialDepositRatio,
		p.BurnVoteQuorum,
		p.BurnProposalDepositPrevote,
		p.BurnVoteVeto,
	)
}

func DefaultEGFParams() *EGFParams {
	return NewEGFParam(
		sdk.NewCoin(fxtypes.DefaultDenom, DefaultEgfDepositThreshold),
		DefaultClaimRatio.String(),
	)
}

// Erc20ProposalParams  register default erc20 parameters
func Erc20ProposalParams(minDeposit []sdk.Coin, minInitialDeposit sdk.Coin, votingPeriod *time.Duration, quorum string,
	maxDepositPeriod *time.Duration, threshold string, vetoThreshold, minInitialDepositRatio string, burnVoteQuorum, burnProposalDepositPrevote, burnVoteVeto bool,
) []*Params {
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
	baseParams := make([]*Params, 0, len(erc20MsgType))
	for _, msgType := range erc20MsgType {
		baseParams = append(baseParams, NewParam(msgType, minDeposit, minInitialDeposit, votingPeriod, quorum,
			maxDepositPeriod, threshold, vetoThreshold, minInitialDepositRatio, burnVoteQuorum, burnProposalDepositPrevote, burnVoteVeto))
	}
	return baseParams
}

// EVMProposalParams register default evm parameters
func EVMProposalParams(minDeposit []sdk.Coin, minInitialDeposit sdk.Coin, votingPeriod *time.Duration, quorum string,
	maxDepositPeriod *time.Duration, threshold string, vetoThreshold, minInitialDepositRatio string, burnVoteQuorum, burnProposalDepositPrevote, burnVoteVeto bool,
) []*Params {
	evmMsgType := []string{
		sdk.MsgTypeURL(&evmtypes.MsgCallContract{}),
	}
	baseParams := make([]*Params, 0, len(evmMsgType))
	for _, msgType := range evmMsgType {
		baseParams = append(baseParams, NewParam(msgType, minDeposit, minInitialDeposit, votingPeriod, quorum,
			maxDepositPeriod, threshold, vetoThreshold, minInitialDepositRatio, burnVoteQuorum, burnProposalDepositPrevote, burnVoteVeto))
	}
	return baseParams
}

// EGFProposalParams register default egf parameters
func EGFProposalParams(minDeposit []sdk.Coin, minInitialDeposit sdk.Coin, votingPeriod *time.Duration, quorum string,
	maxDepositPeriod *time.Duration, threshold string, vetoThreshold, minInitialDepositRatio string, burnVoteQuorum, burnProposalDepositPrevote, burnVoteVeto bool,
) []*Params {
	EGFMsgType := []string{
		"/cosmos.distribution.v1beta1.CommunityPoolSpendProposal",
		// TODO v1 MsgServer MsgCommunityPoolSpend pending
	}
	baseParams := make([]*Params, 0, len(EGFMsgType))
	for _, msgType := range EGFMsgType {
		baseParams = append(baseParams, NewParam(msgType, minDeposit, minInitialDeposit, votingPeriod, quorum,
			maxDepositPeriod, threshold, vetoThreshold, minInitialDepositRatio, burnVoteQuorum, burnProposalDepositPrevote, burnVoteVeto))
	}
	return baseParams
}

// ValidateBasic performs basic validation on governance parameters.
//
//gocyclo:ignore
func (p *Params) ValidateBasic() error {
	if p.MsgType != "" && proto.MessageType(strings.TrimPrefix(p.MsgType, "/")) == nil {
		return fmt.Errorf("proto message un registered: %s", p.MsgType)
	}
	if minDeposit := sdk.Coins(p.MinDeposit); minDeposit.Empty() || !minDeposit.IsValid() {
		return fmt.Errorf("invalid minimum deposit: %s", minDeposit)
	}
	if !p.MinInitialDeposit.IsValid() {
		return fmt.Errorf("invalid minimum initial deposit: %s", p.MinInitialDeposit)
	}
	if p.MaxDepositPeriod == nil {
		return fmt.Errorf("maximum deposit period must not be nil: %d", p.MaxDepositPeriod)
	}
	if p.MaxDepositPeriod.Seconds() <= 0 {
		return fmt.Errorf("maximum deposit period must be positive: %v", p.MaxDepositPeriod.String())
	}
	quorum, err := sdk.NewDecFromStr(p.Quorum)
	if err != nil {
		return fmt.Errorf("invalid quorum string: %w", err)
	}
	if quorum.IsNegative() {
		return fmt.Errorf("quorom cannot be negative: %s", quorum)
	}
	if quorum.GT(sdk.OneDec()) {
		return fmt.Errorf("quorom too large: %s", p.Quorum)
	}
	threshold, err := sdk.NewDecFromStr(p.Threshold)
	if err != nil {
		return fmt.Errorf("invalid threshold string: %w", err)
	}
	if !threshold.IsPositive() {
		return fmt.Errorf("vote threshold must be positive: %s", threshold)
	}
	if threshold.GT(sdk.OneDec()) {
		return fmt.Errorf("vote threshold too large: %s", threshold)
	}
	vetoThreshold, err := sdk.NewDecFromStr(p.VetoThreshold)
	if err != nil {
		return fmt.Errorf("invalid vetoThreshold string: %w", err)
	}
	if !vetoThreshold.IsPositive() {
		return fmt.Errorf("veto threshold must be positive: %s", vetoThreshold)
	}
	if vetoThreshold.GT(sdk.OneDec()) {
		return fmt.Errorf("veto threshold too large: %s", vetoThreshold)
	}
	if p.VotingPeriod == nil {
		return fmt.Errorf("voting period must not be nil: %d", p.VotingPeriod)
	}
	if p.VotingPeriod.Seconds() <= 0 {
		return fmt.Errorf("voting period must be positive: %s", p.VotingPeriod)
	}

	return nil
}

func (p *EGFParams) ValidateBasic() error {
	if !p.EgfDepositThreshold.IsValid() {
		return fmt.Errorf("invalid Egf Deposit Threshold: %s", p.EgfDepositThreshold)
	}
	ratio, err := sdk.NewDecFromStr(p.ClaimRatio)
	if err != nil {
		return fmt.Errorf("invalid egf claim ratio string: %w", err)
	}
	if ratio.IsNegative() {
		return fmt.Errorf("egf claim ratio cannot be negative: %s", ratio)
	}
	if ratio.GT(sdk.OneDec()) {
		return fmt.Errorf("egf claim ratio too large: %s", ratio)
	}
	return nil
}

func ParamsByMsgTypeKey(msgType string) []byte {
	return append(FxBaseParamsKeyPrefix, []byte(msgType)...)
}

func ExtractMsgTypeURL(msgs []*codectypes.Any) string {
	if len(msgs) == 0 {
		return ""
	}
	msg := msgs[0]
	if strings.EqualFold(msg.TypeUrl, sdk.MsgTypeURL(&govv1.MsgExecLegacyContent{})) {
		legacyContent := msg.GetCachedValue().(*govv1.MsgExecLegacyContent)
		content := legacyContent.GetContent()
		return content.TypeUrl
	}
	return msg.TypeUrl
}

func CheckEGFProposalMsg(msgs []*codectypes.Any) (bool, sdk.Coins) {
	totalCommunityPoolSpendAmount := sdk.NewCoins()
	for _, msg := range msgs {
		if strings.EqualFold(msg.TypeUrl, sdk.MsgTypeURL(&distributiontypes.MsgCommunityPoolSpend{})) {
			communityPoolSpendProposal := msg.GetCachedValue().(*distributiontypes.MsgCommunityPoolSpend)
			totalCommunityPoolSpendAmount = totalCommunityPoolSpendAmount.Add(communityPoolSpendProposal.Amount...)
		} else {
			return false, nil
		}
	}
	return true, totalCommunityPoolSpendAmount
}

func (p *SwitchParams) ValidateBasic() error {
	duplicate := make(map[string]bool)
	for _, precompile := range p.DisablePrecompiles {
		if duplicate[precompile] {
			return fmt.Errorf("duplicate precompile: %s", precompile)
		}
		duplicate[precompile] = true
	}

	duplicate = make(map[string]bool)
	for _, msgType := range p.DisableMsgTypes {
		if duplicate[msgType] {
			return fmt.Errorf("duplicate msg type: %s", msgType)
		}
		duplicate[msgType] = true
	}
	return nil
}
