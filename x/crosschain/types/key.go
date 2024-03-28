package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "crosschain"

	// RouterKey is the module name router key
	RouterKey = ModuleName
)

var (
	// OracleKey key oracle address -> Oracle
	OracleKey = []byte{0x12}

	// OracleAddressByExternalKey key external address -> value oracle address
	OracleAddressByExternalKey = []byte{0x13}

	// OracleAddressByBridgerKey key bridger address -> value oracle address
	OracleAddressByBridgerKey = []byte{0x14}

	// OracleSetRequestKey indexes oracle set requests by nonce
	OracleSetRequestKey = []byte{0x15}

	// OracleSetConfirmKey indexes oracle set confirmations by nonce and the validator account address
	OracleSetConfirmKey = []byte{0x16}

	// OracleAttestationKey attestation details by nonce and validator address
	// An attestation can be thought of as the 'event to be executed' while
	// the Claims are an individual validator saying that they saw an event
	// occur the Attestation is 'the event' that multiple claims vote on and
	// eventually executes
	OracleAttestationKey = []byte{0x17}

	// OutgoingTxPoolKey indexes the last nonce for the outgoing tx pool
	OutgoingTxPoolKey = []byte{0x18}

	// OutgoingTxBatchKey indexes outgoing tx batches under a nonce and token address
	OutgoingTxBatchKey = []byte{0x20}

	// OutgoingTxBatchBlockKey indexes outgoing tx batches under a block height and token address
	OutgoingTxBatchBlockKey = []byte{0x21}

	// BatchConfirmKey indexes oracle confirmations by token contract address
	BatchConfirmKey = []byte{0x22}

	// LastEventNonceByOracleKey indexes latest event nonce by oracle
	LastEventNonceByOracleKey = []byte{0x23}

	// LastObservedEventNonceKey indexes the latest event nonce
	LastObservedEventNonceKey = []byte{0x24}

	// SequenceKeyPrefix indexes different txIds
	SequenceKeyPrefix = []byte{0x25}

	// KeyLastTxPoolID indexes the lastTxPoolID
	KeyLastTxPoolID = append(SequenceKeyPrefix, []byte("lastTxPoolId")...)

	// KeyLastOutgoingBatchID indexes the lastBatchID
	KeyLastOutgoingBatchID = append(SequenceKeyPrefix, []byte("lastBatchId")...)

	// DenomToTokenKey prefixes the index of asset denom to external token
	DenomToTokenKey = []byte{0x26}

	// TokenToDenomKey prefixes the index of assets external token to denom
	TokenToDenomKey = []byte{0x27}

	// LastSlashedOracleSetNonce indexes the latest slashed oracleSet nonce
	LastSlashedOracleSetNonce = []byte{0x28}

	// LatestOracleSetNonce indexes the latest oracleSet nonce
	LatestOracleSetNonce = []byte{0x29}

	// LastSlashedBatchBlock indexes the latest slashed batch block height
	LastSlashedBatchBlock = []byte{0x30}

	// Deprecated: LastProposalBlockHeight
	// LastProposalBlockHeight = []byte{0x31}

	// LastObservedBlockHeightKey indexes the latest observed external block height
	LastObservedBlockHeightKey = []byte{0x32}

	// LastObservedOracleSetKey indexes the latest observed OracleSet nonce
	LastObservedOracleSetKey = []byte{0x33}

	// LastEventBlockHeightByOracleKey indexes latest event blockHeight by oracle
	LastEventBlockHeightByOracleKey = []byte{0x35}

	// Deprecated: PastExternalSignatureCheckpointKey indexes eth signature checkpoints that have existed
	// PastExternalSignatureCheckpointKey = []byte{0x36}

	// LastOracleSlashBlockHeight indexes the last oracle slash block height
	LastOracleSlashBlockHeight = []byte{0x37}

	// ProposalOracleKey -> value ProposalOracle
	ProposalOracleKey = []byte{0x38}

	// LastTotalPowerKey oracle set total power
	LastTotalPowerKey = []byte{0x39}

	// ParamsKey is the prefix for params key
	ParamsKey = []byte{0x40}

	// Deprecated: OutgoingTxRelationKey outgoing tx with evm
	// OutgoingTxRelationKey = []byte{0x41}

	BridgeCallRefundKey = []byte{0x42}

	BridgeCallRefundEventNonceKey = []byte{0x43}

	SnapshotOracleKey = []byte{0x44}

	BridgeCallRefundConfirmKey = []byte{0x45}

	// LastSlashedRefundNonce indexes the latest slashed refund nonce
	LastSlashedRefundNonce = []byte{0x46}

	// TokenTypeToTokenKey prefixes the index of asset token type to external token
	TokenTypeToTokenKey = []byte{0x47}
)

// GetOracleKey returns the following key format
func GetOracleKey(oracle sdk.AccAddress) []byte {
	return append(OracleKey, oracle.Bytes()...)
}

// GetOracleAddressByBridgerKey returns the following key format
func GetOracleAddressByBridgerKey(bridger sdk.AccAddress) []byte {
	return append(OracleAddressByBridgerKey, bridger.Bytes()...)
}

// GetOracleAddressByExternalKey returns the following key format
func GetOracleAddressByExternalKey(externalAddress string) []byte {
	return append(OracleAddressByExternalKey, []byte(externalAddress)...)
}

// GetOracleSetKey returns the following key format
func GetOracleSetKey(nonce uint64) []byte {
	return append(OracleSetRequestKey, sdk.Uint64ToBigEndian(nonce)...)
}

// GetOracleSetConfirmKey returns the following key format
func GetOracleSetConfirmKey(nonce uint64, oracleAddr sdk.AccAddress) []byte {
	return append(OracleSetConfirmKey, append(sdk.Uint64ToBigEndian(nonce), oracleAddr.Bytes()...)...)
}

// GetAttestationKey returns the following key format
// An attestation is an event multiple people are voting on, this function needs the claim
// details because each Attestation is aggregating all claims of a specific event, lets say
// validator X and validator y where making different claims about the same event nonce
// Note that the claim hash does NOT include the claimer address and only identifies an event
func GetAttestationKey(eventNonce uint64, claimHash []byte) []byte {
	return append(OracleAttestationKey, append(sdk.Uint64ToBigEndian(eventNonce), claimHash...)...)
}

// GetOutgoingTxPoolContractPrefix returns the following key format
// This prefix is used for iterating over unbatched transactions for a given contract
func GetOutgoingTxPoolContractPrefix(tokenContract string) []byte {
	return append(OutgoingTxPoolKey, []byte(tokenContract)...)
}

// GetOutgoingTxPoolKey returns the following key format
func GetOutgoingTxPoolKey(fee ERC20Token, id uint64) []byte {
	amount := make([]byte, 32)
	amount = fee.Amount.BigInt().FillBytes(amount)
	return append(OutgoingTxPoolKey, append([]byte(fee.Contract), append(amount, sdk.Uint64ToBigEndian(id)...)...)...)
}

// GetOutgoingTxBatchKey returns the following key format
func GetOutgoingTxBatchKey(tokenContract string, batchNonce uint64) []byte {
	return append(append(OutgoingTxBatchKey, []byte(tokenContract)...), sdk.Uint64ToBigEndian(batchNonce)...)
}

// GetOutgoingTxBatchBlockKey returns the following key format
func GetOutgoingTxBatchBlockKey(blockHeight uint64) []byte {
	return append(OutgoingTxBatchBlockKey, sdk.Uint64ToBigEndian(blockHeight)...)
}

// GetBatchConfirmKey returns the following key format
func GetBatchConfirmKey(tokenContract string, batchNonce uint64, oracleAddr sdk.AccAddress) []byte {
	return append(BatchConfirmKey, append([]byte(tokenContract), append(sdk.Uint64ToBigEndian(batchNonce), oracleAddr.Bytes()...)...)...)
}

// GetLastEventNonceByOracleKey returns the following key format
func GetLastEventNonceByOracleKey(oracleAddr sdk.AccAddress) []byte {
	return append(LastEventNonceByOracleKey, oracleAddr.Bytes()...)
}

// GetLastEventBlockHeightByOracleKey returns the following key format
func GetLastEventBlockHeightByOracleKey(oracleAddr sdk.AccAddress) []byte {
	return append(LastEventBlockHeightByOracleKey, oracleAddr.Bytes()...)
}

// GetDenomToTokenKey returns the following key format
func GetDenomToTokenKey(tokenContract string) []byte {
	return append(DenomToTokenKey, []byte(tokenContract)...)
}

// GetTokenToDenomKey returns the following key format
func GetTokenToDenomKey(denom string) []byte {
	return append(TokenToDenomKey, []byte(denom)...)
}

func GetBridgeCallRefundKey(address string, nonce uint64) []byte {
	return append(BridgeCallRefundKey, append([]byte(address), sdk.Uint64ToBigEndian(nonce)...)...)
}

func ParseBridgeCallRefundNonce(key []byte, address string) (nonce uint64) {
	addrNonce := bytes.TrimPrefix(key, BridgeCallRefundKey)
	nonceBytes := bytes.TrimPrefix(addrNonce, []byte(address))
	return sdk.BigEndianToUint64(nonceBytes)
}

func GetBridgeCallRefundAddressKey(address string) []byte {
	return append(BridgeCallRefundKey, []byte(address)...)
}

func GetBridgeCallRefundEventNonceKey(nonce uint64) []byte {
	return append(BridgeCallRefundEventNonceKey, sdk.Uint64ToBigEndian(nonce)...)
}

func GetSnapshotOracleKey(oracleSetNonce uint64) []byte {
	return append(SnapshotOracleKey, sdk.Uint64ToBigEndian(oracleSetNonce)...)
}

func GetRefundConfirmKey(nonce uint64, addr sdk.AccAddress) []byte {
	return append(BridgeCallRefundConfirmKey, append(sdk.Uint64ToBigEndian(nonce), addr.Bytes()...)...)
}

func GetRefundConfirmKeyByNonce(nonce uint64) []byte {
	return append(BridgeCallRefundConfirmKey, sdk.Uint64ToBigEndian(nonce)...)
}

func GetRefundConfirmNonceKey(nonce uint64) []byte {
	return append(BridgeCallRefundConfirmKey, sdk.Uint64ToBigEndian(nonce)...)
}

func GetTokenTypeToTokenKey(tokenContract string) []byte {
	return append(TokenTypeToTokenKey, []byte(tokenContract)...)
}
