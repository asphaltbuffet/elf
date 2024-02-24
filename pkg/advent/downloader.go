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
	ErrInvalidLanguage  = errors.New("invalid language")
)

type Downloader struct {
	*Exercise
	exerciseBaseDir string
	cacheDir        string
	cfgDir          string
	inputFileName   string
	rClient         *resty.Client
	token           string
	overwrites      *Overwrites
	skipImpl        bool
}

type Overwrites struct {
	Input bool
}

func NewDownloader(config krampus.DownloadConfiguration, options ...func(*Downloader)) (*Downloader, error) {
	if config == nil {
		return nil, ErrNilConfiguration
	}

	d := &Downloader{
		Exercise: &Exercise{
			ID:       "",
			Title:    "",
			Language: config.GetLanguage(),
			Year:     0,
			Day:      0,
			URL:      "",
			Data:     nil,
			Path:     "",
			runner:   nil, // not used when downloading
			appFs:    config.GetFs(),
			logger:   config.GetLogger(),
		},
		cacheDir:        config.GetCacheDir(),
		cfgDir:          config.GetConfigDir(),
		exerciseBaseDir: config.GetBaseDir(),
		rClient:         resty.New().SetBaseURL("https://adventofcode.com"),
		token:           config.GetToken(),
		inputFileName:   config.GetInputFilename(),
	}

	for _, option := range options {
		option(d)
	}

	if err := d.validate(); err != nil {
		return nil, err
	}

	return d, nil
}

// WithDownloadLanguage sets the language for the exercise implementation.
// This will override any language set in the configuration.
func WithDownloadLanguage(lang string) func(*Downloader) {
	return func(d *Downloader) {
		if lang != "" {
			// expect to check for valid language later
			d.Language = lang
		}
	}
}

// WithURL sets the exercise URL to download.
func WithURL(url string) func(*Downloader) {
	return func(d *Downloader) {
		d.URL = url
	}
}

// WithOverwrites sets the files that can be overwritten if already in place.
func WithOverwrites(o *Overwrites) func(*Downloader) {
	return func(d *Downloader) {
		if o == nil {
			d.overwrites = &Overwrites{}
		} else {
			d.overwrites = o
		}
	}
}

// WithSkipImpl sets the downloader to skip creating implementation files and structure.
func WithSkipImpl(skip bool) func(*Downloader) {
	return func(d *Downloader) {
		d.skipImpl = skip
	}
}

func (d *Downloader) validate() error {
	var err []error

	if d.rClient == nil {
		err = append(err, fmt.Errorf("http client: %w", ErrNotConfigured))
	}

	if d.appFs == nil {
		err = append(err, fmt.Errorf("filesystem: %w", ErrNotConfigured))
	}

	// the token cannot be empty if we're downloading the input
	if d.token == "" {
		err = append(err, fmt.Errorf("advent user token: %w", ErrNotConfigured))
	}

	if !d.skipImpl && d.Language == "" {
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
	year, day, err := ParseURL(d.URL)
	if err != nil {
		return err
	}

	// update client with year and day
	d.rClient.
		SetHeader("User-Agent", "github.com/asphaltbuffet/elf").
		SetPathParams(map[string]string{
			"year": strconv.Itoa(year),
			"day":  strconv.Itoa(day),
		})

	exPath, ok := d.getExercisePath(year, day)
	if ok {
		d.Exercise.Path = exPath
		err = d.loadInfo(d.appFs)
	} else {
		err = d.loadFromURL(year, day)
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

	d.logger.Debug("exercise added", slog.String("dir", d.Path))

	return nil
}

func (d *Downloader) loadFromURL(year, day int) error {
	logger := d.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "loadFromURL"))
	logger.Debug("loading exercise")

	var (
		page  []byte
		title string
		err   error
	)

	page, err = d.getPage(year, day)
	if err != nil {
		logger.Debug("getting page data", slog.String("url", d.URL), tint.Err(err))
		return fmt.Errorf("get page data %d-%02d: %w", year, day, err)
	}

	title, err = extractTitle(page)
	if err != nil {
		logger.Debug("extracting title", slog.Int("page-size", len(page)), tint.Err(err))
		return fmt.Errorf("extract %d-%02d title: %w", year, day, err)
	}

	d.Exercise.ID = makeExerciseID(year, day)
	d.Exercise.Title = title
	d.Exercise.Year = year
	d.Exercise.Day = day
	d.Exercise.Path = makeExercisePath(d.exerciseBaseDir, year, day, title)

	logger.Debug("loaded exercise", slog.Any("exercise", d.LogValue()))

	return nil
}

func (d *Downloader) getExercisePath(year, day int) (string, bool) {
	logger := d.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "getExercisePath"))

	var exPath string
	dayPrefix := fmt.Sprintf("%02d-", day)

	logger.Debug("searching for exercise dir", slog.String("root", d.exerciseBaseDir), slog.String("prefix", dayPrefix))

	_ = afero.Walk(d.appFs, d.exerciseBaseDir, func(path string, info fs.FileInfo, err error) error {
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

	return exPath, exPath != ""
}

func extractTitle(page []byte) (string, error) {
	doc, err := html.Parse(bytes.NewReader(page))
	if err != nil {
		return "", err
	}

	extract, err := getH2NodeFromHTML(doc)
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
	logger := d.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "getPage"))

	pageData, ok := d.getCachedPage(year, day)
	if ok {
		logger.Debug("using cached puzzle page", slog.Int("size", len(pageData)))
		return pageData, nil
	}

	logger.Info("no cached page")

	return d.downloadPage(year, day)
}

func ParseURL(url string) (int, int, error) {
	var year, day int

	// regex here is validating year/day are integers, if this changes, add validation below
	re := regexp.MustCompile(`^https?://(www\.)?adventofcode\.com/(?P<year>\d{4})/day/(?P<day>\d{1,2})`)

	matches := findNamedMatches(re, url)
	if len(matches) != 2 { //nolint:gomnd // we expect 2 matches
		return 0, 0, fmt.Errorf("parse %s: %w", url, ErrInvalidURL)
	}

	// ignore errors; we already validated type via regex
	year, _ = strconv.Atoi(matches["year"])
	day, _ = strconv.Atoi(matches["day"])

	return year, day, nil
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
	if err := d.appFs.MkdirAll(pageCacheDir, 0o750); err != nil {
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

	if len(matches) != 2 { //nolint:gomnd // we expect 2 matches
		logger.Debug("extracting page data", slog.String("url", resp.Request.URL), slog.Any("found", matches))

		return nil, fmt.Errorf("extracting page data: no match")
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

	err := d.appFs.MkdirAll(filepath.Join(d.cacheDir, "inputs"), 0o750)
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

	logger.Info("no cached page")

	return d.downloadInput(year, day)
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

	implPath := filepath.Join(d.Path, d.Language)

	if err = d.appFs.MkdirAll(implPath, 0o750); err != nil {
		logger.Error("add exercise implementation path", tint.Err(err))
		return fmt.Errorf("creating %s implementation directory: %w", d.Language, err)
	}

	// TODO: give user option to overwrite existing files
	if err = d.writeInputFile(); err != nil {
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

	switch d.Language {
	case "go":
		tmpls = append(tmpls, tmplFile{
			Name:     "go",
			Path:     "go",
			Data:     goTemplate,
			FileName: "exercise.go",
			Replace:  false,
		})

	case "py":
		tmpls = append(tmpls, tmplFile{
			Name:     "py",
			Path:     "py",
			Data:     pyTemplate,
			FileName: "__init__.py",
			Replace:  false,
		})

	default:
		return fmt.Errorf("template %s files: %w", d.Language, ErrInvalidLanguage)
	}

	for _, t := range tmpls {
		logger.Debug("add template file", slog.Any("template", t.LogValue()))

		err = d.addTemplatedFile(t)
		if err != nil {
			return fmt.Errorf("adding %q template: %w", t.FileName, err)
		}
	}

	return nil
}

func (d *Downloader) writeInputFile() error {
	logger := d.logger.With(slog.String("fn", "writeInputFile"))

	fp := filepath.Join(d.Path, d.inputFileName)

	// check if the file exists already
	exists, err := afero.Exists(d.appFs, fp)
	if err != nil {
		return err
	}

	if exists && !d.overwrites.Input {
		logger.Info("found %s, overwrite by using '--force-input'", slog.String("file", fp))
		return nil
	}

	inputFile, err := d.getInput(d.Year, d.Day)
	if err != nil {
		return fmt.Errorf("loading input: %w", err)
	}

	d.Exercise.Data = &Data{
		InputData:     string(inputFile),
		InputFileName: d.inputFileName,
		TestCases: TestCase{
			One: []*Test{{Input: "", Expected: ""}},
			Two: []*Test{{Input: "", Expected: ""}},
		},
		Answers: Answer{
			One: "",
			Two: "",
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

	fp := filepath.Join(d.Path, "info.json") // TODO: filename should be in config

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
	data, err := json.MarshalIndent(d.Exercise, "", "  ")
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
	fp := filepath.Join(d.Path, templateFile.Path, templateFile.FileName)
	logger := d.logger.With(slog.String("fn", "addTemplatedFile"))

	// only write if file doesn't exist or if we're replacing it
	exists, err := afero.Exists(d.appFs, fp)
	if err != nil {
		return fmt.Errorf("checking for %q: %w", fp, err)
	}

	if exists && !templateFile.Replace {
		logger.Debug("file exists, skipping", "template", templateFile.LogValue())

		fmt.Printf("%s already exists, overwrite by using --force\n", fp)

		return nil
	}

	t := template.Must(template.New(templateFile.Name).Parse(string(templateFile.Data)))
	b := new(bytes.Buffer)

	if err = t.Execute(b, d); err != nil {
		return fmt.Errorf("template %q: %w", templateFile.Name, err)
	}

	return afero.WriteFile(d.appFs, fp, b.Bytes(), 0o600)
}

func makeExercisePath(baseDir string, year, day int, title string) string {
	return filepath.Join(
		baseDir,
		strconv.Itoa(year),
		fmt.Sprintf("%02d-%s", day, utilities.ToCamel(title)),
	)
}
