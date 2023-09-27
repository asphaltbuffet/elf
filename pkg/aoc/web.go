package aoc

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/afero"
)

var rClient = resty.New()

func downloadPuzzlePage(year int, day int) ([]byte, error) {
	// make sure we can write the cached file before we download it
	err := appFs.MkdirAll(filepath.Join(cfgDir, "puzzle_pages"), 0o750)
	if err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	res, err := rClient.R().Get(fmt.Sprintf(adventPuzzleURL, year, day))
	if err != nil {
		return nil, fmt.Errorf("getting puzzle page: %w", err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("getting puzzle page: %s", res.Status())
	}

	re := regexp.MustCompile(`(?s)<article.*?>(.*)</article>`)

	matches := re.FindSubmatch(res.Body())
	if len(matches) != 2 {
		// save the raw output to a file for debugging/error reporting
		err = appFs.MkdirAll(filepath.Join(cfgDir, "logs"), 0o750)
		if err != nil {
			return nil, fmt.Errorf("creating cache directory: %w", err)
		}

		dumpFile := filepath.Join(cfgDir, "puzzle_pages", fmt.Sprintf("%d-%d-ERROR.dump", year, day))
		_ = afero.WriteFile(appFs, dumpFile, res.Body(), 0o600)

		return nil, fmt.Errorf("parsing puzzle page, raw output saved to: %s", dumpFile)
	}

	data := bytes.TrimSpace(matches[1])

	cacheFile := filepath.Join(cfgDir, "puzzle_pages", fmt.Sprintf("%d-%d.txt", year, day))

	err = afero.WriteFile(appFs, cacheFile, data, 0o644)
	if err != nil {
		return nil, fmt.Errorf("caching puzzle page to %s: %w", cacheFile, err)
	}

	return data, nil
}

func downloadInput(year, day int) ([]byte, error) {
	res, err := rClient.R().Get(fmt.Sprintf(adventInputURL, year, day))
	if err != nil {
		return nil, fmt.Errorf("accessing input site: %w", err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("getting input data: %s", res.Status())
	}

	err = appFs.MkdirAll(filepath.Join(cfgDir, "inputs"), 0o750)
	if err != nil {
		return nil, fmt.Errorf("creating inputs directory: %w", err)
	}

	inputPath := filepath.Join(cfgDir, "inputs", fmt.Sprintf("%d-%d.txt", year, day))

	err = afero.WriteFile(appFs, inputPath, res.Body(), 0o600)
	if err != nil {
		return nil, fmt.Errorf("caching puzzle page to %s: %w", inputPath, err)
	}

	return bytes.TrimSpace(res.Body()), nil
}
