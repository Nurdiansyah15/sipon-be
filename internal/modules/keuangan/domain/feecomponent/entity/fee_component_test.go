package entity_test

import (
	"testing"

	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	"sipon-be/internal/modules/keuangan/domain/feecomponent/entity"
)

func TestNewFeeComponent(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		code      string
		fn        string
		feeType   feeConst.FeeComponentType
		amount    float64
		createdBy string
		wantErr   bool
	}{
		{"valid component", "fc-1", "SPP", "SPP Bulanan", feeConst.FeeTypeSPP, 150000, "user-1", false},
		{"empty id", "", "SPP", "SPP Bulanan", feeConst.FeeTypeSPP, 150000, "user-1", true},
		{"empty code", "fc-2", "", "SPP Bulanan", feeConst.FeeTypeSPP, 150000, "user-1", true},
		{"empty name", "fc-3", "SPP", "", feeConst.FeeTypeSPP, 150000, "user-1", true},
		{"empty createdBy", "fc-4", "SPP", "SPP Bulanan", feeConst.FeeTypeSPP, 150000, "", true},
		{"invalid type", "fc-5", "SPP", "SPP Bulanan", feeConst.FeeComponentType("invalid"), 150000, "user-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc, err := entity.NewFeeComponent(tt.id, tt.code, tt.fn, tt.feeType, tt.amount, tt.createdBy)
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
			if !fc.IsActive {
				t.Error("new component should be active")
			}
		})
	}
}

func TestFeeComponentDeactivate(t *testing.T) {
	fc, _ := entity.NewFeeComponent("fc-1", "SPP", "SPP", feeConst.FeeTypeSPP, 150000, "user-1")
	fc.Deactivate()
	if fc.IsActive {
		t.Error("fee component should be inactive after deactivation")
	}
}

func TestFeeComponentActivate(t *testing.T) {
	fc, _ := entity.NewFeeComponent("fc-1", "SPP", "SPP", feeConst.FeeTypeSPP, 150000, "user-1")
	fc.Deactivate()
	fc.Activate()
	if !fc.IsActive {
		t.Error("fee component should be active after activation")
	}
}

func TestFeeComponentSoftDelete(t *testing.T) {
	fc, _ := entity.NewFeeComponent("fc-1", "SPP", "SPP", feeConst.FeeTypeSPP, 150000, "user-1")
	fc.SoftDelete()
	if fc.DeletedAt == nil {
		t.Error("deletedAt should not be nil after soft delete")
	}
}

func TestFeeComponentUpdate(t *testing.T) {
	fc, _ := entity.NewFeeComponent("fc-1", "SPP", "SPP", feeConst.FeeTypeSPP, 150000, "user-1")
	oldUpdatedAt := fc.UpdatedAt
	desc := "updated description"
	pt := feeConst.PeriodType("monthly")
	fc.Update("SPP Updated", 200000, true, &pt, &desc)
	if fc.Name != "SPP Updated" {
		t.Error("name should be updated")
	}
	if fc.Amount != 200000 {
		t.Error("amount should be updated")
	}
	if !fc.UpdatedAt.After(oldUpdatedAt) {
		t.Error("updatedAt should be updated")
	}
}
