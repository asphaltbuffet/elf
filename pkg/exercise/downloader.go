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

func NewDownloader(cfg config.DownloadConfiguration, options ...func(*Downloader)) (*Downloader, error) {
	if cfg == nil {
		return nil, ErrNilConfiguration
	}

	d := &Downloader{
		Exercise: &Exercise{
			ID:       "",
			Title:    "",
			Language: cfg.GetLanguage(),
			Year:     0,
			Day:      0,
			URL:      "",
			Data:     nil,
			Path:     "",
			runner:   nil, // not used when downloading
			appFs:    cfg.GetFs(),
			logger:   cfg.GetLogger(),
		},
		cacheDir:        cfg.GetCacheDir(),
		cfgDir:          cfg.GetConfigDir(),
		exerciseBaseDir: cfg.GetBaseDir(),
		rClient:         resty.New().SetBaseURL("https://adventofcode.com"),
		token:           cfg.GetToken(),
		inputFileName:   cfg.GetInputFilename(),
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
		err = d.loadInfo()
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

func (d *Downloader) FilePath() string {
	return d.Path
}
