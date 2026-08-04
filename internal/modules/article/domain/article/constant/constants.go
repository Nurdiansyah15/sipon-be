package constant

import "sipon-be/internal/shared/kernel"

type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusPublished ArticleStatus = "published"
	ArticleStatusArchived  ArticleStatus = "archived"
)

func (s ArticleStatus) IsValid() bool {
	switch s {
	case ArticleStatusDraft, ArticleStatusPublished, ArticleStatusArchived:
		return true
	}
	return false
}

const (
	CodeArticleNotFound          kernel.Code = "ARTICLE_NOT_FOUND"
	CodeArticlePersistenceFailed kernel.Code = "ARTICLE_PERSISTENCE_FAILED"
	CodeArticleQueryFailed       kernel.Code = "ARTICLE_QUERY_FAILED"
	CodeArticleInvalidStatus     kernel.Code = "ARTICLE_INVALID_STATUS"
	CodeArticleCannotEdit        kernel.Code = "ARTICLE_CANNOT_EDIT"
	CodeArticleCannotDelete      kernel.Code = "ARTICLE_CANNOT_DELETE"
	CodeArticleTitleRequired     kernel.Code = "ARTICLE_TITLE_REQUIRED"
	CodeArticleAuthorRequired    kernel.Code = "ARTICLE_AUTHOR_REQUIRED"

	CodeCategoryNotFound          kernel.Code = "CATEGORY_NOT_FOUND"
	CodeCategoryDuplicateSlug     kernel.Code = "CATEGORY_DUPLICATE_SLUG"
	CodeCategoryNameRequired      kernel.Code = "CATEGORY_NAME_REQUIRED"
	CodeCategorySlugRequired      kernel.Code = "CATEGORY_SLUG_REQUIRED"
	CodeCategoryPersistenceFailed kernel.Code = "CATEGORY_PERSISTENCE_FAILED"
	CodeCategoryQueryFailed       kernel.Code = "CATEGORY_QUERY_FAILED"
)
