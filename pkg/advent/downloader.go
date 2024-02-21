package advent

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
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

	"github.com/asphaltbuffet/elf/pkg/krampus"
	"github.com/asphaltbuffet/elf/pkg/utilities"
)

var (
	ErrNotConfigured    = errors.New("not configured")
	ErrNilConfiguration = errors.New("nil configuration")
	ErrHTTPRequest      = errors.New("http request")
	ErrHTTPResponse     = errors.New("http response")
	ErrInvalidURL       = errors.New("invalid URL")
)

type Downloader struct {
	appFs           afero.Fs
	exerciseBaseDir string
	cacheDir        string
	cfgDir          string
	exercise        *Exercise
	lang            string
	logger          *slog.Logger
	rClient         *resty.Client
	token           string
	url             string
}

func NewDownloader(config *krampus.Config, url, lang string) (*Downloader, error) {
	if config == nil {
		return nil, ErrNilConfiguration
	}

	// use the language from the config if none is provided
	if lang == "" {
		// we'll validate this was set to something later
		lang = config.GetLanguage()
	}

	// set up logger; if not provided, make one default
	logger := config.GetLogger()
	if logger == nil {
		logger = slog.New(tint.NewHandler(os.Stderr, nil))
	}
	logger = logger.With(slog.String("action", "download"))

	d := &Downloader{
		appFs:           config.GetFs(),
		cacheDir:        config.GetCacheDir(),
		exerciseBaseDir: config.GetBaseDir(),
		cfgDir:          config.GetConfigDir(),
		exercise:        nil,
		lang:            lang,
		logger:          logger,
		rClient:         resty.New().SetBaseURL("https://adventofcode.com"),
		token:           config.GetToken(),
		url:             url,
	}

	if err := d.validate(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Downloader) validate() error {
	var err []error

	if d.rClient == nil {
		err = append(err, fmt.Errorf("http client: %w", ErrNotConfigured))
	}

	if d.appFs == nil {
		err = append(err, fmt.Errorf("filesystem: %w", ErrNotConfigured))
	}

	if d.token == "" {
		err = append(err, fmt.Errorf("advent user token: %w", ErrNotConfigured))
	}

	if d.lang == "" {
		err = append(err, fmt.Errorf("implementation language: %w", ErrNotConfigured))
	}

	if d.cfgDir == "" {
		err = append(err, fmt.Errorf("user config directory: %w", ErrNotConfigured))
	}

	if d.cacheDir == "" {
		err = append(err, fmt.Errorf("cache directory: %w", ErrNotConfigured))
	}

	if d.exerciseBaseDir == "" {
		err = append(err, fmt.Errorf("advent solution root: %w", ErrNotConfigured))
	} else if d.appFs != nil {
		if _, statErr := d.appFs.Stat(d.exerciseBaseDir); err != nil {
			err = append(err, statErr)
		}
	}

	return errors.Join(err...)
}

func (d *Downloader) Download() error {
	year, day, err := ParseURL(d.url)
	if err != nil {
		return err
	}

	// update client with exercise-specific data
	d.rClient.
		SetHeader("User-Agent", "github.com/asphaltbuffet/elf").
		SetPathParams(map[string]string{
			"year": strconv.Itoa(year),
			"day":  strconv.Itoa(day),
		})

	d.exercise = &Exercise{}

	exPath, ok := d.getExercisePath(year, day)
	if ok {
		d.exercise.Path = exPath
		err = d.exercise.loadInfo(d.appFs)
	} else {
		d.exercise, err = d.loadFromURL(year, day)
	}
	if err != nil {
		d.logger.Error("loading exercise", tint.Err(err))
		return err
	}

	// the basic exercise information is here; add missing elements
	if err = d.addMissingFiles(); err != nil {
		d.logger.Error("add missing files", slog.Int("year", year), slog.Int("day", day), tint.Err(err))
		return err
	}

	d.logger.Debug("exercise added", slog.String("dir", d.exercise.Path))

	return nil
}

func (d *Downloader) Path() string {
	return d.exercise.Path
}

func (d *Downloader) loadFromURL(year, day int) (*Exercise, error) {
	logger := d.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "loadFromURL"))

	var (
		page  []byte
		title string
		err   error
		e     *Exercise
	)

	page, err = d.getPage(year, day)
	if err != nil {
		logger.Error("getting page data", tint.Err(err))

		return nil, err
	}

	title, err = extractTitle(page)
	if err != nil {
		slog.Error("extracting title from page data", tint.Err(err))

		return nil, err
	}

	e = &Exercise{
		ID:       makeExerciseID(year, day),
		Title:    title,
		Year:     year,
		Language: d.lang,
		Day:      day,
		URL:      d.url,
		Data:     nil, // this should be empty, we only load this from info.json
		Path:     makeExercisePath(d.exerciseBaseDir, year, day, title),
	}

	return e, nil
}

func (d *Downloader) getExercisePath(year, day int) (string, bool) {
	logger := d.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "getExercisePath"))

	var exPath string
	dayPrefix := fmt.Sprintf("%02d-", day)
	logger.Debug("searching for exercise dir", slog.String("root", d.exerciseBaseDir), slog.String("prefix", dayPrefix))

	err := afero.Walk(d.appFs, d.exerciseBaseDir, func(path string, info fs.FileInfo, err error) error {
		switch {
		case err != nil:
			return nil //nolint:nilerr // errors are used to abort walking

		case !info.IsDir():
			logger.Debug("skipping non-directory", slog.String("path", path))
			fallthrough
		case path == d.exerciseBaseDir:
			return nil

		case strings.HasPrefix(info.Name(), dayPrefix):
			logger.Debug("found exercise directory", slog.String("path", path))
			exPath = path

			// we found the directory we're looking for, stop walking
			return filepath.SkipAll

		case info.Name() == strconv.Itoa(year):
			logger.Debug("found year directory", slog.String("path", path))
			return nil

		default:
			logger.Debug("skipping non-year directory", slog.String("path", path))
			// we only recurse into the specified year directory until we find the wanted day
			return filepath.SkipDir
		}
	})
	switch {
	case errors.Is(err, filepath.SkipAll):
		logger.Debug("found exercise", slog.String("found", exPath))

	default:
		logger.Debug("did not find exercise", tint.Err(err))
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
		return "", fmt.Errorf("%w: no match", ErrInvalidData)
	}

	return matches[1], nil
}

func (d *Downloader) getPage(year, day int) ([]byte, error) {
	pageData, err := d.getCachedPuzzlePage(year, day) // we're never going to see the error this returns
	if err == nil {
		slog.Debug("using cached puzzle page",
			slog.Int("year", year),
			slog.Int("day", day),
			slog.Int("size", len(pageData)))
		return pageData, nil
	}

	slog.Debug("downloading puzzle page", slog.Int("year", year), slog.Int("day", day))

	return d.downloadPuzzlePage(year, day)
}

func ParseURL(url string) (int, int, error) {
	var y, d int

	// regex here is validating year/day are integers, if this changes, add validation below
	re := regexp.MustCompile(`^https?://(www\.)?adventofcode\.com/(?P<year>\d{4})/day/(?P<day>\d{1,2})`)

	matches := findNamedMatches(re, url)
	if len(matches) != 2 { //nolint:gomnd // we expect 2 matches
		slog.Debug("parsing URL", slog.String("url", url), slog.Any("found", matches))
		return 0, 0, fmt.Errorf("parse %s: %w", url, ErrInvalidURL)
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

func (d *Downloader) getCachedPuzzlePage(year, day int) ([]byte, error) {
	if d.cacheDir == "" {
		return nil, fmt.Errorf("cache directory not set")
	}

	fp := filepath.Join(d.cacheDir, "pages", makeExerciseID(year, day))

	f, err := afero.ReadFile(d.appFs, fp)
	if err != nil {
		return nil, fmt.Errorf("reading puzzle page: %w", err)
	}

	return f, nil
}

func (d *Downloader) getCachedInput(year, day int) ([]byte, error) {
	if d.cacheDir == "" {
		return nil, fmt.Errorf("cache directory: %w", ErrNotConfigured)
	}

	fp := filepath.Join(d.cacheDir, "inputs", makeExerciseID(year, day))

	f, err := afero.ReadFile(d.appFs, fp)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (d *Downloader) downloadPuzzlePage(year, day int) ([]byte, error) {
	logger := d.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "downloadPuzzlePage"))

	if d.cacheDir == "" {
		return nil, fmt.Errorf("cache directory: %w", ErrNotConfigured)
	}

	// make sure we can write the cached file before we download it
	err := d.appFs.MkdirAll(filepath.Join(d.cacheDir, "pages"), 0o750)
	if err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	logger.Debug("downloading puzzle page",
		slog.String("file", filepath.Join(d.cacheDir, "pages", makeExerciseID(year, day))))

	req := d.rClient.R().SetPathParams(map[string]string{
		"year": strconv.Itoa(year),
		"day":  strconv.Itoa(day),
	})

	resp, err := req.Get("/{year}/day/{day}")
	if err != nil {
		return nil, errors.Join(ErrHTTPRequest, err)
	}

	if resp.StatusCode() != http.StatusOK {
		slog.Debug("download page response",
			slog.String("url", resp.Request.URL),
			slog.String("status", http.StatusText(resp.StatusCode())),
			slog.Int("code", resp.StatusCode()))

		return nil, fmt.Errorf("%w: %s: %s", ErrHTTPResponse, resp.Request.Method, resp.Status())
	}

	// only keep relevant parts of the page
	re := regexp.MustCompile(`(?s)<article.*?>(.*)</article>`)
	matches := re.FindSubmatch(resp.Body())
	if len(matches) != 2 { //nolint:gomnd // we expect 2 matches
		slog.Debug("extracting page data", slog.String("url", resp.Request.URL), slog.Any("found", matches))
		return nil, fmt.Errorf("extracting page data: no match")
	}

	pd := bytes.TrimSpace(matches[1])

	// write response to disk
	err = afero.WriteFile(d.appFs, filepath.Join(d.cacheDir, "pages", makeExerciseID(year, day)), pd, 0o600)
	if err != nil {
		slog.Debug("writing to cache", slog.String("url", resp.Request.URL), tint.Err(err))
		return nil, fmt.Errorf("writing cached puzzle page: %w", err)
	}

	return pd, nil
}

func (d *Downloader) downloadInput(year, day int) ([]byte, error) {
	logger := d.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "downloadInput"))

	if d.cacheDir == "" {
		return nil, fmt.Errorf("cache directory: %w", ErrNotConfigured)
	}

	err := d.appFs.MkdirAll(filepath.Join(d.cacheDir, "inputs"), 0o750)
	if err != nil {
		return nil, fmt.Errorf("creating inputs directory: %w", err)
	}

	logger.Debug("downloading input")

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

		return nil, fmt.Errorf("%w: %s: %s", ErrHTTPResponse, resp.Request.Method, resp.Status())
	}

	data := bytes.TrimSpace(resp.Body())

	// write response to disk
	err = afero.WriteFile(d.appFs, filepath.Join(d.cacheDir, "inputs", makeExerciseID(year, day)), data, 0o600)
	if err != nil {
		logger.Error("writing to cache", tint.Err(err))
		return nil, fmt.Errorf("writing cached input: %w", err)
	}

	return data, nil
}

func (d *Downloader) getInput(year, day int) ([]byte, error) {
	logger := d.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "getInput"))

	data, err := d.getCachedInput(year, day)
	if err != nil { // TODO: we should only continue if the error is "file not found"
		logger.Debug("failed to get cached input data", tint.Err(err))
		return d.downloadInput(year, day)
	}

	logger.Debug("got cached input")
	return data, nil
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

func (d *Downloader) addMissingFiles() error {
	logger := d.logger.With(slog.String("fn", "addMissingFiles"))

	var err error

	if d.exercise.Language == "" || d.exercise.Dir() == "" {
		return fmt.Errorf("incomplete exercise: missing language or directory")
	}

	implPath := filepath.Join(d.exercise.Path, d.exercise.Language)

	if err = d.appFs.MkdirAll(implPath, 0o750); err != nil {
		logger.Error("add exercise implementation path", tint.Err(err))
		return fmt.Errorf("creating %s implementation directory: %w", d.lang, err)
	}

	// TODO: give user option to overwrite existing files
	if err = d.writeInputFile(false); err != nil {
		return fmt.Errorf("writing input file: %w", err)
	}

	// TODO: give user option to overwrite existing files
	if err = d.writeInfoFile(false); err != nil {
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

	if d.lang == "go" {
		tmpls = append(tmpls, tmplFile{
			Name:     "go",
			Path:     "go",
			Data:     goTemplate,
			FileName: "exercise.go",
			Replace:  false,
		})
	} else if d.lang == "py" {
		tmpls = append(tmpls, tmplFile{
			Name:     "py",
			Path:     "py",
			Data:     pyTemplate,
			FileName: "__init__.py",
			Replace:  false,
		})
	}

	for _, t := range tmpls {
		logger.Debug("add template file", slog.Any("template", t.LogValue()))

		err = d.addTemplatedFile(t)
		if err != nil {
			return fmt.Errorf("adding %s template: %w", t.FileName, err)
		}
	}

	return nil
}

func (d *Downloader) writeInputFile(replace bool) error {
	logger := d.logger.With(slog.String("fn", "writeInputFile"))

	fp := filepath.Join(d.exercise.Path, "input.txt") // TODO: this should be configurable

	// check if the file exists already
	exists, err := afero.Exists(d.appFs, fp)
	if err != nil {
		return err
	}

	if exists && !replace {
		logger.Warn("input file already exists, overwrite by using --force", slog.String("file", fp))
		return nil
	}

	inputFile, err := d.getInput(d.exercise.Year, d.exercise.Day)
	if err != nil {
		return fmt.Errorf("loading input: %w", err)
	}

	d.exercise.Data = &Data{
		InputData:     string(inputFile),
		InputFileName: "input.txt",
		TestCases: TestCase{
			One: []*Test{{Input: "", Expected: ""}},
			Two: []*Test{{Input: "", Expected: ""}},
		},
	}

	if err = afero.WriteFile(d.appFs, fp, inputFile, 0o600); err != nil {
		return fmt.Errorf("writing input file: %w", err)
	}

	logger.Debug("wrote input file", slog.String("path", fp))

	return nil
}

func (d *Downloader) writeInfoFile(replace bool) error {
	logger := d.logger.With(slog.String("fn", "writeInfoFile"))

	fp := filepath.Join(d.exercise.Path, "info.json") // TODO: filename should be in config

	// check if the file exists already
	exists, err := afero.Exists(d.appFs, fp)
	if err != nil {
		return fmt.Errorf("checking for info file: %w", err)
	}

	if exists && !replace {
		logger.Warn("info file already exists, overwrite by using --force",
			slog.String("file", fp))
		return nil
	}

	// marshall exercise data
	data, err := json.MarshalIndent(d.exercise, "", "  ")
	if err != nil {
		return err
	}

	if err = afero.WriteFile(d.appFs, fp, data, 0o600); err != nil {
		return fmt.Errorf("write info file: %w", err)
	}

	logger.Debug("wrote info file", slog.String("path", fp))

	return nil
}

func (d *Downloader) addTemplatedFile(templateFile tmplFile) error {
	fp := filepath.Join(d.exercise.Path, templateFile.Path, templateFile.FileName)

	// only write if file doesn't exist or if we're replacing it
	exists, _ := afero.Exists(d.appFs, fp) // TODO: handle error
	if exists && !templateFile.Replace {
		slog.Debug("file exists, skipping", "template", templateFile.LogValue())

		fmt.Printf("%s already exists, overwrite by using --force\n", fp)

		return nil
	}

	t := template.Must(template.New(templateFile.Name).Parse(string(templateFile.Data)))
	b := new(bytes.Buffer)

	if err := t.Execute(b, d.exercise); err != nil {
		return fmt.Errorf("executing %s template: %w", templateFile.Name, err)
	}

	return afero.WriteFile(d.appFs, fp, b.Bytes(), 0o600)
}

func makeExercisePath(baseDir string, year, day int, title string) string {
	return filepath.Join(baseDir, strconv.Itoa(year), fmt.Sprintf("%02d-%s", day, utilities.ToCamel(title)))
}
