package fxcore

import (
	"io"
	"os"
	"path/filepath"

	ibcclienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/spf13/cast"
	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/functionx/fx-core/x/crosschain"
	crosschaintypes "github.com/functionx/fx-core/x/crosschain/types"
	"github.com/functionx/fx-core/x/ibc/applications/transfer"
	ibctransferkeeper "github.com/functionx/fx-core/x/ibc/applications/transfer/keeper"
	ibctransfertypes "github.com/functionx/fx-core/x/ibc/applications/transfer/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ibc "github.com/cosmos/cosmos-sdk/x/ibc/core"
	ibcclient "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client"
	porttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/05-port/types"
	ibchost "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	ibckeeper "github.com/cosmos/cosmos-sdk/x/ibc/core/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	// this line is used by starport scaffolding # stargate/app/moduleImport

	"github.com/functionx/fx-core/app"

	"github.com/functionx/fx-core/x/gravity"
	gravitykeeper "github.com/functionx/fx-core/x/gravity/keeper"
	gravitytypes "github.com/functionx/fx-core/x/gravity/types"

	"github.com/functionx/fx-core/x/bsc"
	bsctypes "github.com/functionx/fx-core/x/bsc/types"

	"github.com/functionx/fx-core/x/polygon"
	polygontypes "github.com/functionx/fx-core/x/polygon/types"

	crosschainkeeper "github.com/functionx/fx-core/x/crosschain/keeper"
	"github.com/functionx/fx-core/x/other"

	"github.com/functionx/fx-core/x/tron"
	tronkeeper "github.com/functionx/fx-core/x/tron/keeper"
	trontypes "github.com/functionx/fx-core/x/tron/types"
)

var ChainID = "fxcore"

const Name = "fxcore"
const MintDenom = "FX"
const AddressPrefix = "fx"

func getGovProposalHandlers() []govclient.ProposalHandler {
	return []govclient.ProposalHandler{
		paramsclient.ProposalHandler,
		distrclient.ProposalHandler,
		upgradeclient.ProposalHandler,
		upgradeclient.CancelProposalHandler,
	}
}

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(getGovProposalHandlers()...),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		ibc.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
		vesting.AppModuleBasic{},
		// this line is used by starport scaffolding # stargate/app/moduleBasic
		gravity.AppModuleBasic{},
		other.AppModuleBasic{},
		crosschain.AppModuleBasic{},
		bsc.AppModuleBasic{},
		polygon.AppModuleBasic{},
		tron.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		gravitytypes.ModuleName:        {authtypes.Minter, authtypes.Burner},
		bsctypes.ModuleName:            {authtypes.Minter, authtypes.Burner},
		polygontypes.ModuleName:        {authtypes.Minter, authtypes.Burner},
		trontypes.ModuleName:           {authtypes.Minter, authtypes.Burner},
	}
)

var (
	_ CosmosApp               = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
)

func init() {
	fxHome := os.ExpandEnv("$FXHOME")
	if len(fxHome) > 0 {
		DefaultNodeHome = fxHome
		return
	}
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+Name)
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	cdc               *codec.LegacyAmino
	appCodec          codec.Marshaler
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*sdk.KVStoreKey
	tkeys   map[string]*sdk.TransientStoreKey
	memKeys map[string]*sdk.MemoryStoreKey

	// keepers
	AccountKeeper    authkeeper.AccountKeeper
	BankKeeper       bankkeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    stakingkeeper.Keeper
	SlashingKeeper   slashingkeeper.Keeper
	MintKeeper       mintkeeper.Keeper
	DistrKeeper      distrkeeper.Keeper
	GovKeeper        govkeeper.Keeper
	CrisisKeeper     crisiskeeper.Keeper
	UpgradeKeeper    upgradekeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper
	IBCKeeper        *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	EvidenceKeeper   evidencekeeper.Keeper
	TransferKeeper   ibctransferkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper

	// this line is used by starport scaffolding # stargate/app/keeperDeclaration
	GravityKeeper    gravitykeeper.Keeper
	CrosschainKeeper crosschainkeeper.RouterKeeper
	BscKeeper        crosschainkeeper.Keeper
	PolygonKeeper    crosschainkeeper.Keeper
	TronKeeper       crosschainkeeper.Keeper

	// the module manager
	mm *module.Manager
}

func New(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, skipUpgradeHeights map[int64]bool,
	homePath string, invCheckPeriod uint, encodingConfig app.EncodingConfig,
	appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp),
) *App {

	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey,
		// this line is used by starport scaffolding # stargate/myApp/storeKey
		gravitytypes.StoreKey,
		bsctypes.StoreKey,
		polygontypes.StoreKey,
		trontypes.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	myApp := &App{
		BaseApp:           bApp,
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	myApp.ParamsKeeper = initParamsKeeper(appCodec, cdc, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	// set the BaseApp's parameter store
	bApp.SetParamStore(myApp.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))

	// add capability keeper and ScopeToModule for ibc module
	myApp.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := myApp.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedTransferKeeper := myApp.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)

	// add keepers
	myApp.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec, keys[authtypes.StoreKey], myApp.GetSubspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
	)
	myApp.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec, keys[banktypes.StoreKey], myApp.AccountKeeper, myApp.GetSubspace(banktypes.ModuleName), myApp.ModuleAccountAddrs(),
	)
	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec, keys[stakingtypes.StoreKey], myApp.AccountKeeper, myApp.BankKeeper, myApp.GetSubspace(stakingtypes.ModuleName),
	)
	myApp.MintKeeper = mintkeeper.NewKeeper(
		appCodec, keys[minttypes.StoreKey], myApp.GetSubspace(minttypes.ModuleName), &stakingKeeper,
		myApp.AccountKeeper, myApp.BankKeeper, authtypes.FeeCollectorName,
	)
	myApp.DistrKeeper = distrkeeper.NewKeeper(
		appCodec, keys[distrtypes.StoreKey], myApp.GetSubspace(distrtypes.ModuleName), myApp.AccountKeeper, myApp.BankKeeper,
		&stakingKeeper, authtypes.FeeCollectorName, myApp.ModuleAccountAddrs(),
	)
	myApp.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, keys[slashingtypes.StoreKey], &stakingKeeper, myApp.GetSubspace(slashingtypes.ModuleName),
	)
	myApp.CrisisKeeper = crisiskeeper.NewKeeper(
		myApp.GetSubspace(crisistypes.ModuleName), invCheckPeriod, myApp.BankKeeper, authtypes.FeeCollectorName,
	)
	myApp.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, keys[upgradetypes.StoreKey], appCodec, homePath)

	// Create IBC Keeper
	myApp.IBCKeeper = ibckeeper.NewKeeper(
		appCodec, keys[ibchost.StoreKey], myApp.GetSubspace(ibchost.ModuleName), stakingKeeper, scopedIBCKeeper,
	)

	// Create Transfer Keepers
	myApp.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec, keys[ibctransfertypes.StoreKey], myApp.GetSubspace(ibctransfertypes.ModuleName),
		myApp.IBCKeeper.ChannelKeeper, &myApp.IBCKeeper.PortKeeper,
		myApp.AccountKeeper, myApp.BankKeeper, scopedTransferKeeper,
	)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, keys[evidencetypes.StoreKey], stakingKeeper, myApp.SlashingKeeper,
	)
	// If evidence needs to be handled for the myApp, set routes in router here and seal
	myApp.EvidenceKeeper = *evidenceKeeper

	// this line is used by starport scaffolding # stargate/myApp/keeperDefinition

	myApp.GravityKeeper = gravitykeeper.NewKeeper(
		appCodec,
		keys[gravitytypes.StoreKey],
		myApp.GetSubspace(gravitytypes.ModuleName),
		stakingKeeper,
		myApp.BankKeeper,
		myApp.AccountKeeper,
		myApp.SlashingKeeper,
		myApp.TransferKeeper,
		myApp.IBCKeeper.ChannelKeeper,
	)

	// ## init cross chain module

	// ### init bsc cross chain module
	myApp.BscKeeper = crosschainkeeper.NewKeeper(appCodec, bsctypes.ModuleName, keys[bsctypes.StoreKey], myApp.GetSubspace(bsctypes.ModuleName),
		myApp.BankKeeper, myApp.AccountKeeper, myApp.TransferKeeper, myApp.IBCKeeper.ChannelKeeper)

	myApp.PolygonKeeper = crosschainkeeper.NewKeeper(appCodec, polygontypes.ModuleName, keys[polygontypes.StoreKey], myApp.GetSubspace(polygontypes.ModuleName),
		myApp.BankKeeper, myApp.AccountKeeper, myApp.TransferKeeper, myApp.IBCKeeper.ChannelKeeper)

	myApp.TronKeeper = crosschainkeeper.NewKeeper(appCodec, trontypes.ModuleName, keys[trontypes.StoreKey], myApp.GetSubspace(trontypes.ModuleName),
		myApp.BankKeeper, myApp.AccountKeeper, myApp.TransferKeeper, myApp.IBCKeeper.ChannelKeeper)

	crosschainRouter := crosschainkeeper.NewRouter()
	//// add cross-chain router
	crosschainRouter.
		AddRoute(bsctypes.ModuleName, &crosschainkeeper.ModuleHandler{
			QueryServer: myApp.BscKeeper,
			MsgServer:   crosschainkeeper.NewMsgServerImpl(myApp.BscKeeper),
		}).
		AddRoute(polygontypes.ModuleName, &crosschainkeeper.ModuleHandler{
			QueryServer: myApp.PolygonKeeper,
			MsgServer:   crosschainkeeper.NewMsgServerImpl(myApp.PolygonKeeper),
		}).
		AddRoute(trontypes.ModuleName, &crosschainkeeper.ModuleHandler{
			QueryServer: myApp.TronKeeper,
			MsgServer:   tronkeeper.NewMsgServerImpl(myApp.TronKeeper),
		})

	myApp.CrosschainKeeper = crosschainkeeper.NewRouterKeeper(crosschainRouter)

	// register the proposal types
	govRouter := govtypes.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(myApp.ParamsKeeper)).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(myApp.DistrKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(myApp.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientUpdateProposalHandler(myApp.IBCKeeper.ClientKeeper)).
		AddRoute(crosschaintypes.RouterKey, crosschain.NewCrossChainProposalHandler(myApp.CrosschainKeeper))

	myApp.GovKeeper = govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey], myApp.GetSubspace(govtypes.ModuleName), myApp.AccountKeeper, myApp.BankKeeper,
		&stakingKeeper, govRouter,
	)

	ibcTransferRouter := ibctransfertypes.NewRouter()
	ibcTransferRouter.AddRoute(gravitytypes.ModuleName, gravity.NewAppModule(myApp.GravityKeeper, myApp.BankKeeper))
	ibcTransferRouter.AddRoute(bsctypes.ModuleName, bsc.NewAppModule(myApp.BscKeeper, myApp.BankKeeper))
	ibcTransferRouter.AddRoute(polygontypes.ModuleName, polygon.NewAppModule(myApp.PolygonKeeper, myApp.BankKeeper))
	ibcTransferRouter.AddRoute(trontypes.ModuleName, tron.NewAppModule(myApp.TronKeeper, myApp.BankKeeper))
	myApp.TransferKeeper.SetRouter(ibcTransferRouter)
	transferModule := transfer.NewAppModule(myApp.TransferKeeper)
	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferModule)
	// this line is used by starport scaffolding # ibc/myApp/router
	myApp.IBCKeeper.SetRouter(ibcRouter)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	myApp.StakingKeeper = *stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			myApp.DistrKeeper.Hooks(),
			myApp.SlashingKeeper.Hooks(),
			myApp.GravityKeeper.Hooks(),
		),
	)

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	var skipGenesisInvariants = cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	myApp.mm = module.NewManager(
		genutil.NewAppModule(
			myApp.AccountKeeper, myApp.StakingKeeper, myApp.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, myApp.AccountKeeper, nil),
		vesting.NewAppModule(myApp.AccountKeeper, myApp.BankKeeper),
		bank.NewAppModule(appCodec, myApp.BankKeeper, myApp.AccountKeeper),
		capability.NewAppModule(appCodec, *myApp.CapabilityKeeper),
		crisis.NewAppModule(&myApp.CrisisKeeper, skipGenesisInvariants),
		gov.NewAppModule(appCodec, myApp.GovKeeper, myApp.AccountKeeper, myApp.BankKeeper),
		mint.NewAppModule(appCodec, myApp.MintKeeper, myApp.AccountKeeper),
		slashing.NewAppModule(appCodec, myApp.SlashingKeeper, myApp.AccountKeeper, myApp.BankKeeper, myApp.StakingKeeper),
		distr.NewAppModule(appCodec, myApp.DistrKeeper, myApp.AccountKeeper, myApp.BankKeeper, myApp.StakingKeeper),
		staking.NewAppModule(appCodec, myApp.StakingKeeper, myApp.AccountKeeper, myApp.BankKeeper),
		upgrade.NewAppModule(myApp.UpgradeKeeper),
		evidence.NewAppModule(myApp.EvidenceKeeper),
		ibc.NewAppModule(myApp.IBCKeeper),
		params.NewAppModule(myApp.ParamsKeeper),
		transferModule,
		// this line is used by starport scaffolding # stargate/myApp/appModule
		gravity.NewAppModule(myApp.GravityKeeper, myApp.BankKeeper),
		other.NewAppModule(appCodec),
		crosschain.NewAppModuleByRouter(myApp.CrosschainKeeper),
		bsc.NewAppModule(myApp.BscKeeper, myApp.BankKeeper),
		polygon.NewAppModule(myApp.PolygonKeeper, myApp.BankKeeper),
		tron.NewAppModule(myApp.TronKeeper, myApp.BankKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	myApp.mm.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		ibchost.ModuleName,
	)

	myApp.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		gravitytypes.ModuleName,
		bsctypes.ModuleName,
		polygontypes.ModuleName,
		trontypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	myApp.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		gravitytypes.ModuleName,
		crisistypes.ModuleName,
		ibchost.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		ibctransfertypes.ModuleName,
		// this line is used by starport scaffolding # stargate/myApp/initGenesis
	)

	myApp.mm.RegisterInvariants(&myApp.CrisisKeeper)
	myApp.mm.RegisterRoutes(myApp.Router(), myApp.QueryRouter(), encodingConfig.Amino)
	myApp.mm.RegisterServices(module.NewConfigurator(myApp.MsgServiceRouter(), myApp.GRPCQueryRouter()))

	// initialize stores
	myApp.MountKVStores(keys)
	myApp.MountTransientStores(tkeys)
	myApp.MountMemoryStores(memKeys)

	// initialize BaseApp
	myApp.SetInitChainer(myApp.InitChainer)
	myApp.SetBeginBlocker(myApp.BeginBlocker)
	myApp.SetAnteHandler(
		app.NewAnteHandler(
			myApp.AccountKeeper, myApp.BankKeeper, ante.DefaultSigVerificationGasConsumer,
			encodingConfig.TxConfig.SignModeHandler(),
		),
	)
	myApp.SetEndBlocker(myApp.EndBlocker)

	rootmulti.AddIgnoreCommitKey(app.CrossChainSupportBscBlock(), bsctypes.StoreKey)
	rootmulti.AddIgnoreCommitKey(app.CrossChainSupportPolygonBlock(), polygontypes.StoreKey)
	rootmulti.AddIgnoreCommitKey(app.CrossChainSupportTronBlock(), trontypes.StoreKey)

	if loadLatest {
		if err := myApp.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}

		// Initialize and seal the capability keeper so all persistent capabilities
		// are loaded in-memory and prevent any further modules from creating scoped
		// sub-keepers.
		// This must be done during creation of baseapp rather than in InitChain so
		// that in-memory capabilities get regenerated on myApp restart.
		// Note that since this reads from the store, we can only perform it when
		// `loadLatest` is set to true.
		ctx := myApp.BaseApp.NewUncachedContext(true, tmproto.Header{})
		myApp.CapabilityKeeper.InitializeAndSeal(ctx)
	}

	myApp.ScopedIBCKeeper = scopedIBCKeeper
	myApp.ScopedTransferKeeper = scopedTransferKeeper
	return myApp
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState AppGenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Marshaler {
	return app.appCodec
}

// InterfaceRegistry returns InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *sdk.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *sdk.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *sdk.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	rpc.RegisterRoutes(clientCtx, apiSvr.Router)
	// Register legacy tx routes.
//	authrest.RegisterTxRoutes(clientCtx, apiSvr.Router)
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
//	ModuleBasics.RegisterRESTRoutes(clientCtx, apiSvr.Router)
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.interfaceRegistry)
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryMarshaler, legacyAmino *codec.LegacyAmino, key, tkey sdk.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypes.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	// this line is used by starport scaffolding # stargate/app/paramSubspace
	paramsKeeper.Subspace(gravitytypes.ModuleName)
	paramsKeeper.Subspace(bsctypes.ModuleName)
	paramsKeeper.Subspace(polygontypes.ModuleName)
	paramsKeeper.Subspace(trontypes.ModuleName)

	return paramsKeeper
}
