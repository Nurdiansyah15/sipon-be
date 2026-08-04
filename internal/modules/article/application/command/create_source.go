package command

import (
	"context"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articleentity "sipon-be/internal/modules/article/domain/article/entity"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"

	"github.com/google/uuid"
)

type CreateSourceUseCase struct {
	sourceRepo          articlerepo.SourceRepository
	selectorRepo        articlerepo.SourceSelectorRepository
}

func NewCreateSourceUseCase(sourceRepo articlerepo.SourceRepository, selectorRepo articlerepo.SourceSelectorRepository) *CreateSourceUseCase {
	return &CreateSourceUseCase{sourceRepo: sourceRepo, selectorRepo: selectorRepo}
}

func (uc *CreateSourceUseCase) Execute(ctx context.Context, req dto.CreateSourceRequest, userID string) (*dto.SourceMutationResponse, error) {
	s, err := articleentity.NewSource(articleentity.SourceParams{
		ID:          uuid.NewString(),
		Key:         req.Key,
		Name:        req.Name,
		BaseURL:     req.BaseURL,
		AutoPublish: req.AutoPublish,
		IsActive:    req.IsActive,
		CreatedBy:   &userID,
	})
	if err != nil {
		return nil, err
	}

	if err := uc.sourceRepo.Save(ctx, s); err != nil {
		return nil, application.WrapConflictErr(err, articleconst.CodeSourceDuplicateKey)
	}

	if req.Selectors != nil {
		sel := articleentity.NewSourceSelector(articleentity.SourceSelectorParams{
			ID:              uuid.NewString(),
			SourceID:        s.ID,
			ContentSelector: req.Selectors.ContentSelector,
			AuthorSelector:  req.Selectors.AuthorSelector,
			TagsSelector:    req.Selectors.TagsSelector,
		})
		if err := uc.selectorRepo.Save(ctx, sel); err != nil {
			return nil, application.WrapRepoErr(err, articleconst.CodeSelectorPersistenceFailed)
		}
	}

	return &dto.SourceMutationResponse{ID: s.ID, Key: s.Key}, nil
}
