package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmos "github.com/cometbft/cometbft/libs/os"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	sigtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/gogoproto/proto"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/ethereum/go-ethereum/eth/tracers/js"
	_ "github.com/ethereum/go-ethereum/eth/tracers/native"
	evmante "github.com/evmos/ethermint/app/ante"
	ethermintcodec "github.com/evmos/ethermint/crypto/codec"
	srvflags "github.com/evmos/ethermint/server/flags"
	etherminttypes "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"

	fxante "github.com/functionx/fx-core/v8/ante"
	"github.com/functionx/fx-core/v8/app/keepers"
	_ "github.com/functionx/fx-core/v8/docs/statik"
	fxcfg "github.com/functionx/fx-core/v8/server/config"
	fxauth "github.com/functionx/fx-core/v8/server/grpc/auth"
	gaspricev1 "github.com/functionx/fx-core/v8/server/grpc/gasprice/legacy/v1"
	gaspricev2 "github.com/functionx/fx-core/v8/server/grpc/gasprice/legacy/v2"
	fxtypes "github.com/functionx/fx-core/v8/types"
	"github.com/functionx/fx-core/v8/x/crosschain"
	"github.com/functionx/fx-core/v8/x/crosschain/keeper"
	crosschaintypes "github.com/functionx/fx-core/v8/x/crosschain/types"
	fxevmtypes "github.com/functionx/fx-core/v8/x/evm/types"
	ibcmiddlewaretypes "github.com/functionx/fx-core/v8/x/ibc/middleware/types"
)

var _ servertypes.Application = (*App)(nil)

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp
	keepers.AppKeepers

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	mm           *module.Manager
	ModuleBasics module.BasicManager

	configurator       module.Configurator
	pendingTxListeners []evmante.PendingTxListener
}

func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	legacyAmino := codec.NewLegacyAmino()
	interfaceRegistry := NewInterfaceRegistry()
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	ethermintcodec.RegisterInterfaces(interfaceRegistry)
	etherminttypes.RegisterInterfaces(interfaceRegistry)

	crosschaintypes.RegisterLegacyAminoCodec(legacyAmino)
	crosschaintypes.RegisterInterfaces(interfaceRegistry)

	fxevmtypes.RegisterCryptoEthSecp256k1(legacyAmino)

	ibcmiddlewaretypes.RegisterLegacyAminoCodec(legacyAmino)
	ibcmiddlewaretypes.RegisterInterfaces(interfaceRegistry)

	legacy.Cdc = legacyAmino

	// App Opts
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))
	invCheckPeriod := cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))

	bApp := baseapp.NewBaseApp(
		fxtypes.Name,
		logger,
		db,
		txConfig.TxDecoder(),
		baseAppOptions...)

	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	app := &App{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		txConfig:          txConfig,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
	}

	// Setup keepers
	app.AppKeepers = keepers.NewAppKeeper(
		appCodec,
		bApp,
		legacyAmino,
		maccPerms,
		BlockedAccountAddrs(),
		skipUpgradeHeights,
		homePath,
		invCheckPeriod,
		logger,
		appOpts,
	)

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(appModules(app, appCodec, txConfig, skipGenesisInvariants)...)
	app.ModuleBasics = newBasicManagerFromManager(app)

	enabledSignModes := append([]sigtypes.SignMode(nil), authtx.DefaultSignModes...)
	enabledSignModes = append(enabledSignModes, sigtypes.SignMode_SIGN_MODE_TEXTUAL)

	txConfigOpts := authtx.ConfigOptions{
		EnabledSignModes:           enabledSignModes,
		TextualCoinMetadataQueryFn: txmodule.NewBankKeeperCoinMetadataQueryFn(app.BankKeeper),
	}
	var err error
	txConfig, err = authtx.NewTxConfigWithOptions(
		appCodec,
		txConfigOpts,
	)
	if err != nil {
		panic(err)
	}
	app.txConfig = txConfig

	// NOTE: upgrade module is required to be prioritized
	app.mm.SetOrderPreBlockers(
		upgradetypes.ModuleName,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	// NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)
	app.mm.SetOrderBeginBlockers(orderBeginBlockers()...)

	app.mm.SetOrderEndBlockers(orderEndBlockers()...)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(orderInitBlockers()...)
	app.mm.SetOrderExportGenesis(orderInitBlockers()...)

	app.mm.RegisterInvariants(app.CrisisKeeper)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	if err = app.RegisterServices(app.configurator); err != nil {
		panic(err)
	}

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.mm.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// initialize stores
	app.MountKVStores(app.GetKVStoreKey())
	app.MountTransientStores(app.GetTransientStoreKey())
	app.MountMemoryStores(app.GetMemoryStoreKey())
	app.MountObjectStores(app.GetObjectStoreKey())

	app.setAnteHandler(appOpts)

	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	if manager := app.SnapshotManager(); manager != nil {
		err = manager.RegisterExtensions()
		if err != nil {
			panic("failed to register snapshot extension: " + err.Error())
		}
	}

	app.setupUpgradeHandlers()
	app.setupUpgradeStoreLoaders()

	app.setPostHandler()

	// At startup, after all modules have been registered, check that all prot
	// annotations are correct.
	protoFiles, err := proto.MergedRegistry()
	if err != nil {
		panic(err)
	}
	err = msgservice.ValidateProtoAnnotations(protoFiles)
	if err != nil {
		// Once we switch to using protoreflect-based antehandlers, we might
		// want to panic here instead of logging a warning.
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}

	if loadLatest {
		if err = app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	return app
}

func (app *App) setAnteHandler(appOpts servertypes.AppOptions) {
	maxGasWanted := cast.ToUint64(appOpts.Get(srvflags.EVMMaxTxGasWanted))
	BypassMinFeeMsgTypes := cast.ToStringSlice(appOpts.Get(fxcfg.BypassMinFeeMsgTypesKey))
	MaxBypassMinFeeMsgGasUsage := cast.ToUint64(appOpts.Get(fxcfg.BypassMinFeeMsgMaxGasUsageKey))
	anteOptions := fxante.HandlerOptions{
		AccountKeeper:   app.AccountKeeper,
		BankKeeper:      app.BankKeeper,
		EvmKeeper:       app.EvmKeeper,
		FeeMarketKeeper: app.FeeMarketKeeper,
		IbcKeeper:       app.IBCKeeper,
		GovKeeper:       app.GovKeeper,
		SignModeHandler: app.txConfig.SignModeHandler(),
		SigGasConsumer:  fxante.DefaultSigVerificationGasConsumer,
		MaxTxGasWanted:  maxGasWanted,
		TxFeeChecker:    fxante.NewCheckTxFeees(BypassMinFeeMsgTypes, MaxBypassMinFeeMsgGasUsage).Check,
		DisabledAuthzMsgs: []string{
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}),
		},
		PendingTxListener: app.onPendingTx,
	}

	if err := anteOptions.Validate(); err != nil {
		panic(fmt.Errorf("failed to ante options validate: %w", err))
	}

	app.SetAnteHandler(fxante.NewAnteHandler(anteOptions))
}

func (app *App) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}
	app.SetPostHandler(postHandler)
}

// Name returns the name of the App
func (app *App) Name() string {
	return app.BaseApp.Name()
}

// PreBlocker application updates every pre block
func (app *App) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.mm.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.mm.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.mm.EndBlock(ctx)
}

// Configurator returns the configurator of the app
func (app *App) Configurator() module.Configurator {
	return app.configurator
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	if err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap()); err != nil {
		panic(err)
	}

	response, err := app.mm.InitGenesis(ctx, app.appCodec, genesisState)
	if err != nil {
		panic(err)
	}

	return response, nil
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// InterfaceRegistry returns InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetTxConfig returns App's TxConfig
func (app *App) GetTxConfig() client.TxConfig {
	return app.txConfig
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (app *App) DefaultGenesis() map[string]json.RawMessage {
	return NewDefAppGenesisByDenom(app.appCodec, app.ModuleBasics)
}

// RegisterServices registers all module services
func (app *App) RegisterServices(cfg module.Configurator) error {
	if err := app.mm.RegisterServices(cfg); err != nil {
		return err
	}

	// Deprecated
	gaspricev1.RegisterQueryServer(cfg.QueryServer(), gaspricev1.Querier{})
	gaspricev2.RegisterQueryServer(cfg.QueryServer(), gaspricev2.Querier{})

	crosschaintypes.RegisterQueryServer(cfg.QueryServer(), app.CrosschainRouterKeeper)
	crosschaintypes.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerRouterImpl(app.CrosschainRouterKeeper))

	fxauth.RegisterQueryServer(cfg.QueryServer(), fxauth.Querier{})
	return nil
}

// RegisterAPIRoutes registers all application module routes with the provided API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	// Deprecated: Register gas price queries routes from grpc-gateway.
	gaspricev1.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	gaspricev2.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register node gRPC service for grpc-gateway.
	// query that exposes operator configuration, most notably the operator's configured minimum gas price
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register crosschain queries routes from grpc-gateway.
	crosschain.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register fx auth queries routes from grpc-gateway.
	fxauth.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	app.ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		statikFS, err := fs.New()
		if err != nil {
			panic(err)
		}
		staticServer := http.FileServer(statikFS)
		apiSvr.Router.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.interfaceRegistry, app.Query)
}

// RegisterNodeService registers the node gRPC service on the provided application gRPC query router.
func (app *App) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// RegisterPendingTxListener is used by json-rpc server to listen to pending transactions callback.
func (app *App) RegisterPendingTxListener(listener evmante.PendingTxListener) {
	app.pendingTxListeners = append(app.pendingTxListeners, listener)
}

func (app *App) onPendingTx(hash common.Hash) {
	for _, listener := range app.pendingTxListeners {
		listener(hash)
	}
}

func (app *App) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.mm.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.mm.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

// << only for test, do not use in production >>

// SimulationManager NOTE: This is solely to be used for testing purposes.
func (app *App) SimulationManager() *module.SimulationManager {
	return &module.SimulationManager{}
}

// LegacyAmino NOTE: This is solely to be used for testing purposes.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec NOTE: This is solely to be used for testing purposes.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// GetModules NOTE: This is solely to be used for testing purposes.
func (app *App) GetModules() map[string]module.AppModule {
	modules := make(map[string]module.AppModule, len(app.mm.Modules))
	for name, mod := range app.mm.Modules {
		modules[name] = mod.(module.AppModule)
	}
	return modules
}

// GetOrderBeginBlockersModules NOTE: This is solely to be used for testing purposes..
func (app *App) GetOrderBeginBlockersModules() []string {
	return app.mm.OrderBeginBlockers
}

// GetOrderEndBlockersModules NOTE: This is solely to be used for testing purposes.
func (app *App) GetOrderEndBlockersModules() []string {
	return app.mm.OrderEndBlockers
}

// GetOrderInitGenesisModules NOTE: This is solely to be used for testing purposes.
func (app *App) GetOrderInitGenesisModules() []string {
	return app.mm.OrderInitGenesis
}

// EmptyAppOptions is a stub implementing AppOptions
type EmptyAppOptions struct{}

// Get implements AppOptions
func (ao EmptyAppOptions) Get(string) interface{} {
	return nil
}
