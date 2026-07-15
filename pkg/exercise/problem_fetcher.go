package exercise

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
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

// eulerBaseURL is the Project Euler site root; problem requests are relative to it.
//
//nolint:unused // wired into ProblemAdder in a follow-up task
const eulerBaseURL = "https://projecteuler.net"

// problemFetcher fetches a Project Euler problem's title from projecteuler.net.
// Unlike the AoC pageFetcher it needs no token and does no on-disk caching:
// `add euler` is a one-shot-per-problem operation. It owns the projecteuler.net
// URL shape and nothing about the exercise directory.
type problemFetcher struct {
	rClient *resty.Client
}

// newProblemFetcher builds a problemFetcher whose client targets projecteuler.net
// and identifies elf via the shared User-Agent, per site etiquette.
//
//nolint:unused // wired into ProblemAdder in a follow-up task
func newProblemFetcher() *problemFetcher {
	return &problemFetcher{
		rClient: resty.New().SetBaseURL(eulerBaseURL).SetHeader("User-Agent", userAgent),
	}
}

// fetchTitle GETs the problem page and returns its title. A transport error or a
// non-200 response is a transient failure (wrapped ErrHTTPRequest/ErrHTTPResponse)
// that callers may degrade past with a placeholder. A 200 whose body has no <h2>
// means the number does not exist and is returned as an ErrInvalidData-wrapped
// error, which callers treat as a hard, non-degradable failure.
func (f *problemFetcher) fetchTitle(number int) (string, error) {
	resp, err := f.rClient.R().
		SetPathParam("number", strconv.Itoa(number)).
		Get("/problem={number}")
	if err != nil {
		return "", errors.Join(ErrHTTPRequest, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("%w: %s: %s", ErrHTTPResponse, resp.Request.Method, resp.Status())
	}

	return extractProblemTitle(resp.Body())
}
