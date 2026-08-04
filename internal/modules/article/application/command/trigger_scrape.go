package command

import (
	"context"
	"log/slog"
	"time"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
	"sipon-be/internal/modules/article/infrastructure/scraper"
)

type TriggerScrapeAllUseCase struct {
	sourceRepo         articlerepo.SourceRepository
	categoryRepo       articlerepo.SourceCategoryRepository
	scraperSourceRepo  articlerepo.ScraperSourceRepo
	pipeline           *scraper.Pipeline
}

func NewTriggerScrapeAllUseCase(
	sourceRepo articlerepo.SourceRepository,
	categoryRepo articlerepo.SourceCategoryRepository,
	scraperSourceRepo articlerepo.ScraperSourceRepo,
	pipeline *scraper.Pipeline,
) *TriggerScrapeAllUseCase {
	return &TriggerScrapeAllUseCase{
		sourceRepo:        sourceRepo,
		categoryRepo:      categoryRepo,
		scraperSourceRepo: scraperSourceRepo,
		pipeline:          pipeline,
	}
}

func (uc *TriggerScrapeAllUseCase) Execute(ctx context.Context, sourceID string) (*dto.ScrapeResult, error) {
	_, err := uc.sourceRepo.FindByID(ctx, sourceID)
	if err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeSourceNotFound)
	}

	cats, err := uc.categoryRepo.FindAllBySource(ctx, sourceID)
	if err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeSourceCategoryQueryFailed)
	}

	result := &dto.ScrapeResult{}
	var totalScraped, totalSkipped int

	for _, cat := range cats {
		if !cat.IsActive {
			continue
		}

		src, err := uc.scraperSourceRepo.FindActiveCategoryByID(ctx, cat.ID)
		if err != nil || src == nil {
			result.Categories = append(result.Categories, dto.ScrapeCategoryItem{
				CategoryKey: cat.CategoryKey,
				Error:       "source or category not found or inactive",
			})
			continue
		}

		saved, scrapeErr := uc.pipeline.ScrapeCategory(
			ctx,
			src.BaseURL,
			src.Category,
			src.AutoPublish,
			src.ID,
			cat.ID,
			cat.ArticleCategoryID,
			scraper.Selectors{
				ContentSelector: src.Selectors.ContentSelector,
				AuthorSelector:  src.Selectors.AuthorSelector,
				TagsSelector:    src.Selectors.TagsSelector,
			},
		)

		catItem := dto.ScrapeCategoryItem{
			CategoryKey: cat.CategoryKey,
			Saved:       saved,
		}
		if scrapeErr != nil {
			catItem.Error = scrapeErr.Error()
		}
		result.Categories = append(result.Categories, catItem)
		totalScraped += saved
		totalSkipped += 0

		_ = uc.scraperSourceRepo.UpdateLastScrapedCategory(ctx, cat.ID, time.Now())
	}

	result.Scraped = totalScraped
	result.Skipped = totalSkipped

	slog.Info("scrape all completed", "source_id", sourceID, "scraped", totalScraped)
	return result, nil
}
