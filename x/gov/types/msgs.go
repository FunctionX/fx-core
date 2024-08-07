package types

import (
	"encoding/hex"
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	_ sdk.Msg = &MsgUpdateFXParams{}
	_ sdk.Msg = &MsgUpdateEGFParams{}
	_ sdk.Msg = &MsgUpdateStore{}
	_ sdk.Msg = &MsgUpdateSwitchParams{}
)

const (
	TypeMsgUpdateParams       = "fx_update_params"
	TypeMsgUpdateEGFParams    = "fx_update_egf_params"
	TypeMsgUpdateStore        = "fx_update_store"
	TypeMsgUpdateSwitchParams = "fx_update_switch_params"
)

func NewMsgUpdateFXParams(authority string, params Params) *MsgUpdateFXParams {
	return &MsgUpdateFXParams{Authority: authority, Params: params}
}

// Route returns the MsgUpdateParams message route.
func (m *MsgUpdateFXParams) Route() string { return types.ModuleName }

// Type returns the MsgUpdateParams message type.
func (m *MsgUpdateFXParams) Type() string { return TypeMsgUpdateParams }

// GetSignBytes returns the raw bytes for a MsgUpdateParams message that
// the expected signer needs to sign.
func (m *MsgUpdateFXParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateFXParams) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

func (m *MsgUpdateFXParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "authority")
	}
	if err := m.Params.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "params")
	}
	return nil
}

func NewMsgUpdateEGFParams(authority string, params EGFParams) *MsgUpdateEGFParams {
	return &MsgUpdateEGFParams{Authority: authority, Params: params}
}

// Route returns the MsgUpdateParams message route.
func (m *MsgUpdateEGFParams) Route() string { return types.ModuleName }

// Type returns the MsgUpdateParams message type.
func (m *MsgUpdateEGFParams) Type() string { return TypeMsgUpdateEGFParams }

// GetSignBytes returns the raw bytes for a MsgUpdateParams message that
// the expected signer needs to sign.
func (m *MsgUpdateEGFParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateEGFParams) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

func (m *MsgUpdateEGFParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "authority")
	}
	if err := m.Params.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "params")
	}
	return nil
}

func NewMsgUpdateStore(authority string, updateStores []UpdateStore) *MsgUpdateStore {
	return &MsgUpdateStore{Authority: authority, UpdateStores: updateStores}
}

func (m *MsgUpdateStore) Route() string { return types.ModuleName }

func (m *MsgUpdateStore) Type() string { return TypeMsgUpdateStore }

func (m *MsgUpdateStore) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgUpdateStore) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

func (m *MsgUpdateStore) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "authority")
	}
	if len(m.UpdateStores) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("stores are empty")
	}
	for _, updateStore := range m.UpdateStores {
		if len(updateStore.Space) == 0 {
			return sdkerrors.ErrInvalidRequest.Wrap("store space is empty")
		}
		if len(updateStore.Key) == 0 {
			return sdkerrors.ErrInvalidRequest.Wrap("store key is empty")
		}
		if _, err := hex.DecodeString(updateStore.Key); err != nil {
			return sdkerrors.ErrInvalidRequest.Wrap("invalid store key")
		}
		if len(updateStore.OldValue) > 0 {
			if _, err := hex.DecodeString(updateStore.OldValue); err != nil {
				return sdkerrors.ErrInvalidRequest.Wrap("invalid old store value")
			}
		}
		if len(updateStore.Value) > 0 {
			if _, err := hex.DecodeString(updateStore.Value); err != nil {
				return sdkerrors.ErrInvalidRequest.Wrap("invalid store value")
			}
		}
	}
	return nil
}

func (us *UpdateStore) String() string {
	out, _ := json.Marshal(us)
	return string(out)
}

func (us *UpdateStore) KeyToBytes() []byte {
	b, err := hex.DecodeString(us.Key)
	if err != nil {
		panic(err)
	}
	return b
}

func (us *UpdateStore) OldValueToBytes() []byte {
	if len(us.OldValue) == 0 {
		return []byte{}
	}
	b, err := hex.DecodeString(us.OldValue)
	if err != nil {
		panic(err)
	}
	return b
}

func (us *UpdateStore) ValueToBytes() []byte {
	if len(us.Value) == 0 {
		return []byte{}
	}
	b, err := hex.DecodeString(us.Value)
	if err != nil {
		panic(err)
	}
	return b
}

func NewMsgUpdateSwitchParams(authority string, params SwitchParams) *MsgUpdateSwitchParams {
	return &MsgUpdateSwitchParams{Authority: authority, Params: params}
}

// Route returns the MsgUpdateParams message route.
func (m *MsgUpdateSwitchParams) Route() string { return types.ModuleName }

// Type returns the MsgUpdateParams message type.
func (m *MsgUpdateSwitchParams) Type() string { return TypeMsgUpdateSwitchParams }

// GetSignBytes returns the raw bytes for a MsgUpdateParams message that
// the expected signer needs to sign.
func (m *MsgUpdateSwitchParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateSwitchParams) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

func (m *MsgUpdateSwitchParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "authority")
	}
	if err := m.Params.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "params")
	}
	return nil
}
