package scraper

import "strings"

func FilterByKeywords(items []FeedItem, keywords []string) []FeedItem {
	cleaned := make([]string, 0, len(keywords))
	for _, kw := range keywords {
		if kw = strings.ToLower(strings.TrimSpace(kw)); kw != "" {
			cleaned = append(cleaned, kw)
		}
	}
	if len(cleaned) == 0 {
		return items
	}

	out := items[:0]
	for _, it := range items {
		title := strings.ToLower(it.Title)
		desc := strings.ToLower(it.Description)
		for _, kw := range cleaned {
			if strings.Contains(title, kw) || strings.Contains(desc, kw) {
				out = append(out, it)
				break
			}
		}
	}
	return out
}
