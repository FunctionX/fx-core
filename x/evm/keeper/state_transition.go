package keeper

import (
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	"github.com/evmos/ethermint/x/evm/statedb"
	"github.com/evmos/ethermint/x/evm/types"
)

// ApplyTransaction runs and attempts to perform a state transition with the given transaction (i.e Message), that will
// only be persisted (committed) to the underlying KVStore if the transaction does not fail.
//
// # Gas tracking
//
// Ethereum consumes gas according to the EVM opcodes instead of general reads and writes to store. Because of this, the
// state transition needs to ignore the SDK gas consumption mechanism defined by the GasKVStore and instead consume the
// amount of gas used by the VM execution. The amount of gas used is tracked by the EVM and returned in the execution
// result.
//
// Prior to the execution, the starting tx gas meter is saved and replaced with an infinite gas meter in a new context
// in order to ignore the SDK gas consumption config values (read, write, has, delete).
// After the execution, the gas used from the message execution will be added to the starting gas consumed, taking into
// consideration the amount of gas returned. Finally, the context is updated with the EVM gas consumed value prior to
// returning.
//
// For relevant discussion see: https://github.com/cosmos/cosmos-sdk/discussions/9072
func (k *Keeper) ApplyTransaction(ctx sdk.Context, tx *ethtypes.Transaction) (*types.MsgEthereumTxResponse, error) {
	var (
		bloom        *big.Int
		bloomReceipt ethtypes.Bloom
	)

	cfg, err := k.EVMConfig(ctx)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to load evm config")
	}
	txConfig := k.TxConfig(ctx, tx.Hash())

	// get the signer according to the chain rules from the config and block height
	signer := ethtypes.MakeSigner(cfg.ChainConfig, big.NewInt(ctx.BlockHeight()))
	msg, err := tx.AsMessage(signer, cfg.BaseFee)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to return ethereum transaction as core message")
	}

	// snapshot to contain the tx processing and post processing in same scope
	var commit func()
	tmpCtx := ctx
	if k.hasHooks {
		// Create a cache context to revert state when tx hooks fails,
		// the cache context is only committed when both tx and hooks executed successfully.
		// Didn't use `Snapshot` because the context stack has exponential complexity on certain operations,
		// thus restricted to be used only inside `ApplyMessage`.
		tmpCtx, commit = ctx.CacheContext()
	}

	// pass true to commit the StateDB
	res, err := k.ApplyMessageWithConfig(tmpCtx, msg, nil, true, cfg, txConfig)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to apply ethereum core message")
	}

	logs := types.LogsToEthereum(res.Logs)

	// Compute block bloom filter
	if len(logs) > 0 {
		bloom = k.GetBlockBloomTransient(ctx)
		bloom.Or(bloom, big.NewInt(0).SetBytes(ethtypes.LogsBloom(logs)))
		bloomReceipt = ethtypes.BytesToBloom(bloom.Bytes())
	}

	cumulativeGasUsed := res.GasUsed
	if ctx.BlockGasMeter() != nil {
		limit := ctx.BlockGasMeter().Limit()
		consumed := ctx.BlockGasMeter().GasConsumed()
		cumulativeGasUsed = uint64(math.Min(float64(cumulativeGasUsed+consumed), float64(limit)))
	}

	var contractAddr common.Address
	if msg.To() == nil {
		contractAddr = crypto.CreateAddress(msg.From(), msg.Nonce())
	}

	receipt := &ethtypes.Receipt{
		Type:              tx.Type(),
		PostState:         nil, // TODO: intermediate state root
		CumulativeGasUsed: cumulativeGasUsed,
		Bloom:             bloomReceipt,
		Logs:              logs,
		TxHash:            txConfig.TxHash,
		ContractAddress:   contractAddr,
		GasUsed:           res.GasUsed,
		BlockHash:         txConfig.BlockHash,
		BlockNumber:       big.NewInt(ctx.BlockHeight()),
		TransactionIndex:  txConfig.TxIndex,
	}

	if !res.Failed() {
		receipt.Status = ethtypes.ReceiptStatusSuccessful
		// Only call hooks if tx executed successfully.
		if err = k.PostTxProcessing(tmpCtx, msg, receipt); err != nil {
			// If hooks return error, revert the whole tx.
			res.VmError = types.ErrPostTxProcessing.Error()
			k.Logger(ctx).Error("tx post processing failed", "error", err)

			// If the tx failed in post processing hooks, we should clear the logs
			res.Logs = nil
		} else if commit != nil {
			// PostTxProcessing is successful, commit the tmpCtx
			commit()
			// Since the post processing can alter the log, we need to update the result
			res.Logs = types.NewLogsFromEth(receipt.Logs)
			ctx.EventManager().EmitEvents(tmpCtx.EventManager().Events())
		}
	}

	// refund gas in order to match the Ethereum gas consumption instead of the default SDK one.
	if err = k.RefundGas(ctx, msg, msg.Gas()-res.GasUsed, cfg.Params.EvmDenom); err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to refund gas leftover gas to sender %s", msg.From())
	}

	if len(receipt.Logs) > 0 {
		// Update transient block bloom filter
		k.SetBlockBloomTransient(ctx, receipt.Bloom.Big())
		k.SetLogSizeTransient(ctx, uint64(txConfig.LogIndex)+uint64(len(receipt.Logs)))
	}

	k.SetTxIndexTransient(ctx, uint64(txConfig.TxIndex)+1)

	totalGasUsed, err := k.AddTransientGasUsed(ctx, res.GasUsed)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to add transient gas used")
	}

	// reset the gas meter for current cosmos transaction
	k.ResetGasMeterAndConsumeGas(ctx, totalGasUsed)
	return res, nil
}

// ApplyMessageWithConfig computes the new state by applying the given message against the existing state.
// If the message fails, the VM execution error with the reason will be returned to the client
// and the transaction won't be committed to the store.
//
// # Reverted state
//
// The snapshot and rollback are supported by the `statedb.StateDB`.
//
// # Different Callers
//
// It's called in three scenarios:
// 1. `ApplyTransaction`, in the transaction processing flow.
// 2. `EthCall/EthEstimateGas` grpc query handler.
// 3. Called by other native modules directly.
//
// # Prechecks and Preprocessing
//
// All relevant state transition prechecks for the MsgEthereumTx are performed on the AnteHandler,
// prior to running the transaction against the state. The prechecks run are the following:
//
// 1. the nonce of the message caller is correct
// 2. caller has enough balance to cover transaction fee(gaslimit * gasprice)
// 3. the amount of gas required is available in the block
// 4. the purchased gas is enough to cover intrinsic usage
// 5. there is no overflow when calculating intrinsic gas
// 6. caller has enough balance to cover asset transfer for **topmost** call
//
// The preprocessing steps performed by the AnteHandler are:
//
// 1. set up the initial access list (iff fork > Berlin)
//
// # Tracer parameter
//
// It should be a `vm.Tracer` object or nil, if pass `nil`, it'll create a default one based on keeper options.
//
// # Commit parameter
//
// If commit is true, the `StateDB` will be committed, otherwise discarded.
func (k *Keeper) ApplyMessageWithConfig(ctx sdk.Context, msg core.Message, tracer vm.EVMLogger, commit bool, cfg *types.EVMConfig, txConfig statedb.TxConfig) (*types.MsgEthereumTxResponse, error) {
	var (
		ret   []byte // return bytes from evm execution
		vmErr error  // vm errors do not effect consensus and are therefore not assigned to err
	)

	// return error if contract creation or call are disabled through governance
	if !cfg.Params.EnableCreate && msg.To() == nil {
		return nil, sdkerrors.Wrap(types.ErrCreateDisabled, "failed to create new contract")
	} else if !cfg.Params.EnableCall && msg.To() != nil {
		return nil, sdkerrors.Wrap(types.ErrCallDisabled, "failed to call contract")
	}

	stateDB := statedb.New(ctx, k, txConfig)
	evm := k.NewEVM(ctx, msg, cfg, tracer, stateDB)

	leftoverGas := msg.Gas()
	// Allow the tracer captures the tx level events, mainly the gas consumption.
	if evm.Config.Debug {
		evm.Config.Tracer.CaptureTxStart(leftoverGas)
		defer func() {
			evm.Config.Tracer.CaptureTxEnd(leftoverGas)
		}()
	}

	sender := vm.AccountRef(msg.From())
	contractCreation := msg.To() == nil
	isLondon := cfg.ChainConfig.IsLondon(evm.Context.BlockNumber)

	intrinsicGas, err := k.GetEthIntrinsicGas(ctx, msg, cfg.ChainConfig, contractCreation)
	if err != nil {
		// should have already been checked on Ante Handler
		return nil, sdkerrors.Wrap(err, "intrinsic gas failed")
	}

	// Should check again even if it is checked on Ante Handler, because eth_call don't go through Ante Handler.
	if leftoverGas < intrinsicGas {
		// eth_estimateGas will check for this exact error
		return nil, sdkerrors.Wrap(core.ErrIntrinsicGas, "apply message")
	}
	leftoverGas -= intrinsicGas

	// access list preparation is moved from ante handler to here, because it's needed when `ApplyMessage` is called
	// under contexts where ante handlers are not run, for example `eth_call` and `eth_estimateGas`.
	if rules := cfg.ChainConfig.Rules(big.NewInt(ctx.BlockHeight()), cfg.ChainConfig.MergeNetsplitBlock != nil); rules.IsBerlin {
		stateDB.PrepareAccessList(msg.From(), msg.To(), vm.ActivePrecompiles(rules), msg.AccessList())
	}

	if contractCreation {
		// take over the nonce management from evm:
		// - reset sender's nonce to msg.Nonce() before calling evm.
		// - increase sender's nonce by one no matter the result.
		stateDB.SetNonce(sender.Address(), msg.Nonce())
		ret, _, leftoverGas, vmErr = evm.Create(sender, msg.Data(), leftoverGas, msg.Value())
		stateDB.SetNonce(sender.Address(), msg.Nonce()+1)
	} else {
		ret, leftoverGas, vmErr = evm.Call(sender, *msg.To(), msg.Data(), leftoverGas, msg.Value())
	}

	refundQuotient := params.RefundQuotient

	// After EIP-3529: refunds are capped to gasUsed / 5
	if isLondon {
		refundQuotient = params.RefundQuotientEIP3529
	}

	// calculate gas refund
	if msg.Gas() < leftoverGas {
		return nil, sdkerrors.Wrap(types.ErrGasOverflow, "apply message")
	}
	// refund gas
	leftoverGas += evmkeeper.GasToRefund(stateDB.GetRefund(), msg.Gas()-leftoverGas, refundQuotient)

	// EVM execution error needs to be available for the JSON-RPC client
	var vmError string
	if vmErr != nil {
		vmError = vmErr.Error()
	}

	// The dirty states in `StateDB` is either committed or discarded after return
	if commit {
		if err := stateDB.Commit(); err != nil {
			return nil, sdkerrors.Wrap(err, "failed to commit stateDB")
		}
	}

	// calculate a minimum amount of gas to be charged to sender if GasLimit
	// is considerably higher than GasUsed to stay more aligned with Tendermint gas mechanics
	// for more info https://github.com/evmos/ethermint/issues/1085
	gasLimit := sdk.NewDec(int64(msg.Gas()))
	minGasMultiplier := k.GetMinGasMultiplier(ctx)
	minimumGasUsed := gasLimit.Mul(minGasMultiplier)

	if msg.Gas() < leftoverGas {
		return nil, sdkerrors.Wrapf(types.ErrGasOverflow, "message gas limit < leftover gas (%d < %d)", msg.Gas(), leftoverGas)
	}
	temporaryGasUsed := msg.Gas() - leftoverGas
	gasUsed := sdk.MaxDec(minimumGasUsed, sdk.NewDec(int64(temporaryGasUsed))).TruncateInt().Uint64()
	// reset leftoverGas, to be used by the tracer
	leftoverGas = msg.Gas() - gasUsed

	return &types.MsgEthereumTxResponse{
		GasUsed: gasUsed,
		VmError: vmError,
		Ret:     ret,
		Logs:    types.NewLogsFromEth(stateDB.Logs()),
		Hash:    txConfig.TxHash.Hex(),
	}, nil
}

// ApplyMessage calls ApplyMessageWithConfig with default EVMConfig
func (k *Keeper) ApplyMessage(ctx sdk.Context, msg core.Message, tracer vm.EVMLogger, commit bool) (*types.MsgEthereumTxResponse, error) {
	cfg, err := k.EVMConfig(ctx)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to load evm config")
	}
	txConfig := statedb.NewEmptyTxConfig(common.BytesToHash(ctx.HeaderHash()))
	return k.ApplyMessageWithConfig(ctx, msg, tracer, commit, cfg, txConfig)
}