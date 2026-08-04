package command

import (
	"context"
	"time"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	ports "sipon-be/internal/modules/article/application/ports"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateArticleUseCase struct {
	articleRepo  articlerepo.ArticleRepository
	fileUploader ports.FileUploader
	transactor   ports.Transactor
}

func NewUpdateArticleUseCase(
	articleRepo articlerepo.ArticleRepository,
	fileUploader ports.FileUploader,
	transactor ports.Transactor,
) *UpdateArticleUseCase {
	return &UpdateArticleUseCase{
		articleRepo:  articleRepo,
		fileUploader: fileUploader,
		transactor:   transactor,
	}
}

func (uc *UpdateArticleUseCase) Execute(ctx context.Context, articleID, userID string, req dto.UpdateArticleRequest) (*dto.ArticleMutationResponse, error) {
	article, err := uc.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeArticleNotFound)
	}
	if err := article.EnsureEditable(); err != nil {
		return nil, kernel.Wrap(application.ErrCodeForbidden, err)
	}

	if req.Title != nil {
		article.Title = *req.Title
	}
	if req.Content != nil {
		article.Content = *req.Content
	}
	if req.Summary != nil {
		article.Summary = req.Summary
	}
	if req.CategoryID != nil {
		article.CategoryID = req.CategoryID
	}
	if req.Author != nil {
		article.Author = *req.Author
	}
	if req.ThumbnailURL != nil {
		article.ThumbnailURL = normalizeThumbnailKeyPtr(uc.fileUploader, req.ThumbnailURL)
	}
	if req.IsFeatured != nil {
		article.IsFeatured = *req.IsFeatured
	}

	article.UpdatedBy = &userID
	article.UpdatedAt = time.Now()

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.articleRepo.Update(txCtx, article)
	}); err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeArticleNotFound)
	}

	if req.ThumbnailURL != nil {
		confirmThumbnailKey(ctx, uc.fileUploader, article.ThumbnailURL)
	}

	return &dto.ArticleMutationResponse{ID: articleID}, nil
}
