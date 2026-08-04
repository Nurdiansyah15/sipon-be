package entity

import (
	"strings"
	"time"

	"sipon-be/internal/modules/article/domain/article/constant"
	"sipon-be/internal/shared/kernel"
)

type Source struct {
	ID            string
	Key           string
	Name          string
	BaseURL       string
	AutoPublish   bool
	IsActive      bool
	LastScrapedAt *time.Time
	CreatedBy     *string
	UpdatedBy     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time

	Selectors  *SourceSelector
	Categories []*SourceCategory
}

type SourceParams struct {
	ID          string
	Key         string
	Name        string
	BaseURL     string
	AutoPublish bool
	IsActive    bool
	CreatedBy   *string
}

func NewSource(params SourceParams) (*Source, error) {
	key := strings.TrimSpace(params.Key)
	if key == "" {
		return nil, kernel.New(constant.CodeSourceKeyRequired)
	}

	baseURL := strings.TrimSpace(params.BaseURL)
	if baseURL == "" {
		return nil, kernel.New(constant.CodeSourceURLRequired)
	}

	now := time.Now()
	return &Source{
		ID:          params.ID,
		Key:         key,
		Name:        strings.TrimSpace(params.Name),
		BaseURL:     baseURL,
		AutoPublish: params.AutoPublish,
		IsActive:    params.IsActive,
		CreatedBy:   normalizeSourceStr(params.CreatedBy),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *Source) SoftDelete() {
	now := time.Now()
	s.DeletedAt = &now
	s.UpdatedAt = now
}

func normalizeSourceStr(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
