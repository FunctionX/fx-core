package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/functionx/fx-core/v7/contract"
	crosschaintypes "github.com/functionx/fx-core/v7/x/crosschain/types"
	"github.com/functionx/fx-core/v7/x/erc20/types"
)

// Keeper of this module maintains collections of erc20.
type Keeper struct {
	storeKey          storetypes.StoreKey
	cdc               codec.BinaryCodec
	accountKeeper     types.AccountKeeper
	bankKeeper        types.BankKeeper
	evmKeeper         types.EVMKeeper
	ibcTransferKeeper types.IBCTransferKeeper

	moduleAddress common.Address

	authority  string
	chainsName []string
}

// NewKeeper creates new instances of the erc20 Keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	evmKeeper types.EVMKeeper,
	ibcTransferKeeper types.IBCTransferKeeper,
	authority string,
) Keeper {
	moduleAddress := ak.GetModuleAddress(types.ModuleName)
	if moduleAddress == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	return Keeper{
		storeKey:          storeKey,
		cdc:               cdc,
		accountKeeper:     ak,
		bankKeeper:        bk,
		evmKeeper:         evmKeeper,
		ibcTransferKeeper: ibcTransferKeeper,
		moduleAddress:     common.BytesToAddress(moduleAddress),
		authority:         authority,
		chainsName:        crosschaintypes.GetSupportChains(),
	}
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// ModuleAddress return erc20 module address
func (k Keeper) ModuleAddress() common.Address {
	return k.moduleAddress
}

// TransferAfter ibc transfer after
func (k Keeper) TransferAfter(ctx sdk.Context, sender sdk.AccAddress, receive string, coin, fee sdk.Coin, _, _ bool) error {
	if err := contract.ValidateEthereumAddress(receive); err != nil {
		return errortypes.ErrInvalidAddress.Wrapf("invalid receive address: %s", err.Error())
	}
	_, err := k.ConvertCoin(sdk.WrapSDKContext(ctx), &types.MsgConvertCoin{
		Coin:     coin.Add(fee),
		Receiver: receive,
		Sender:   sender.String(),
	})
	return err
}

func (k Keeper) HasDenomAlias(ctx sdk.Context, denom string) (banktypes.Metadata, bool) {
	md, found := k.bankKeeper.GetDenomMetaData(ctx, denom)
	// not register metadata
	if !found {
		return banktypes.Metadata{}, false
	}
	// not have denom units
	if len(md.DenomUnits) == 0 {
		return banktypes.Metadata{}, false
	}
	// not have alias
	if len(md.DenomUnits[0].Aliases) == 0 {
		return banktypes.Metadata{}, false
	}
	return md, true
}

func (k Keeper) GetValidMetadata(ctx sdk.Context, denom string) (banktypes.Metadata, bool) {
	md, found := k.bankKeeper.GetDenomMetaData(ctx, denom)
	// not register metadata
	if !found {
		return banktypes.Metadata{}, false
	}
	// not have denom units
	if len(md.DenomUnits) == 0 {
		return banktypes.Metadata{}, false
	}
	return md, true
}
