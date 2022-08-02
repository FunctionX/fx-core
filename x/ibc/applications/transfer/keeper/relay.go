package keeper

import (
	"fmt"
	"strings"

	coretypes "github.com/cosmos/ibc-go/v3/modules/core/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	"github.com/functionx/fx-core/v2/x/ibc/applications/transfer/types"

	"github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
)

// SendFxTransfer handles transfer sending logic. There are 2 possible cases:
//
// 1. Sender chain is acting as the source zone. The coins are transferred
// to an escrow address (i.e locked) on the sender chain and then transferred
// to the receiving chain through IBC TAO logic. It is expected that the
// receiving chain will mint vouchers to the receiving address.
//
// 2. Sender chain is acting as the sink zone. The coins (vouchers) are burned
// on the sender chain and then transferred to the receiving chain though IBC
// TAO logic. It is expected that the receiving chain, which had previously
// sent the original denomination, will unescrow the fungible token and send
// it to the receiving address.
//
// Another way of thinking of source and sink zones is through the token's
// timeline. Each send to any chain other than the one it was previously
// received from is a movement forwards in the token's timeline. This causes
// trace to be added to the token's history and the destination port and
// destination channel to be prefixed to the denomination. In these instances
// the sender chain is acting as the source zone. When the token is sent back
// to the chain it previously received from, the prefix is removed. This is
// a backwards movement in the token's timeline and the sender chain
// is acting as the sink zone.
//
// Example:
// These steps of transfer occur: A -> B -> C -> A -> C -> B -> A
//
// 1. A -> B : sender chain is source zone. Denom upon receiving: 'B/denom'
// 2. B -> C : sender chain is source zone. Denom upon receiving: 'C/B/denom'
// 3. C -> A : sender chain is source zone. Denom upon receiving: 'A/C/B/denom'
// 4. A -> C : sender chain is sink zone. Denom upon receiving: 'C/B/denom'
// 5. C -> B : sender chain is sink zone. Denom upon receiving: 'B/denom'
// 6. B -> A : sender chain is sink zone. Denom upon receiving: 'denom'
func (k Keeper) SendFxTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	token sdk.Coin,
	sender sdk.AccAddress,
	receiver string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	router string,
	fee sdk.Coin,
) error {

	if !k.GetSendEnabled(ctx) {
		return transfertypes.ErrSendDisabled
	}

	sourceChannelEnd, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", sourcePort, sourceChannel)
	}

	destinationPort := sourceChannelEnd.GetCounterparty().GetPortID()
	destinationChannel := sourceChannelEnd.GetCounterparty().GetChannelID()

	// get the next sequence
	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", sourcePort, sourceChannel,
		)
	}

	// begin createOutgoingPacket logic
	// See spec for this logic: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	// NOTE: denomination and hex hash correctness checked during msg.ValidateBasic
	fullDenomPath := token.Denom

	var err error

	// deconstruct the token denomination into the denomination trace info
	// to determine if the sender is the source chain
	if strings.HasPrefix(token.Denom, "ibc/") {
		fullDenomPath, err = k.DenomPathFromHash(ctx, token.Denom)
		if err != nil {
			return err
		}
	}

	labels := []metrics.Label{
		telemetry.NewLabel(coretypes.LabelDestinationPort, destinationPort),
		telemetry.NewLabel(coretypes.LabelDestinationChannel, destinationChannel),
	}

	packetData := types.NewFungibleTokenPacketData(
		fullDenomPath, token.Amount.String(), sender.String(), receiver, router, fee.Amount.String(),
	)

	// If the router address is specified, the number of token + fee is deducted
	if router != "" {
		token = token.Add(sdk.NewCoin(token.Denom, fee.Amount))
	}
	// NOTE: SendTransfer simply sends the denomination as it exists on its own
	// chain inside the packet data. The receiving chain will perform denom
	// prefixing as necessary.
	if transfertypes.SenderChainIsSource(sourcePort, sourceChannel, fullDenomPath) {
		labels = append(labels, telemetry.NewLabel(coretypes.LabelSource, "true"))

		// create the escrow address for the tokens
		escrowAddress := transfertypes.GetEscrowAddress(sourcePort, sourceChannel)

		// escrow source tokens. It fails if balance insufficient.
		if err = k.bankKeeper.SendCoins(
			ctx, sender, escrowAddress, sdk.NewCoins(token),
		); err != nil {
			return err
		}

	} else {
		labels = append(labels, telemetry.NewLabel(coretypes.LabelSource, "false"))

		// transfer the coins to the module account and burn them
		if err = k.bankKeeper.SendCoinsFromAccountToModule(
			ctx, sender, transfertypes.ModuleName, sdk.NewCoins(token),
		); err != nil {
			return err
		}

		if err = k.bankKeeper.BurnCoins(
			ctx, transfertypes.ModuleName, sdk.NewCoins(token),
		); err != nil {
			// NOTE: should not happen as the module account was
			// retrieved on the step above and it has enough balace
			// to burn.
			panic(fmt.Sprintf("cannot burn coins after a successful send to a module account: %v", err))
		}
	}

	packet := channeltypes.NewPacket(
		packetData.GetBytes(),
		sequence,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		timeoutHeight,
		timeoutTimestamp,
	)

	if err = k.ics4Wrapper.SendPacket(ctx, channelCap, packet); err != nil {
		return err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"ibc", transfertypes.ModuleName, "send"},
			1,
			labels,
		)
	}()

	return nil
}

// FxOnRecvPacket processes a cross chain fungible token transfer. If the
// sender chain is the source of minted tokens then vouchers will be minted
// and sent to the receiving address. Otherwise if the sender chain is sending
// back tokens this chain originally transferred to it, the tokens are
// unescrowed and sent to the receiving address.
func (k Keeper) FxOnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, data types.FungibleTokenPacketData) error {
	// validate packet data upon receiving
	if err := data.ValidateBasic(); err != nil {
		return err
	}

	receiver, transferAmount, feeAmount, err := parseReceiveAndAmountByPacket(data)
	if err != nil {
		return err
	}

	receiveAmount := transferAmount.Add(feeAmount)
	if err = k.OnRecvPacket(ctx, packet, transfertypes.NewFungibleTokenPacketData(data.GetDenom(), receiveAmount.String(), data.GetSender(), receiver.String())); err != nil {
		return err
	}

	receiveDenom := parseIBCCoinDenom(packet, data.GetDenom())

	receiveCoin := sdk.NewCoin(receiveDenom, receiveAmount)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReceive,
			sdk.NewAttribute(transfertypes.AttributeKeyReceiver, receiver.String()),
			sdk.NewAttribute(transfertypes.AttributeKeyAmount, receiveCoin.String()),
		),
	)

	if data.Router == "" || !k.Router.HasRoute(data.Router) {
		return nil
	}
	if route, exists := k.Router.GetRoute(data.Router); exists {
		ibcAmount := sdk.NewCoin(receiveDenom, transferAmount)
		ibcFee := sdk.NewCoin(receiveDenom, feeAmount)
		ctx.Logger().Info("IBCTransfer", "transfer route sourceChannel", packet.GetSourceChannel(),
			"destChannel", packet.GetDestChannel(), "sequence", packet.GetSequence(), "sender", receiver.String(),
			"receive", data.Receiver, "amount", ibcAmount, "fee", ibcFee, "router", data.Router)
		cacheCtx, writeFn := ctx.CacheContext()
		err = route.TransferAfter(cacheCtx, receiver.String(), data.Receiver, ibcAmount, ibcFee)
		routerEvent := sdk.NewEvent(types.EventTypeReceiveRoute,
			sdk.NewAttribute(types.AttributeKeyRouteSuccess, fmt.Sprintf("%t", err == nil)),
		)
		switch err {
		case nil:
			writeFn()
			ctx.EventManager().EmitEvents(cacheCtx.EventManager().Events())
		default:
			ctx.Logger().Error("IBCTransfer", "transfer after route err!!!sourceChannel", packet.GetSourceChannel(), "destChannel", packet.GetDestChannel(), "sequence", packet.GetSequence(), "err", err)
			routerEvent = routerEvent.AppendAttributes(sdk.NewAttribute(types.AttributeKeyRouteError, err.Error()))
		}
		ctx.EventManager().EmitEvent(routerEvent)

		return nil
	}
	return nil
}

// OnAcknowledgementPacket responds to the the success or failure of a packet
// acknowledgement written on the receiving chain. If the acknowledgement
// was a success then nothing occurs. If the acknowledgement failed, then
// the sender is refunded their tokens using the refundPacketToken function.
func (k Keeper) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, data types.FungibleTokenPacketData, ack channeltypes.Acknowledgement) error {
	switch ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:
		return k.refundPacketToken(ctx, packet, data)
	default:
		// the acknowledgement succeeded on the receiving chain so nothing
		// needs to be executed and no error needs to be returned
		return nil
	}
}

// OnTimeoutPacket refunds the sender since the original packet sent was
// never received and has been timed out.
func (k Keeper) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, data types.FungibleTokenPacketData) error {
	return k.refundPacketToken(ctx, packet, data)
}

// refundPacketToken will unescrow and send back the tokens back to sender
// if the sending chain was the source chain. Otherwise, the sent tokens
// were burnt in the original send so new tokens are minted and sent to
// the sending address.
func (k Keeper) refundPacketToken(ctx sdk.Context, packet channeltypes.Packet, data types.FungibleTokenPacketData) error {
	// NOTE: packet data type already checked in handler.go

	// parse the denomination from the full denom path
	trace := transfertypes.ParseDenomTrace(data.Denom)

	amount, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		return sdkerrors.Wrapf(transfertypes.ErrInvalidAmount, "unable to parse transfer amount (%s) into sdk.Int", data.Amount)
	}
	// If the IBC router is not empty, the feeAmount refund is added.
	if data.Router != "" {
		feeAmount, ok := sdk.NewIntFromString(data.Fee)
		if !ok {
			return sdkerrors.Wrapf(transfertypes.ErrInvalidAmount, "unable to parse transfer fee (%s) into sdk.Int", data.Amount)
		}
		amount = amount.Add(feeAmount)
	}
	token := sdk.NewCoin(trace.IBCDenom(), amount)

	// decode the sender address
	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		return err
	}

	if transfertypes.SenderChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) {
		// unescrow tokens back to sender
		escrowAddress := transfertypes.GetEscrowAddress(packet.GetSourcePort(), packet.GetSourceChannel())
		if err = k.bankKeeper.SendCoins(ctx, escrowAddress, sender, sdk.NewCoins(token)); err != nil {
			// NOTE: this error is only expected to occur given an unexpected bug or a malicious
			// counterparty module. The bug may occur in bank or any part of the code that allows
			// the escrow address to be drained. A malicious counterparty module could drain the
			// escrow address by allowing more tokens to be sent back then were escrowed.
			return sdkerrors.Wrap(err, "unable to unescrow tokens, this may be caused by a malicious counterparty module or a bug: please open an issue on counterparty module")
		}
	} else {
		// mint vouchers back to sender
		if err = k.bankKeeper.MintCoins(
			ctx, transfertypes.ModuleName, sdk.NewCoins(token),
		); err != nil {
			return err
		}

		if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, transfertypes.ModuleName, sender, sdk.NewCoins(token)); err != nil {
			panic(fmt.Sprintf("unable to send coins from module to account despite previously minting coins to module account: %v", err))
		}
	}

	if k.RefundHook != nil {
		ctx.Logger().Info("ibc refund hook", "sourcePort", packet.SourcePort, "sourceChannel",
			packet.SourceChannel, "sequence", fmt.Sprintf("%d", packet.Sequence), "sender", sender.String(), "token", token.String())
		if err = k.RefundHook.RefundAfter(ctx, packet.SourcePort, packet.SourceChannel, packet.Sequence, sender, data.Receiver, token); err != nil {
			ctx.Logger().Error("refundPacketToken", "refund hook err!!!sourceChannel", packet.GetSourceChannel(), "destChannel", packet.GetDestChannel(), "sequence", packet.GetSequence(), "err", err)
		}
	}
	return nil
}
