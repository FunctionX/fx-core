package types

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/functionx/fx-core/v7/contract"
)

var _ MsgValidateBasic = &MsgValidate{}

type MsgValidate struct{}

func (b MsgValidate) MsgBondedOracleValidate(m *MsgBondedOracle) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.OracleAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid oracle address: %s", err)
	}
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if err = contract.ValidateEthereumAddress(m.ExternalAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid external address: %s", err)
	}
	if !m.DelegateAmount.IsValid() || m.DelegateAmount.IsNegative() {
		return errortypes.ErrInvalidRequest.Wrap("invalid delegation amount")
	}
	if m.OracleAddress == m.BridgerAddress {
		return errortypes.ErrInvalidRequest.Wrap("same address")
	}
	return nil
}

func (b MsgValidate) MsgAddDelegateValidate(m *MsgAddDelegate) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.OracleAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid oracle address: %s", err)
	}
	if !m.Amount.IsValid() || !m.Amount.IsPositive() {
		return errortypes.ErrInvalidRequest.Wrap("invalid amount")
	}
	return nil
}

func (b MsgValidate) MsgReDelegateValidate(m *MsgReDelegate) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.OracleAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid oracle address: %s", err)
	}
	if _, err = sdk.ValAddressFromBech32(m.ValidatorAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}
	return nil
}

func (b MsgValidate) MsgEditBridgerValidate(m *MsgEditBridger) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.OracleAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid oracle address: %s", err)
	}
	if _, err = sdk.ValAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if m.OracleAddress == m.BridgerAddress {
		return errortypes.ErrInvalidRequest.Wrap("same address")
	}
	return nil
}

func (b MsgValidate) MsgWithdrawRewardValidate(m *MsgWithdrawReward) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.OracleAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid oracle address: %s", err)
	}
	return nil
}

func (b MsgValidate) MsgUnbondedOracleValidate(m *MsgUnbondedOracle) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.OracleAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid oracle address: %s", err)
	}
	return nil
}

func (b MsgValidate) MsgOracleSetConfirmValidate(m *MsgOracleSetConfirm) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if err = contract.ValidateEthereumAddress(m.ExternalAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid external address: %s", err)
	}
	if len(m.Signature) == 0 {
		return errortypes.ErrInvalidRequest.Wrap("empty signature")
	}
	if _, err = hex.DecodeString(m.Signature); err != nil {
		return errortypes.ErrInvalidRequest.Wrap("could not hex decode signature")
	}
	return nil
}

func (b MsgValidate) MsgOracleSetUpdatedClaimValidate(m *MsgOracleSetUpdatedClaim) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if len(m.Members) == 0 {
		return errortypes.ErrInvalidRequest.Wrap("empty members")
	}
	for _, member := range m.Members {
		if err = contract.ValidateEthereumAddress(member.ExternalAddress); err != nil {
			return errortypes.ErrInvalidAddress.Wrapf("invalid external address: %s", err)
		}
		if member.Power == 0 {
			return errortypes.ErrInvalidRequest.Wrap("zero power")
		}
	}
	if m.EventNonce == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero event nonce")
	}
	if m.BlockHeight == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero block height")
	}
	return nil
}

func (b MsgValidate) MsgBridgeTokenClaimValidate(m *MsgBridgeTokenClaim) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if err = contract.ValidateEthereumAddress(m.TokenContract); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid token contract: %s", err)
	}
	if _, err = hex.DecodeString(m.ChannelIbc); len(m.ChannelIbc) > 0 && err != nil {
		return errortypes.ErrInvalidRequest.Wrap("could not decode hex channelIbc string")
	}
	if len(m.Name) == 0 {
		return errortypes.ErrInvalidRequest.Wrap("empty token name")
	}
	if len(m.Symbol) == 0 {
		return errortypes.ErrInvalidRequest.Wrap("empty token symbol")
	}
	if m.EventNonce == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero event nonce")
	}
	if m.BlockHeight == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero block height")
	}
	return ValidateBridgeTokenType(m.TokenType)
}

func (b MsgValidate) MsgSendToExternalClaimValidate(m *MsgSendToExternalClaim) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if err = contract.ValidateEthereumAddress(m.TokenContract); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid token contract: %s", err)
	}
	if m.EventNonce == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero event nonce")
	}
	if m.BlockHeight == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero block height")
	}
	if m.BatchNonce == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero batch nonce")
	}
	return nil
}

func (b MsgValidate) MsgSendToFxClaimValidate(m *MsgSendToFxClaim) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if err = contract.ValidateEthereumAddress(m.Sender); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	if err = contract.ValidateEthereumAddress(m.TokenContract); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid token contract: %s", err)
	}
	if _, err = sdk.AccAddressFromBech32(m.Receiver); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid receiver address: %s", err)
	}
	if m.Amount.IsNil() || m.Amount.IsNegative() {
		return errortypes.ErrInvalidRequest.Wrap("invalid amount")
	}
	if _, err = hex.DecodeString(m.TargetIbc); len(m.TargetIbc) > 0 && err != nil {
		return errortypes.ErrInvalidRequest.Wrap("could not decode hex targetIbc")
	}
	if m.EventNonce == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero event nonce")
	}
	if m.BlockHeight == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero block height")
	}
	return nil
}

func (b MsgValidate) MsgBridgeCallClaimValidate(m *MsgBridgeCallClaim) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if m.DstChainId == "" {
		return errortypes.ErrInvalidRequest.Wrap("empty dst chain id")
	}
	if err = contract.ValidateEthereumAddress(m.Sender); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	if len(m.To) > 0 {
		if err = contract.ValidateEthereumAddress(m.To); err != nil {
			return errortypes.ErrInvalidAddress.Wrapf("invalid to contract: %s", err)
		}
	}
	if err = contract.ValidateEthereumAddress(m.Receiver); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid receiver address: %s", err)
	}
	if m.Value.IsNil() || m.Value.IsNegative() {
		return errortypes.ErrInvalidRequest.Wrap("invalid value")
	}
	if len(m.Message) > 0 {
		if _, err := hex.DecodeString(m.Message); err != nil {
			return errortypes.ErrInvalidRequest.Wrap("invalid message")
		}
	}
	if m.EventNonce == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero event nonce")
	}
	if m.BlockHeight == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero block height")
	}
	return nil
}

func (b MsgValidate) MsgRefundTokenClaimValidate(m *MsgRefundTokenClaim) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if m.RefundNonce == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero refund nonce")
	}
	if m.EventNonce == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero event nonce")
	}
	if m.BlockHeight == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero block height")
	}
	return nil
}

func (b MsgValidate) MsgSendToExternalValidate(m *MsgSendToExternal) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	if err = contract.ValidateEthereumAddress(m.Dest); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid dest address: %s", err)
	}
	if !m.Amount.IsValid() || !m.Amount.IsPositive() {
		return errortypes.ErrInvalidRequest.Wrap("invalid amount")
	}
	if m.Amount.Denom != m.BridgeFee.Denom {
		return errortypes.ErrInvalidRequest.Wrap("bridge fee denom not equal amount denom")
	}
	if !m.BridgeFee.IsValid() || !m.BridgeFee.IsPositive() {
		return errortypes.ErrInvalidRequest.Wrap("invalid bridge fee")
	}
	return nil
}

func (b MsgValidate) MsgCancelSendToExternalValidate(m *MsgCancelSendToExternal) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	if m.TransactionId == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero transaction id")
	}
	return nil
}

func (b MsgValidate) MsgIncreaseBridgeFeeValidate(m *MsgIncreaseBridgeFee) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	if m.TransactionId == 0 {
		return errortypes.ErrInvalidRequest.Wrap("zero transaction id")
	}
	if !m.AddBridgeFee.IsValid() || !m.AddBridgeFee.IsPositive() {
		return errortypes.ErrInvalidRequest.Wrap("invalid bridge fee")
	}
	return nil
}

func (b MsgValidate) MsgRequestBatchValidate(m *MsgRequestBatch) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	if len(m.Denom) == 0 {
		return errortypes.ErrInvalidRequest.Wrap("empty denom")
	}
	if m.MinimumFee.IsNil() || !m.MinimumFee.IsPositive() {
		return errortypes.ErrInvalidRequest.Wrap("invalid minimum fee")
	}
	if err = contract.ValidateEthereumAddress(m.FeeReceive); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid fee receive address: %s", err)
	}
	if m.BaseFee.IsNil() || m.BaseFee.IsNegative() {
		return errortypes.ErrInvalidRequest.Wrap("invalid base fee")
	}
	return nil
}

func (b MsgValidate) MsgConfirmBatchValidate(m *MsgConfirmBatch) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if err = contract.ValidateEthereumAddress(m.ExternalAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid external address: %s", err)
	}
	if err = contract.ValidateEthereumAddress(m.TokenContract); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid token contract: %s", err)
	}
	if len(m.Signature) == 0 {
		return errortypes.ErrInvalidRequest.Wrap("empty signature")
	}
	if _, err = hex.DecodeString(m.Signature); err != nil {
		return errortypes.ErrInvalidRequest.Wrap("could not hex decode signature")
	}
	return nil
}

func (b MsgValidate) MsgConfirmRefundValidate(m *MsgConfirmRefund) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if err = contract.ValidateEthereumAddress(m.ExternalAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid external address: %s", err)
	}
	if len(m.Signature) == 0 {
		return errortypes.ErrInvalidRequest.Wrap("empty signature")
	}
	if _, err = hex.DecodeString(m.Signature); err != nil {
		return errortypes.ErrInvalidRequest.Wrap("could not hex decode signature")
	}
	return nil
}

func (b MsgValidate) ValidateExternalAddress(addr string) error {
	return contract.ValidateEthereumAddress(addr)
}

func (b MsgValidate) ExternalAddressToAccAddress(addr string) (sdk.AccAddress, error) {
	if err := contract.ValidateEthereumAddress(addr); err != nil {
		return nil, err
	}
	return common.HexToAddress(addr).Bytes(), nil
}
