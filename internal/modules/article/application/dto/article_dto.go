package dto

import "time"

type PaginationParams struct {
	Page     int    `form:"page" json:"page" binding:"min=1"`
	Limit    int    `form:"limit" json:"limit" binding:"min=1,max=100"`
	SortBy   string `form:"sort_by" json:"sort_by"`
	SortType string `form:"sort_type" json:"sort_type"`
}

type Meta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	TotalPages  int `json:"total_pages"`
}

// ── Article Requests ──────────────────────────────────────────────────────────

type CreateArticleRequest struct {
	Title        string  `json:"title" binding:"required"`
	Content      string  `json:"content" binding:"required"`
	Summary      *string `json:"summary"`
	CategoryID   *string `json:"category_id"`
	Author       string  `json:"author" binding:"required"`
	ThumbnailURL *string `json:"thumbnail_url"`
	IsFeatured   bool    `json:"is_featured"`
	Status       *string `json:"status"`
}

type UpdateArticleRequest struct {
	Title        *string `json:"title"`
	Content      *string `json:"content"`
	Summary      *string `json:"summary"`
	CategoryID   *string `json:"category_id"`
	Author       *string `json:"author"`
	ThumbnailURL *string `json:"thumbnail_url"`
	IsFeatured   *bool   `json:"is_featured"`
}

// ── Article Query ─────────────────────────────────────────────────────────────

type ListArticlesQuery struct {
	Status     *string `form:"status"`
	CategoryID *string `form:"category_id"`
	Search     *string `form:"q"`
	PaginationParams
}

// ── Article Response ──────────────────────────────────────────────────────────

type ArticleMutationResponse struct {
	ID string `json:"id"`
}

type ArticleListItem struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Status       string     `json:"status"`
	CategoryID   *string    `json:"category_id,omitempty"`
	CategoryName *string    `json:"category_name,omitempty"`
	Author       string     `json:"author"`
	ThumbnailURL *string    `json:"thumbnail_url,omitempty"`
	IsFeatured   bool       `json:"is_featured"`
	ViewCount    int        `json:"view_count"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type ArticleDetailResponse struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	Summary      *string    `json:"summary,omitempty"`
	CategoryID   *string    `json:"category_id,omitempty"`
	CategoryName *string    `json:"category_name,omitempty"`
	Status       string     `json:"status"`
	Author       string     `json:"author"`
	ThumbnailURL *string    `json:"thumbnail_url,omitempty"`
	ViewCount    int        `json:"view_count"`
	IsFeatured   bool       `json:"is_featured"`
	CreatedBy    *string    `json:"created_by,omitempty"`
	UpdatedBy    *string    `json:"updated_by,omitempty"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	ArchivedAt   *time.Time `json:"archived_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ── Category ──────────────────────────────────────────────────────────────────

type CreateCategoryRequest struct {
	Name      string `json:"name" binding:"required"`
	Slug      string `json:"slug" binding:"required"`
	SortOrder *int   `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Name      *string `json:"name"`
	Slug      *string `json:"slug"`
	IsActive  *bool   `json:"is_active"`
	SortOrder *int    `json:"sort_order"`
}

type CategoryMutationResponse struct {
	ID string `json:"id"`
}

type CategoryItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	IsActive  bool   `json:"is_active"`
	SortOrder int    `json:"sort_order"`
}

// ── Media ─────────────────────────────────────────────────────────────────────

type PresignRequest struct {
	ContentType string `json:"content_type" binding:"required"`
}

type PresignResponse struct {
	PresignURL string `json:"presign_url"`
	Key        string `json:"key"`
	PublicURL  string `json:"public_url"`
	ExpiresIn  int    `json:"expires_in"`
}

type ConfirmUploadRequest struct {
	Key string `json:"key" binding:"required"`
}

type DeleteThumbnailRequest struct {
	Key string `json:"key" binding:"required"`
}
