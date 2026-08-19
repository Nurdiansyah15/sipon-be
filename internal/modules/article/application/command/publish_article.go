package command

import (
	"context"
	"encoding/json"
	"log/slog"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	ports "sipon-be/internal/modules/article/application/ports"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
	"sipon-be/internal/shared/kernel"
)

type PublishArticleUseCase struct {
	articleRepo articlerepo.ArticleRepository
	outboxWriter ports.OutboxWriter
}

func NewPublishArticleUseCase(articleRepo articlerepo.ArticleRepository) *PublishArticleUseCase {
	return &PublishArticleUseCase{articleRepo: articleRepo}
}

func (uc *PublishArticleUseCase) SetOutboxWriter(w ports.OutboxWriter) {
	uc.outboxWriter = w
}

func (uc *PublishArticleUseCase) publishEvent(ctx context.Context, routingKey, articleID, title string) {
	if uc.outboxWriter == nil {
		return
	}
	payload, _ := json.Marshal(map[string]string{
		"article_id": articleID,
		"title":      title,
	})
	if err := uc.outboxWriter.Save(ctx, routingKey, payload); err != nil {
		slog.Warn("article: gagal publish event", "routing_key", routingKey, "article_id", articleID, "error", err)
	}
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

	uc.publishEvent(ctx, RoutingArticlePublished, articleID, article.Title)

	return &dto.ArticleMutationResponse{ID: articleID}, nil
}
