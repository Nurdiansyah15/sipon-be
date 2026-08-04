package entity

import (
	"strings"
	"time"

	"sipon-be/internal/modules/article/domain/article/constant"
	"sipon-be/internal/shared/kernel"
)

type Article struct {
	ID           string
	Title        string
	Content      string
	Summary      *string
	CategoryID   *string
	CategoryName *string
	Status       constant.ArticleStatus
	Author       string
	ThumbnailURL *string
	ViewCount    int
	IsFeatured   bool
	CreatedBy    *string
	UpdatedBy    *string
	PublishedAt  *time.Time
	ArchivedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type ArticleParams struct {
	ID           string
	Title        string
	Content      string
	Summary      *string
	CategoryID   *string
	Author       string
	ThumbnailURL *string
	IsFeatured   bool
	CreatedBy    *string
}

func NewArticle(params ArticleParams) (*Article, error) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, kernel.New(constant.CodeArticleNotFound)
	}
	if strings.TrimSpace(params.Title) == "" {
		return nil, kernel.New(constant.CodeArticleTitleRequired)
	}
	if strings.TrimSpace(params.Author) == "" {
		return nil, kernel.New(constant.CodeArticleAuthorRequired)
	}

	now := time.Now()
	a := &Article{
		ID:           params.ID,
		Title:        strings.TrimSpace(params.Title),
		Content:      strings.TrimSpace(params.Content),
		Summary:      normalizeOptionalStr(params.Summary),
		CategoryID:   normalizeOptionalStr(params.CategoryID),
		Status:       constant.ArticleStatusDraft,
		Author:       strings.TrimSpace(params.Author),
		ThumbnailURL: normalizeOptionalStr(params.ThumbnailURL),
		IsFeatured:   params.IsFeatured,
		CreatedBy:    normalizeOptionalStr(params.CreatedBy),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return a, nil
}

func (a *Article) Publish() error {
	if a.Status != constant.ArticleStatusDraft {
		return kernel.New(constant.CodeArticleInvalidStatus)
	}
	now := time.Now()
	a.Status = constant.ArticleStatusPublished
	a.PublishedAt = &now
	a.UpdatedAt = now
	return nil
}

func (a *Article) Archive() error {
	if a.Status != constant.ArticleStatusPublished {
		return kernel.New(constant.CodeArticleInvalidStatus)
	}
	now := time.Now()
	a.Status = constant.ArticleStatusArchived
	a.ArchivedAt = &now
	a.UpdatedAt = now
	return nil
}

func (a *Article) CanEdit() bool {
	return a.Status == constant.ArticleStatusDraft ||
		a.Status == constant.ArticleStatusPublished
}

func (a *Article) CanDelete() bool {
	return a.Status == constant.ArticleStatusDraft
}

func (a *Article) EnsureEditable() error {
	if !a.CanEdit() {
		return kernel.New(constant.CodeArticleCannotEdit)
	}
	return nil
}

func (a *Article) EnsureDeletable() error {
	if !a.CanDelete() {
		return kernel.New(constant.CodeArticleCannotDelete)
	}
	return nil
}

func (a *Article) SoftDelete() {
	now := time.Now()
	a.DeletedAt = &now
	a.UpdatedAt = now
}

func normalizeOptionalStr(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
