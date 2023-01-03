package keeper

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	fxtypes "github.com/functionx/fx-core/v3/types"
	"github.com/functionx/fx-core/v3/x/erc20/types"
)

// QueryERC20 returns the data of a deployed ERC20 contract
func (k Keeper) QueryERC20(ctx sdk.Context, contract common.Address) (types.ERC20Data, error) {
	erc20 := fxtypes.GetERC20().ABI

	// Name
	res, err := k.CallEVM(ctx, erc20, types.ModuleAddress, contract, false, "name")
	if err != nil {
		return types.ERC20Data{}, err
	}
	var nameRes struct{ Value string }
	if err := erc20.UnpackIntoInterface(&nameRes, "name", res.Ret); err != nil {
		return types.ERC20Data{}, sdkerrors.Wrapf(types.ErrABIUnpack, "failed to unpack name: %s", err.Error())
	}

	// Symbol
	res, err = k.CallEVM(ctx, erc20, types.ModuleAddress, contract, false, "symbol")
	if err != nil {
		return types.ERC20Data{}, err
	}
	var symbolRes struct{ Value string }
	if err := erc20.UnpackIntoInterface(&symbolRes, "symbol", res.Ret); err != nil {
		return types.ERC20Data{}, sdkerrors.Wrapf(types.ErrABIUnpack, "failed to unpack symbol: %s", err.Error())
	}

	// Decimals
	res, err = k.CallEVM(ctx, erc20, types.ModuleAddress, contract, false, "decimals")
	if err != nil {
		return types.ERC20Data{}, err
	}
	var decimalRes struct{ Value uint8 }
	if err := erc20.UnpackIntoInterface(&decimalRes, "decimals", res.Ret); err != nil {
		return types.ERC20Data{}, sdkerrors.Wrapf(types.ErrABIUnpack, "failed to unpack decimals: %s", err.Error())
	}

	return types.NewERC20Data(nameRes.Value, symbolRes.Value, decimalRes.Value), nil
}

// BalanceOf returns the balance of an address for ERC20 contract
func (k Keeper) BalanceOf(ctx sdk.Context, contract, addr common.Address) (*big.Int, error) {
	erc20 := fxtypes.GetERC20().ABI

	res, err := k.CallEVM(ctx, erc20, types.ModuleAddress, contract, false, "balanceOf", addr)
	if err != nil {
		return nil, err
	}

	var balanceRes struct{ Value *big.Int }
	if err := erc20.UnpackIntoInterface(&balanceRes, "balanceOf", res.Ret); err != nil {
		return nil, err
	}
	return balanceRes.Value, nil
}

// CallEVM performs a smart contract method call using given args
func (k Keeper) CallEVM(
	ctx sdk.Context,
	abi abi.ABI,
	from, contract common.Address,
	commit bool,
	method string,
	args ...interface{},
) (*evmtypes.MsgEthereumTxResponse, error) {
	data, err := abi.Pack(method, args...)
	if err != nil {
		return nil, sdkerrors.Wrap(
			types.ErrABIPack,
			sdkerrors.Wrap(err, "failed to create transaction data").Error(),
		)
	}

	resp, err := k.evmKeeper.CallEVMWithData(ctx, from, &contract, data, commit)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "contract call failed: method '%s', contract '%s'", method, contract)
	}
	return resp, nil
}

// monitorApprovalEvent returns an error if the given transactions logs include
// an unexpected `approve` event
func (k Keeper) monitorApprovalEvent(res *evmtypes.MsgEthereumTxResponse) error {
	if res == nil || len(res.Logs) == 0 {
		return nil
	}

	logApprovalSig := []byte("Approval(address,address,uint256)")
	logApprovalSigHash := crypto.Keccak256Hash(logApprovalSig)

	for _, log := range res.Logs {
		if log.Topics[0] == logApprovalSigHash.Hex() {
			return sdkerrors.Wrapf(
				types.ErrUnexpectedEvent, "Approval event",
			)
		}
	}

	return nil
}
