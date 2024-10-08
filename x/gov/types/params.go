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

	erc20types "github.com/functionx/fx-core/v8/x/erc20/types"
	evmtypes "github.com/functionx/fx-core/v8/x/evm/types"
)

var (
	DefaultCustomParamVotingPeriod    = time.Hour * 24 * 7  // Default period for deposits & voting  7 days
	DefaultEGFCustomParamVotingPeriod = time.Hour * 24 * 14 // Default egf period for deposits & voting  14 days

	EGFCustomParamDepositRatio     = sdkmath.LegacyNewDecWithPrec(1, 1) // 10%
	DefaultCustomParamDepositRatio = sdkmath.LegacyZeroDec()

	DefaultCustomParamQuorum40 = sdkmath.LegacyNewDecWithPrec(4, 1)
	DefaultCustomParamQuorum25 = sdkmath.LegacyNewDecWithPrec(25, 2) // 25%
)

var (
	// todo delete this store
	// FxBaseParamsKeyPrefix is the key to query all base params
	FxBaseParamsKeyPrefix = []byte("0x90")
	// todo delete this store
	// FxEGFParamsKey is the key to query all EGF params
	FxEGFParamsKey = []byte("0x91")

	FxSwitchParamsKey = []byte{0x92}
	CustomParamsKey   = []byte{0x93}
)

type InitGenesisCustomParams struct {
	MsgType string
	Params  CustomParams
}

func NewInitGenesisCustomParams(msgType string, params CustomParams) InitGenesisCustomParams {
	return InitGenesisCustomParams{
		MsgType: msgType,
		Params:  params,
	}
}

func NewCustomParams(depositRatio string, votingPeriod time.Duration, quorum string) *CustomParams {
	return &CustomParams{
		DepositRatio: depositRatio,
		VotingPeriod: &votingPeriod,
		Quorum:       quorum,
	}
}

func DefaultInitGenesisCustomParams() []InitGenesisCustomParams {
	var customParams []InitGenesisCustomParams
	customParams = append(customParams, newEGFCustomParams())
	customParams = append(customParams, newOtherCustomParams()...)
	return customParams
}

func newEGFCustomParams() InitGenesisCustomParams {
	return NewInitGenesisCustomParams(
		sdk.MsgTypeURL(&distributiontypes.MsgCommunityPoolSpend{}),
		*NewCustomParams(EGFCustomParamDepositRatio.String(), DefaultEGFCustomParamVotingPeriod, DefaultCustomParamQuorum40.String()),
	)
}

func newOtherCustomParams() []InitGenesisCustomParams {
	params := []InitGenesisCustomParams{}
	defaultParams := *NewCustomParams(DefaultCustomParamDepositRatio.String(), DefaultCustomParamVotingPeriod, DefaultCustomParamQuorum25.String())
	customMsgTypes := []string{
		// erc20 proposal
		sdk.MsgTypeURL(&erc20types.MsgRegisterCoin{}),
		sdk.MsgTypeURL(&erc20types.MsgRegisterERC20{}),
		sdk.MsgTypeURL(&erc20types.MsgToggleTokenConversion{}),
		sdk.MsgTypeURL(&erc20types.MsgUpdateDenomAlias{}),

		// evm proposal
		sdk.MsgTypeURL(&evmtypes.MsgCallContract{}),
	}
	for _, msgType := range customMsgTypes {
		params = append(params, NewInitGenesisCustomParams(msgType, defaultParams))
	}

	return params
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
