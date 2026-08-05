package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	"sipon-be/internal/shared/kernel"
)

type BillingScheme struct {
	ID          string
	Name        string
	Description *string
	IsActive    bool
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Items       []*BillingSchemeItem
}

func NewBillingScheme(id, name string, createdBy string) (*BillingScheme, error) {
	if id == "" || name == "" || createdBy == "" {
		return nil, kernel.New(constant.CodeBillingSchemeNotFound)
	}
	now := time.Now()
	return &BillingScheme{
		ID:        id,
		Name:      name,
		IsActive:  true,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *BillingScheme) Update(name string, description *string) {
	s.Name = name
	s.Description = description
	s.UpdatedAt = time.Now()
}

func (s *BillingScheme) Deactivate() {
	s.IsActive = false
	s.UpdatedAt = time.Now()
}

func (s *BillingScheme) Activate() {
	s.IsActive = true
	s.UpdatedAt = time.Now()
}

func (s *BillingScheme) AddItem(item *BillingSchemeItem) error {
	for _, existing := range s.Items {
		if existing.FeeComponentID == item.FeeComponentID {
			return kernel.New(constant.CodeSchemeItemDuplicate)
		}
	}
	s.Items = append(s.Items, item)
	return nil
}

func (s *BillingScheme) RemoveItem(itemID string) error {
	for i, item := range s.Items {
		if item.ID == itemID {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			return nil
		}
	}
	return kernel.New(constant.CodeSchemeItemNotFound)
}
