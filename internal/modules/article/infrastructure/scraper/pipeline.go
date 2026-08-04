package scraper

import (
	"context"
	"log/slog"
	"sync/atomic"

	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
	"sipon-be/internal/modules/article/application/thumbnail"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type Pipeline struct {
	articleRepo articlerepo.ArticleRepository
	maxParallel int64
}

func NewPipeline(articleRepo articlerepo.ArticleRepository, maxParallel int64) *Pipeline {
	return &Pipeline{
		articleRepo: articleRepo,
		maxParallel: maxParallel,
	}
}

func (p *Pipeline) ScrapeCategory(
	ctx context.Context,
	baseURL string,
	cat articlerepo.ScraperCategory,
	autoPublish bool,
	sourceID string,
	categoryID string,
	articleCategoryID *string,
	sel Selectors,
) (int, error) {
	feedURL := BuildFeedURL(baseURL, cat.URLSuffix, cat.URLOverride)

	items, err := FetchFeed(ctx, feedURL)
	if err != nil {
		return 0, err
	}

	items = FilterByKeywords(items, cat.Keywords)

	urls := make([]string, len(items))
	for i, it := range items {
		urls[i] = it.Link
	}
	existing, err := p.articleRepo.ExistingURLs(ctx, urls)
	if err != nil {
		existing = map[string]bool{}
	}
	items = FilterNew(ctx, items, existing)

	if cat.ArticleLimit > 0 && len(items) > cat.ArticleLimit {
		items = items[:cat.ArticleLimit]
	}
	if len(items) == 0 {
		return 0, nil
	}

	sem := semaphore.NewWeighted(p.maxParallel)
	g, gctx := errgroup.WithContext(ctx)
	var saved atomic.Int64

	for _, it := range items {
		it := it
		if err := sem.Acquire(gctx, 1); err != nil {
			break
		}
		g.Go(func() error {
			defer sem.Release(1)

			detail := FetchDetail(gctx, it.Link, sel)
			contentHTML := detail.ContentHTML
			if contentHTML == "" {
				contentHTML = detail.Content
			}
			if contentHTML == "" {
				slog.Warn("fetch detail yielded no content", "url", it.Link)
				return nil
			}

			var thumbnailURL *string
			if it.Thumbnail != nil && *it.Thumbnail != "" {
				thumbnailURL = thumbnail.ToStorage(it.Thumbnail)
			}

			articleID, err := p.articleRepo.SaveScraped(gctx, articlerepo.SaveScrapedInput{
				SourceID:    sourceID,
				CategoryKey: cat.CategoryKey,
				Title:       it.Title,
				Content:     contentHTML,
				Summary:     it.Description,
				Author:      firstNonEmpty(detail.Author, "Unknown"),
				Thumbnail:   thumbnailURL,
				Tags:        detail.Tags,
				PubDate:     it.PubDate,
				OriginalURL: firstNonEmpty(detail.URL, it.Link),
				CategoryID:  articleCategoryID,
				AutoPublish: autoPublish,
			})
			if err != nil {
				slog.Warn("save scraped article failed", "url", it.Link, "err", err)
				return nil
			}
			_ = articleID
			saved.Add(1)
			return nil
		})
	}
	_ = g.Wait()
	slog.Info("scrape category completed",
		"category", cat.CategoryKey, "fetched", len(items), "saved", int(saved.Load()))

	return int(saved.Load()), nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
