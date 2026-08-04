package command

import (
	"context"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	ports "sipon-be/internal/modules/article/application/ports"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articleentity "sipon-be/internal/modules/article/domain/article/entity"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type CreateArticleUseCase struct {
	articleRepo  articlerepo.ArticleRepository
	fileUploader ports.FileUploader
	transactor   ports.Transactor
}

func NewCreateArticleUseCase(
	articleRepo articlerepo.ArticleRepository,
	fileUploader ports.FileUploader,
	transactor ports.Transactor,
) *CreateArticleUseCase {
	return &CreateArticleUseCase{
		articleRepo:  articleRepo,
		fileUploader: fileUploader,
		transactor:   transactor,
	}
}

func (uc *CreateArticleUseCase) Execute(ctx context.Context, req dto.CreateArticleRequest, userID string) (*dto.ArticleMutationResponse, error) {
	articleID := uuid.NewString()

	article, err := articleentity.NewArticle(articleentity.ArticleParams{
		ID:           articleID,
		Title:        req.Title,
		Content:      req.Content,
		Summary:      req.Summary,
		CategoryID:   req.CategoryID,
		Author:       req.Author,
		ThumbnailURL: normalizeThumbnailKeyPtr(uc.fileUploader, req.ThumbnailURL),
		IsFeatured:   req.IsFeatured,
		CreatedBy:    &userID,
	})
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if req.Status != nil && *req.Status == "published" {
		if err := article.Publish(); err != nil {
			return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
		}
	}

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.articleRepo.Save(txCtx, article)
	}); err != nil {
		return nil, application.WrapConflictErr(err, articleconst.CodeArticlePersistenceFailed)
	}

	confirmThumbnailKey(ctx, uc.fileUploader, article.ThumbnailURL)

	return &dto.ArticleMutationResponse{ID: articleID}, nil
}

func normalizeThumbnailKeyPtr(uploader ports.FileUploader, raw *string) *string {
	if raw == nil {
		return nil
	}
	v := *raw
	if v == "" {
		return nil
	}
	return &v
}

func confirmThumbnailKey(ctx context.Context, uploader ports.FileUploader, key *string) {
	if uploader == nil || key == nil || *key == "" {
		return
	}
	_ = uploader.ConfirmUpload(ctx, *key)
}
