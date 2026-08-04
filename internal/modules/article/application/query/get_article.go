package query

import (
	"context"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	ports "sipon-be/internal/modules/article/application/ports"
	"sipon-be/internal/modules/article/application/thumbnail"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articleentity "sipon-be/internal/modules/article/domain/article/entity"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
)

type GetArticleUseCase struct {
	articleRepo  articlerepo.ArticleRepository
	fileUploader ports.FileUploader
}

func NewGetArticleUseCase(articleRepo articlerepo.ArticleRepository, fileUploader ports.FileUploader) *GetArticleUseCase {
	return &GetArticleUseCase{articleRepo: articleRepo, fileUploader: fileUploader}
}

func (uc *GetArticleUseCase) Execute(ctx context.Context, articleID string) (*dto.ArticleDetailResponse, error) {
	article, err := uc.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeArticleNotFound)
	}
	return mapArticleToDetail(article, uc.fileUploader), nil
}

func mapArticleToDetail(a *articleentity.Article, uploader ports.FileUploader) *dto.ArticleDetailResponse {
	return &dto.ArticleDetailResponse{
		ID:           a.ID,
		Title:        a.Title,
		Content:      a.Content,
		Summary:      a.Summary,
		CategoryID:   a.CategoryID,
		CategoryName: a.CategoryName,
		Status:       string(a.Status),
		Author:       a.Author,
		ThumbnailURL: resolveThumbnailURL(uploader, a.ThumbnailURL),
		OriginalURL:  a.OriginalURL,
		ViewCount:    a.ViewCount,
		IsFeatured:   a.IsFeatured,
		CreatedBy:    a.CreatedBy,
		UpdatedBy:    a.UpdatedBy,
		PublishedAt:  a.PublishedAt,
		ArchivedAt:   a.ArchivedAt,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

func resolveThumbnailURL(uploader ports.FileUploader, storedKey *string) *string {
	if storedKey == nil || *storedKey == "" {
		return storedKey
	}

	var s3Resolver func(key string) string
	if uploader != nil {
		s3Resolver = uploader.PublicURL
	}

	return thumbnail.FromStorage(storedKey, s3Resolver)
}
