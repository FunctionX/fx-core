package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	migratetypes "github.com/functionx/fx-core/v2/x/migrate/types"
)

type BankMigrate struct {
	bankKeeper migratetypes.BankKeeper
}

func NewBankMigrate(bk migratetypes.BankKeeper) MigrateI {
	return &BankMigrate{bankKeeper: bk}
}

func (m *BankMigrate) Validate(_ sdk.Context, _ Keeper, _ sdk.AccAddress, _ common.Address) error {
	return nil
}

func (m *BankMigrate) Execute(ctx sdk.Context, _ Keeper, from sdk.AccAddress, to common.Address) error {
	balances := m.bankKeeper.GetAllBalances(ctx, from)
	if balances.IsZero() {
		return nil
	}
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			migratetypes.EventTypeMigrateBankSend,
			sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
		),
	})
	return m.bankKeeper.SendCoins(ctx, from, to.Bytes(), balances)
}
