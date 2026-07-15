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
// <h2> yields an ErrInvalidData error, which fetchTitle treats as a secondary
// bad-number guard — the primary signal is a redirect (see fetchTitle), since
// projecteuler.net redirects an out-of-range number to /archives rather than
// serving a heading-less page. It shares the generic getH2NodeFromHTML walk with
// the AoC extractor (extractTitle) but keeps its own interpretation — euler's
// <h2> is the bare title, so site-specific markup drift touches only this line.
func extractProblemTitle(page []byte) (string, error) {
	doc, err := html.Parse(bytes.NewReader(page))
	if err != nil {
		return "", fmt.Errorf("parsing problem page: %w", err)
	}

	// Reuse the shared <h2> walk; euler differs from AoC only in interpretation —
	// its <h2> holds the bare title, so we take the text directly instead of
	// stripping AoC's "--- Day N: ... ---" decoration.
	h2, err := getH2NodeFromHTML(doc)
	if err != nil {
		return "", err
	}

	var text string
	if h2.FirstChild != nil && h2.FirstChild.Type == html.TextNode {
		text = h2.FirstChild.Data
	}

	return strings.TrimSpace(text), nil
}

// eulerBaseURL is the Project Euler site root; problem requests are relative to it.
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
func newProblemFetcher() *problemFetcher {
	return &problemFetcher{
		rClient: resty.New().
			SetBaseURL(eulerBaseURL).
			SetHeader("User-Agent", userAgent).
			SetRedirectPolicy(resty.NoRedirectPolicy()),
	}
}

// fetchTitle GETs the problem page and returns its title. projecteuler.net
// redirects out-of-range problem numbers to /archives (a page that itself has
// an <h2>, so a redirect is the only reliable bad-number signal); the client
// disables auto-redirect so an attempted redirect surfaces as
// resty.ErrAutoRedirectDisabled, which is reported as an ErrInvalidData-wrapped
// hard failure. A genuine transport error or a non-redirect non-200 response is
// a transient failure (wrapped ErrHTTPRequest/ErrHTTPResponse) that callers may
// degrade past with a placeholder. A 200 whose body has no <h2> is a secondary
// bad-number guard, also reported as ErrInvalidData.
func (f *problemFetcher) fetchTitle(number int) (string, error) {
	resp, err := f.rClient.R().
		SetPathParam("number", strconv.Itoa(number)).
		Get("/problem={number}")
	if err != nil {
		if errors.Is(err, resty.ErrAutoRedirectDisabled) {
			return "", fmt.Errorf("%w: problem %d does not exist (redirected)", ErrInvalidData, number)
		}

		return "", errors.Join(ErrHTTPRequest, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("%w: %s: %s", ErrHTTPResponse, resp.Request.Method, resp.Status())
	}

	return extractProblemTitle(resp.Body())
}
