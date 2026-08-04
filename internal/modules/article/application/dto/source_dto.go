package dto

import "time"

type SourceSelectorInput struct {
	ContentSelector *string `json:"content_selector"`
	AuthorSelector  *string `json:"author_selector"`
	TagsSelector    *string `json:"tags_selector"`
}

type SourceSelectorItem struct {
	ContentSelector *string `json:"content_selector"`
	AuthorSelector  *string `json:"author_selector"`
	TagsSelector    *string `json:"tags_selector"`
}

type SourceCategoryItem struct {
	ID                string     `json:"id"`
	CategoryKey       string     `json:"category_key"`
	URLSuffix         *string    `json:"url_suffix"`
	URLOverride       *string    `json:"url_override"`
	ArticleLimit      int        `json:"article_limit"`
	IsActive          bool       `json:"is_active"`
	ArticleCategoryID *string    `json:"article_category_id"`
	Keywords          []string   `json:"keywords,omitempty"`
	LastScrapedAt     *time.Time `json:"last_scraped_at"`
}

type CreateSourceRequest struct {
	Name        string                `json:"name" binding:"required"`
	Key         string                `json:"key" binding:"required"`
	BaseURL     string                `json:"base_url" binding:"required"`
	AutoPublish bool                  `json:"auto_publish"`
	IsActive    bool                  `json:"is_active"`
	Selectors   *SourceSelectorInput  `json:"selectors"`
}

type UpdateSourceRequest struct {
	Name        *string               `json:"name"`
	Key         *string               `json:"key"`
	BaseURL     *string               `json:"base_url"`
	AutoPublish *bool                 `json:"auto_publish"`
	IsActive    *bool                 `json:"is_active"`
	Selectors   *SourceSelectorInput  `json:"selectors"`
}

type SourceMutationResponse struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

type SourceListItem struct {
	ID            string               `json:"id"`
	Key           string               `json:"key"`
	Name          string               `json:"name"`
	BaseURL       string               `json:"base_url"`
	AutoPublish   bool                 `json:"auto_publish"`
	IsActive      bool                 `json:"is_active"`
	LastScrapedAt *time.Time           `json:"last_scraped_at"`
	Selectors     *SourceSelectorItem  `json:"selectors"`
	Categories    []SourceCategoryItem `json:"categories"`
	CreatedAt     time.Time            `json:"created_at"`
}

type CreateSourceCategoryRequest struct {
	CategoryKey       string   `json:"category_key" binding:"required"`
	URLSuffix         *string  `json:"url_suffix"`
	URLOverride       *string  `json:"url_override"`
	ArticleLimit      int      `json:"article_limit"`
	IsActive          bool     `json:"is_active"`
	ArticleCategoryID *string  `json:"article_category_id"`
	Keywords          []string `json:"keywords,omitempty"`
}

type UpdateSourceCategoryRequest struct {
	CategoryKey       *string   `json:"category_key"`
	URLSuffix         *string   `json:"url_suffix"`
	URLOverride       *string   `json:"url_override"`
	ArticleLimit      *int      `json:"article_limit"`
	IsActive          *bool     `json:"is_active"`
	ArticleCategoryID *string   `json:"article_category_id"`
	Keywords          *[]string `json:"keywords"`
}

type SourceCategoryMutationResponse struct {
	ID string `json:"id"`
}

type ScrapeResult struct {
	Scraped     int                  `json:"scraped"`
	Skipped     int                  `json:"skipped"`
	Categories  []ScrapeCategoryItem `json:"categories"`
}

type ScrapeCategoryItem struct {
	CategoryKey string `json:"category_key"`
	Fetched     int    `json:"fetched"`
	Saved       int    `json:"saved"`
	Skipped     int    `json:"skipped"`
	Error       string `json:"error,omitempty"`
}
