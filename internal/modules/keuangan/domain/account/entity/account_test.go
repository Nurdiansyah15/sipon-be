package entity_test

import (
	"testing"

	"sipon-be/internal/modules/keuangan/domain/account/constant"
	"sipon-be/internal/modules/keuangan/domain/account/entity"
)

func TestNewAccount(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		code          string
		accName       string
		accType       constant.AccountType
		parentID      *string
		level         int
		normalBalance constant.NormalBalance
		subType       *constant.AccountSubType
		createdBy     string
		wantErr       bool
	}{
		{
			name: "valid account", id: "acc-1", code: "1101", accName: "Kas",
			accType: constant.TypeAsset, parentID: nil, level: 1,
			normalBalance: constant.BalanceDebit, subType: subTypePtr(constant.SubTypeCashBank), createdBy: "user-1", wantErr: false,
		},
		{"empty id", "", "1101", "Kas", constant.TypeAsset, nil, 1, constant.BalanceDebit, subTypePtr(constant.SubTypeCashBank), "user-1", true},
		{"empty code", "acc-2", "", "Kas", constant.TypeAsset, nil, 1, constant.BalanceDebit, subTypePtr(constant.SubTypeCashBank), "user-1", true},
		{"empty name", "acc-3", "1101", "", constant.TypeAsset, nil, 1, constant.BalanceDebit, subTypePtr(constant.SubTypeCashBank), "user-1", true},
		{"empty createdBy", "acc-4", "1101", "Kas", constant.TypeAsset, nil, 1, constant.BalanceDebit, subTypePtr(constant.SubTypeCashBank), "", true},
		{"level 0 not postable", "acc-5", "1", "Header", constant.TypeAsset, nil, 0, constant.BalanceDebit, nil, "user-1", false},
		{"with parent", "acc-6", "1102", "Bank", constant.TypeAsset, strPtr("acc-1"), 2, constant.BalanceDebit, subTypePtr(constant.SubTypeCashBank), "user-1", false},
		{"postable without sub type", "acc-7", "1101", "Kas", constant.TypeAsset, nil, 1, constant.BalanceDebit, nil, "user-1", true},
		{"sub type invalid for type", "acc-8", "9999", "Akun Aneh", constant.TypeRevenue, nil, 1, constant.BalanceCredit, subTypePtr(constant.SubTypeCashBank), "user-1", true},
		{"revenue operating revenue", "acc-9", "4100", "Pendapatan SPP", constant.TypeRevenue, nil, 1, constant.BalanceCredit, subTypePtr(constant.SubTypeOperatingRevenue), "user-1", false},
		{"liability payable", "acc-10", "2101", "Utang Usaha", constant.TypeLiability, nil, 1, constant.BalanceCredit, subTypePtr(constant.SubTypePayable), "user-1", false},
		{"equity capital", "acc-11", "3100", "Modal", constant.TypeEquity, nil, 1, constant.BalanceCredit, subTypePtr(constant.SubTypeCapital), "user-1", false},
		{"expense operating expense", "acc-12", "5100", "Beban Gaji", constant.TypeExpense, nil, 1, constant.BalanceDebit, subTypePtr(constant.SubTypeOperatingExpense), "user-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := entity.NewAccount(tt.id, tt.code, tt.accName, tt.accType, tt.parentID, tt.level, tt.normalBalance, tt.subType, tt.createdBy)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if acc.IsSystem {
				t.Error("new account should not be system")
			}
			if !acc.IsActive {
				t.Error("new account should be active")
			}
			if tt.subType != nil && (acc.SubType == nil || *acc.SubType != *tt.subType) {
				t.Error("sub type should be set")
			}
		})
	}
}

func TestAccountUpdate(t *testing.T) {
	t.Run("non-system account", func(t *testing.T) {
		acc, _ := entity.NewAccount("acc-1", "1101", "Kas", constant.TypeAsset, nil, 1, constant.BalanceDebit, subTypePtr(constant.SubTypeCashBank), "user-1")
		desc := "new description"
		if err := acc.Update("Kas Updated", &desc, true, subTypePtr(constant.SubTypeReceivable)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if acc.Name != "Kas Updated" {
			t.Error("name should be updated")
		}
		if acc.SubType == nil || *acc.SubType != constant.SubTypeReceivable {
			t.Error("sub type should be updated")
		}
	})

	t.Run("update to invalid sub type for type", func(t *testing.T) {
		acc, _ := entity.NewAccount("acc-1", "1101", "Kas", constant.TypeAsset, nil, 1, constant.BalanceDebit, subTypePtr(constant.SubTypeCashBank), "user-1")
		if err := acc.Update("Kas Updated", nil, true, subTypePtr(constant.SubTypeOperatingRevenue)); err == nil {
			t.Error("expected error when updating to invalid sub type")
		}
	})

	t.Run("postable account must keep sub type", func(t *testing.T) {
		acc, _ := entity.NewAccount("acc-1", "1101", "Kas", constant.TypeAsset, nil, 1, constant.BalanceDebit, subTypePtr(constant.SubTypeCashBank), "user-1")
		if err := acc.Update("Kas Updated", nil, true, nil); err == nil {
			t.Error("expected error when postable account loses sub type")
		}
	})

	t.Run("system account", func(t *testing.T) {
		acc := &entity.Account{
			ID: "acc-sys", Code: "1101", Name: "Kas",
			Type: constant.TypeAsset, IsSystem: true, IsActive: true,
		}
		desc := "new description"
		if err := acc.Update("Kas Updated", &desc, true, subTypePtr(constant.SubTypeCashBank)); err == nil {
			t.Error("expected error when updating system account")
		}
	})
}

func TestAccountDeactivate(t *testing.T) {
	t.Run("non-system account", func(t *testing.T) {
		acc, _ := entity.NewAccount("acc-1", "1101", "Kas", constant.TypeAsset, nil, 1, constant.BalanceDebit, subTypePtr(constant.SubTypeCashBank), "user-1")
		if err := acc.Deactivate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if acc.IsActive {
			t.Error("account should be inactive after deactivation")
		}
	})

	t.Run("system account", func(t *testing.T) {
		acc := &entity.Account{
			ID: "acc-sys", Code: "1101", Name: "Kas",
			Type: constant.TypeAsset, IsSystem: true, IsActive: true,
		}
		if err := acc.Deactivate(); err == nil {
			t.Error("expected error when deactivating system account")
		}
	})
}

func TestAccountActivate(t *testing.T) {
	acc, _ := entity.NewAccount("acc-1", "1101", "Kas", constant.TypeAsset, nil, 1, constant.BalanceDebit, subTypePtr(constant.SubTypeCashBank), "user-1")
	if err := acc.Deactivate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	acc.Activate()
	if !acc.IsActive {
		t.Error("account should be active after activation")
	}
}

func TestAccountSoftDelete(t *testing.T) {
	t.Run("non-system account", func(t *testing.T) {
		acc, _ := entity.NewAccount("acc-1", "1101", "Kas", constant.TypeAsset, nil, 1, constant.BalanceDebit, subTypePtr(constant.SubTypeCashBank), "user-1")
		if err := acc.SoftDelete(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if acc.DeletedAt == nil {
			t.Error("deletedAt should be set after soft delete")
		}
	})

	t.Run("system account", func(t *testing.T) {
		acc := &entity.Account{
			ID: "acc-sys", Code: "1101", Name: "Kas",
			Type: constant.TypeAsset, IsSystem: true, IsActive: true,
		}
		if err := acc.SoftDelete(); err == nil {
			t.Error("expected error when soft deleting system account")
		}
	})
}

func TestAccountEnsurePostable(t *testing.T) {
	t.Run("postable and active", func(t *testing.T) {
		acc, _ := entity.NewAccount("acc-1", "1101", "Kas", constant.TypeAsset, nil, 1, constant.BalanceDebit, subTypePtr(constant.SubTypeCashBank), "user-1")
		if err := acc.EnsurePostable(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("not postable", func(t *testing.T) {
		acc, _ := entity.NewAccount("acc-1", "1", "Header", constant.TypeAsset, nil, 0, constant.BalanceDebit, nil, "user-1")
		if err := acc.EnsurePostable(); err == nil {
			t.Error("expected error for non-postable account")
		}
	})

	t.Run("not active", func(t *testing.T) {
		acc, _ := entity.NewAccount("acc-1", "1101", "Kas", constant.TypeAsset, nil, 1, constant.BalanceDebit, subTypePtr(constant.SubTypeCashBank), "user-1")
		if err := acc.Deactivate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := acc.EnsurePostable(); err == nil {
			t.Error("expected error for inactive account")
		}
	})
}

func TestIsValidSubTypeForType(t *testing.T) {
	if !constant.IsValidSubTypeForType(constant.TypeAsset, constant.SubTypeCashBank) {
		t.Error("cash_bank should be valid for asset")
	}
	if !constant.IsValidSubTypeForType(constant.TypeRevenue, constant.SubTypeOperatingRevenue) {
		t.Error("operating_revenue should be valid for revenue")
	}
	if constant.IsValidSubTypeForType(constant.TypeRevenue, constant.SubTypeCashBank) {
		t.Error("cash_bank should not be valid for revenue")
	}
	if constant.IsValidSubTypeForType(constant.TypeExpense, constant.SubTypeCapital) {
		t.Error("capital should not be valid for expense")
	}
}

func strPtr(s string) *string {
	return &s
}

func subTypePtr(st constant.AccountSubType) *constant.AccountSubType {
	return &st
}
