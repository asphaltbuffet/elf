package advent

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/lmittmann/tint"
	"github.com/spf13/afero"
)

func (d *Downloader) getPage(year, day int) ([]byte, error) {
	logger := d.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "getPage"))

	pageData, ok := d.getCachedPage(year, day)
	if ok {
		logger.Debug("using cached puzzle page", slog.Int("size", len(pageData)))
		return pageData, nil
	}

	logger.Info("no cached page")

	return d.downloadPage(year, day)
}

func (d *Downloader) getCachedPage(year, day int) ([]byte, bool) {
	fp := filepath.Join(d.cacheDir, "pages", makeExerciseID(year, day))
	data, err := afero.ReadFile(d.appFs, fp)

	return data, err == nil
}

func (d *Downloader) getCachedInput(year, day int) ([]byte, bool) {
	fp := filepath.Join(d.cacheDir, "inputs", makeExerciseID(year, day))
	data, err := afero.ReadFile(d.appFs, fp)

	return data, err == nil
}

func (d *Downloader) downloadPage(year, day int) ([]byte, error) {
	pageCacheDir := filepath.Join(d.cacheDir, "pages")
	logger := d.logger.With(
		slog.String("fn", "downloadPage"),
		slog.Int("year", year),
		slog.Int("day", day),
		slog.String("dir", pageCacheDir),
	)

	// make sure we can write the cached file before we download it
	if err := d.appFs.MkdirAll(pageCacheDir, dirPerm); err != nil {
		return nil, fmt.Errorf("create %q: %w", pageCacheDir, err)
	}

	req := d.rClient.R().SetPathParams(map[string]string{
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
	err = afero.WriteFile(d.appFs, filepath.Join(pageCacheDir, makeExerciseID(year, day)), pd, 0o600)
	if err != nil {
		logger.Debug("writing page to cache", slog.String("url", resp.Request.URL), tint.Err(err))

		return nil, fmt.Errorf("writing cached puzzle page: %w", err)
	}

	return pd, nil
}

func (d *Downloader) downloadInput(year, day int) ([]byte, error) {
	logger := d.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "downloadInput"))

	err := d.appFs.MkdirAll(filepath.Join(d.cacheDir, "inputs"), dirPerm)
	if err != nil {
		return nil, fmt.Errorf("creating inputs directory: %w", err)
	}

	resp, err := d.rClient.R().
		SetPathParams(map[string]string{
			"year": strconv.Itoa(year),
			"day":  strconv.Itoa(day),
		}).
		SetCookie(&http.Cookie{
			Name:   "session",
			Value:  d.token,
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

	// write response to disk
	err = afero.WriteFile(d.appFs, filepath.Join(d.cacheDir, "inputs", makeExerciseID(year, day)), data, 0o600)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *Downloader) getInput(year, day int) ([]byte, error) {
	logger := d.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "getInput"))

	data, ok := d.getCachedInput(year, day)
	if ok {
		return data, nil
	}

	logger.Info("no cached input")

	return d.downloadInput(year, day)
}
