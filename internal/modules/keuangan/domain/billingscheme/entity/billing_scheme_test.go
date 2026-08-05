package entity_test

import (
	"testing"

	"sipon-be/internal/modules/keuangan/domain/billingscheme/entity"
)

func TestNewBillingScheme(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		schemeName string
		createdBy string
		wantErr   bool
	}{
		{"valid scheme", "bs-1", "Skema SPP", "user-1", false},
		{"empty id", "", "Skema SPP", "user-1", true},
		{"empty name", "bs-2", "", "user-1", true},
		{"empty createdBy", "bs-3", "Skema SPP", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs, err := entity.NewBillingScheme(tt.id, tt.schemeName, tt.createdBy)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bs.IsActive {
				t.Error("new scheme should be active")
			}
		})
	}
}

func TestBillingSchemeUpdate(t *testing.T) {
	bs, _ := entity.NewBillingScheme("bs-1", "Skema SPP", "user-1")
	desc := "description"
	bs.Update("Skema Updated", &desc)
	if bs.Name != "Skema Updated" {
		t.Error("name should be updated")
	}
	if bs.Description == nil || *bs.Description != "description" {
		t.Error("description should be updated")
	}
}

func TestBillingSchemeDeactivate(t *testing.T) {
	bs, _ := entity.NewBillingScheme("bs-1", "Skema SPP", "user-1")
	bs.Deactivate()
	if bs.IsActive {
		t.Error("scheme should be inactive after deactivation")
	}
}

func TestBillingSchemeActivate(t *testing.T) {
	bs, _ := entity.NewBillingScheme("bs-1", "Skema SPP", "user-1")
	bs.Deactivate()
	bs.Activate()
	if !bs.IsActive {
		t.Error("scheme should be active after activation")
	}
}

func TestBillingSchemeAddItem(t *testing.T) {
	bs, _ := entity.NewBillingScheme("bs-1", "Skema SPP", "user-1")

	t.Run("add items", func(t *testing.T) {
		item1, _ := entity.NewBillingSchemeItem("bsi-1", "bs-1", "fc-1", nil, true, 1)
		item2, _ := entity.NewBillingSchemeItem("bsi-2", "bs-1", "fc-2", nil, false, 2)

		if err := bs.AddItem(item1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := bs.AddItem(item2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(bs.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(bs.Items))
		}
	})

	t.Run("add duplicate", func(t *testing.T) {
		item3, _ := entity.NewBillingSchemeItem("bsi-3", "bs-1", "fc-1", nil, true, 3)
		if err := bs.AddItem(item3); err == nil {
			t.Error("expected error when adding duplicate fee component")
		}
	})
}

func TestBillingSchemeRemoveItem(t *testing.T) {
	bs, _ := entity.NewBillingScheme("bs-1", "Skema SPP", "user-1")
	item1, _ := entity.NewBillingSchemeItem("bsi-1", "bs-1", "fc-1", nil, true, 1)
	item2, _ := entity.NewBillingSchemeItem("bsi-2", "bs-1", "fc-2", nil, false, 2)
	bs.AddItem(item1)
	bs.AddItem(item2)

	if err := bs.RemoveItem("bsi-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bs.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(bs.Items))
	}
	if bs.Items[0].ID != "bsi-2" {
		t.Error("wrong item removed")
	}

	if err := bs.RemoveItem("bsi-non-existent"); err == nil {
		t.Error("expected error when removing non-existent item")
	}
}

func TestNewBillingSchemeItem(t *testing.T) {
	amt := 200000.0
	tests := []struct {
		name            string
		id              string
		billingSchemeID string
		feeComponentID  string
		amountOverride  *float64
		isRequired      bool
		sortOrder       int
		wantErr         bool
	}{
		{"valid item", "bsi-1", "bs-1", "fc-1", nil, true, 1, false},
		{"with amount override", "bsi-2", "bs-1", "fc-2", &amt, false, 2, false},
		{"empty id", "", "bs-1", "fc-1", nil, true, 1, true},
		{"empty billingSchemeID", "bsi-3", "", "fc-1", nil, true, 1, true},
		{"empty feeComponentID", "bsi-4", "bs-1", "", nil, true, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := entity.NewBillingSchemeItem(tt.id, tt.billingSchemeID, tt.feeComponentID, tt.amountOverride, tt.isRequired, tt.sortOrder)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if item.ID != tt.id {
				t.Errorf("expected ID %s, got %s", tt.id, item.ID)
			}
		})
	}
}

func TestBillingSchemeItemGetEffectiveAmount(t *testing.T) {
	t.Run("with override", func(t *testing.T) {
		amt := 200000.0
		item, _ := entity.NewBillingSchemeItem("bsi-1", "bs-1", "fc-1", &amt, true, 1)
		effective := item.GetEffectiveAmount(150000)
		if effective != 200000 {
			t.Errorf("expected 200000, got %f", effective)
		}
	})

	t.Run("no override", func(t *testing.T) {
		item, _ := entity.NewBillingSchemeItem("bsi-1", "bs-1", "fc-1", nil, true, 1)
		effective := item.GetEffectiveAmount(150000)
		if effective != 150000 {
			t.Errorf("expected 150000, got %f", effective)
		}
	})
}
