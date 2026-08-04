package scraper

import "time"

type FeedItem struct {
	Title       string
	Link        string
	Description string
	PubDate     *time.Time
	Thumbnail   *string
}

type ArticleDetail struct {
	URL         string
	Content     string
	ContentHTML string
	Author      string
	Tags        []string
}
