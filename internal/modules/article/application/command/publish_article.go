package command

import (
	"context"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
	"sipon-be/internal/shared/kernel"
)

type PublishArticleUseCase struct {
	articleRepo articlerepo.ArticleRepository
}

func NewPublishArticleUseCase(articleRepo articlerepo.ArticleRepository) *PublishArticleUseCase {
	return &PublishArticleUseCase{articleRepo: articleRepo}
}

func (uc *PublishArticleUseCase) Execute(ctx context.Context, articleID string) (*dto.ArticleMutationResponse, error) {
	article, err := uc.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeArticleNotFound)
	}
	if err := article.Publish(); err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if err := uc.articleRepo.Update(ctx, article); err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeArticleNotFound)
	}
	return &dto.ArticleMutationResponse{ID: articleID}, nil
}
