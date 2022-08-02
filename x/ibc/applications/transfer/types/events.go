package types

// IBC transfer events
const (
	EventTypeReceive      = "ibc_receive"
	EventTypeReceiveRoute = "ibc_receive_route"

	AttributeKeyAckError     = "error"
	AttributeKeyRouteSuccess = "success"
	AttributeKeyRoute        = "route"
	AttributeKeyRouteError   = "error"
)
