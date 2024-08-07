package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	fxtypes "github.com/functionx/fx-core/v7/types"
	"github.com/functionx/fx-core/v7/x/crosschain/types"
	erc20types "github.com/functionx/fx-core/v7/x/erc20/types"
)

var _ types.MsgServer = MsgServer{}

type MsgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the gov MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &MsgServer{Keeper: keeper}
}

func (s MsgServer) BondedOracle(c context.Context, msg *types.MsgBondedOracle) (*types.MsgBondedOracleResponse, error) {
	oracleAddr, err := sdk.AccAddressFromBech32(msg.OracleAddress)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "oracle address")
	}
	bridgerAddr, err := sdk.AccAddressFromBech32(msg.BridgerAddress)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "bridger address")
	}
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "validator address")
	}
	ctx := sdk.UnwrapSDKContext(c)
	if !s.IsProposalOracle(ctx, msg.OracleAddress) {
		return nil, types.ErrNoFoundOracle
	}
	// check oracle has set bridger address
	if s.HasOracle(ctx, oracleAddr) {
		return nil, errorsmod.Wrap(types.ErrInvalid, "oracle existed bridger address")
	}
	// check bridger address is bound to oracle
	if s.HasOracleAddrByBridgerAddr(ctx, bridgerAddr) {
		return nil, errorsmod.Wrap(types.ErrInvalid, "bridger address is bound to oracle")
	}
	// check external address is bound to oracle
	if s.HasOracleAddrByExternalAddr(ctx, msg.ExternalAddress) {
		return nil, errorsmod.Wrap(types.ErrInvalid, "external address is bound to oracle")
	}
	threshold := s.GetOracleDelegateThreshold(ctx)
	oracle := types.Oracle{
		OracleAddress:     oracleAddr.String(),
		BridgerAddress:    bridgerAddr.String(),
		ExternalAddress:   msg.ExternalAddress,
		DelegateAmount:    msg.DelegateAmount.Amount,
		StartHeight:       ctx.BlockHeight(),
		Online:            true,
		DelegateValidator: msg.ValidatorAddress,
		SlashTimes:        0,
	}
	if threshold.Denom != msg.DelegateAmount.Denom {
		return nil, errorsmod.Wrapf(types.ErrInvalid, "delegate denom got %s, expected %s", msg.DelegateAmount.Denom, threshold.Denom)
	}
	if msg.DelegateAmount.IsLT(threshold) {
		return nil, types.ErrDelegateAmountBelowMinimum
	}
	if msg.DelegateAmount.Amount.GT(threshold.Amount.Mul(sdkmath.NewInt(s.GetOracleDelegateMultiple(ctx)))) {
		return nil, types.ErrDelegateAmountAboveMaximum
	}

	delegateAddr := oracle.GetDelegateAddress(s.moduleName)
	if err = s.bankKeeper.SendCoins(ctx, oracleAddr, delegateAddr, sdk.NewCoins(msg.DelegateAmount)); err != nil {
		return nil, err
	}
	msgDelegate := stakingtypes.NewMsgDelegate(delegateAddr, valAddr, msg.DelegateAmount)
	if _, err = s.stakingMsgServer.Delegate(sdk.WrapSDKContext(ctx), msgDelegate); err != nil {
		return nil, err
	}

	s.SetOracle(ctx, oracle)
	s.SetOracleAddrByBridgerAddr(ctx, bridgerAddr, oracleAddr)
	s.SetOracleAddrByExternalAddr(ctx, msg.ExternalAddress, oracleAddr)
	s.SetLastTotalPower(ctx)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, msg.ChainName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.OracleAddress),
	))

	return &types.MsgBondedOracleResponse{}, nil
}

//gocyclo:ignore
func (s MsgServer) AddDelegate(c context.Context, msg *types.MsgAddDelegate) (*types.MsgAddDelegateResponse, error) {
	oracleAddr, err := sdk.AccAddressFromBech32(msg.OracleAddress)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "oracle address")
	}
	ctx := sdk.UnwrapSDKContext(c)
	if !s.IsProposalOracle(ctx, msg.OracleAddress) {
		return nil, types.ErrNoFoundOracle
	}
	oracle, found := s.GetOracle(ctx, oracleAddr)
	if !found {
		return nil, types.ErrNoFoundOracle
	}

	threshold := s.GetOracleDelegateThreshold(ctx)

	if threshold.Denom != msg.Amount.Denom {
		return nil, errorsmod.Wrapf(types.ErrInvalid, "delegate denom got %s, expected %s", msg.Amount.Denom, threshold.Denom)
	}

	slashAmount := types.NewDelegateAmount(oracle.GetSlashAmount(s.GetSlashFraction(ctx)))
	if slashAmount.IsPositive() && msg.Amount.Amount.LT(slashAmount.Amount) {
		return nil, errorsmod.Wrap(types.ErrInvalid, "not sufficient slash amount")
	}

	delegateCoin := types.NewDelegateAmount(msg.Amount.Amount.Sub(slashAmount.Amount))

	oracle.DelegateAmount = oracle.DelegateAmount.Add(delegateCoin.Amount)
	if oracle.DelegateAmount.Sub(threshold.Amount).IsNegative() {
		return nil, types.ErrDelegateAmountBelowMinimum
	}
	if oracle.DelegateAmount.GT(threshold.Amount.Mul(sdkmath.NewInt(s.GetOracleDelegateMultiple(ctx)))) {
		return nil, types.ErrDelegateAmountAboveMaximum
	}

	if slashAmount.IsPositive() {
		if err = s.bankKeeper.SendCoinsFromAccountToModule(ctx, oracleAddr, s.moduleName, sdk.NewCoins(slashAmount)); err != nil {
			return nil, err
		}
		if err = s.bankKeeper.BurnCoins(ctx, s.moduleName, sdk.NewCoins(slashAmount)); err != nil {
			return nil, err
		}
	}

	if delegateCoin.IsPositive() {
		delegateAddr := oracle.GetDelegateAddress(s.moduleName)
		if err = s.bankKeeper.SendCoins(ctx, oracleAddr, delegateAddr, sdk.NewCoins(delegateCoin)); err != nil {
			return nil, err
		}
		msgDelegate := stakingtypes.NewMsgDelegate(delegateAddr, oracle.GetValidator(), delegateCoin)
		if _, err = s.stakingMsgServer.Delegate(c, msgDelegate); err != nil {
			return nil, err
		}
	}

	if !oracle.Online {
		oracle.Online = true
		oracle.StartHeight = ctx.BlockHeight()
		if !ctx.IsCheckTx() {
			telemetry.SetGaugeWithLabels(
				[]string{types.ModuleName, "oracle_status"},
				float32(0),
				[]metrics.Label{
					telemetry.NewLabel("module", s.moduleName),
					telemetry.NewLabel("address", oracle.OracleAddress),
				},
			)
		}
	}
	oracle.SlashTimes = 0

	s.SetOracle(ctx, oracle)
	s.SetLastTotalPower(ctx)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, msg.ChainName),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.OracleAddress),
		),
	)

	return &types.MsgAddDelegateResponse{}, nil
}

func (s MsgServer) ReDelegate(c context.Context, msg *types.MsgReDelegate) (*types.MsgReDelegateResponse, error) {
	oracleAddr, err := sdk.AccAddressFromBech32(msg.OracleAddress)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "oracle address")
	}
	valDstAddress, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "validator address")
	}
	ctx := sdk.UnwrapSDKContext(c)
	oracle, found := s.GetOracle(ctx, oracleAddr)
	if !found {
		return nil, types.ErrNoFoundOracle
	}
	if !oracle.Online {
		return nil, types.ErrOracleNotOnLine
	}
	if oracle.DelegateValidator == msg.ValidatorAddress {
		return nil, errorsmod.Wrap(types.ErrInvalid, "validator address is not changed")
	}
	delegateAddr := oracle.GetDelegateAddress(s.moduleName)
	valSrcAddress := oracle.GetValidator()
	delegateToken, err := s.GetOracleDelegateToken(ctx, delegateAddr, valSrcAddress)
	if err != nil {
		return nil, err
	}
	msgBeginRedelegate := stakingtypes.NewMsgBeginRedelegate(delegateAddr, valSrcAddress, valDstAddress, types.NewDelegateAmount(delegateToken))
	if _, err = s.stakingMsgServer.BeginRedelegate(c, msgBeginRedelegate); err != nil {
		return nil, err
	}
	return &types.MsgReDelegateResponse{}, err
}

func (s MsgServer) EditBridger(c context.Context, msg *types.MsgEditBridger) (*types.MsgEditBridgerResponse, error) {
	oracleAddr, err := sdk.AccAddressFromBech32(msg.OracleAddress)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "oracle address")
	}
	bridgerAddr, err := sdk.AccAddressFromBech32(msg.BridgerAddress)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "oracle address")
	}
	ctx := sdk.UnwrapSDKContext(c)
	oracle, found := s.GetOracle(ctx, oracleAddr)
	if !found {
		return nil, types.ErrNoFoundOracle
	}
	if !oracle.Online {
		return nil, types.ErrOracleNotOnLine
	}
	if oracle.BridgerAddress == msg.BridgerAddress {
		return nil, errorsmod.Wrap(types.ErrInvalid, "bridger address is not changed")
	}
	if s.HasOracleAddrByBridgerAddr(ctx, bridgerAddr) {
		return nil, errorsmod.Wrap(types.ErrInvalid, "bridger address is bound to oracle")
	}
	s.DelOracleAddrByBridgerAddr(ctx, oracle.GetBridger())
	oracle.BridgerAddress = msg.BridgerAddress
	s.SetOracle(ctx, oracle)
	s.SetOracleAddrByBridgerAddr(ctx, bridgerAddr, oracleAddr)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, msg.ChainName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.OracleAddress),
	))
	return &types.MsgEditBridgerResponse{}, nil
}

func (s MsgServer) WithdrawReward(c context.Context, msg *types.MsgWithdrawReward) (*types.MsgWithdrawRewardResponse, error) {
	oracleAddr, err := sdk.AccAddressFromBech32(msg.OracleAddress)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "oracle address")
	}
	ctx := sdk.UnwrapSDKContext(c)
	oracle, found := s.GetOracle(ctx, oracleAddr)
	if !found {
		return nil, types.ErrNoFoundOracle
	}
	if !oracle.Online {
		return nil, types.ErrOracleNotOnLine
	}

	delegateAddr := oracle.GetDelegateAddress(s.moduleName)
	msgWithdrawDelegatorReward := distributiontypes.NewMsgWithdrawDelegatorReward(delegateAddr, oracle.GetValidator())
	if _, err = s.distributionKeeper.WithdrawDelegatorReward(c, msgWithdrawDelegatorReward); err != nil {
		return nil, err
	}
	balances := s.bankKeeper.GetAllBalances(ctx, delegateAddr)
	if !balances.IsAllPositive() {
		return nil, errorsmod.Wrap(types.ErrInvalid, "rewards is empty")
	}
	if err = s.bankKeeper.SendCoins(ctx, delegateAddr, oracleAddr, balances); err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, msg.ChainName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.OracleAddress),
	))
	return &types.MsgWithdrawRewardResponse{}, nil
}

func (s MsgServer) UnbondedOracle(c context.Context, msg *types.MsgUnbondedOracle) (*types.MsgUnbondedOracleResponse, error) {
	oracleAddr, err := sdk.AccAddressFromBech32(msg.OracleAddress)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "oracle address")
	}
	ctx := sdk.UnwrapSDKContext(c)
	if s.IsProposalOracle(ctx, msg.OracleAddress) {
		return nil, errorsmod.Wrap(types.ErrInvalid, "need to pass a proposal to unbind")
	}
	oracle, found := s.GetOracle(ctx, oracleAddr)
	if !found {
		return nil, types.ErrNoFoundOracle
	}
	if oracle.Online {
		return nil, errorsmod.Wrap(types.ErrInvalid, "oracle on line")
	}
	delegateAddr := oracle.GetDelegateAddress(s.moduleName)
	validatorAddr := oracle.GetValidator()
	if _, found = s.stakingKeeper.GetUnbondingDelegation(ctx, delegateAddr, validatorAddr); found {
		return nil, errorsmod.Wrap(types.ErrInvalid, "exist unbonding delegation")
	}
	balances := s.bankKeeper.GetAllBalances(ctx, delegateAddr)
	slashAmount := types.NewDelegateAmount(oracle.GetSlashAmount(s.GetSlashFraction(ctx)))
	if slashAmount.IsPositive() {
		if balances.AmountOf(slashAmount.Denom).LT(slashAmount.Amount) {
			return nil, errorsmod.Wrap(types.ErrInvalid, "not sufficient slash amount")
		}
		if err = s.bankKeeper.SendCoinsFromAccountToModule(ctx, delegateAddr, s.moduleName, sdk.NewCoins(slashAmount)); err != nil {
			return nil, err
		}
		if err = s.bankKeeper.BurnCoins(ctx, s.moduleName, sdk.NewCoins(slashAmount)); err != nil {
			return nil, err
		}
	}
	sendCoins := balances.Sub(sdk.NewCoins(slashAmount)...)
	for i := 0; i < len(sendCoins); i++ {
		if !sendCoins[i].IsPositive() {
			sendCoins = append(sendCoins[:i], sendCoins[i+1:]...)
			i--
		}
	}
	if sendCoins.IsAllPositive() {
		if err = s.bankKeeper.SendCoins(ctx, delegateAddr, oracleAddr, sendCoins); err != nil {
			return nil, err
		}
	}

	s.DelOracleAddrByExternalAddr(ctx, oracle.ExternalAddress)
	s.DelOracleAddrByBridgerAddr(ctx, oracle.GetBridger())
	s.DelOracle(ctx, oracle.GetOracle())
	s.DelLastEventNonceByOracle(ctx, oracleAddr)

	return &types.MsgUnbondedOracleResponse{}, nil
}

func (s MsgServer) SendToExternal(c context.Context, msg *types.MsgSendToExternal) (*types.MsgSendToExternalResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "sender address")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// convert denom to many
	fxTarget := fxtypes.ParseFxTarget(s.moduleName)
	targetCoin, err := s.erc20Keeper.ConvertDenomToTarget(ctx, sender, msg.Amount.Add(msg.BridgeFee), fxTarget)
	if err != nil && !erc20types.IsInsufficientLiquidityErr(err) {
		return nil, err
	}
	msg.Amount.Denom = targetCoin.Denom
	msg.BridgeFee.Denom = targetCoin.Denom

	var txID uint64
	if erc20types.IsInsufficientLiquidityErr(err) {
		txID, err = s.AddToOutgoingPendingPool(ctx, sender, msg.Dest, msg.Amount, msg.BridgeFee)
	} else {
		txID, err = s.AddToOutgoingPool(ctx, sender, msg.Dest, msg.Amount, msg.BridgeFee)
	}
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, msg.ChainName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	return &types.MsgSendToExternalResponse{
		OutgoingTxId: txID,
	}, nil
}

func (s MsgServer) CancelSendToExternal(c context.Context, msg *types.MsgCancelSendToExternal) (*types.MsgCancelSendToExternalResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "sender address")
	}

	ctx := sdk.UnwrapSDKContext(c)

	if _, err = s.RemoveFromOutgoingPoolAndRefund(ctx, msg.TransactionId, sender); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, msg.ChainName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	return &types.MsgCancelSendToExternalResponse{}, nil
}

func (s MsgServer) IncreaseBridgeFee(c context.Context, msg *types.MsgIncreaseBridgeFee) (*types.MsgIncreaseBridgeFeeResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "sender address")
	}

	ctx := sdk.UnwrapSDKContext(c)

	if err = s.AddUnbatchedTxBridgeFee(ctx, msg.TransactionId, sender, msg.AddBridgeFee); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, msg.ChainName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	return &types.MsgIncreaseBridgeFeeResponse{}, nil
}

func (s MsgServer) RequestBatch(c context.Context, msg *types.MsgRequestBatch) (*types.MsgRequestBatchResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "sender address")
	}

	ctx := sdk.UnwrapSDKContext(c)

	bridgeToken := s.GetDenomBridgeToken(ctx, msg.Denom)
	if bridgeToken == nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "bridge token is not exist")
	}

	if !s.HasOracleAddrByBridgerAddr(ctx, sender) {
		if !s.IsProposalOracle(ctx, msg.Sender) {
			return nil, errorsmod.Wrap(types.ErrEmpty, "sender must be oracle or bridger")
		}
	}

	batch, err := s.BuildOutgoingTxBatch(ctx, bridgeToken.Token, msg.FeeReceive, types.OutgoingTxBatchSize, msg.MinimumFee, msg.BaseFee)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, msg.ChainName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	return &types.MsgRequestBatchResponse{
		BatchNonce: batch.BatchNonce,
	}, nil
}

func (s MsgServer) ConfirmBatch(c context.Context, msg *types.MsgConfirmBatch) (*types.MsgConfirmBatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// fetch the outgoing batch given the nonce
	batch := s.GetOutgoingTxBatch(ctx, msg.TokenContract, msg.Nonce)
	if batch == nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "couldn't find batch")
	}

	checkpoint, err := batch.GetCheckpoint(s.GetGravityID(ctx))
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, err.Error())
	}

	oracleAddr, err := s.confirmHandlerCommon(ctx, msg.BridgerAddress, msg.ExternalAddress, msg.Signature, checkpoint)
	if err != nil {
		return nil, err
	}
	// check if we already have this confirm
	if s.GetBatchConfirm(ctx, msg.TokenContract, msg.Nonce, oracleAddr) != nil {
		return nil, errorsmod.Wrap(types.ErrDuplicate, "signature")
	}
	s.SetBatchConfirm(ctx, oracleAddr, msg)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, msg.ChainName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.BridgerAddress),
	))

	return &types.MsgConfirmBatchResponse{}, nil
}

func (s MsgServer) OracleSetConfirm(c context.Context, msg *types.MsgOracleSetConfirm) (*types.MsgOracleSetConfirmResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	oracleSet := s.GetOracleSet(ctx, msg.Nonce)
	if oracleSet == nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "couldn't find oracleSet")
	}

	checkpoint, err := oracleSet.GetCheckpoint(s.GetGravityID(ctx))
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, err.Error())
	}
	oracleAddr, err := s.confirmHandlerCommon(ctx, msg.BridgerAddress, msg.ExternalAddress, msg.Signature, checkpoint)
	if err != nil {
		return nil, err
	}
	// check if we already have this confirm
	if s.GetOracleSetConfirm(ctx, msg.Nonce, oracleAddr) != nil {
		return nil, errorsmod.Wrap(types.ErrDuplicate, "signature")
	}
	s.SetOracleSetConfirm(ctx, oracleAddr, msg)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, msg.ChainName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.BridgerAddress),
	))

	return &types.MsgOracleSetConfirmResponse{}, nil
}

func (s MsgServer) BridgeCallConfirm(c context.Context, msg *types.MsgBridgeCallConfirm) (*types.MsgBridgeCallConfirmResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	outgoingBridgeCall, found := s.GetOutgoingBridgeCallByNonce(ctx, msg.Nonce)
	if !found {
		return nil, errorsmod.Wrap(types.ErrInvalid, "couldn't find outgoing bridge call")
	}

	checkpoint, err := outgoingBridgeCall.GetCheckpoint(s.GetGravityID(ctx))
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, err.Error())
	}

	oracleAddr, err := s.confirmHandlerCommon(ctx, msg.BridgerAddress, msg.ExternalAddress, msg.Signature, checkpoint)
	if err != nil {
		return nil, err
	}

	if s.HasBridgeCallConfirm(ctx, msg.Nonce, oracleAddr) {
		return nil, errorsmod.Wrap(types.ErrDuplicate, "signature")
	}
	s.SetBridgeCallConfirm(ctx, oracleAddr, msg)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, msg.ChainName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.BridgerAddress),
	))

	return &types.MsgBridgeCallConfirmResponse{}, nil
}

func (s MsgServer) SendToExternalClaim(c context.Context, msg *types.MsgSendToExternalClaim) (*types.MsgSendToExternalClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if err := s.claimHandlerCommon(ctx, msg); err != nil {
		return nil, err
	}
	return &types.MsgSendToExternalClaimResponse{}, nil
}

func (s MsgServer) SendToFxClaim(c context.Context, msg *types.MsgSendToFxClaim) (*types.MsgSendToFxClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if err := s.claimHandlerCommon(ctx, msg); err != nil {
		return nil, err
	}
	return &types.MsgSendToFxClaimResponse{}, nil
}

func (s MsgServer) BridgeCallClaim(c context.Context, msg *types.MsgBridgeCallClaim) (*types.MsgBridgeCallClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if err := s.claimHandlerCommon(ctx, msg); err != nil {
		return nil, err
	}
	return &types.MsgBridgeCallClaimResponse{}, nil
}

func (s MsgServer) BridgeCallResultClaim(c context.Context, msg *types.MsgBridgeCallResultClaim) (*types.MsgBridgeCallResultClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if err := s.claimHandlerCommon(ctx, msg); err != nil {
		return nil, err
	}
	return &types.MsgBridgeCallResultClaimResponse{}, nil
}

func (s MsgServer) BridgeTokenClaim(c context.Context, msg *types.MsgBridgeTokenClaim) (*types.MsgBridgeTokenClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if err := s.claimHandlerCommon(ctx, msg); err != nil {
		return nil, err
	}
	return &types.MsgBridgeTokenClaimResponse{}, nil
}

func (s MsgServer) BridgeCall(c context.Context, msg *types.MsgBridgeCall) (*types.MsgBridgeCallResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "sender address")
	}

	ctx := sdk.UnwrapSDKContext(c)
	tokens, notLiquidCoins, err := s.BridgeCallCoinsToERC20Token(ctx, sender, msg.Coins)
	if err != nil {
		return nil, err
	}

	var outCallNonce uint64
	if len(notLiquidCoins) > 0 {
		outCallNonce, err = s.AddPendingOutgoingBridgeCall(ctx, msg.GetSenderAddr(), msg.GetRefundAddr(), tokens, msg.GetToAddr(), msg.MustData(), msg.MustMemo(), 0, notLiquidCoins)
	} else {
		outCallNonce, err = s.AddOutgoingBridgeCall(ctx, msg.GetSenderAddr(), msg.GetRefundAddr(), tokens, msg.GetToAddr(), msg.MustData(), msg.MustMemo(), 0)
	}
	if err != nil {
		return nil, err
	}

	// bridge call from msg
	s.SetBridgeCallFromMsg(ctx, outCallNonce)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, msg.ChainName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	return &types.MsgBridgeCallResponse{}, nil
}

func (s MsgServer) CancelPendingBridgeCall(c context.Context, msg *types.MsgCancelPendingBridgeCall) (*types.MsgCancelPendingBridgeCallResponse, error) {
	sender := sdk.MustAccAddressFromBech32(msg.Sender)
	ctx := sdk.UnwrapSDKContext(c)

	if _, err := s.HandleCancelPendingOutgoingBridgeCall(ctx, msg.Nonce, sender); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, msg.ChainName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))
	return &types.MsgCancelPendingBridgeCallResponse{}, nil
}

func (s MsgServer) AddPendingPoolRewards(c context.Context, msg *types.MsgAddPendingPoolRewards) (*types.MsgAddPendingPoolRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	sender := sdk.MustAccAddressFromBech32(msg.Sender)
	if err := s.Keeper.AddPendingPoolRewards(ctx, msg.Id, sender, msg.Rewards); err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, msg.ChainName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))
	return &types.MsgAddPendingPoolRewardsResponse{}, nil
}

// OracleSetUpdateClaim handles claims for executing a oracle set update on Ethereum
func (s MsgServer) OracleSetUpdateClaim(c context.Context, msg *types.MsgOracleSetUpdatedClaim) (*types.MsgOracleSetUpdatedClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	for _, member := range msg.Members {
		if !s.HasOracleAddrByExternalAddr(ctx, member.ExternalAddress) {
			return nil, errorsmod.Wrap(types.ErrInvalid, "external address")
		}
	}

	if err := s.claimHandlerCommon(ctx, msg); err != nil {
		return nil, err
	}
	return &types.MsgOracleSetUpdatedClaimResponse{}, nil
}

func (s MsgServer) UpdateParams(c context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if s.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", s.authority, req.Authority)
	}
	ctx := sdk.UnwrapSDKContext(c)
	if err := s.SetParams(ctx, &req.Params); err != nil {
		return nil, err
	}
	return &types.MsgUpdateParamsResponse{}, nil
}

func (s MsgServer) UpdateChainOracles(c context.Context, req *types.MsgUpdateChainOracles) (*types.MsgUpdateChainOraclesResponse, error) {
	if s.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", s.authority, req.Authority)
	}
	ctx := sdk.UnwrapSDKContext(c)
	if err := s.UpdateProposalOracles(ctx, req.Oracles); err != nil {
		return nil, err
	}
	return &types.MsgUpdateChainOraclesResponse{}, nil
}

func (s MsgServer) checkBridgerIsOracle(ctx sdk.Context, bridgerAddr sdk.AccAddress) (oracleAddr sdk.AccAddress, err error) {
	oracleAddr, found := s.GetOracleAddrByBridgerAddr(ctx, bridgerAddr)
	if !found {
		return oracleAddr, types.ErrNoFoundOracle
	}
	oracle, found := s.GetOracle(ctx, oracleAddr)
	if !found {
		return oracleAddr, types.ErrNoFoundOracle
	}
	if !oracle.Online {
		return oracleAddr, types.ErrOracleNotOnLine
	}
	return oracleAddr, nil
}

// claimHandlerCommon is an internal function that provides common code for processing claims once they are
// translated from the message to the Ethereum claim interface
func (s MsgServer) claimHandlerCommon(ctx sdk.Context, msg types.ExternalClaim) (err error) {
	bridgerAddr := msg.GetClaimer()
	oracleAddr, err := s.checkBridgerIsOracle(ctx, bridgerAddr)
	if err != nil {
		return err
	}

	// Add the claim to the store
	if _, err := s.Attest(ctx, oracleAddr, msg); err != nil {
		return err
	}

	// Emit the handle message event
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, s.moduleName),
		sdk.NewAttribute(sdk.AttributeKeySender, bridgerAddr.String()),
	))

	return nil
}

func (s MsgServer) confirmHandlerCommon(ctx sdk.Context, bridgerAddr, signatureAddr, signature string, checkpoint []byte) (oracleAddr sdk.AccAddress, err error) {
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, "signature decoding")
	}

	oracleAddr, found := s.GetOracleAddrByExternalAddr(ctx, signatureAddr)
	if !found {
		return nil, types.ErrNoFoundOracle
	}

	oracle, found := s.GetOracle(ctx, oracleAddr)
	if !found {
		return nil, types.ErrNoFoundOracle
	}

	if oracle.ExternalAddress != signatureAddr {
		return nil, errorsmod.Wrapf(types.ErrInvalid, "got %s, expected %s", signatureAddr, oracle.ExternalAddress)
	}
	if oracle.BridgerAddress != bridgerAddr {
		return nil, errorsmod.Wrapf(types.ErrInvalid, "got %s, expected %s", bridgerAddr, oracle.BridgerAddress)
	}
	if err = types.ValidateEthereumSignature(checkpoint, sigBytes, oracle.ExternalAddress); err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalid, fmt.Sprintf("signature verification failed expected sig by %s with checkpoint %s found %s", oracle.ExternalAddress, hex.EncodeToString(checkpoint), signature))
	}
	return oracleAddr, nil
}
