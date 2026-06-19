package exercise

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/lmittmann/tint"
	"github.com/spf13/afero"
	"golang.org/x/net/html"

	"github.com/asphaltbuffet/elf/internal/utilities"
	"github.com/asphaltbuffet/elf/pkg/config"
)

const (
	// dirPerm is the permission mode for directories (rwxr-x---).
	dirPerm = 0o750
)

// Sentinel errors for downloader construction and HTTP operations.
var (
	ErrNotConfigured   = errors.New("not configured")
	ErrHTTPRequest     = errors.New("http request")
	ErrHTTPResponse    = errors.New("http response")
	ErrEmptyInput      = errors.New("empty puzzle input")
	ErrInvalidURL      = errors.New("invalid URL")
	ErrInvalidLanguage = errors.New("invalid language")
)

// Downloader coordinates fetching a challenge from the AoC website and laying it out on disk. It
// assembles a finished Exercise value (via the page fetcher or an existing info.json) and hands it
// to the scaffold to write — it never builds an Exercise up in place.
type Downloader struct {
	exerciseBaseDir string
	cfgDir          string
	language        string
	url             string
	skipImpl        bool
	appFs           afero.Fs
	logger          *slog.Logger

	fetcher  *pageFetcher
	scaffold *exerciseScaffold

	// path is the resolved exercise directory, set once Download assembles or loads the Exercise.
	path string
}

// Overwrites controls which existing exercise files are overwritten during download.
type Overwrites struct {
	Input bool
}

// NewDownloader creates a Downloader, applies options, and validates the configuration.
func NewDownloader(cfg config.Config, options ...func(*Downloader)) (*Downloader, error) {
	d := &Downloader{
		language:        cfg.GetLanguage(),
		cfgDir:          cfg.GetConfigDir(),
		exerciseBaseDir: cfg.GetBaseDir(),
		appFs:           cfg.GetFs(),
		logger:          cfg.GetLogger(),
		fetcher: &pageFetcher{
			rClient:  resty.New().SetBaseURL("https://adventofcode.com"),
			token:    cfg.GetToken(),
			cacheDir: cfg.GetCacheDir(),
			fs:       cfg.GetFs(),
			logger:   cfg.GetLogger(),
		},
		scaffold: &exerciseScaffold{
			fs:            cfg.GetFs(),
			inputFileName: cfg.GetInputFilename(),
			logger:        cfg.GetLogger(),
		},
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
			d.language = lang
		}
	}
}

// WithURL sets the exercise URL to download.
func WithURL(url string) func(*Downloader) {
	return func(d *Downloader) {
		d.url = url
	}
}

// WithOverwrites sets the files that can be overwritten if already in place.
func WithOverwrites(o *Overwrites) func(*Downloader) {
	return func(d *Downloader) {
		if o == nil {
			d.scaffold.overwrites = &Overwrites{}
		} else {
			d.scaffold.overwrites = o
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

	if d.fetcher.rClient == nil {
		err = append(err, fmt.Errorf("http client: %w", ErrNotConfigured))
	}

	if d.appFs == nil {
		err = append(err, fmt.Errorf("filesystem: %w", ErrNotConfigured))
	}

	// the token cannot be empty if we're downloading the input
	if d.fetcher.token == "" {
		err = append(err, fmt.Errorf("advent user token: %w", ErrNotConfigured))
	}

	if !d.skipImpl && d.language == "" {
		err = append(err, fmt.Errorf("implementation language: %w", ErrNotConfigured))
	}

	if d.cfgDir == "" {
		err = append(err, fmt.Errorf("user config directory: %w", ErrNotConfigured))
	}

	if d.fetcher.cacheDir == "" {
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

// Download fetches challenge metadata, puzzle input, and implementation templates from the AoC website.
func (d *Downloader) Download() error {
	year, day, err := ParseURL(d.url)
	if err != nil {
		return err
	}

	// update client with year and day
	d.fetcher.rClient.
		SetHeader("User-Agent", "github.com/asphaltbuffet/elf").
		SetPathParams(map[string]string{
			"year": strconv.Itoa(year),
			"day":  strconv.Itoa(day),
		})

	var ex *Exercise

	if exPath, ok := d.getExercisePath(year, day); ok {
		ex = &Exercise{Path: exPath, Language: d.language}
		if err = ex.loadInfo(d.appFs, d.logger); err == nil {
			// loadInfo populates metadata from info.json but not the puzzle input; fetch it so the
			// scaffold writes real input instead of an empty input.txt.
			var input []byte
			if input, err = d.fetcher.fetchInput(year, day); err == nil {
				ex.Data.InputData = string(input)
			}
		}
	} else {
		ex, err = d.assemble(year, day)
	}
	if err != nil {
		d.logger.Error("loading exercise", tint.Err(err))
		return err
	}

	d.path = ex.Path

	// the exercise is fully assembled; lay it out on disk
	if err = d.scaffold.write(ex); err != nil {
		d.logger.Error("add missing files", slog.Int("year", year), slog.Int("day", day), tint.Err(err))
		return err
	}

	d.logger.Debug("exercise added", slog.String("dir", ex.Path))

	return nil
}

// assemble fetches the puzzle page and input and returns a finished Exercise value. It never
// mutates the Downloader — the result is a fresh value ready for the scaffold.
func (d *Downloader) assemble(year, day int) (*Exercise, error) {
	logger := d.logger.With(slog.Int("year", year), slog.Int("day", day), slog.String("fn", "assemble"))
	logger.Debug("loading exercise")

	page, err := d.fetcher.fetchPage(year, day)
	if err != nil {
		logger.Debug("getting page data", slog.String("url", d.url), tint.Err(err))
		return nil, fmt.Errorf("get page data %d-%02d: %w", year, day, err)
	}

	title, err := extractTitle(page)
	if err != nil {
		logger.Debug("extracting title", slog.Int("page-size", len(page)), tint.Err(err))
		return nil, fmt.Errorf("extract %d-%02d title: %w", year, day, err)
	}

	input, err := d.fetcher.fetchInput(year, day)
	if err != nil {
		return nil, fmt.Errorf("loading input: %w", err)
	}

	ex := &Exercise{
		ID:       makeExerciseID(year, day),
		Title:    title,
		Language: d.language,
		Year:     year,
		Day:      day,
		URL:      d.url,
		Path:     makeExercisePath(d.exerciseBaseDir, year, day, title),
		Data: &Data{
			InputData:     string(input),
			InputFileName: d.scaffold.inputFileName,
			TestCases: TestCase{
				One: []*Test{{Input: "", Expected: ""}},
				Two: []*Test{{Input: "", Expected: ""}},
			},
			Answers: Answer{One: "", Two: ""},
		},
	}

	logger.Debug("loaded exercise", slog.Any("exercise", ex.LogValue()))

	return ex, nil
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

// ParseURL extracts the year and day from an Advent of Code puzzle URL.
func ParseURL(url string) (int, int, error) {
	var year, day int

	// regex here is validating year/day are integers, if this changes, add validation below
	re := regexp.MustCompile(`^https?://(www\.)?adventofcode\.com/(?P<year>\d{4})/day/(?P<day>\d{1,2})`)

	matches := findNamedMatches(re, url)
	if len(matches) != 2 { //nolint:mnd // we expect 2 matches
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
	if len(matches) != 2 { //nolint:mnd // we expect 2 matches
		return "", fmt.Errorf("%w: no match", ErrInvalidData)
	}

	return matches[1], nil
}

func makeExercisePath(baseDir string, year, day int, title string) string {
	return filepath.Join(
		baseDir,
		strconv.Itoa(year),
		fmt.Sprintf("%02d-%s", day, utilities.ToCamel(title)),
	)
}

// FilePath returns the local filesystem path where the downloaded exercise will be stored.
func (d *Downloader) FilePath() string {
	return d.path
}
