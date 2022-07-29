package types

import (
	"strings"
	"time"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// DefaultRelativePacketTimeoutTimestamp is the default packet timeout timestamp (in nanoseconds)
	// relative to the current block timestamp of the counterparty chain provided by the client
	// state. The timeout is disabled when set to 0. The default is currently set to a 12 hour
	// timeout.
	DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(12) * time.Hour).Nanoseconds())
)

// NewFungibleTokenPacketData contructs a new FungibleTokenPacketData instance
func NewFungibleTokenPacketData(denom, amount, sender, receiver, router string, fee string,
) FungibleTokenPacketData {
	return FungibleTokenPacketData{
		Denom:    denom,
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Router:   router,
		Fee:      fee,
	}
}

// ValidateBasic is used for validating the token transfer.
// NOTE: The addresses formats are not validated as the sender and recipient can have different
// formats defined by their corresponding chains that are not known to IBC.
func (ftpd FungibleTokenPacketData) ValidateBasic() error {
	amount, ok := sdk.NewIntFromString(ftpd.Amount)
	if !ok {
		return sdkerrors.Wrapf(transfertypes.ErrInvalidAmount, "unable to parse transfer amount (%s) into sdk.Int", ftpd.Amount)
	}
	if !amount.IsPositive() {
		return sdkerrors.Wrapf(transfertypes.ErrInvalidAmount, "amount must be strictly positive: got %d", amount)
	}
	if strings.TrimSpace(ftpd.Sender) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be blank")
	}
	if strings.TrimSpace(ftpd.Receiver) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "receiver address cannot be blank")
	}
	fee, ok := sdk.NewIntFromString(ftpd.Fee)
	if !ok {
		return sdkerrors.Wrapf(transfertypes.ErrInvalidAmount, "unable to parse transfer fee (%s) into sdk.Int", ftpd.Fee)
	}
	if fee.IsNegative() {
		return sdkerrors.Wrapf(transfertypes.ErrInvalidAmount, "fee must be strictly not negative: got %d", fee)
	}
	return transfertypes.ValidatePrefixedDenom(ftpd.Denom)
}

// GetBytes is a helper for serialising
func (ftpd FungibleTokenPacketData) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&ftpd))
}
