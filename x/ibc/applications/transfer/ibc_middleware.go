package transfer

import (
	"fmt"
	"strings"

	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	"github.com/functionx/fx-core/v2/x/ibc"

	fxtypes "github.com/functionx/fx-core/v2/types"

	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"time"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	coretypes "github.com/cosmos/ibc-go/v3/modules/core/types"

	"github.com/functionx/fx-core/v2/x/ibc/applications/transfer/keeper"
	"github.com/functionx/fx-core/v2/x/ibc/applications/transfer/types"
)

const (
	ForwardPacketTimeHour time.Duration = 12
)

var _ porttypes.Middleware = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 interface for transfer given the transfer keeper.
type IBCMiddleware struct {
	*ibc.Module
	keeper keeper.Keeper
}

func (im IBCMiddleware) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI) error {
	return im.keeper.SendTransfer(ctx, chanCap, packet)
}

func (im IBCMiddleware) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	//TODO implement me
	return im.keeper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application
func NewIBCMiddleware(k keeper.Keeper, app porttypes.IBCModule) IBCMiddleware {
	return IBCMiddleware{
		Module: ibc.NewModule(app),
		keeper: k,
	}
}

// OnRecvPacket implements the IBCModule interface
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})

	var data types.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		var ibcPacketData ibctransfertypes.FungibleTokenPacketData
		if err = types.ModuleCdc.UnmarshalJSON(packet.GetData(), &ibcPacketData); err != nil {
			ack = channeltypes.NewErrorAcknowledgement("cannot unmarshal ICS-20 transfer packet data")
		} else {
			data = types.FungibleTokenPacketData{
				Denom:    ibcPacketData.GetDenom(),
				Amount:   ibcPacketData.GetAmount(),
				Sender:   ibcPacketData.GetSender(),
				Receiver: ibcPacketData.GetReceiver(),
				Router:   "",
				Fee:      sdk.ZeroInt().String(),
			}
		}
	}

	// only attempt the application logic if the packet data
	// was successfully decoded
	if ack.Success() {
		if len(data.GetFee()) == 0 {
			data.Fee = sdk.ZeroInt().String()
		}
		var err error
		// if router set, route packet
		if ctx.BlockHeight() >= fxtypes.IBCRouteBlock() && data.Router != "" {
			err = im.keeper.OnRecvPacket(ctx, packet, data)
		} else {
			err = handlerForwardTransferPacket(ctx, im, packet, data)
		}

		if err != nil {
			ack = types.NewErrorAcknowledgement(err)
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePacket,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyReceiver, data.Receiver),
			sdk.NewAttribute(types.AttributeKeyDenom, data.Denom),
			sdk.NewAttribute(types.AttributeKeyAmount, data.Amount),
			sdk.NewAttribute(types.AttributeKeyAckSuccess, fmt.Sprintf("%t", ack.Success())),
		),
	)

	// NOTE: acknowledgement will be written synchronously during IBC handler execution.
	return ack
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := types.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}
	var data types.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	if err := im.keeper.OnAcknowledgementPacket(ctx, packet, data, ack); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePacket,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyReceiver, data.Receiver),
			sdk.NewAttribute(types.AttributeKeyDenom, data.Denom),
			sdk.NewAttribute(types.AttributeKeyAmount, data.Amount),
			sdk.NewAttribute(types.AttributeKeyAck, ack.String()),
		),
	)

	switch resp := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypePacket,
				sdk.NewAttribute(types.AttributeKeyAckSuccess, string(resp.Result)),
			),
		)
	case *channeltypes.Acknowledgement_Error:
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypePacket,
				sdk.NewAttribute(types.AttributeKeyAckError, resp.Error),
			),
		)
	}

	return nil
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	var data types.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}
	// refund tokens
	if err := im.keeper.OnTimeoutPacket(ctx, packet, data); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTimeout,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyRefundReceiver, data.Sender),
			sdk.NewAttribute(types.AttributeKeyRefundDenom, data.Denom),
			sdk.NewAttribute(types.AttributeKeyRefundAmount, data.Amount),
		),
	)

	return nil
}

func handlerForwardTransferPacket(ctx sdk.Context, im IBCMiddleware, packet channeltypes.Packet, data types.FungibleTokenPacketData) error {
	// parse out any forwarding info
	receiver, finalDest, port, channel, err := ParseIncomingTransferField(data.Receiver)
	switch {

	// if this isn't a packet to forward, just use the transfer module normally
	case finalDest == "" && port == "" && channel == "" && err == nil:
		return im.keeper.OnRecvPacket(ctx, packet, data)

	// If the parsing fails return a failure ack
	case err != nil:
		return fmt.Errorf("cannot parse packet fowrading information")

	// Otherwise we have a packet to forward
	default:
		// Modify packet data to process packet transfer for this chain, omitting forwarding info
		newData := data
		newData.Receiver = receiver.String()
		bz, err := types.ModuleCdc.MarshalJSON(&newData)
		if err != nil {
			return err
		}
		newPacket := packet
		newPacket.Data = bz

		if err := im.keeper.OnRecvPacket(ctx, newPacket, newData); err != nil {
			return err
		}
		// recalculate denom, skip checks that were already done in app.OnRecvPacket
		denom := GetDenomByIBCPacket(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetDestPort(), packet.GetDestChannel(), newData.GetDenom())
		// parse the transfer amount
		transferAmount, ok := sdk.NewIntFromString(data.Amount)
		if !ok {
			return sdkerrors.Wrapf(types.ErrInvalidAmount, "unable to parse forward transfer amount (%s) into sdk.Int", data.Amount)
		}

		var token = sdk.NewCoin(denom, transferAmount)

		if err = im.keeper.SendFxTransfer(ctx, port, channel, token, receiver, finalDest,
			clienttypes.Height{}, uint64(ctx.BlockTime().Add(ForwardPacketTimeHour*time.Hour).UnixNano()),
			"", sdk.NewCoin(denom, sdk.ZeroInt())); err != nil {
			return fmt.Errorf("failed to forward transfer packet")
		}
		defer func() {
			telemetry.IncrCounterWithLabels(
				[]string{"ibc", types.ModuleName, "packet", "forward"},
				1,
				append(
					[]metrics.Label{
						telemetry.NewLabel(coretypes.LabelSourcePort, packet.GetSourcePort()),
						telemetry.NewLabel(coretypes.LabelSourceChannel, packet.GetSourceChannel()),
					},
					telemetry.NewLabel(coretypes.LabelSource, "true"),
				),
			)
		}()
	}
	return nil
}

func GetDenomByIBCPacket(sourcePort, sourceChannel, destPort, destChannel, packetDenom string) string {
	var denom string

	if types.ReceiverChainIsSource(sourcePort, sourceChannel, packetDenom) {
		voucherPrefix := types.GetDenomPrefix(sourcePort, sourceChannel)
		unPrefixedDenom := packetDenom[len(voucherPrefix):]

		// coin denomination used in sending from the escrow address
		denom = unPrefixedDenom

		// The denomination used to send the coins is either the native denom or the hash of the path
		// if the denomination is not native.
		denomTrace := types.ParseDenomTrace(unPrefixedDenom)
		if denomTrace.Path != "" {
			denom = denomTrace.IBCDenom()
		}
	} else {
		// since SendPacket did not prefix the denomination, we must prefix denomination here
		sourcePrefix := types.GetDenomPrefix(destPort, destChannel)
		// NOTE: sourcePrefix contains the trailing "/"
		prefixedDenom := sourcePrefix + packetDenom

		// construct the denomination trace from the full raw denomination
		denom = types.ParseDenomTrace(prefixedDenom).IBCDenom()
	}
	return denom
}

// ParseIncomingTransferField For now this assumes one hop, should be better parsing
func ParseIncomingTransferField(receiverData string) (thisChainAddr sdk.AccAddress, finalDestination, port, channel string, err error) {
	sep1 := strings.Split(receiverData, ":")
	switch {
	case len(sep1) == 1 && sep1[0] != "":
		thisChainAddr, err = sdk.AccAddressFromBech32(receiverData)
		return
	case len(sep1) >= 2 && sep1[len(sep1)-1] != "":
		finalDestination = strings.Join(sep1[1:], ":")
	default:
		return nil, "", "", "", fmt.Errorf("unparsable receiver field, need: '{address_on_this_chain}|{portid}/{channelid}:{final_dest_address}', got: '%s'", receiverData)
	}
	sep2 := strings.Split(sep1[0], "|")
	if len(sep2) != 2 {
		err = fmt.Errorf("formatting incorrect, need: '{address_on_this_chain}|{portid}/{channelid}:{final_dest_address}', got: '%s'", receiverData)
		return
	}
	thisChainAddr, err = sdk.AccAddressFromBech32(sep2[0])
	if err != nil {
		return
	}
	sep3 := strings.Split(sep2[1], "/")
	if len(sep3) != 2 {
		err = fmt.Errorf("formatting incorrect, need: '{address_on_this_chain}|{portid}/{channelid}:{final_dest_address}', got: '%s'", receiverData)
		return
	}
	port = sep3[0]
	channel = sep3[1]
	return
}
