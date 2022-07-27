package keeper

import (
	ibctransferkeeper "github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/functionx/fx-core/v2/x/ibc/applications/transfer/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
)

// Keeper defines the IBC fungible transfer keeper
type Keeper struct {
	ibctransferkeeper.Keeper
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace

	ics4Wrapper   porttypes.ICS4Wrapper
	channelKeeper types.ChannelKeeper
	portKeeper    types.PortKeeper
	authKeeper    types.AccountKeeper
	bankKeeper    types.BankKeeper
	scopedKeeper  capabilitykeeper.ScopedKeeper
	Router        *types.Router
	RefundHook    types.RefundHook
}

// NewKeeper creates a new IBC transfer Keeper instance
func NewKeeper(keeper ibctransferkeeper.Keeper,
	cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	ics4Wrapper porttypes.ICS4Wrapper, channelKeeper types.ChannelKeeper, portKeeper types.PortKeeper,
	authKeeper types.AccountKeeper, bankKeeper types.BankKeeper, scopedKeeper capabilitykeeper.ScopedKeeper,
) Keeper {

	// ensure ibc transfer module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the FX IBC transfer module account has not been set")
	}

	// set KeyTable if it has not already been set
	//if !paramSpace.HasKeyTable() {
	//	paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	//}

	return Keeper{
		Keeper:        keeper,
		cdc:           cdc,
		storeKey:      key,
		paramSpace:    paramSpace,
		ics4Wrapper:   ics4Wrapper,
		channelKeeper: channelKeeper,
		portKeeper:    portKeeper,
		authKeeper:    authKeeper,
		bankKeeper:    bankKeeper,
		scopedKeeper:  scopedKeeper,
	}
}

// SetRouter sets the Router in IBC Transfer Keeper and seals it. The method panics if
// there is an existing router that's already sealed.
func (k *Keeper) SetRouter(rtr *types.Router) {
	if k.Router != nil && k.Router.Sealed() {
		panic("cannot reset a sealed router")
	}
	k.Router = rtr
	k.Router.Seal()
}

func (k Keeper) GetRouter() *types.Router {
	return k.Router
}

func (k *Keeper) SetRefundHook(hook types.RefundHook) {
	k.RefundHook = hook
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+host.ModuleName+"-"+types.ModuleName)
}
