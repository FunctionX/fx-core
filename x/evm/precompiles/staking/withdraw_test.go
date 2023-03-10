package staking_test

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	fxtypes "github.com/functionx/fx-core/v3/types"
	"github.com/functionx/fx-core/v3/x/evm/precompiles/staking"
)

func (suite *PrecompileTestSuite) TestWithdraw() {
	testCases := []struct {
		name     string
		malleate func(val sdk.ValAddress, shares sdk.Dec) (*evmtypes.MsgEthereumTx, []string)
		error    func(errArgs []string) string
		result   bool
	}{
		{
			name: "ok",
			malleate: func(val sdk.ValAddress, shares sdk.Dec) (*evmtypes.MsgEthereumTx, []string) {
				pack, err := fxtypes.MustABIJson(StakingTestABI).Pack(staking.WithdrawMethodName, val.String())
				suite.Require().NoError(err)
				return suite.PackEthereumTx(suite.signer, suite.precompileStaking, big.NewInt(0), pack), nil
			},
			result: true,
		},
		{
			name: "failed invalid validator address",
			malleate: func(val sdk.ValAddress, shares sdk.Dec) (*evmtypes.MsgEthereumTx, []string) {
				newVal := val.String() + "1"
				pack, err := fxtypes.MustABIJson(StakingTestABI).Pack(staking.WithdrawMethodName, newVal)
				suite.Require().NoError(err)
				return suite.PackEthereumTx(suite.signer, suite.precompileStaking, big.NewInt(0), pack), []string{newVal}
			},
			error: func(errArgs []string) string {
				return fmt.Sprintf("withdraw failed: invalid validator address: %s", errArgs[0])
			},
			result: false,
		},
		{
			name: "failed validator not found",
			malleate: func(val sdk.ValAddress, shares sdk.Dec) (*evmtypes.MsgEthereumTx, []string) {
				newVal := sdk.ValAddress(suite.signer.Address().Bytes()).String()
				pack, err := fxtypes.MustABIJson(StakingTestABI).Pack(staking.WithdrawMethodName, newVal)
				suite.Require().NoError(err)
				return suite.PackEthereumTx(suite.signer, suite.precompileStaking, big.NewInt(0), pack), []string{newVal}
			},
			error: func(errArgs []string) string {
				return "withdraw failed: no validator distribution info"
			},
			result: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			vals := suite.app.StakingKeeper.GetValidators(suite.ctx, 10)
			val0 := vals[0]

			delAmt := sdkmath.NewInt(1000).Mul(sdkmath.NewInt(1e18))
			pack, err := fxtypes.MustABIJson(StakingTestABI).Pack(staking.DelegateMethodName, val0.GetOperator().String(), delAmt.BigInt())
			suite.Require().NoError(err)
			delegateEthTx := suite.PackEthereumTx(suite.signer, suite.precompileStaking, delAmt.BigInt(), pack)
			res, err := suite.app.EvmKeeper.EthereumTx(sdk.WrapSDKContext(suite.ctx), delegateEthTx)
			suite.Require().NoError(err)
			suite.Require().False(res.Failed(), res.VmError)

			delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, suite.precompileStaking.Bytes(), val0.GetOperator())
			suite.Require().True(found)

			suite.Commit()

			ethTx, errArgs := tc.malleate(val0.GetOperator(), delegation.Shares)
			res, err = suite.app.EvmKeeper.EthereumTx(sdk.WrapSDKContext(suite.ctx), ethTx)

			if tc.result {
				suite.Require().NoError(err)
				suite.Require().False(res.Failed(), res.VmError)
			} else {
				suite.Require().True(err != nil || res.Failed())
				if err != nil {
					suite.Require().Equal(tc.error(errArgs), err.Error())
				}
				if res.Failed() {
					if res.VmError != vm.ErrExecutionReverted.Error() {
						suite.Require().Equal(tc.error(errArgs), res.VmError)
					} else {
						if len(res.Ret) > 0 {
							reason, err := abi.UnpackRevert(common.CopyBytes(res.Ret))
							suite.Require().NoError(err)

							suite.Require().Equal(tc.error(errArgs), reason)
						} else {
							suite.Require().Equal(tc.error(errArgs), vm.ErrExecutionReverted.Error())
						}
					}
				} else {
					suite.Require().Equal(tc.error(errArgs), err.Error())
				}
			}
		})
	}
}
