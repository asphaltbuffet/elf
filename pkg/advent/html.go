package advent

import (
	"bytes"
	"fmt"

	"golang.org/x/net/html"
)

func H2(doc *html.Node) (*html.Node, error) {
	var (
		h2      *html.Node
		crawler func(*html.Node)
	)

	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "h2" {
			h2 = node
			return
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			crawler(c)
		}
	}

	crawler(doc)

	if h2 == nil {
		return nil, fmt.Errorf("%w: no <h2> found", ErrInvalidData)
	}

	return h2, nil
}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer

	if err := html.Render(&buf, n); err != nil {
		panic("render html node")
	}

	return buf.String()
}
