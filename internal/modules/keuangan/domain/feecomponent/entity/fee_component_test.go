package entity_test

import (
	"testing"

	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	"sipon-be/internal/modules/keuangan/domain/feecomponent/entity"
)

const (
	testRevenueAccountID    = "rev-acc-1"
	testReceivableAccountID = "rec-acc-1"
)

func TestNewFeeComponent(t *testing.T) {
	tests := []struct {
		name                string
		id                  string
		code                string
		fn                  string
		revenueAccountID    string
		receivableAccountID string
		amount              float64
		createdBy           string
		wantErr             bool
	}{
		{"valid component", "fc-1", "SPP", "SPP Bulanan", testRevenueAccountID, testReceivableAccountID, 150000, "user-1", false},
		{"empty id", "", "SPP", "SPP Bulanan", testRevenueAccountID, testReceivableAccountID, 150000, "user-1", true},
		{"empty code", "fc-2", "", "SPP Bulanan", testRevenueAccountID, testReceivableAccountID, 150000, "user-1", true},
		{"empty name", "fc-3", "SPP", "", testRevenueAccountID, testReceivableAccountID, 150000, "user-1", true},
		{"empty createdBy", "fc-4", "SPP", "SPP Bulanan", testRevenueAccountID, testReceivableAccountID, 150000, "", true},
		{"empty revenue account", "fc-5", "SPP", "SPP Bulanan", "", testReceivableAccountID, 150000, "user-1", true},
		{"empty receivable account", "fc-6", "SPP", "SPP Bulanan", testRevenueAccountID, "", 150000, "user-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc, err := entity.NewFeeComponent(tt.id, tt.code, tt.fn, tt.revenueAccountID, tt.receivableAccountID, tt.amount, tt.createdBy)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fc.ID != tt.id {
				t.Errorf("expected ID %s, got %s", tt.id, fc.ID)
			}
			if fc.RevenueAccountID != tt.revenueAccountID {
				t.Errorf("expected revenue account ID %s, got %s", tt.revenueAccountID, fc.RevenueAccountID)
			}
			if fc.ReceivableAccountID != tt.receivableAccountID {
				t.Errorf("expected receivable account ID %s, got %s", tt.receivableAccountID, fc.ReceivableAccountID)
			}
			if !fc.IsActive {
				t.Error("new component should be active")
			}
		})
	}
}

func TestFeeComponentDeactivate(t *testing.T) {
	fc, _ := entity.NewFeeComponent("fc-1", "SPP", "SPP", testRevenueAccountID, testReceivableAccountID, 150000, "user-1")
	fc.Deactivate()
	if fc.IsActive {
		t.Error("fee component should be inactive after deactivation")
	}
}

func TestFeeComponentActivate(t *testing.T) {
	fc, _ := entity.NewFeeComponent("fc-1", "SPP", "SPP", testRevenueAccountID, testReceivableAccountID, 150000, "user-1")
	fc.Deactivate()
	fc.Activate()
	if !fc.IsActive {
		t.Error("fee component should be active after activation")
	}
}

func TestFeeComponentSoftDelete(t *testing.T) {
	fc, _ := entity.NewFeeComponent("fc-1", "SPP", "SPP", testRevenueAccountID, testReceivableAccountID, 150000, "user-1")
	fc.SoftDelete()
	if fc.DeletedAt == nil {
		t.Error("deletedAt should not be nil after soft delete")
	}
}

func TestFeeComponentUpdate(t *testing.T) {
	fc, _ := entity.NewFeeComponent("fc-1", "SPP", "SPP", testRevenueAccountID, testReceivableAccountID, 150000, "user-1")
	oldUpdatedAt := fc.UpdatedAt
	desc := "updated description"
	pt := feeConst.PeriodType("monthly")
	fc.Update("rev-acc-2", "rec-acc-2", "SPP Updated", 200000, true, &pt, &desc)
	if fc.Name != "SPP Updated" {
		t.Error("name should be updated")
	}
	if fc.Amount != 200000 {
		t.Error("amount should be updated")
	}
	if fc.RevenueAccountID != "rev-acc-2" {
		t.Error("revenue account id should be updated")
	}
	if fc.ReceivableAccountID != "rec-acc-2" {
		t.Error("receivable account id should be updated")
	}
	if !fc.UpdatedAt.After(oldUpdatedAt) {
		t.Error("updatedAt should be updated")
	}
}
