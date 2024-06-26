package precompile

import (
	"errors"
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	fxcontract "github.com/functionx/fx-core/v7/contract"
	fxtypes "github.com/functionx/fx-core/v7/types"
	crosschaintypes "github.com/functionx/fx-core/v7/x/crosschain/types"
	evmtypes "github.com/functionx/fx-core/v7/x/evm/types"
)

type AddPendingPoolRewardsMethod struct {
	*Keeper
	abi.Method
	abi.Event
}

func NewAddPendingPoolRewardsMethod(keeper *Keeper) *AddPendingPoolRewardsMethod {
	return &AddPendingPoolRewardsMethod{
		Keeper: keeper,
		Method: crosschaintypes.GetABI().Methods["addPendingPoolRewards"],
		Event:  crosschaintypes.GetABI().Events["AddPendingPoolRewardsEvent"],
	}
}

func (m *AddPendingPoolRewardsMethod) IsReadonly() bool {
	return false
}

func (m *AddPendingPoolRewardsMethod) GetMethodId() []byte {
	return m.Method.ID
}

func (m *AddPendingPoolRewardsMethod) RequiredGas() uint64 {
	return 40_000
}

func (m *AddPendingPoolRewardsMethod) Run(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract) ([]byte, error) {
	if m.router == nil {
		return nil, errors.New("cross chain router empty")
	}

	args, err := m.UnpackInput(contract.Input)
	if err != nil {
		return nil, err
	}

	fxTarget := fxtypes.ParseFxTarget(args.Chain)
	route, has := m.router.GetRoute(fxTarget.GetTarget())
	if !has {
		return nil, fmt.Errorf("chain not support: %s", args.Chain)
	}

	value := contract.Value()
	sender := contract.Caller()
	totalCoin := sdk.Coin{}
	if value.Cmp(big.NewInt(0)) == 1 && fxcontract.IsZeroEthAddress(args.Token) {
		if args.Reward.Cmp(value) != 0 {
			return nil, errors.New("add bridge fee not equal msg.value")
		}
		totalCoin, err = m.handlerOriginToken(ctx, evm, sender, args.Reward)
		if err != nil {
			return nil, err
		}
	} else {
		totalCoin, err = m.handlerERC20Token(ctx, evm, sender, args.Token, args.Reward)
		if err != nil {
			return nil, err
		}
	}

	// convert token to bridge reward token
	rewardCoin := sdk.NewCoin(totalCoin.Denom, sdkmath.NewIntFromBigInt(args.Reward))
	addReward, err := m.erc20Keeper.ConvertDenomToTarget(ctx, sender.Bytes(), rewardCoin, fxTarget)
	if err != nil {
		return nil, err
	}

	if err = route.PrecompileAddPendingPoolRewards(ctx, args.TxID.Uint64(), sender.Bytes(), addReward); err != nil {
		return nil, err
	}

	data, topic, err := m.NewAddPendingPoolRewardsEvent(sender, args.Token, args.Chain, args.TxID, args.Reward)
	if err != nil {
		return nil, err
	}
	EmitEvent(evm, data, topic)

	return m.PackOutput(true)
}

func (m *AddPendingPoolRewardsMethod) NewAddPendingPoolRewardsEvent(sender common.Address, token common.Address, chain string, txId, reward *big.Int) (data []byte, topic []common.Hash, err error) {
	data, topic, err = evmtypes.PackTopicData(m.Event, []common.Hash{sender.Hash(), token.Hash()}, chain, txId, reward)
	if err != nil {
		return nil, nil, err
	}
	return data, topic, nil
}

func (m *AddPendingPoolRewardsMethod) PackInput(args crosschaintypes.AddPendingPoolRewardArgs) ([]byte, error) {
	data, err := m.Method.Inputs.Pack(args.Chain, args.TxID, args.Token, args.Reward)
	if err != nil {
		return nil, err
	}
	return append(m.GetMethodId(), data...), nil
}

func (m *AddPendingPoolRewardsMethod) UnpackInput(data []byte) (*crosschaintypes.AddPendingPoolRewardArgs, error) {
	args := new(crosschaintypes.AddPendingPoolRewardArgs)
	if err := evmtypes.ParseMethodArgs(m.Method, args, data[4:]); err != nil {
		return nil, err
	}
	return args, nil
}

func (m *AddPendingPoolRewardsMethod) PackOutput(success bool) ([]byte, error) {
	pack, err := m.Method.Outputs.Pack(success)
	if err != nil {
		return nil, err
	}
	return pack, nil
}
