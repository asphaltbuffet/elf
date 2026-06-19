package exercise

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/lmittmann/tint"
	"github.com/spf13/afero"
)

// pageFetcher fetches puzzle page HTML and puzzle input from Advent of Code, caching both on disk.
// It owns the HTTP client, the session token, and the cache directory. It knows the AoC URL shape
// and the cacheDir/pages + cacheDir/inputs layout, and nothing about the exercise directory.
type pageFetcher struct {
	rClient  *resty.Client
	token    string
	cacheDir string
	fs       afero.Fs
	logger   *slog.Logger
}

func (f *pageFetcher) fetchPage(year, day int) ([]byte, error) {
	logger := f.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "getPage"))

	pageData, ok := f.getCachedPage(year, day)
	if ok {
		logger.Debug("using cached puzzle page", slog.Int("size", len(pageData)))
		return pageData, nil
	}

	logger.Info("no cached page")

	return f.downloadPage(year, day)
}

func (f *pageFetcher) getCachedPage(year, day int) ([]byte, bool) {
	fp := filepath.Join(f.cacheDir, "pages", makeExerciseID(year, day))
	data, err := afero.ReadFile(f.fs, fp)

	return data, err == nil
}

func (f *pageFetcher) getCachedInput(year, day int) ([]byte, bool) {
	fp := filepath.Join(f.cacheDir, "inputs", makeExerciseID(year, day))
	data, err := afero.ReadFile(f.fs, fp)

	// A 0-byte cache file is not a usable hit — treat it as a miss so the input
	// is re-fetched rather than silently producing an empty puzzle input.
	return data, err == nil && len(data) > 0
}

func (f *pageFetcher) downloadPage(year, day int) ([]byte, error) {
	pageCacheDir := filepath.Join(f.cacheDir, "pages")
	logger := f.logger.With(
		slog.String("fn", "downloadPage"),
		slog.Int("year", year),
		slog.Int("day", day),
		slog.String("dir", pageCacheDir),
	)

	// make sure we can write the cached file before we download it
	if err := f.fs.MkdirAll(pageCacheDir, dirPerm); err != nil {
		return nil, fmt.Errorf("create %q: %w", pageCacheDir, err)
	}

	req := f.rClient.R().SetPathParams(map[string]string{
		"year": strconv.Itoa(year),
		"day":  strconv.Itoa(day),
	})

	resp, err := req.Get("/{year}/day/{day}")
	if err != nil {
		return nil, errors.Join(ErrHTTPRequest, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("%w: %s: %s", ErrHTTPResponse, resp.Request.Method, resp.Status())
	}

	logger.Debug("download page response",
		slog.String("url", resp.Request.URL),
		slog.String("status", http.StatusText(resp.StatusCode())),
		slog.Int("code", resp.StatusCode()))

	// only keep relevant parts of the page
	re := regexp.MustCompile(`(?s)<article.*?>(.*)</article>`)
	matches := re.FindSubmatch(resp.Body())

	if len(matches) != 2 { //nolint:mnd // we expect 2 matches
		logger.Debug("extracting page data", slog.String("url", resp.Request.URL), slog.Any("found", matches))

		return nil, errors.New("extracting page data: no match")
	}

	pd := bytes.TrimSpace(matches[1])

	// write response to disk
	err = afero.WriteFile(f.fs, filepath.Join(pageCacheDir, makeExerciseID(year, day)), pd, 0o600)
	if err != nil {
		logger.Debug("writing page to cache", slog.String("url", resp.Request.URL), tint.Err(err))

		return nil, fmt.Errorf("writing cached puzzle page: %w", err)
	}

	return pd, nil
}

func (f *pageFetcher) downloadInput(year, day int) ([]byte, error) {
	logger := f.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "downloadInput"))

	err := f.fs.MkdirAll(filepath.Join(f.cacheDir, "inputs"), dirPerm)
	if err != nil {
		return nil, fmt.Errorf("creating inputs directory: %w", err)
	}

	resp, err := f.rClient.R().
		SetPathParams(map[string]string{
			"year": strconv.Itoa(year),
			"day":  strconv.Itoa(day),
		}).
		SetCookie(&http.Cookie{
			Name:   "session",
			Value:  f.token,
			Domain: ".adventofcode.com",
		}).
		Get("/{year}/day/{day}/input")
	if err != nil {
		return nil, errors.Join(ErrHTTPRequest, err)
	}

	if resp.StatusCode() != http.StatusOK {
		logger.Debug("getting input data",
			slog.Group("request",
				slog.String("method", resp.Request.Method),
				slog.String("url", resp.Request.URL),
				slog.Any("cookies", resp.Request.Cookies)),
			slog.String("status", resp.Status()),
			slog.Int("code", resp.StatusCode()))

		return nil, fmt.Errorf("%w: %s", ErrHTTPResponse, resp.Status())
	}

	data := bytes.TrimSpace(resp.Body())

	// A 200 with an empty body means the request was unauthenticated (e.g. a
	// missing or expired token); fail loudly instead of caching an empty input.
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: %d-%02d", ErrEmptyInput, year, day)
	}

	// write response to disk
	err = afero.WriteFile(f.fs, filepath.Join(f.cacheDir, "inputs", makeExerciseID(year, day)), data, 0o600)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (f *pageFetcher) fetchInput(year, day int) ([]byte, error) {
	logger := f.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "getInput"))

	data, ok := f.getCachedInput(year, day)
	if ok {
		return data, nil
	}

	logger.Info("no cached input")

	return f.downloadInput(year, day)
}
