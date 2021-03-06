package keeper

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/libs/log"

	fxtypes "github.com/functionx/fx-core/v2/types"

	"github.com/functionx/fx-core/v2/x/erc20/types"

	"github.com/evmos/ethermint/x/evm/statedb"
)

// Keeper of this module maintains collections of erc20.
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramstore paramtypes.Subspace

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	evmKeeper     types.EVMKeeper
	// fetch EIP1559 base fee and parameters

	ibcTransferKeeper types.IBCTransferKeeper
	ibcChannelKeeper  types.IBCChannelKeeper

	Router *types.Router
}

// NewKeeper creates new instances of the erc20 Keeper
func NewKeeper(
	storeKey sdk.StoreKey,
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	evmKeeper types.EVMKeeper,
	ibcTransferKeeper types.IBCTransferKeeper,
	ibcChannelKeeper types.IBCChannelKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:          storeKey,
		cdc:               cdc,
		paramstore:        ps,
		accountKeeper:     ak,
		bankKeeper:        bk,
		evmKeeper:         evmKeeper,
		ibcTransferKeeper: ibcTransferKeeper,
		ibcChannelKeeper:  ibcChannelKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) RefundAfter(ctx sdk.Context, sourcePort, sourceChannel string, sequence uint64, sender sdk.AccAddress, receiver string, amount sdk.Coin) error {
	//check tx
	if !k.HashIBCTransferHash(ctx, sourcePort, sourceChannel, sequence) {
		ctx.Logger().Info("ignore refund, transaction not belong to evm ibc transfer", "module", types.ModuleName)
		return nil
	}
	return k.RelayConvertCoin(ctx, sender, common.BytesToAddress(sender.Bytes()), amount)
}

func (k Keeper) TransferAfter(ctx sdk.Context, sender, receive string, coin, fee sdk.Coin) error {
	sendAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return fmt.Errorf("invalid sender address %s, error %s", sender, err.Error())
	}
	if err = fxtypes.ValidateEthereumAddress(receive); err != nil {
		return fmt.Errorf("invalid receiver address %s", err.Error())
	}
	return k.RelayConvertCoin(ctx, sendAddr, common.HexToAddress(receive), coin.Add(fee))
}

func (k Keeper) RelayConvertCoin(ctx sdk.Context, sender sdk.AccAddress, receiver common.Address, coin sdk.Coin) error {
	if !k.IsDenomRegistered(ctx, coin.Denom) {
		return fmt.Errorf("denom(%s) not registered", coin.Denom)
	}
	msg := &types.MsgConvertCoin{
		Coin:     coin,
		Receiver: receiver.Hex(),
		Sender:   sender.String(),
	}
	_, err := k.ConvertCoin(sdk.WrapSDKContext(ctx), msg)
	return err
}

func (k Keeper) RelayConvertDenomToOne(ctx sdk.Context, from sdk.AccAddress, coin sdk.Coin) (sdk.Coin, error) {
	return k.convertDenomToOne(ctx, from, coin)
}
func (k Keeper) RelayConvertDenomToMany(ctx sdk.Context, from sdk.AccAddress, coin sdk.Coin, target string) (sdk.Coin, error) {
	return k.convertDenomToMany(ctx, from, coin, target)
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

func (k *Keeper) SetIBCTransferKeeperForTest(t types.IBCTransferKeeper) {
	k.ibcTransferKeeper = t
}

func (k *Keeper) SetIBCChannelKeeperForTest(t types.IBCChannelKeeper) {
	k.ibcChannelKeeper = t
}

func (k Keeper) CreateContractWithCode(ctx sdk.Context, addr common.Address, code []byte) error {
	k.Logger(ctx).Debug("create contract with code", "address", addr.String(), "code", hex.EncodeToString(code[:32])+"..."+hex.EncodeToString(code[len(code)-32:]))
	codeHash := crypto.Keccak256Hash(code)
	acc := k.evmKeeper.GetAccount(ctx, addr)
	if acc == nil {
		k.Logger(ctx).Info("create contract with code", "address", addr.String(), "action", "create")
		acc = statedb.NewEmptyAccount()
		acc.CodeHash = codeHash.Bytes()
		k.evmKeeper.SetCode(ctx, acc.CodeHash, code)
		return k.evmKeeper.SetAccount(ctx, addr, *acc)
	}
	k.Logger(ctx).Info("create contract with code", "address", addr.String(), "action", "update")
	acc.CodeHash = codeHash.Bytes()
	k.evmKeeper.SetCode(ctx, acc.CodeHash, code)
	return k.evmKeeper.SetAccount(ctx, addr, *acc)
}

func (k Keeper) IsManyToOneDenom(ctx sdk.Context, denom string) (bool, error) {
	md, found := k.bankKeeper.GetDenomMetaData(ctx, denom)
	if !found {
		return false, fmt.Errorf("denom %s not found", denom)
	}
	return types.IsManyToOneMetadata(md), nil
}
