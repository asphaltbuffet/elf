package advent

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"text/template"

	"golang.org/x/net/html"

	"github.com/go-resty/resty/v2"
	"github.com/lmittmann/tint"
	"github.com/spf13/afero"
)

var (
	rClient = resty.New().SetBaseURL("https://adventofcode.com")
	cfgDir  string
	appFs   = afero.NewOsFs()
	logger  = slog.With(slog.String("action", "download"))
)

func Download(url string, lang string, _ bool) (string, error) {
	if cfgDir == "" {
		logger.Error("no config directory")
		return "", fmt.Errorf("cache directory not set")
	}

	year, day, err := ParseURL(url)
	if err != nil {
		logger.Error("creating exercise from URL", slog.String("url", url), slog.String("lang", lang), tint.Err(err))
		return "", fmt.Errorf("creating exercise from URL: %w", err)
	}

	// update client with exercise-specific data
	rClient.
		SetOutputDirectory(cfgDir).
		SetHeader("User-Agent", "github.com/asphaltbuffet/elf").
		SetPathParams(map[string]string{
			"year": strconv.Itoa(year),
			"day":  strconv.Itoa(day),
		})

	var e *Exercise

	exPath, ok := getExercisePath(year, day)
	if ok {
		e, err = loadExisting(exPath)
	} else {
		e, err = loadFromURL(url, year, day, lang)
	}
	if err != nil {
		logger.Error("loading exercise", slog.String("url", url), slog.String("lang", lang), tint.Err(err))
		return "", fmt.Errorf("loading exercise: %w", err)
	}

	// the basic exercise information is here; add missing elements
	if err = e.addMissingFiles(); err != nil {
		logger.Error("adding missing files", slog.Any("exercise", e), tint.Err(err))
		return "", fmt.Errorf("adding missing files: %w", err)
	}

	logger.Info("exercise added", slog.String("url", e.URL), slog.String("dir", e.Dir()))

	return e.path, nil
}

func loadExisting(path string) (*Exercise, error) {
	var (
		err error
		e   *Exercise
	)

	infoPath := filepath.Join(path, "info.json")

	_, err = appFs.Stat(infoPath)
	if err == nil {
		// exercise exists, we may need to update it
		logger.Info("update existing exercise", slog.String("dir", path))

		// TODO: a bad info.json will cause this to behave unpredictably
		// TODO: if this fails, try to create a new exercise, or tell user to delete file(s)
		e, err = NewExerciseFromInfo(path)
		if err != nil {
			logger.Error("creating exercise from info", slog.String("dir", path), tint.Err(err))
			return nil, fmt.Errorf("loading exercise from info: %w", err)
		}
	}

	return e, nil
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
		logger.Error("getting page data",
			slog.String("url", url),
			slog.Int("year", year),
			slog.Int("day", day),
			tint.Err(err))

		return nil, fmt.Errorf("getting page data: %w", err)
	}

	title, err = extractTitle(page)
	if err != nil {
		logger.Error("extracting title from page data",
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
		path:     makeExercisePath(year, day, title),
	}

	return e, nil
}

func getExercisePath(year, day int) (string, bool) {
	// it may be better to use filepath.WalkDir here...
	dirEntries, err := os.ReadDir(strconv.Itoa(year))
	if err != nil {
		return "", false
	}

	for _, entry := range dirEntries {
		if entry.IsDir() && entry.Name()[:2] == fmt.Sprintf("%02d-", day) {
			fp, fpErr := filepath.Abs(filepath.Join(strconv.Itoa(year), entry.Name()))
			return fp, fpErr == nil
		}
	}

	return "", false
}

func extractTitle(page []byte) (string, error) {
	doc, _ := html.Parse(bytes.NewReader(page))

	tn, err := H2(doc)
	if err != nil {
		return "", fmt.Errorf("extracting title: %w", err)
	}

	tt := renderNode(tn)

	re := regexp.MustCompile(`--- Day \d{1,2}: (.*) ---`)

	matches := re.FindStringSubmatch(tt)
	if len(matches) != 2 { //nolint:gomnd // we expect 2 matches
		return "", fmt.Errorf("getting title from page data: no match")
	}

	return matches[1], nil
}

func (e *Exercise) PuzzlePage() ([]byte, error) {
	return getPage(e.Year, e.Day)
}

func getPage(year, day int) ([]byte, error) {
	d, err := getCachedPuzzlePage(year, day)
	if err == nil {
		logger.Debug("using cached puzzle page", slog.Int("year", year), slog.Int("day", day), slog.Int("size", len(d)))
		return d, nil
	}

	logger.Info("downloading puzzle page", slog.Int("year", year), slog.Int("day", day))

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
	fp := filepath.Join(cfgDir, "pages", makeExerciseID(year, day))

	f, err := afero.ReadFile(appFs, fp)
	if err != nil {
		logger.Warn("reading cached puzzle page", slog.String("file", fp), tint.Err(err))
		return nil, fmt.Errorf("reading puzzle page: %w", err)
	}

	return f, nil
}

func (e *Exercise) getCachedInput() ([]byte, error) {
	fp := filepath.Join(cfgDir, "inputs", e.ID)

	f, err := afero.ReadFile(appFs, fp)
	if err != nil {
		logger.Warn("read cached input", slog.String("file", fp), tint.Err(err))
		return nil, fmt.Errorf("read cached input: %w", err)
	}

	return f, nil
}

func downloadPuzzlePage(year, day int) ([]byte, error) {
	if cfgDir == "" {
		return nil, fmt.Errorf("cache directory not set")
	}

	// make sure we can write the cached file before we download it
	err := appFs.MkdirAll(filepath.Join(cfgDir, "pages"), 0o750)
	if err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	logger.Info("downloading puzzle page",
		slog.String("file", filepath.Join(cfgDir, "pages", makeExerciseID(year, day))))

	req := rClient.R().SetPathParams(map[string]string{
		"year": strconv.Itoa(year),
		"day":  strconv.Itoa(day),
	})

	resp, err := req.Get("/{year}/day/{day}")
	if err != nil {
		logger.Error("getting puzzle page", tint.Err(err))

		return nil, fmt.Errorf("requesting page: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		logger.Error("download page response",
			slog.String("url", resp.Request.URL),
			slog.String("status", http.StatusText(resp.StatusCode())),
			slog.Int("code", resp.StatusCode()))

		return nil, fmt.Errorf("processing page response: %s", http.StatusText(resp.StatusCode()))
	}

	// only keep relevant parts of the page
	re := regexp.MustCompile(`(?s)<article.*?>(.*)</article>`)
	matches := re.FindSubmatch(resp.Body())
	if len(matches) != 2 { //nolint:gomnd // we expect 2 matches
		logger.Error("extracting page data", slog.String("url", resp.Request.URL), slog.Any("found", matches))
		return nil, fmt.Errorf("extracting page data: no match")
	}

	pd := bytes.TrimSpace(matches[1])

	// write response to disk
	err = os.WriteFile(filepath.Join(cfgDir, "pages", makeExerciseID(year, day)), pd, 0o600)
	if err != nil {
		logger.Error("writing to cache", slog.String("url", resp.Request.URL), tint.Err(err))
		return nil, fmt.Errorf("writing cached puzzle page: %w", err)
	}

	return pd, nil
}

func (e *Exercise) downloadInput() ([]byte, error) {
	err := appFs.MkdirAll(filepath.Join(cfgDir, "inputs"), 0o750)
	if err != nil {
		return nil, fmt.Errorf("creating inputs directory: %w", err)
	}

	resp, err := rClient.R().
		SetOutput(filepath.Join("inputs", e.ID)).
		SetCookie(&http.Cookie{
			Name:   "session",
			Value:  os.Getenv("ELF_SESSION"),
			Domain: ".adventofcode.com",
		}).
		Get("/{year}/day/{day}/input")
	if err != nil {
		return nil, fmt.Errorf("accessing input data page: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		logger.Error("getting input data",
			slog.Group("request",
				slog.String("method", resp.Request.Method),
				slog.String("url", resp.Request.URL),
				slog.Any("cookies", resp.Request.Cookies)),
			slog.String("status", resp.Status()),
			slog.Int("code", resp.StatusCode()))

		return nil, fmt.Errorf("downloading input data: %s", resp.Status())
	}

	return bytes.TrimSpace(resp.Body()), nil
}

func (e *Exercise) getInput() ([]byte, error) {
	d, err := e.getCachedInput()
	if err == nil {
		return d, nil
	}

	return e.downloadInput()
}

//go:embed templates/info.tmpl
var infoTemplate []byte

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
	var (
		err             error
		replaceInfo     = true // TODO: use flag value
		replaceReadme   = true // TODO: use flag value
		replaceInput    = true // TODO: use flag value
		replaceLanguage = true // TODO: use flag value
	)

	addLogger := logger.With(slog.Any("exercise", e))

	if e.Language == "" || e.Dir() == "" {
		err = fmt.Errorf("incomplete exercise: missing language or directory")
		addLogger.Error("add files", tint.Err(err))

		return err
	}

	implPath := filepath.Join(e.Dir(), e.Language)

	if err = appFs.MkdirAll(implPath, 0o750); err != nil {
		addLogger.Error("add exercise implementation path", tint.Err(err))
		return fmt.Errorf("creating %s implementation directory: %w", e.Language, err)
	}

	if replaceInput { // TODO: should be if replacing OR if it doesn't exist
		// read cached data or download puzzle input
		inputFile, inErr := e.getInput()
		if inErr != nil {
			addLogger.Error("load input data", tint.Err(inErr))
			return fmt.Errorf("loading input: %w", inErr)
		}

		inErr = afero.WriteFile(appFs, filepath.Join(e.Dir(), "input.txt"), inputFile, 0o600)
		if inErr != nil {
			addLogger.Error("write input file", tint.Err(inErr))
			return fmt.Errorf("writing input file: %w", inErr)
		}
	}

	tmpls := []tmplFile{
		{
			Name:     "info",
			Path:     "",
			Data:     infoTemplate,
			FileName: "info.json",
			Replace:  replaceInfo,
		},
		{
			Name:     "readme",
			Path:     "",
			Data:     readmeTemplate,
			FileName: "README.md",
			Replace:  replaceReadme,
		},
	}

	if e.Language == "go" {
		tmpls = append(tmpls, tmplFile{
			Name:     "go",
			Path:     "go",
			Data:     goTemplate,
			FileName: "exercise.go",
			Replace:  replaceLanguage,
		})
	} else if e.Language == "py" {
		tmpls = append(tmpls, tmplFile{
			Name:     "py",
			Path:     "py",
			Data:     pyTemplate,
			FileName: "__init__.py",
			Replace:  replaceLanguage,
		})
	}

	for _, t := range tmpls {
		addLogger.LogAttrs(context.TODO(), slog.LevelDebug, "add template file", slog.Any("template", t.LogValue()))

		err = e.addTemplatedFile(appFs, t)
		if err != nil {
			addLogger.Error("adding template", slog.Any("template", t), tint.Err(err))
			return fmt.Errorf("adding %s template: %w", t.FileName, err)
		}
	}

	return nil
}

func (e *Exercise) addTemplatedFile(fs afero.Fs, tf tmplFile) error {
	// only write if file doesn't exist or if we're replacing it
	fp := filepath.Join(e.Dir(), tf.Path, tf.FileName)

	exists, _ := afero.Exists(fs, fp)
	if exists && !tf.Replace {
		logger.Info("file exists, skipping", "template", tf.LogValue())

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
