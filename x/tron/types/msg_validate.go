package types

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	tronaddress "github.com/fbsobreira/gotron-sdk/pkg/address"

	crosschaintypes "github.com/functionx/fx-core/v7/x/crosschain/types"
)

var _ crosschaintypes.MsgValidateBasic = &TronMsgValidate{}

type TronMsgValidate struct {
	crosschaintypes.MsgValidate
}

func (b TronMsgValidate) MsgBondedOracleValidate(m *crosschaintypes.MsgBondedOracle) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.OracleAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid oracle address: %s", err)
	}
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if err = ValidateTronAddress(m.ExternalAddress); err != nil {
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

func (b TronMsgValidate) MsgOracleSetConfirmValidate(m *crosschaintypes.MsgOracleSetConfirm) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if err = ValidateTronAddress(m.ExternalAddress); err != nil {
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

func (b TronMsgValidate) MsgOracleSetUpdatedClaimValidate(m *crosschaintypes.MsgOracleSetUpdatedClaim) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if len(m.Members) == 0 {
		return errortypes.ErrInvalidRequest.Wrap("empty members")
	}
	for _, member := range m.Members {
		if err = ValidateTronAddress(member.ExternalAddress); err != nil {
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

func (b TronMsgValidate) MsgBridgeTokenClaimValidate(m *crosschaintypes.MsgBridgeTokenClaim) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if err = ValidateTronAddress(m.TokenContract); err != nil {
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
	return crosschaintypes.ValidateBridgeTokenType(m.TokenType)
}

func (b TronMsgValidate) MsgSendToExternalClaimValidate(m *crosschaintypes.MsgSendToExternalClaim) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if err = ValidateTronAddress(m.TokenContract); err != nil {
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

func (b TronMsgValidate) MsgSendToFxClaimValidate(m *crosschaintypes.MsgSendToFxClaim) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if err = ValidateTronAddress(m.Sender); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	if err = ValidateTronAddress(m.TokenContract); err != nil {
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

func (b TronMsgValidate) MsgBridgeCallClaimValidate(m *crosschaintypes.MsgBridgeCallClaim) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if m.DstChainId == "" {
		return errortypes.ErrInvalidRequest.Wrap("empty dst chain id")
	}
	if err = ValidateTronAddress(m.Sender); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	if len(m.To) > 0 {
		if err = ValidateTronAddress(m.To); err != nil {
			return errortypes.ErrInvalidAddress.Wrapf("invalid to contract: %s", err)
		}
	}
	if err = ValidateTronAddress(m.Receiver); err != nil {
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

func (b TronMsgValidate) MsgSendToExternalValidate(m *crosschaintypes.MsgSendToExternal) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	if err = ValidateTronAddress(m.Dest); err != nil {
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

func (b TronMsgValidate) MsgRequestBatchValidate(m *crosschaintypes.MsgRequestBatch) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	if len(m.Denom) == 0 {
		return errortypes.ErrInvalidRequest.Wrap("empty denom")
	}
	if m.MinimumFee.IsNil() || !m.MinimumFee.IsPositive() {
		return errortypes.ErrInvalidRequest.Wrap("invalid minimum fee")
	}
	if err = ValidateTronAddress(m.FeeReceive); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid fee receive address: %s", err)
	}
	if m.BaseFee.IsNil() || m.BaseFee.IsNegative() {
		return errortypes.ErrInvalidRequest.Wrap("invalid base fee")
	}
	return nil
}

func (b TronMsgValidate) MsgConfirmBatchValidate(m *crosschaintypes.MsgConfirmBatch) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if err = ValidateTronAddress(m.ExternalAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid external address: %s", err)
	}
	if err = ValidateTronAddress(m.TokenContract); err != nil {
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

func (b TronMsgValidate) MsgConfirmRefundValidate(m *crosschaintypes.MsgConfirmRefund) (err error) {
	if _, err = sdk.AccAddressFromBech32(m.BridgerAddress); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid bridger address: %s", err)
	}
	if err = ValidateTronAddress(m.ExternalAddress); err != nil {
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

func (b TronMsgValidate) ValidateExternalAddress(addr string) error {
	return ValidateTronAddress(addr)
}

func (b TronMsgValidate) ExternalAddressToAccAddress(addr string) (sdk.AccAddress, error) {
	tronAddr, err := tronaddress.Base58ToAddress(addr)
	if err != nil {
		return nil, err
	}
	return tronAddr.Bytes()[1:], nil
}
