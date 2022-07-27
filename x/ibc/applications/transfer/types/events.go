package types

// IBC transfer events
const (
	//EventTypeTimeout      = "timeout"
	//EventTypePacket       = "fungible_token_packet"
	//EventTypeTransfer     = "ibc_transfer"
	EventTypeReceive      = "ibc_receive"
	EventTypeReceiveRoute = "ibc_receive_route"

	//AttributeKeyReceiver       = "receiver"
	//AttributeKeyDenom          = "denom"
	//AttributeKeyAmount         = "amount"
	//AttributeKeyRefundReceiver = "refund_receiver"
	//AttributeKeyRefundDenom    = "refund_denom"
	//AttributeKeyRefundAmount   = "refund_amount"
	//AttributeKeyAckSuccess     = "success"
	//AttributeKeyAck            = "acknowledgement"
	AttributeKeyAckError     = "error"
	AttributeKeyRouteSuccess = "success"
	AttributeKeyRouteError   = "error"
)
