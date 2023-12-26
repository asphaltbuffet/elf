package advent

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"golang.org/x/net/html"

	"github.com/go-resty/resty/v2"
	"github.com/lmittmann/tint"
	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"github.com/asphaltbuffet/elf/pkg/krampus"
)

var cfg *viper.Viper
var (
	rClient  = resty.New().SetBaseURL("https://adventofcode.com")
	cfgDir   string
	cacheDir string
	appFs    = afero.NewOsFs()
)

func Download(url string, lang string, _ bool) (string, error) {
	var err error
	cfg, err = krampus.New()
	if err != nil {
		return "", err
	}

	cfgDir = cfg.GetString("config-dir")
	cacheDir = cfg.GetString("cache-dir")

	if cacheDir == "" {
		slog.Error("empty cache directory path", slog.String("url", url), slog.String("lang", lang), slog.String("defaultCacheDir", cfg.GetString("cache_dir")))
		return "", fmt.Errorf("cache directory not set")
	}

	if cfg.GetString("advent.token") == "" {
		slog.Error("empty session token")
		return "", fmt.Errorf("session token not set")
	}

	year, day, err := ParseURL(url)
	if err != nil {
		slog.Error("creating exercise from URL", slog.String("url", url), slog.String("lang", lang), tint.Err(err))
		return "", fmt.Errorf("creating exercise from URL: %w", err)
	}

	// update client with exercise-specific data
	rClient.
		SetHeader("User-Agent", "github.com/asphaltbuffet/elf").
		SetPathParams(map[string]string{
			"year": strconv.Itoa(year),
			"day":  strconv.Itoa(day),
		})

	e := &Exercise{}

	exPath, ok := getExercisePath(year, day)
	if ok {
		e.Path = exPath
		err = e.loadInfo()
	} else {
		e, err = loadFromURL(url, year, day, lang)
	}
	if err != nil {
		slog.Error("loading exercise", slog.String("url", url), slog.String("lang", lang), tint.Err(err))
		return "", fmt.Errorf("loading exercise: %w", err)
	}

	// the basic exercise information is here; add missing elements
	if err = e.addMissingFiles(); err != nil {
		slog.Error("adding missing files", slog.Any("exercise", e), tint.Err(err))
		return "", fmt.Errorf("adding missing files: %w", err)
	}

	slog.Info("exercise added", slog.String("url", e.URL), slog.String("dir", e.Dir()))

	return e.Path, nil
}

func loadFromURL(url string, year, day int, lang string) (*Exercise, error) {
	var (
		page  []byte
		title string
		err   error
		e     *Exercise
	)

	page, err = getPage(year, day)
	if err != nil {
		slog.Error("getting page data",
			slog.String("url", url),
			slog.Int("year", year),
			slog.Int("day", day),
			tint.Err(err))

		return nil, fmt.Errorf("getting page data: %w", err)
	}

	title, err = extractTitle(page)
	if err != nil {
		slog.Error("extracting title from page data",
			slog.String("url", url),
			slog.Int("year", year),
			slog.Int("day", day),
			tint.Err(err))

		return nil, fmt.Errorf("extracting title from page data: %w", err)
	}

	e = &Exercise{
		ID:       makeExerciseID(year, day),
		Title:    title,
		Year:     year,
		Language: lang,
		Day:      day,
		URL:      url,
		Data:     nil, // this should be empty, we only load this from info.json
		Path:     makeExercisePath(year, day, title),
	}

	return e, nil
}

func getExercisePath(year, day int) (string, bool) {
	slog.Debug("searching for exercise directory",
		slog.Int("year", year),
		slog.Int("day", day),
		slog.String("dir", exerciseBaseDir))

	var exPath string
	dayPrefix := fmt.Sprintf("%02d-", day)

	err := filepath.WalkDir(exerciseBaseDir, func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err

		case !d.IsDir():
			slog.Debug("skipping non-directory", slog.String("path", path))
			fallthrough
		case path == exerciseBaseDir:
			return nil

		case strings.HasPrefix(d.Name(), dayPrefix):
			slog.Info("found exercise directory", slog.String("path", path))
			exPath = path

			// we found the directory we're looking for, stop walking
			return filepath.SkipAll

		case d.Name() == strconv.Itoa(year):
			slog.Debug("found year directory", slog.String("path", path))
			return nil

		default:
			slog.Debug("skipping non-year directory", slog.String("path", path))
			// we only recurse into the specified year directory until we find the wanted day
			return filepath.SkipDir
		}
	})
	if err != nil {
		slog.Error("walking exercise directory", slog.Int("year", year), slog.Int("day", day), tint.Err(err))
		return "", false
	}

	return exPath, exPath != ""
}

func extractTitle(page []byte) (string, error) {
	doc, _ := html.Parse(bytes.NewReader(page))

	extract, err := H2(doc)
	if err != nil {
		return "", fmt.Errorf("extracting title: %w", err)
	}

	rendNode := renderNode(extract)

	re := regexp.MustCompile(`--- Day \d{1,2}: (.*) ---`)

	matches := re.FindStringSubmatch(rendNode)
	if len(matches) != 2 { //nolint:gomnd // we expect 2 matches
		return "", fmt.Errorf("getting title from page data: no match")
	}

	return matches[1], nil
}

func getPage(year, day int) ([]byte, error) {
	pageData, err := getCachedPuzzlePage(year, day)
	if err == nil {
		slog.Debug("using cached puzzle page",
			slog.Int("year", year),
			slog.Int("day", day),
			slog.Int("size", len(pageData)))
		return pageData, nil
	}

	slog.Info("downloading puzzle page", slog.Int("year", year), slog.Int("day", day))

	return downloadPuzzlePage(year, day)
}

func ParseURL(url string) (int, int, error) {
	var y, d int

	// regex here is validating year/day are integers, if this changes, add validation below
	re := regexp.MustCompile(`^https?://(www\.)?adventofcode\.com/(?P<year>\d{4})/day/(?P<day>\d{1,2})`)

	matches := findNamedMatches(re, url)
	if len(matches) != 2 { //nolint:gomnd // we expect 2 matches
		slog.Error("parsing URL", slog.String("url", url), slog.Any("found", matches))
		return 0, 0, fmt.Errorf("parsing %s: invalid URL format", url)
	}

	// ignore errors; we already validated type via regex
	y, _ = strconv.Atoi(matches["year"])
	d, _ = strconv.Atoi(matches["day"])

	return y, d, nil
}

func findNamedMatches(re *regexp.Regexp, s string) map[string]string {
	match := re.FindStringSubmatch(s)
	if len(match) == 0 {
		return nil
	}

	result := make(map[string]string)

	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	return result
}

func getCachedPuzzlePage(year, day int) ([]byte, error) {
	if cacheDir == "" {
		return nil, fmt.Errorf("cache directory not set")
	}

	fp := filepath.Join(cacheDir, "pages", makeExerciseID(year, day))

	f, err := afero.ReadFile(appFs, fp)
	if err != nil {
		return nil, fmt.Errorf("reading puzzle page: %w", err)
	}

	return f, nil
}

func (e *Exercise) getCachedInput() ([]byte, error) {
	if cacheDir == "" {
		return nil, fmt.Errorf("cache directory not set")
	}

	fp := filepath.Join(cacheDir, "inputs", e.ID)

	f, err := afero.ReadFile(appFs, fp)
	if err != nil {
		return nil, fmt.Errorf("read cached input: %w", err)
	}

	return f, nil
}

func downloadPuzzlePage(year, day int) ([]byte, error) {
	if cacheDir == "" {
		return nil, fmt.Errorf("cache directory not set")
	}

	// make sure we can write the cached file before we download it
	err := appFs.MkdirAll(filepath.Join(cacheDir, "pages"), 0o750)
	if err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	slog.Info("downloading puzzle page",
		slog.String("file", filepath.Join(cacheDir, "pages", makeExerciseID(year, day))))

	req := rClient.R().SetPathParams(map[string]string{
		"year": strconv.Itoa(year),
		"day":  strconv.Itoa(day),
	})

	resp, err := req.Get("/{year}/day/{day}")
	if err != nil {
		slog.Error("getting puzzle page", tint.Err(err))

		return nil, fmt.Errorf("requesting page: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		slog.Error("download page response",
			slog.String("url", resp.Request.URL),
			slog.String("status", http.StatusText(resp.StatusCode())),
			slog.Int("code", resp.StatusCode()))

		return nil, fmt.Errorf("processing page response: %s", http.StatusText(resp.StatusCode()))
	}

	// only keep relevant parts of the page
	re := regexp.MustCompile(`(?s)<article.*?>(.*)</article>`)
	matches := re.FindSubmatch(resp.Body())
	if len(matches) != 2 { //nolint:gomnd // we expect 2 matches
		slog.Error("extracting page data", slog.String("url", resp.Request.URL), slog.Any("found", matches))
		return nil, fmt.Errorf("extracting page data: no match")
	}

	pd := bytes.TrimSpace(matches[1])

	// write response to disk
	err = os.WriteFile(filepath.Join(cacheDir, "pages", makeExerciseID(year, day)), pd, 0o600)
	if err != nil {
		slog.Error("writing to cache", slog.String("url", resp.Request.URL), tint.Err(err))
		return nil, fmt.Errorf("writing cached puzzle page: %w", err)
	}

	return pd, nil
}

func (e *Exercise) downloadInput() ([]byte, error) {
	if cacheDir == "" {
		return nil, fmt.Errorf("cache directory not set")
	}

	err := appFs.MkdirAll(filepath.Join(cacheDir, "inputs"), 0o750)
	if err != nil {
		return nil, fmt.Errorf("creating inputs directory: %w", err)
	}

	slog.Info("downloading input",
		slog.String("file", filepath.Join(cacheDir, "inputs", e.ID)))

	resp, err := rClient.R().
		SetPathParams(map[string]string{
			"year": strconv.Itoa(e.Year),
			"day":  strconv.Itoa(e.Day),
		}).
		SetCookie(&http.Cookie{
			Name:   "session",
			Value:  cfg.GetString("advent.token"),
			Domain: ".adventofcode.com",
		}).
		Get("/{year}/day/{day}/input")
	if err != nil {
		return nil, fmt.Errorf("accessing input data page: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		slog.Error("getting input data",
			slog.Group("request",
				slog.String("method", resp.Request.Method),
				slog.String("url", resp.Request.URL),
				slog.Any("cookies", resp.Request.Cookies)),
			slog.String("status", resp.Status()),
			slog.Int("code", resp.StatusCode()))

		return nil, fmt.Errorf("downloading input data: %s", resp.Status())
	}

	data := bytes.TrimSpace(resp.Body())

	// write response to disk
	err = os.WriteFile(filepath.Join(cacheDir, "inputs", e.ID), data, 0o600)
	if err != nil {
		slog.Error("writing to cache", slog.String("url", resp.Request.URL), tint.Err(err))
		return nil, fmt.Errorf("writing cached input: %w", err)
	}

	return data, nil
}

func (e *Exercise) getInput() ([]byte, error) {
	d, err := e.getCachedInput()
	if err == nil {
		slog.Debug("using cached input", "exercise", e)
		return d, nil
	}

	slog.Debug("no cached input found; downloading input data", "exercise", e)
	return e.downloadInput()
}

//go:embed templates/readme.tmpl
var readmeTemplate []byte

//go:embed templates/go.tmpl
var goTemplate []byte

//go:embed templates/py.tmpl
var pyTemplate []byte

type tmplFile struct {
	Name     string
	Path     string
	Data     []byte
	FileName string
	Replace  bool
}

func (t *tmplFile) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("file", t.FileName),
		slog.String("name", t.Name),
		slog.String("path", t.Path),
		slog.Int("size", len(t.Data)),
		slog.Bool("replace", t.Replace),
	)
}

func (e *Exercise) addMissingFiles() error {
	var err error

	if e.Language == "" || e.Dir() == "" {
		return fmt.Errorf("incomplete exercise: missing language or directory")
	}

	implPath := filepath.Join(e.Path, e.Language)

	if err = appFs.MkdirAll(implPath, 0o750); err != nil {
		slog.Error("add exercise implementation path", tint.Err(err))
		return fmt.Errorf("creating %s implementation directory: %w", e.Language, err)
	}

	// TODO: give user option to overwrite existing files
	if err = e.writeInputFile(appFs, false); err != nil {
		return fmt.Errorf("writing input file: %w", err)
	}

	// TODO: give user option to overwrite existing files
	if err = e.writeInfoFile(appFs, false); err != nil {
		return fmt.Errorf("writing info file: %w", err)
	}

	tmpls := []tmplFile{
		{
			Name:     "readme",
			Path:     "",
			Data:     readmeTemplate,
			FileName: "README.md",
			Replace:  false,
		},
	}

	if e.Language == "go" {
		tmpls = append(tmpls, tmplFile{
			Name:     "go",
			Path:     "go",
			Data:     goTemplate,
			FileName: "exercise.go",
			Replace:  false,
		})
	} else if e.Language == "py" {
		tmpls = append(tmpls, tmplFile{
			Name:     "py",
			Path:     "py",
			Data:     pyTemplate,
			FileName: "__init__.py",
			Replace:  false,
		})
	}

	for _, t := range tmpls {
		slog.LogAttrs(context.TODO(), slog.LevelDebug, "add template file", slog.Any("template", t.LogValue()))

		err = e.addTemplatedFile(appFs, t)
		if err != nil {
			return fmt.Errorf("adding %s template: %w", t.FileName, err)
		}
	}

	return nil
}

func (e *Exercise) writeInputFile(fs afero.Fs, replace bool) error {
	fp := filepath.Join(e.Path, "input.txt")

	// check if the file exists already
	exists, err := afero.Exists(fs, fp)
	if err != nil {
		return fmt.Errorf("checking for input file: %w", err)
	}

	if exists && !replace {
		fmt.Fprintln(os.Stderr, "input file already exists, overwrite by using --force")
		return nil
	}

	inputFile, err := e.getInput()
	if err != nil {
		return fmt.Errorf("loading input: %w", err)
	}

	e.Data = &Data{
		Input:     string(inputFile),
		InputFile: "input.txt",
		TestCases: TestCase{
			One: []*Test{{Input: "", Expected: ""}},
			Two: []*Test{{Input: "", Expected: ""}},
		},
	}

	if err = afero.WriteFile(appFs, fp, inputFile, 0o600); err != nil {
		return fmt.Errorf("writing input file: %w", err)
	}

	slog.Debug("wrote input file", slog.String("path", fp))

	return nil
}

func (e *Exercise) writeInfoFile(fs afero.Fs, replace bool) error {
	fp := filepath.Join(e.Path, "info.json")

	// check if the file exists already
	exists, err := afero.Exists(fs, fp)
	if err != nil {
		return fmt.Errorf("checking for info file: %w", err)
	}

	if exists && !replace {
		fmt.Fprintln(os.Stderr, "info file already exists, overwrite by using --force")
		return nil
	}

	// marshall exercise data
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal exercise data: %w", err)
	}

	if err = afero.WriteFile(appFs, fp, data, 0o600); err != nil {
		return fmt.Errorf("write info file: %w", err)
	}

	slog.Debug("wrote info file", slog.String("path", fp))

	return nil
}

func (e *Exercise) addTemplatedFile(fs afero.Fs, tf tmplFile) error {
	// only write if file doesn't exist or if we're replacing it
	fp := filepath.Join(e.Path, tf.Path, tf.FileName)

	exists, _ := afero.Exists(fs, fp)
	if exists && !tf.Replace {
		slog.Info("file exists, skipping", "template", tf.LogValue())

		fmt.Printf("%s already exists, overwrite by using --force\n", fp)

		return nil
	}

	t := template.Must(template.New(tf.Name).Parse(string(tf.Data)))
	b := new(bytes.Buffer)

	if err := t.Execute(b, e); err != nil {
		return fmt.Errorf("executing %s template: %w", tf.Name, err)
	}

	return afero.WriteFile(fs, fp, b.Bytes(), 0o600)
}
