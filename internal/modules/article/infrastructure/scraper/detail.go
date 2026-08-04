package scraper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	readability "github.com/go-shiori/go-readability"
)

const maxPages = 10

type Selectors struct {
	ContentSelector *string
	AuthorSelector  *string
	TagsSelector    *string
}

func FetchDetail(ctx context.Context, articleURL string, sel Selectors) ArticleDetail {
	var allContent strings.Builder
	var allContentHTML strings.Builder
	var firstAuthor string
	var firstTags []string
	var finalURL string
	visited := map[string]bool{}
	current := articleURL
	pageCount := 0

	for current != "" && !visited[current] && pageCount < maxPages {
		visited[current] = true
		pageCount++

		doc, rawHTML, resolvedURL, err := httpGet(ctx, current)
		if err != nil {
			if pageCount == 1 {
				return ArticleDetail{}
			}
			break
		}

		if pageCount == 1 {
			finalURL = resolvedURL
		}

		parsedURL, _ := url.Parse(resolvedURL)
		article, rerr := readability.FromReader(strings.NewReader(rawHTML), parsedURL)
		if rerr == nil {
			if allContent.Len() > 0 {
				allContent.WriteString("\n\n")
			}
			allContent.WriteString(strings.TrimSpace(article.TextContent))

			if article.Content != "" {
				if allContentHTML.Len() > 0 {
					allContentHTML.WriteString("\n\n")
				}
				allContentHTML.WriteString(strings.TrimSpace(article.Content))
			}
		}

		if pageCount == 1 {
			firstAuthor = extractAuthor(doc, sel.AuthorSelector, article)
			firstTags = extractTags(doc, sel.TagsSelector)
		}

		current = detectNextPage(doc, resolvedURL, pageCount)
	}

	if allContent.Len() == 0 {
		return ArticleDetail{}
	}

	contentHTML := strings.TrimSpace(allContentHTML.String())
	if contentHTML == "" {
		contentHTML = strings.TrimSpace(allContent.String())
	}

	return ArticleDetail{
		URL:         finalURL,
		Content:     strings.TrimSpace(allContent.String()),
		ContentHTML: contentHTML,
		Author:      firstAuthor,
		Tags:        firstTags,
	}
}

func httpGet(ctx context.Context, target string) (*goquery.Document, string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, "", "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", "https://www.google.com/")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	finalURL := resp.Request.URL.String()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, "", "", err
	}
	rawHTML, _ := doc.Html()
	return doc, rawHTML, finalURL, nil
}

func extractAuthor(doc *goquery.Document, selector *string, article readability.Article) string {
	if article.Byline != "" {
		return normalizeSpace(article.Byline)
	}
	sel := ".author, .writer, .byline, .reporter, .penulis"
	if selector != nil && *selector != "" {
		sel = *selector
	}
	return normalizeSpace(doc.Find(sel).First().Text())
}

func extractTags(doc *goquery.Document, selector *string) []string {
	sel := ".tags a, .tag-links a, .detail-tag a, [href*='/tag/']"
	if selector != nil && *selector != "" {
		sel = *selector
	}
	var tags []string
	seen := map[string]bool{}
	doc.Find(sel).Each(func(_ int, s *goquery.Selection) {
		t := strings.TrimSpace(s.Text())
		if t != "" && !seen[t] {
			seen[t] = true
			tags = append(tags, t)
		}
	})
	return tags
}

func detectNextPage(doc *goquery.Document, current string, pageCount int) string {
	if strings.Contains(current, "tribunnews.com") && !strings.Contains(current, "page=all") {
		if strings.Contains(current, "?") {
			return current + "&page=all"
		}
		return current + "?page=all"
	}

	var next string
	doc.Find("a").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		text := strings.ToLower(strings.TrimSpace(s.Text()))
		href, exists := s.Attr("href")
		if !exists {
			return true
		}
		if text == "selanjutnya" || text == "next" || text == fmt.Sprint(pageCount+1) {
			if u, err := resolveURL(current, href); err == nil {
				next = u
				return false
			}
		}
		return true
	})
	return next
}

func resolveURL(base, href string) (string, error) {
	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	r, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	return b.ResolveReference(r).String(), nil
}

func normalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
