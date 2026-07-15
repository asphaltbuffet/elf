package exercise

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// extractProblemTitle returns the trimmed text of the first <h2> in a
// projecteuler.net problem page, which is the problem title. Project Euler
// problem pages contain exactly one <h2> and it holds the title. A page with no
// <h2> means the problem number does not exist (projecteuler.net serves a 200
// with no problem heading for out-of-range numbers), which callers treat as a
// hard error. This is deliberately separate from the AoC title extractor
// (extractTitle) so drift in either site's markup touches only its own path.
func extractProblemTitle(page []byte) (string, error) {
	doc, err := html.Parse(bytes.NewReader(page))
	if err != nil {
		return "", fmt.Errorf("parsing problem page: %w", err)
	}

	var (
		text    string
		found   bool
		crawler func(*html.Node)
	)

	crawler = func(node *html.Node) {
		if found {
			return
		}

		if node.Type == html.ElementNode && node.Data == "h2" {
			if node.FirstChild != nil && node.FirstChild.Type == html.TextNode {
				text = node.FirstChild.Data
			}
			found = true
			return
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			crawler(c)
		}
	}

	crawler(doc)

	if !found {
		return "", fmt.Errorf("%w: no problem title found", ErrInvalidData)
	}

	return strings.TrimSpace(text), nil
}
