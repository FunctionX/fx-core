package keeper

import (
	"context"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	"github.com/functionx/fx-core/v2/x/ibc/applications/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ types.MsgServer   = Keeper{}
	_ types.QueryServer = Keeper{}
)

// See createOutgoingPacket in spec:https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay

// Transfer defines a rpc handler method for MsgTransfer.
func (k Keeper) Transfer(goCtx context.Context, msg *types.MsgTransfer) (*types.MsgTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	if err = k.SendFxTransfer(
		ctx, msg.SourcePort, msg.SourceChannel, msg.Token, sender, msg.Receiver, msg.TimeoutHeight, msg.TimeoutTimestamp, msg.Router, sdk.NewCoin(msg.Token.Denom, msg.Fee.Amount),
	); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("IBC fungible token transfer", "token", msg.Token.Denom, "amount", msg.Token.Amount.String(), "sender", msg.Sender, "receiver", msg.Receiver)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			transfertypes.EventTypeTransfer,
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(transfertypes.AttributeKeyReceiver, msg.Receiver),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	})

	return &types.MsgTransferResponse{}, nil
}
