package types

import (
	"fmt"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
)

const (
	TypeMsgTransfer = "transfer"
)

// NewMsgTransfer creates a new MsgTransfer instance
//nolint:interfacer
func NewMsgTransfer(
	sourcePort, sourceChannel string,
	token sdk.Coin, sender string, receiver string,
	timeoutHeight clienttypes.Height, timeoutTimestamp uint64, router string, fee sdk.Coin,
) *MsgTransfer {
	return &MsgTransfer{
		SourcePort:       sourcePort,
		SourceChannel:    sourceChannel,
		Token:            token,
		Sender:           sender,
		Receiver:         receiver,
		TimeoutHeight:    timeoutHeight,
		TimeoutTimestamp: timeoutTimestamp,
		Router:           router,
		Fee:              fee,
	}
}

// Route implements sdk.Msg
func (MsgTransfer) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgTransfer) Type() string {
	return TypeMsgTransfer
}

// ValidateBasic performs a basic check of the MsgTransfer fields.
// NOTE: timeout height or timestamp values can be 0 to disable the timeout.
// NOTE: The recipient addresses format is not validated as the format defined by
// the chain is not known to IBC.
func (msg MsgTransfer) ValidateBasic() error {
	if err := host.PortIdentifierValidator(msg.SourcePort); err != nil {
		return sdkerrors.Wrap(err, "invalid source port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.SourceChannel); err != nil {
		return sdkerrors.Wrap(err, "invalid source channel ID")
	}
	if !msg.Token.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Token.String())
	}
	if !msg.Token.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, msg.Token.String())
	}
	// NOTE: sender format must be validated as it is required by the GetSigners function.
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "string could not be parsed as address: %v", err)
	}
	if strings.TrimSpace(msg.Receiver) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing recipient address")
	}
	if msg.Fee.Amount.IsNil() || !msg.Fee.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "fees")
	}
	if msg.Fee.Denom != msg.Token.Denom {
		return sdkerrors.Wrap(ErrFeeDenomNotMatchTokenDenom, fmt.Sprintf("token denom:%s, fee denom:%s", msg.Token.Denom, msg.Fee.Denom))
	}
	return transfertypes.ValidateIBCDenom(msg.Token.Denom)
}

// GetSignBytes implements sdk.Msg.
func (msg MsgTransfer) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgTransfer) GetSigners() []sdk.AccAddress {
	valAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{valAddr}
}
