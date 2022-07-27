package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
)

// IBC transfer sentinel errors
var (
	ErrInvalidPacketTimeout       = ibctransfertypes.ErrInvalidPacketTimeout
	ErrInvalidDenomForTransfer    = ibctransfertypes.ErrInvalidDenomForTransfer
	ErrInvalidVersion             = ibctransfertypes.ErrInvalidVersion
	ErrInvalidAmount              = ibctransfertypes.ErrInvalidAmount
	ErrTraceNotFound              = ibctransfertypes.ErrTraceNotFound
	ErrSendDisabled               = ibctransfertypes.ErrSendDisabled
	ErrReceiveDisabled            = ibctransfertypes.ErrReceiveDisabled
	ErrMaxTransferChannels        = ibctransfertypes.ErrMaxTransferChannels
	ErrFeeDenomNotMatchTokenDenom = sdkerrors.Register(ModuleName, 10, "invalid fee denom, must match token denom")
)
