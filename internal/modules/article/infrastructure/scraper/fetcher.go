package scraper

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

var feedClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		TLSNextProto:    make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	},
}

func BuildFeedURL(base string, suffix, override *string) string {
	if override != nil && *override != "" {
		return *override
	}
	if suffix != nil && *suffix != "" {
		return base + *suffix
	}
	return base
}

func FetchFeed(ctx context.Context, feedURL string) ([]FeedItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := feedClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed returned status %d", resp.StatusCode)
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}

	items := make([]FeedItem, 0, len(feed.Items))
	for _, it := range feed.Items {
		fi := FeedItem{
			Link:        it.Link,
			Title:       it.Title,
			Description: cleanHTML(it.Description),
		}
		if it.PublishedParsed != nil {
			fi.PubDate = it.PublishedParsed
		}
		if len(it.Enclosures) > 0 {
			fi.Thumbnail = &it.Enclosures[0].URL
		} else if it.Image != nil {
			fi.Thumbnail = &it.Image.URL
		}
		items = append(items, fi)
	}
	return items, nil
}

func cleanHTML(s string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		return s
	}
	return strings.TrimSpace(doc.Text())
}
