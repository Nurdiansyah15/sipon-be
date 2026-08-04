package command

import (
	"context"

	"sipon-be/internal/modules/article/application"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
	"sipon-be/internal/shared/kernel"
)

type DeleteArticleUseCase struct {
	articleRepo articlerepo.ArticleRepository
}

func NewDeleteArticleUseCase(articleRepo articlerepo.ArticleRepository) *DeleteArticleUseCase {
	return &DeleteArticleUseCase{articleRepo: articleRepo}
}

func (uc *DeleteArticleUseCase) Execute(ctx context.Context, articleID string) error {
	article, err := uc.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		return application.WrapRepoErr(err, articleconst.CodeArticleNotFound)
	}
	if err := article.EnsureDeletable(); err != nil {
		return kernel.Wrap(application.ErrCodeForbidden, err)
	}

	if err := uc.articleRepo.Delete(ctx, articleID); err != nil {
		return application.WrapRepoErr(err, articleconst.CodeArticleNotFound)
	}
	return nil
}
