package entity

import (
	"strings"
	"time"

	"sipon-be/internal/modules/article/domain/article/constant"
	"sipon-be/internal/shared/kernel"
)

type SourceCategory struct {
	ID                string
	SourceID          string
	CategoryKey       string
	URLSuffix         *string
	URLOverride       *string
	ArticleLimit      int
	IsActive          bool
	ArticleCategoryID *string
	Keywords          []string
	LastScrapedAt     *time.Time
	UpdatedBy         *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type SourceCategoryParams struct {
	ID                string
	SourceID          string
	CategoryKey       string
	URLSuffix         *string
	URLOverride       *string
	ArticleLimit      int
	IsActive          bool
	ArticleCategoryID *string
	Keywords          []string
}

func NewSourceCategory(params SourceCategoryParams) (*SourceCategory, error) {
	if strings.TrimSpace(params.SourceID) == "" {
		return nil, kernel.New(constant.CodeSourceNotFound)
	}

	key := strings.TrimSpace(params.CategoryKey)
	if key == "" {
		return nil, kernel.New(constant.CodeSourceCategoryKeyRequired)
	}

	limit := params.ArticleLimit
	if limit <= 0 {
		limit = 10
	}

	now := time.Now()
	return &SourceCategory{
		ID:                params.ID,
		SourceID:          params.SourceID,
		CategoryKey:       key,
		URLSuffix:         normalizeSourceStr(params.URLSuffix),
		URLOverride:       normalizeSourceStr(params.URLOverride),
		ArticleLimit:      limit,
		IsActive:          params.IsActive,
		ArticleCategoryID: normalizeSourceStr(params.ArticleCategoryID),
		Keywords:          params.Keywords,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}
