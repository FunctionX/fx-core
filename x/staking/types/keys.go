package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

const GrantPrivilegeSignaturePrefix = "GrantPrivilege:"

var (
	AllowanceKey                 = []byte{0x90}
	ValidatorOperatorKey         = []byte{0x91}
	ValidatorNewConsensusPubKey  = []byte{0x92}
	ValidatorOldConsensusAddrKey = []byte{0x93}
	ValidatorDelConsensusAddrKey = []byte{0x94}
)

func GetAllowanceKey(valAddr sdk.ValAddress, owner, spender sdk.AccAddress) []byte {
	// key is of the form AllowanceKey || valAddrLen (1 byte) || valAddr || ownerAddrLen (1 byte) || ownerAddr || spenderAddrLen (1 byte) || spenderAddr
	offset := len(AllowanceKey)
	key := make([]byte, offset+3+len(valAddr)+len(owner)+len(spender))
	copy(key[0:offset], AllowanceKey)
	key[offset] = byte(len(valAddr))
	copy(key[offset+1:offset+1+len(valAddr)], valAddr.Bytes())
	key[offset+1+len(valAddr)] = byte(len(owner))
	copy(key[offset+2+len(valAddr):offset+2+len(valAddr)+len(owner)], owner.Bytes())
	key[offset+2+len(valAddr)+len(owner)] = byte(len(spender))
	copy(key[offset+3+len(valAddr)+len(owner):], spender.Bytes())

	return key
}

func GetValidatorOperatorKey(addr []byte) []byte {
	return append(ValidatorOperatorKey, addr...)
}

func GetValidatorNewConsensusPubKey(addr sdk.ValAddress) []byte {
	return append(ValidatorNewConsensusPubKey, addr...)
}

func GetValidatorOldConsensusAddrKey(addr sdk.ValAddress) []byte {
	return append(ValidatorOldConsensusAddrKey, addr...)
}

func GetValidatorDelConsensusAddrKey(addr sdk.ValAddress) []byte {
	return append(ValidatorDelConsensusAddrKey, addr...)
}

func AddressFromValidatorNewConsensusPubKey(key []byte) []byte {
	kv.AssertKeyAtLeastLength(key, 3)
	return key[1:] // remove prefix bytes and address length
}

func AddressFromValidatorNewConsensusAddrKey(key []byte) []byte {
	kv.AssertKeyAtLeastLength(key, 3)
	return key[1:] // remove prefix bytes and address length
}

func AddressFromValidatorDelConsensusAddrKey(key []byte) []byte {
	kv.AssertKeyAtLeastLength(key, 3)
	return key[1:] // remove prefix bytes and address length
}
