package scraper

import "context"

func FilterNew(ctx context.Context, items []FeedItem, existing map[string]bool) []FeedItem {
	out := items[:0]
	for _, it := range items {
		if !existing[it.Link] {
			out = append(out, it)
		}
	}
	return out
}
