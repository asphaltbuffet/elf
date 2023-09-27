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
	if err := checkOrAddYear(year); err != nil {
		return nil, fmt.Errorf("checking/adding year: %w", err)
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

	err = addMissingFiles(e, language, year, day)
	if err != nil {
		return nil, fmt.Errorf("adding missing files: %w", err)
	}

	return e, nil
}

func checkOrAddYear(year int) error {
	if _, ok := exercises[year]; ok {
		return nil
	}

	yearPath := filepath.Join(baseExercisesDir, fmt.Sprintf("%d", year))

	if err := appFs.MkdirAll(yearPath, 0o750); err != nil {
		return fmt.Errorf("creating year directory: %w", err)
	}

	return nil
}

func addMissingFiles(e *exercise.Exercise, language string, year int, day int) error {
	implPath := filepath.Join(e.Path, language)

	fi, err := appFs.Stat(implPath)
	if err == nil {
		return fmt.Errorf("exercise already exists: %s", fi.Name())
	}

	if err = appFs.MkdirAll(filepath.Join(e.Path, language), 0o750); err != nil {
		return fmt.Errorf("creating %s implementation directory: %w", language, err)
	}

	// download puzzle input
	inputFile, err := downloadOrGetCachedInput(year, day)
	if err != nil {
		return fmt.Errorf("getting puzzle input: %w", err)
	}

	err = afero.WriteFile(appFs, filepath.Join(e.Path, "input.txt"), inputFile, 0o600)
	if err != nil {
		return fmt.Errorf("writing input file: %w", err)
	}

	var (
		t *template.Template
		b *bytes.Buffer
	)

	// add templated implementation
	switch language {
	case "go":
		t = template.Must(template.New("implementation").Parse(string(goTemplate)))
		implPath = filepath.Join(implPath, "exercise.go")
	case "py":
		t = template.Must(template.New("implementation").Parse(string(pyTemplate)))
		implPath = filepath.Join(implPath, "__init__.py")
	default:
		return fmt.Errorf("language not supported: %s", language)
	}

	b = new(bytes.Buffer)

	err = t.Execute(b, e)
	if err != nil {
		return fmt.Errorf("executing %s template: %w", language, err)
	}

	err = afero.WriteFile(appFs, implPath, b.Bytes(), 0o600)
	if err != nil {
		return fmt.Errorf("writing %s implementaton file: %w", language, err)
	}

	// add templated info.json
	t = template.Must(template.New("info").Parse(string(infoTemplate)))
	b = new(bytes.Buffer)

	err = t.Execute(b, e)
	if err != nil {
		return fmt.Errorf("executing info template: %w", err)
	}

	err = afero.WriteFile(appFs, filepath.Join(e.Path, "info.json"), b.Bytes(), 0o600)
	if err != nil {
		return fmt.Errorf("writing info file: %w", err)
	}

	// add templated README.md
	t = template.Must(template.New("readme").Parse(string(readmeTemplate)))
	b = new(bytes.Buffer)

	err = t.Execute(b, e)
	if err != nil {
		return fmt.Errorf("executing readme template: %w", err)
	}

	err = afero.WriteFile(appFs, filepath.Join(e.Path, "README.md"), b.Bytes(), 0o600)
	if err != nil {
		return fmt.Errorf("writing info file: %w", err)
	}

	return nil
}

//go:embed templates/info.tmpl
var infoTemplate []byte

//go:embed templates/readme.tmpl
var readmeTemplate []byte

//go:embed templates/go.tmpl
var goTemplate []byte

//go:embed templates/py.tmpl
var pyTemplate []byte

func addDay(year int, day int) (*exercise.Exercise, error) {
	yearDir := filepath.Join(baseExercisesDir, fmt.Sprintf("%d", year))

	title := getTitle(year, day)
	if title == "" {
		return nil, fmt.Errorf("getting title for day %d", day)
	}

	exerciseDir := fmt.Sprintf("%02d-%s", day, strcase.ToLowerCamel(title))
	exercisePath := filepath.Join(yearDir, exerciseDir)

	if err := appFs.MkdirAll(exercisePath, 0o750); err != nil {
		return nil, fmt.Errorf("creating day directory: %w", err)
	}

	e := &exercise.Exercise{
		Year:  year,
		Day:   day,
		Title: title,
		Dir:   exerciseDir,
		Path:  exercisePath,
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

func downloadOrGetCachedInput(year int, day int) ([]byte, error) {
	d, err := getCachedInput(year, day)
	if err == nil {
		return d, nil
	}

	return downloadInput(year, day)
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
