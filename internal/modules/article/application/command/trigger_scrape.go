package command

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
	ports "sipon-be/internal/modules/article/application/ports"
	"sipon-be/internal/modules/article/infrastructure/scraper"
)

type TriggerScrapeAllUseCase struct {
	sourceRepo         articlerepo.SourceRepository
	categoryRepo       articlerepo.SourceCategoryRepository
	scraperSourceRepo  articlerepo.ScraperSourceRepo
	pipeline           *scraper.Pipeline
	outboxWriter       ports.OutboxWriter
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

func (uc *TriggerScrapeAllUseCase) SetOutboxWriter(w ports.OutboxWriter) {
	uc.outboxWriter = w
}

func (uc *TriggerScrapeAllUseCase) publishEvent(ctx context.Context, routingKey string, payload any) {
	if uc.outboxWriter == nil {
		return
	}
	data, _ := json.Marshal(payload)
	if err := uc.outboxWriter.Save(ctx, routingKey, data); err != nil {
		slog.Warn("article: gagal publish scrape event", "routing_key", routingKey, "error", err)
	}
}

func (uc *TriggerScrapeAllUseCase) Execute(ctx context.Context, sourceID string) (*dto.ScrapeResult, error) {
	source, err := uc.sourceRepo.FindByID(ctx, sourceID)
	if err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeSourceNotFound)
	}

	cats, err := uc.categoryRepo.FindAllBySource(ctx, sourceID)
	if err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeSourceCategoryQueryFailed)
	}

	result := &dto.ScrapeResult{}
	var totalScraped, totalSkipped int
	var allTitles []string

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

		scrapeRes, scrapeErr := uc.pipeline.ScrapeCategory(
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
			Saved:       scrapeRes.Saved,
		}
		if scrapeErr != nil {
			catItem.Error = scrapeErr.Error()
		}
		result.Categories = append(result.Categories, catItem)
		totalScraped += scrapeRes.Saved
		totalSkipped += 0
		allTitles = append(allTitles, scrapeRes.Titles...)

		_ = uc.scraperSourceRepo.UpdateLastScrapedCategory(ctx, cat.ID, time.Now())
	}

	result.Scraped = totalScraped
	result.Skipped = totalSkipped

	if totalScraped > 0 {
		uc.publishEvent(ctx, RoutingArticlesScraped, map[string]any{
			"source_id":   sourceID,
			"source_name": source.Name,
			"count":       totalScraped,
			"titles":      allTitles,
		})
	}

	slog.Info("scrape all completed", "source_id", sourceID, "scraped", totalScraped)
	return result, nil
}

func truncateTitles(titles []string, max int) string {
	if len(titles) == 0 {
		return ""
	}
	if len(titles) <= max {
		return strings.Join(titles, ", ")
	}
	joined := strings.Join(titles[:max], ", ")
	return joined + " dan lainnya"
}
