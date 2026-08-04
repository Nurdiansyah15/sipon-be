package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articleentity "sipon-be/internal/modules/article/domain/article/entity"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
)

type UpdateSourceUseCase struct {
	sourceRepo   articlerepo.SourceRepository
	selectorRepo articlerepo.SourceSelectorRepository
}

func NewUpdateSourceUseCase(sourceRepo articlerepo.SourceRepository, selectorRepo articlerepo.SourceSelectorRepository) *UpdateSourceUseCase {
	return &UpdateSourceUseCase{sourceRepo: sourceRepo, selectorRepo: selectorRepo}
}

func (uc *UpdateSourceUseCase) Execute(ctx context.Context, sourceID, userID string, req dto.UpdateSourceRequest) (*dto.SourceMutationResponse, error) {
	s, err := uc.sourceRepo.FindByID(ctx, sourceID)
	if err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeSourceNotFound)
	}

	if req.Key != nil {
		s.Key = *req.Key
	}
	if req.Name != nil {
		s.Name = *req.Name
	}
	if req.BaseURL != nil {
		s.BaseURL = *req.BaseURL
	}
	if req.AutoPublish != nil {
		s.AutoPublish = *req.AutoPublish
	}
	if req.IsActive != nil {
		s.IsActive = *req.IsActive
	}
	s.UpdatedBy = &userID
	s.UpdatedAt = time.Now()

	if err := uc.sourceRepo.Update(ctx, s); err != nil {
		return nil, application.WrapConflictErr(err, articleconst.CodeSourceDuplicateKey)
	}

	if req.Selectors != nil {
		sel, _ := uc.selectorRepo.FindBySourceID(ctx, sourceID)
		if sel == nil {
			sel = articleentity.NewSourceSelector(articleentity.SourceSelectorParams{
				ID:              generateID(),
				SourceID:        sourceID,
				ContentSelector: req.Selectors.ContentSelector,
				AuthorSelector:  req.Selectors.AuthorSelector,
				TagsSelector:    req.Selectors.TagsSelector,
			})
		} else {
			sel.ContentSelector = req.Selectors.ContentSelector
			sel.AuthorSelector = req.Selectors.AuthorSelector
			sel.TagsSelector = req.Selectors.TagsSelector
			sel.UpdatedAt = time.Now()
		}
		if err := uc.selectorRepo.SaveOrUpdate(ctx, sel); err != nil {
			return nil, application.WrapRepoErr(err, articleconst.CodeSelectorPersistenceFailed)
		}
	}

	return &dto.SourceMutationResponse{ID: s.ID, Key: s.Key}, nil
}

func generateID() string {
	return uuid.NewString()
}
