package entity

import (
	"strings"
	"time"

	"sipon-be/internal/modules/article/domain/article/constant"
	"sipon-be/internal/shared/kernel"
)

type Category struct {
	ID        string
	Name      string
	Slug      string
	IsActive  bool
	SortOrder int
	CreatedBy *string
	UpdatedBy *string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewCategory(id, name, slug string, createdBy *string) (*Category, error) {
	if strings.TrimSpace(id) == "" {
		return nil, kernel.New(constant.CodeCategoryNotFound)
	}
	if strings.TrimSpace(name) == "" {
		return nil, kernel.New(constant.CodeCategoryNameRequired)
	}
	if strings.TrimSpace(slug) == "" {
		return nil, kernel.New(constant.CodeCategorySlugRequired)
	}

	now := time.Now()
	return &Category{
		ID:        id,
		Name:      strings.TrimSpace(name),
		Slug:      strings.TrimSpace(slug),
		IsActive:  true,
		CreatedBy: normalizeCategoryStr(createdBy),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (c *Category) SoftDelete() {
	now := time.Now()
	c.DeletedAt = &now
	c.UpdatedAt = now
}

func normalizeCategoryStr(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
