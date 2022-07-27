package types

const (
	// ModuleName defines the IBC transfer name
	ModuleName = "transfer"

	//CompatibleModuleName is the query and tx module name
	CompatibleModuleName = "fxtransfer"

	// RouterKey is the message route for IBC transfer
	RouterKey = CompatibleModuleName

	// QuerierRoute is the querier route for IBC transfer
	QuerierRoute = CompatibleModuleName
)
