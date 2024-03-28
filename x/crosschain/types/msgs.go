package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"

	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

// cross chain message types
const (
	TypeMsgBondedOracle   = "bonded_oracle"
	TypeMsgAddDelegate    = "add_delegate"
	TypeMsgReDelegate     = "re_delegate"
	TypeMsgEditBridger    = "edit_bridger"
	TypeMsgWithdrawReward = "withdraw_reward"
	TypeMsgUnbondedOracle = "unbonded_oracle"

	TypeMsgOracleSetConfirm      = "valset_confirm"
	TypeMsgOracleSetUpdatedClaim = "valset_updated_claim"

	TypeMsgBridgeTokenClaim = "bridge_token_claim"

	TypeMsgSendToFxClaim    = "send_to_fx_claim"
	TypeMsgBridgeCallClaim  = "bridge_call_claim"
	TypeMsgRefundTokenClaim = "refund_token_claim" //nolint:gosec

	TypeMsgSendToExternal       = "send_to_external"
	TypeMsgCancelSendToExternal = "cancel_send_to_external"
	TypeMsgIncreaseBridgeFee    = "increase_bridge_fee"
	TypeMsgSendToExternalClaim  = "send_to_external_claim"

	TypeMsgRequestBatch  = "request_batch"
	TypeMsgConfirmBatch  = "confirm_batch"
	TypeMsgConfirmRefund = "confirm_refund"

	TypeMsgUpdateParams = "update_params"

	TypeMsgUpdateChainOracles = "update_chain_oracles"
)

type (
	// CrossChainMsg cross msg must implement GetChainName interface.. using in router
	CrossChainMsg interface {
		GetChainName() string
	}
)

var (
	_ sdk.Msg       = &MsgBondedOracle{}
	_ CrossChainMsg = &MsgBondedOracle{}
	_ sdk.Msg       = &MsgAddDelegate{}
	_ CrossChainMsg = &MsgAddDelegate{}
	_ sdk.Msg       = &MsgReDelegate{}
	_ CrossChainMsg = &MsgReDelegate{}
	_ sdk.Msg       = &MsgEditBridger{}
	_ CrossChainMsg = &MsgEditBridger{}
	_ sdk.Msg       = &MsgWithdrawReward{}
	_ CrossChainMsg = &MsgWithdrawReward{}
	_ sdk.Msg       = &MsgUnbondedOracle{}
	_ CrossChainMsg = &MsgUnbondedOracle{}

	_ sdk.Msg       = &MsgOracleSetConfirm{}
	_ CrossChainMsg = &MsgOracleSetConfirm{}
	_ sdk.Msg       = &MsgOracleSetUpdatedClaim{}
	_ CrossChainMsg = &MsgOracleSetUpdatedClaim{}

	_ sdk.Msg       = &MsgBridgeTokenClaim{}
	_ CrossChainMsg = &MsgBridgeTokenClaim{}

	_ sdk.Msg       = &MsgSendToFxClaim{}
	_ CrossChainMsg = &MsgSendToFxClaim{}
	_ sdk.Msg       = &MsgBridgeCallClaim{}
	_ CrossChainMsg = &MsgBridgeCallClaim{}
	_ sdk.Msg       = &MsgRefundTokenClaim{}
	_ CrossChainMsg = &MsgRefundTokenClaim{}

	_ sdk.Msg       = &MsgSendToExternal{}
	_ CrossChainMsg = &MsgSendToExternal{}
	_ sdk.Msg       = &MsgCancelSendToExternal{}
	_ CrossChainMsg = &MsgCancelSendToExternal{}
	_ sdk.Msg       = &MsgIncreaseBridgeFee{}
	_ CrossChainMsg = &MsgIncreaseBridgeFee{}
	_ sdk.Msg       = &MsgSendToExternalClaim{}
	_ CrossChainMsg = &MsgSendToExternalClaim{}

	_ sdk.Msg       = &MsgRequestBatch{}
	_ CrossChainMsg = &MsgRequestBatch{}
	_ sdk.Msg       = &MsgConfirmBatch{}
	_ CrossChainMsg = &MsgConfirmBatch{}

	_ sdk.Msg       = &MsgConfirmRefund{}
	_ CrossChainMsg = &MsgConfirmRefund{}

	_ sdk.Msg       = &MsgUpdateParams{}
	_ CrossChainMsg = &MsgUpdateParams{}

	_ sdk.Msg       = &MsgUpdateChainOracles{}
	_ CrossChainMsg = &MsgUpdateChainOracles{}
)

type MsgValidateBasic interface {
	MsgBondedOracleValidate(m *MsgBondedOracle) (err error)
	MsgAddDelegateValidate(m *MsgAddDelegate) (err error)
	MsgReDelegateValidate(m *MsgReDelegate) (err error)
	MsgEditBridgerValidate(m *MsgEditBridger) (err error)
	MsgWithdrawRewardValidate(m *MsgWithdrawReward) (err error)
	MsgUnbondedOracleValidate(m *MsgUnbondedOracle) (err error)

	MsgOracleSetConfirmValidate(m *MsgOracleSetConfirm) (err error)
	MsgOracleSetUpdatedClaimValidate(m *MsgOracleSetUpdatedClaim) (err error)
	MsgBridgeTokenClaimValidate(m *MsgBridgeTokenClaim) (err error)
	MsgSendToExternalClaimValidate(m *MsgSendToExternalClaim) (err error)

	MsgSendToFxClaimValidate(m *MsgSendToFxClaim) (err error)
	MsgBridgeCallClaimValidate(m *MsgBridgeCallClaim) (err error)
	MsgRefundTokenClaimValidate(m *MsgRefundTokenClaim) (err error)
	MsgSendToExternalValidate(m *MsgSendToExternal) (err error)

	MsgCancelSendToExternalValidate(m *MsgCancelSendToExternal) (err error)
	MsgIncreaseBridgeFeeValidate(m *MsgIncreaseBridgeFee) (err error)
	MsgRequestBatchValidate(m *MsgRequestBatch) (err error)
	MsgConfirmBatchValidate(m *MsgConfirmBatch) (err error)
	MsgConfirmRefundValidate(m *MsgConfirmRefund) (err error)

	ValidateExternalAddress(addr string) error
	ExternalAddressToAccAddress(addr string) (sdk.AccAddress, error)
}

var reModuleName *regexp.Regexp

func init() {
	reModuleNameString := `[a-zA-Z][a-zA-Z0-9/]{1,32}`
	reModuleName = regexp.MustCompile(fmt.Sprintf(`^%s$`, reModuleNameString))
}

// ValidateModuleName is the default validation function for crosschain moduleName.
func ValidateModuleName(moduleName string) error {
	if !reModuleName.MatchString(moduleName) {
		return fmt.Errorf("invalid module name: %s", moduleName)
	}
	return nil
}

var msgValidateBasicRouter = make(map[string]MsgValidateBasic)

func MustGetMsgValidateBasic(chainName string) MsgValidateBasic {
	mvb, ok := msgValidateBasicRouter[chainName]
	if !ok {
		panic(fmt.Sprintf("chain %s validate basic not found", chainName))
	}
	return mvb
}

func GetValidateChains() []string {
	chains := make([]string, 0, len(msgValidateBasicRouter))
	for chainName := range msgValidateBasicRouter {
		chains = append(chains, chainName)
	}
	sort.SliceStable(chains, func(i, j int) bool {
		return chains[i] < chains[j]
	})
	return chains
}

func RegisterValidateBasic(chainName string, validate MsgValidateBasic) {
	if err := ValidateModuleName(chainName); err != nil {
		panic(errortypes.ErrInvalidRequest.Wrapf("invalid chain name: %s", chainName))
	}
	if _, ok := msgValidateBasicRouter[chainName]; ok {
		panic(fmt.Sprintf("duplicate registry msg validateBasic! chainName: %s", chainName))
	}
	msgValidateBasicRouter[chainName] = validate
}

// MsgBondedOracle

func (m *MsgBondedOracle) Route() string { return RouterKey }

func (m *MsgBondedOracle) Type() string { return TypeMsgBondedOracle }

func (m *MsgBondedOracle) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgBondedOracleValidate(m)
	}
}

func (m *MsgBondedOracle) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m *MsgBondedOracle) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.OracleAddress)}
}

// MsgAddDelegate

func (m *MsgAddDelegate) Route() string { return RouterKey }

func (m *MsgAddDelegate) Type() string {
	return TypeMsgAddDelegate
}

func (m *MsgAddDelegate) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgAddDelegateValidate(m)
	}
}

func (m *MsgAddDelegate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m *MsgAddDelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.OracleAddress)}
}

// MsgReDelegate

func (m *MsgReDelegate) Route() string { return RouterKey }

func (m *MsgReDelegate) Type() string {
	return TypeMsgReDelegate
}

func (m *MsgReDelegate) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgReDelegateValidate(m)
	}
}

func (m *MsgReDelegate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m *MsgReDelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.OracleAddress)}
}

// MsgEditBridger

func (m *MsgEditBridger) Route() string { return RouterKey }

func (m *MsgEditBridger) Type() string { return TypeMsgEditBridger }

func (m *MsgEditBridger) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgEditBridgerValidate(m)
	}
}

func (m *MsgEditBridger) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m *MsgEditBridger) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.OracleAddress)}
}

// MsgWithdrawReward

func (m *MsgWithdrawReward) Route() string { return RouterKey }

func (m *MsgWithdrawReward) Type() string { return TypeMsgWithdrawReward }

func (m *MsgWithdrawReward) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgWithdrawRewardValidate(m)
	}
}

func (m *MsgWithdrawReward) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m *MsgWithdrawReward) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.OracleAddress)}
}

// MsgUnbondedOracle

func (m *MsgUnbondedOracle) Route() string { return RouterKey }

func (m *MsgUnbondedOracle) Type() string { return TypeMsgUnbondedOracle }

func (m *MsgUnbondedOracle) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgUnbondedOracleValidate(m)
	}
}

func (m *MsgUnbondedOracle) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m *MsgUnbondedOracle) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.OracleAddress)}
}

// MsgOracleSetConfirm

// Route should return the name of the module
func (m *MsgOracleSetConfirm) Route() string { return RouterKey }

// Type should return the action
func (m *MsgOracleSetConfirm) Type() string { return TypeMsgOracleSetConfirm }

// ValidateBasic performs stateless checks
func (m *MsgOracleSetConfirm) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgOracleSetConfirmValidate(m)
	}
}

// GetSignBytes encodes the message for signing
func (m *MsgOracleSetConfirm) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners defines whose signature is required
func (m *MsgOracleSetConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.BridgerAddress)}
}

// MsgSendToExternal

// Route should return the name of the module
func (m *MsgSendToExternal) Route() string { return RouterKey }

// Type should return the action
func (m *MsgSendToExternal) Type() string { return TypeMsgSendToExternal }

// ValidateBasic runs stateless checks on the message
// Checks if the Eth address is valid
func (m *MsgSendToExternal) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgSendToExternalValidate(m)
	}
}

// GetSignBytes encodes the message for signing
func (m *MsgSendToExternal) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners defines whose signature is required
func (m *MsgSendToExternal) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// MsgRequestBatch

// Route should return the name of the module
func (m *MsgRequestBatch) Route() string { return RouterKey }

// Type should return the action
func (m *MsgRequestBatch) Type() string { return TypeMsgRequestBatch }

// ValidateBasic performs stateless checks
func (m *MsgRequestBatch) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgRequestBatchValidate(m)
	}
}

// GetSignBytes encodes the message for signing
func (m *MsgRequestBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners defines whose signature is required
func (m *MsgRequestBatch) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// MsgConfirmBatch

// Route should return the name of the module
func (m *MsgConfirmBatch) Route() string { return RouterKey }

// Type should return the action
func (m *MsgConfirmBatch) Type() string { return TypeMsgConfirmBatch }

// ValidateBasic performs stateless checks
func (m *MsgConfirmBatch) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgConfirmBatchValidate(m)
	}
}

// GetSignBytes encodes the message for signing
func (m *MsgConfirmBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners defines whose signature is required
func (m *MsgConfirmBatch) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.BridgerAddress)}
}

// MsgConfirmRefund

func (m *MsgConfirmRefund) Route() string { return RouterKey }

func (m *MsgConfirmRefund) Type() string { return TypeMsgConfirmRefund }

func (m *MsgConfirmRefund) ValidateBasic() error {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgConfirmRefundValidate(m)
	}
}

func (m *MsgConfirmRefund) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.BridgerAddress)}
}

func (m *MsgConfirmRefund) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// MsgCancelSendToExternal

// Route should return the name of the module
func (m *MsgCancelSendToExternal) Route() string { return RouterKey }

// Type should return the action
func (m *MsgCancelSendToExternal) Type() string { return TypeMsgCancelSendToExternal }

// ValidateBasic performs stateless checks
func (m *MsgCancelSendToExternal) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgCancelSendToExternalValidate(m)
	}
}

// GetSignBytes encodes the message for signing
func (m *MsgCancelSendToExternal) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners defines whose signature is required
func (m *MsgCancelSendToExternal) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// MsgIncreaseBridgeFee

// Route should return the name of the module
func (m *MsgIncreaseBridgeFee) Route() string { return RouterKey }

// Type should return the action
func (m *MsgIncreaseBridgeFee) Type() string { return TypeMsgIncreaseBridgeFee }

// ValidateBasic performs stateless checks
func (m *MsgIncreaseBridgeFee) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgIncreaseBridgeFeeValidate(m)
	}
}

// GetSignBytes encodes the message for signing
func (m *MsgIncreaseBridgeFee) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners defines whose signature is required
func (m *MsgIncreaseBridgeFee) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// ExternalClaim represents a claim on ethereum state
type ExternalClaim interface {
	proto.Message
	// GetEventNonce All Ethereum claims that we relay from the Gravity contract and into the module
	// have a nonce that is monotonically increasing and unique, since this nonce is
	// issued by the Ethereum contract it is immutable and must be agreed on by all validators
	// any disagreement on what claim goes to what nonce means someone is lying.
	GetEventNonce() uint64
	// GetBlockHeight The block height that the claimed event occurred on. This EventNonce provides sufficient
	// ordering for the execution of all claims. The block height is used only for batchTimeouts + logicTimeouts
	// when we go to create a new batch we set the timeout some number of batches out from the last
	// known height plus projected block progress since then.
	GetBlockHeight() uint64
	// GetClaimer the delegate address of the claimer, for MsgSendToExternalClaim and MsgSendToFxClaim
	// this is sent in as the sdk.AccAddress of the delegated key. it is up to the user
	// to disambiguate this into a sdk.ValAddress
	GetClaimer() sdk.AccAddress
	// GetType Which type of claim this is
	GetType() ClaimType
	ValidateBasic() error
	ClaimHash() []byte
}

var (
	_ ExternalClaim = &MsgSendToFxClaim{}
	_ ExternalClaim = &MsgBridgeCallClaim{}
	_ ExternalClaim = &MsgBridgeTokenClaim{}
	_ ExternalClaim = &MsgSendToExternalClaim{}
	_ ExternalClaim = &MsgOracleSetUpdatedClaim{}
	_ ExternalClaim = &MsgRefundTokenClaim{}
)

func UnpackAttestationClaim(cdc codectypes.AnyUnpacker, att *Attestation) (ExternalClaim, error) {
	var msg ExternalClaim
	err := cdc.UnpackAny(att.Claim, &msg)
	return msg, err
}

// MsgSendToFxClaim

// GetType returns the type of the claim
func (m *MsgSendToFxClaim) GetType() ClaimType {
	return CLAIM_TYPE_SEND_TO_FX
}

// ValidateBasic performs stateless checks
func (m *MsgSendToFxClaim) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgSendToFxClaimValidate(m)
	}
}

// GetSignBytes encodes the message for signing
func (m *MsgSendToFxClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m *MsgSendToFxClaim) GetClaimer() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(m.BridgerAddress)
}

// GetSigners defines whose signature is required
func (m *MsgSendToFxClaim) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.BridgerAddress)}
}

// Type should return the action
func (m *MsgSendToFxClaim) Type() string { return TypeMsgSendToFxClaim }

// Route should return the name of the module
func (m *MsgSendToFxClaim) Route() string { return RouterKey }

// ClaimHash Hash implements BridgeSendToExternal.Hash
func (m *MsgSendToFxClaim) ClaimHash() []byte {
	path := fmt.Sprintf("%d/%d%s/%s/%s/%s/%s", m.BlockHeight, m.EventNonce, m.TokenContract, m.Sender, m.Amount.String(), m.Receiver, m.TargetIbc)
	return tmhash.Sum([]byte(path))
}

// MsgBridgeCallClaim

// GetType returns the type of the claim
func (m *MsgBridgeCallClaim) GetType() ClaimType {
	return CLAIM_TYPE_BRIDGE_CALL
}

// ValidateBasic performs stateless checks
func (m *MsgBridgeCallClaim) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgBridgeCallClaimValidate(m)
	}
}

// GetSignBytes encodes the message for signing
func (m *MsgBridgeCallClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m *MsgBridgeCallClaim) GetClaimer() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(m.BridgerAddress)
}

// GetSigners defines whose signature is required
func (m *MsgBridgeCallClaim) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.BridgerAddress)}
}

// Type should return the action
func (m *MsgBridgeCallClaim) Type() string { return TypeMsgBridgeCallClaim }

// Route should return the name of the module
func (m *MsgBridgeCallClaim) Route() string { return RouterKey }

// ClaimHash Hash implements BridgeSendToExternal.Hash
func (m *MsgBridgeCallClaim) ClaimHash() []byte {
	path := fmt.Sprintf("%d/%d/%s/%s/%s/%s/%s/%s/%s/%d", m.BlockHeight, m.EventNonce, m.DstChainId, m.Sender, m.Receiver, m.To, m.Asset, m.Message, m.Value.String(), m.GasLimit)
	return tmhash.Sum([]byte(path))
}

func (m *MsgBridgeCallClaim) MustSender() common.Address {
	return common.BytesToAddress(ExternalAddressToAccAddress(m.ChainName, m.Sender).Bytes())
}

func (m *MsgBridgeCallClaim) MustReceiver() sdk.AccAddress {
	return ExternalAddressToAccAddress(m.ChainName, m.Receiver)
}

func (m *MsgBridgeCallClaim) MustTo() *common.Address {
	if len(m.To) == 0 {
		return nil
	}
	addr := common.BytesToAddress(ExternalAddressToAccAddress(m.ChainName, m.To).Bytes())
	return &addr
}

func (m *MsgBridgeCallClaim) MustMessage() []byte {
	if len(m.Message) == 0 {
		return []byte{}
	}
	bz, err := hex.DecodeString(m.Message)
	if err != nil {
		panic(err)
	}
	return bz
}

// MsgRefundTokenClaim

// GetType returns the type of the claim
func (m *MsgRefundTokenClaim) GetType() ClaimType {
	return CLAIM_TYPE_REFUND_TOKEN
}

// ValidateBasic performs stateless checks
func (m *MsgRefundTokenClaim) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgRefundTokenClaimValidate(m)
	}
}

// GetSignBytes encodes the message for signing
func (m *MsgRefundTokenClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m *MsgRefundTokenClaim) GetClaimer() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(m.BridgerAddress)
}

// GetSigners defines whose signature is required
func (m *MsgRefundTokenClaim) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.BridgerAddress)}
}

// Type should return the action
func (m *MsgRefundTokenClaim) Type() string { return TypeMsgRefundTokenClaim }

// Route should return the name of the module
func (m *MsgRefundTokenClaim) Route() string { return RouterKey }

// ClaimHash Hash implements BridgeSendToExternal.Hash
func (m *MsgRefundTokenClaim) ClaimHash() []byte {
	path := fmt.Sprintf("%d/%d/%d", m.BlockHeight, m.EventNonce, m.RefundNonce)
	return tmhash.Sum([]byte(path))
}

// MsgSendToExternalClaim

// GetType returns the claim type
func (m *MsgSendToExternalClaim) GetType() ClaimType {
	return CLAIM_TYPE_SEND_TO_EXTERNAL
}

// ValidateBasic performs stateless checks
func (m *MsgSendToExternalClaim) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgSendToExternalClaimValidate(m)
	}
}

// ClaimHash Hash implements SendToFxBatch.Hash
func (m *MsgSendToExternalClaim) ClaimHash() []byte {
	path := fmt.Sprintf("%d/%d/%s/%d/", m.BlockHeight, m.EventNonce, m.TokenContract, m.BatchNonce)
	return tmhash.Sum([]byte(path))
}

// GetSignBytes encodes the message for signing
func (m *MsgSendToExternalClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m *MsgSendToExternalClaim) GetClaimer() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(m.BridgerAddress)
}

// GetSigners defines whose signature is required
func (m *MsgSendToExternalClaim) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.BridgerAddress)}
}

// Route should return the name of the module
func (m *MsgSendToExternalClaim) Route() string { return RouterKey }

// Type should return the action
func (m *MsgSendToExternalClaim) Type() string { return TypeMsgSendToExternalClaim }

// MsgBridgeTokenClaim

func (m *MsgBridgeTokenClaim) Route() string { return RouterKey }

func (m *MsgBridgeTokenClaim) Type() string { return TypeMsgBridgeTokenClaim }

func (m *MsgBridgeTokenClaim) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgBridgeTokenClaimValidate(m)
	}
}

func (m *MsgBridgeTokenClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m *MsgBridgeTokenClaim) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.BridgerAddress)}
}

func (m *MsgBridgeTokenClaim) GetClaimer() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(m.BridgerAddress)
}

func (m *MsgBridgeTokenClaim) GetType() ClaimType {
	return CLAIM_TYPE_BRIDGE_TOKEN
}

func (m *MsgBridgeTokenClaim) ClaimHash() []byte {
	path := fmt.Sprintf("%d/%d%s/%s/%s/%d/%s/%d", m.BlockHeight, m.EventNonce, m.TokenContract, m.Name, m.Symbol, m.Decimals, m.ChannelIbc, m.TokenType)
	return tmhash.Sum([]byte(path))
}

// MsgOracleSetUpdatedClaim

// GetType returns the type of the claim
func (m *MsgOracleSetUpdatedClaim) GetType() ClaimType {
	return CLAIM_TYPE_ORACLE_SET_UPDATED
}

// ValidateBasic performs stateless checks
func (m *MsgOracleSetUpdatedClaim) ValidateBasic() (err error) {
	if router, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	} else {
		return router.MsgOracleSetUpdatedClaimValidate(m)
	}
}

// GetSignBytes encodes the message for signing
func (m *MsgOracleSetUpdatedClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m *MsgOracleSetUpdatedClaim) GetClaimer() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(m.BridgerAddress)
}

// GetSigners defines whose signature is required
func (m *MsgOracleSetUpdatedClaim) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.BridgerAddress)}
}

// Type should return the action
func (m *MsgOracleSetUpdatedClaim) Type() string { return TypeMsgOracleSetUpdatedClaim }

// Route should return the name of the module
func (m *MsgOracleSetUpdatedClaim) Route() string { return RouterKey }

// ClaimHash Hash implements BridgeSendToExternal.Hash
func (m *MsgOracleSetUpdatedClaim) ClaimHash() []byte {
	path := fmt.Sprintf("%d/%d/%d/%v/", m.BlockHeight, m.OracleSetNonce, m.EventNonce, m.Members)
	return tmhash.Sum([]byte(path))
}

func (m *MsgSetOrchestratorAddress) ValidateBasic() (err error) {
	return nil
}

func (m *MsgSetOrchestratorAddress) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.BridgerAddress)}
}

func (m *MsgAddOracleDeposit) ValidateBasic() (err error) {
	return nil
}

func (m *MsgAddOracleDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.OracleAddress)}
}

// Route returns the MsgUpdateParams message route.
func (m *MsgUpdateParams) Route() string { return ModuleName }

// Type returns the MsgUpdateParams message type.
func (m *MsgUpdateParams) Type() string { return TypeMsgUpdateParams }

// GetSignBytes returns the raw bytes for a MsgUpdateParams message that
// the expected signer needs to sign.
func (m *MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "authority")
	}
	if _, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	}
	if err := m.Params.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "params")
	}
	if len(m.Params.Oracles) > 0 {
		return errors.New("deprecated oracles")
	}
	return nil
}

// Route returns the MsgUpdateChainOracles message route.
func (m *MsgUpdateChainOracles) Route() string { return ModuleName }

// Type returns the MsgUpdateChainOracles message type.
func (m *MsgUpdateChainOracles) Type() string { return TypeMsgUpdateChainOracles }

// GetSignBytes returns the raw bytes for a MsgUpdateChainOracles message that
// the expected signer needs to sign.
func (m *MsgUpdateChainOracles) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgUpdateChainOracles) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "authority")
	}
	if _, ok := msgValidateBasicRouter[m.ChainName]; !ok {
		return errortypes.ErrInvalidRequest.Wrap("unrecognized cross chain name")
	}
	if len(m.Oracles) == 0 {
		return errors.New("empty oracles")
	}
	oraclesMap := make(map[string]bool)
	for _, addr := range m.Oracles {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return errorsmod.Wrap(err, "oracle address")
		}
		if oraclesMap[addr] {
			return errors.New("duplicate oracle address")
		}
		oraclesMap[addr] = true
	}
	return nil
}

func (m *MsgUpdateChainOracles) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}
