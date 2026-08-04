package scraper

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const maxScrapedContentWords = 100

func TruncateContentHTML(rawHTML string, maxWords int) string {
	if strings.TrimSpace(rawHTML) == "" {
		return rawHTML
	}

	nodes, err := html.ParseFragment(strings.NewReader(rawHTML), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})
	if err != nil {
		return rawHTML
	}

	wordCount := 0
	truncated := false
	kept := make([]*html.Node, 0, len(nodes))

	for _, n := range nodes {
		if wordCount >= maxWords {
			truncated = true
			break
		}
		kept = append(kept, n)
		if cutNode(n, maxWords, &wordCount, &truncated) {
			break
		}
	}

	var sb strings.Builder
	for _, n := range kept {
		_ = html.Render(&sb, n)
	}
	out := sb.String()
	if truncated {
		out = strings.TrimRight(out, " ") + " ..."
	}
	return out
}

func cutNode(n *html.Node, maxWords int, wordCount *int, truncated *bool) bool {
	if n.Type == html.TextNode {
		if *wordCount >= maxWords {
			n.Data = ""
			return true
		}
		words := strings.Fields(n.Data)
		if *wordCount+len(words) > maxWords {
			remaining := maxWords - *wordCount
			n.Data = strings.Join(words[:remaining], " ")
			*wordCount = maxWords
			*truncated = true
			return true
		}
		*wordCount += len(words)
		return false
	}

	child := n.FirstChild
	for child != nil {
		next := child.NextSibling
		if cutNode(child, maxWords, wordCount, truncated) {
			for sibling := child.NextSibling; sibling != nil; {
				toRemove := sibling
				sibling = sibling.NextSibling
				n.RemoveChild(toRemove)
			}
			return true
		}
		child = next
	}
	return false
}
