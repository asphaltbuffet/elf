package aoc

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/go-resty/resty/v2"
	"github.com/iancoleman/strcase"
	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/exercise"
)

func (ac *AOCClient) AddExercise(year int, day int, language string) (*exercise.Exercise, error) {
	err := isValidYear(year)
	// we don't care about the error, just that it's not a valid year
	if err != nil {
		_, err = addYear(year)
		if err != nil {
			return nil, fmt.Errorf("adding year: %w", err)
		}
	}

	// check for day/exercise
	e, err := ac.GetExercise(year, day)
	if err != nil {
		var dayErr error

		e, dayErr = addDay(year, day)
		if dayErr != nil {
			return nil, fmt.Errorf("adding day: %w", dayErr)
		}
	}

	info, err := fs.Stat(filepath.Join(e.Path, language))
	if err == nil {
		return e, fmt.Errorf("exercise already exists: %s", info.Name())
	}

	err = fs.MkdirAll(filepath.Join(e.Path, language), 0o755)
	if err != nil {
		return nil, fmt.Errorf("creating implementation directory: %w", err)
	}

	// TODO: create appropriate templated files

	return nil, fmt.Errorf("not implemented")
}

func addYear(year int) (string, error) {
	yearPath := filepath.Join(baseExercisesDir, fmt.Sprintf("%d", year))
	if err := fs.MkdirAll(yearPath, 0o755); err != nil {
		return "", fmt.Errorf("creating year directory: %w", err)
	}

	return yearPath, nil
}

//go:embed templates/info.tmpl
var infoTemplate []byte

//go:embed templates/readme.tmpl
var readmeTemplate []byte

func addDay(year int, day int) (*exercise.Exercise, error) {
	yearDir := filepath.Join(baseExercisesDir, fmt.Sprintf("%d", year))

	title := getTitle(year, day)
	if title == "" {
		return nil, fmt.Errorf("getting title for day %d", day)
	}

	exerciseDir := fmt.Sprintf("%02d-%s", day, strcase.ToLowerCamel(title))
	exercisePath := filepath.Join(yearDir, exerciseDir)

	if err := fs.MkdirAll(exercisePath, 0o755); err != nil {
		return nil, fmt.Errorf("creating day directory: %w", err)
	}

	e := &exercise.Exercise{
		Year:  year,
		Day:   day,
		Title: title,
		Dir:   exerciseDir,
		Path:  exercisePath,
	}

	t := template.Must(template.New("info").Parse(string(infoTemplate)))
	b := new(bytes.Buffer)

	err := t.Execute(b, e)
	if err != nil {
		return nil, fmt.Errorf("executing info template: %w", err)
	}

	err = afero.WriteFile(fs, filepath.Join(exercisePath, "info.json"), b.Bytes(), 0o600)
	if err != nil {
		return nil, fmt.Errorf("writing info file: %w", err)
	}

	t = template.Must(template.New("readme").Parse(string(readmeTemplate)))
	b = new(bytes.Buffer)

	err = t.Execute(b, e)
	if err != nil {
		return nil, fmt.Errorf("executing readme template: %w", err)
	}

	err = afero.WriteFile(fs, filepath.Join(exercisePath, "README.md"), b.Bytes(), 0o600)
	if err != nil {
		return nil, fmt.Errorf("writing info file: %w", err)
	}

	return e, nil
}

func getTitle(year int, day int) string {
	puzzlePage, err := getPuzzlePage(year, day)
	if err != nil {
		return ""
	}

	re := regexp.MustCompile(`--- Day \d{1,2}: (.*) ---`)

	matches := re.FindSubmatch(puzzlePage)
	if len(matches) != 2 {
		return ""
	}

	return string(matches[1])
}

func getPuzzlePage(year int, day int) ([]byte, error) {
	d, err := getCachedPuzzlePage(year, day)
	if err == nil {
		return d, nil
	}

	return downloadPuzzlePage(year, day)
}

func getCachedPuzzlePage(year int, day int) ([]byte, error) {
	f, err := afero.ReadFile(fs, filepath.Join(cfgDir, "puzzle_pages", fmt.Sprintf("%d-%d.txt", year, day)))
	if err != nil {
		return nil, fmt.Errorf("reading puzzle page: %w", err)
	}

	return f, nil
}

var rClient = resty.New()

func downloadPuzzlePage(year int, day int) ([]byte, error) {
	// make sure we can write the cached file before we download it
	err := fs.MkdirAll(filepath.Join(cfgDir, "puzzle_pages"), 0o755)
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
		err = fs.MkdirAll(filepath.Join(cfgDir, "logs"), 0o755)
		if err != nil {
			return nil, fmt.Errorf("creating cache directory: %w", err)
		}

		dumpFile := filepath.Join(cfgDir, "puzzle_pages", fmt.Sprintf("%d-%d-ERROR.dump", year, day))
		_ = afero.WriteFile(fs, dumpFile, res.Body(), 0o600)

		return nil, fmt.Errorf("parsing puzzle page, raw output saved to: %s", dumpFile)
	}

	data := bytes.TrimSpace(matches[1])

	cacheFile := filepath.Join(cfgDir, "puzzle_pages", fmt.Sprintf("%d-%d.txt", year, day))

	err = afero.WriteFile(fs, cacheFile, data, 0o644)
	if err != nil {
		return nil, fmt.Errorf("caching puzzle page to %s: %w", cacheFile, err)
	}

	return data, nil
}
