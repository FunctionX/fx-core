package polygon

import (
	"encoding/json"
	"fmt"

	"github.com/functionx/fx-core/v2/x/crosschain"
	crosschaintypes "github.com/functionx/fx-core/v2/x/crosschain/types"
	types "github.com/functionx/fx-core/v2/x/polygon/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	crosschainkeeper "github.com/functionx/fx-core/v2/x/crosschain/keeper"
	crosschainv020 "github.com/functionx/fx-core/v2/x/crosschain/legacy/v020"
)

// type check to ensure the interface is properly implemented
var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic object for module implementation
type AppModuleBasic struct{}

// Name implements app module basic
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec implements app module basic
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

// DefaultGenesis implements app module basic
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis implements app module basic
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, data json.RawMessage) error {
	var state crosschaintypes.GenesisState
	if err := cdc.UnmarshalJSON(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return state.ValidateBasic()
}

// RegisterRESTRoutes implements app module basic
func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {}

// GetQueryCmd implements app module basic
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// GetTxCmd implements app module basic
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the distribution module.
// also implements app modeul basic
func (AppModuleBasic) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}

// RegisterInterfaces implements app bmodule basic
func (AppModuleBasic) RegisterInterfaces(_ codectypes.InterfaceRegistry) {}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule object for module implementation
type AppModule struct {
	AppModuleBasic
	keeper        crosschainkeeper.Keeper
	stakingKeeper crosschainv020.StakingKeeper
	bankKeeper    crosschainv020.BankKeeper
	paramsKey     sdk.StoreKey
	legacyAmino   *codec.LegacyAmino
}

// NewAppModule creates a new AppModule Object
func NewAppModule(keeper crosschainkeeper.Keeper, stakingKeeper crosschainv020.StakingKeeper, bankKeeper crosschainv020.BankKeeper, legacyAmino *codec.LegacyAmino, paramsKey sdk.StoreKey) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
		stakingKeeper:  stakingKeeper,
		bankKeeper:     bankKeeper,
		paramsKey:      paramsKey,
		legacyAmino:    legacyAmino,
	}
}

// RegisterInvariants implements app module
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route implements app module
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// QuerierRoute implements app module
func (am AppModule) QuerierRoute() string { return types.QuerierRoute }

// LegacyQuerierHandler returns the distribution module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return crosschainkeeper.NewQuerier(am.keeper, legacyQuerierCdc)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	m := crosschainkeeper.NewMigrator(am.keeper, am.stakingKeeper, am.bankKeeper, am.legacyAmino, am.paramsKey)
	if err := cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(err)
	}
}

// InitGenesis initializes the genesis state for this module and implements app module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState crosschaintypes.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	crosschain.InitGenesis(ctx, am.keeper, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports the current genesis state to a json.RawMessage
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	state := crosschain.ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(state)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (am AppModule) ConsensusVersion() uint64 {
	return 2
}

// BeginBlock implements app module
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock implements app module
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	crosschain.EndBlocker(ctx, am.keeper)
	return []abci.ValidatorUpdate{}
}
