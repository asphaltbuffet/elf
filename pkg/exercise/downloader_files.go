package exercise

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"text/template"

	"github.com/lmittmann/tint"
	"github.com/spf13/afero"
)

//go:embed templates/readme.tmpl
var readmeTemplate []byte

//go:embed templates/go.tmpl
var goTemplate []byte

//go:embed templates/py.tmpl
var pyTemplate []byte

//go:embed templates/bash.tmpl
var bashTemplate []byte

//go:embed templates/rs-cargo.tmpl
var rsCargoTemplate []byte

//go:embed templates/rs-solution.tmpl
var rsSolutionTemplate []byte

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

// exerciseScaffold lays a finished Exercise out on disk: implementation directory, input file,
// info.json, and language template files. It reads the Exercise and never mutates it. It knows the
// exercise directory layout and nothing about HTTP or the cache.
type exerciseScaffold struct {
	fs            afero.Fs
	inputFileName string
	overwrites    *Overwrites
	logger        *slog.Logger
}

// write lays the given Exercise out on disk. The Exercise must already be fully assembled —
// scaffold invents no data.
func (s *exerciseScaffold) write(ex *Exercise) (Report, error) {
	logger := s.logger.With(slog.String("fn", "addMissingFiles"))

	implPath := filepath.Join(ex.Path, ex.Language)

	if err := s.fs.MkdirAll(implPath, dirPerm); err != nil {
		logger.Error("add exercise implementation path", tint.Err(err))
		return nil, fmt.Errorf("creating %s implementation directory: %w", ex.Language, err)
	}

	report := make(Report, 0, defaultReportCap)

	// TODO: give user option to overwrite existing files
	inputOutcome, err := s.writeInputFile(ex)
	if err != nil {
		return nil, fmt.Errorf("writing input file: %w", err)
	}
	report = append(report, Entry{Path: s.inputFileName, Outcome: inputOutcome})

	// TODO: give user option to overwrite existing files
	infoOutcome, err := s.writeInfoFile(ex, false)
	if err != nil {
		return nil, fmt.Errorf("writing info file: %w", err)
	}
	report = append(report, Entry{Path: "info.json", Outcome: infoOutcome})

	tmpls := []tmplFile{
		{
			Name:     "readme",
			Path:     "",
			Data:     readmeTemplate,
			FileName: "README.md",
			Replace:  false,
		},
	}

	switch ex.Language {
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

	case "bash":
		tmpls = append(tmpls, tmplFile{
			Name:     "bash",
			Path:     "bash",
			Data:     bashTemplate,
			FileName: "exercise.sh",
			Replace:  false,
		})

	case "rs":
		// A cargo crate per exercise: Cargo.toml at the crate root and the
		// solution module under src/. The harness (src/runtime-wrapper.rs) is
		// rendered by the Rust runner's PrepareSpec, not scaffolded here.
		tmpls = append(tmpls,
			tmplFile{
				Name:     "rs-cargo",
				Path:     "rs",
				Data:     rsCargoTemplate,
				FileName: "Cargo.toml",
				Replace:  false,
			},
			tmplFile{
				Name:     "rs-solution",
				Path:     filepath.Join("rs", "src"),
				Data:     rsSolutionTemplate,
				FileName: "solution.rs",
				Replace:  false,
			},
		)

	default:
		return nil, fmt.Errorf("template %s files: %w", ex.Language, ErrInvalidLanguage)
	}

	for _, t := range tmpls {
		logger.Debug("add template file", slog.Any("template", t.LogValue()))

		outcome, tErr := s.addTemplatedFile(ex, t)
		if tErr != nil {
			return nil, fmt.Errorf("adding %q template: %w", t.FileName, tErr)
		}
		report = append(report, Entry{Path: filepath.Join(t.Path, t.FileName), Outcome: outcome})
	}

	return report, nil
}

// writeInputFile writes the already-fetched input data from ex.Data to disk. It does not fetch —
// the Exercise arrives with its input populated by the assemble step.
func (s *exerciseScaffold) writeInputFile(ex *Exercise) (Outcome, error) {
	logger := s.logger.With(slog.String("fn", "writeInputFile"))

	fp := filepath.Join(ex.Path, s.inputFileName)

	// check if the file exists already
	exists, err := afero.Exists(s.fs, fp)
	if err != nil {
		return Skipped, err
	}

	if exists && !s.overwrites.Input {
		logger.Info("found %s, overwrite by using '--force-input'", slog.String("file", fp))
		return Skipped, nil
	}

	if err = afero.WriteFile(s.fs, fp, []byte(ex.Data.InputData), 0o600); err != nil {
		return Skipped, fmt.Errorf("writing input file: %w", err)
	}

	logger.Debug("wrote input file", slog.String("path", fp))

	if exists {
		return Replaced, nil
	}

	return Added, nil
}

func (s *exerciseScaffold) writeInfoFile(ex *Exercise, replace bool) (Outcome, error) {
	logger := s.logger.With(slog.String("fn", "writeInfoFile"))

	fp := filepath.Join(ex.Path, "info.json") // TODO: filename should be in config

	// check if the file exists already
	exists, err := afero.Exists(s.fs, fp)
	if err != nil {
		return Skipped, fmt.Errorf("checking for info file: %w", err)
	}

	if exists && !replace {
		logger.Info("info file already exists, overwrite by using --force",
			slog.String("file", fp))
		return Skipped, nil
	}

	// marshall exercise data
	data, err := json.MarshalIndent(ex, "", "  ")
	if err != nil {
		return Skipped, err
	}

	if err = afero.WriteFile(s.fs, fp, data, 0o600); err != nil {
		return Skipped, fmt.Errorf("write info file: %w", err)
	}

	logger.Debug("wrote info file", slog.String("path", fp))

	if exists {
		return Replaced, nil
	}

	return Added, nil
}

func (s *exerciseScaffold) addTemplatedFile(ex *Exercise, templateFile tmplFile) (Outcome, error) {
	fp := filepath.Join(ex.Path, templateFile.Path, templateFile.FileName)
	logger := s.logger.With(slog.String("fn", "addTemplatedFile"))

	// only write if file doesn't exist or if we're replacing it
	exists, err := afero.Exists(s.fs, fp)
	if err != nil {
		return Skipped, fmt.Errorf("checking for %q: %w", fp, err)
	}

	if exists && !templateFile.Replace {
		logger.Debug("file exists, skipping", "template", templateFile.LogValue())

		return Skipped, nil
	}

	t := template.Must(template.New(templateFile.Name).Parse(string(templateFile.Data)))
	b := new(bytes.Buffer)

	if err = t.Execute(b, ex); err != nil {
		return Skipped, fmt.Errorf("template %q: %w", templateFile.Name, err)
	}

	// Ensure the file's parent directory exists. Most templates live directly in
	// the language dir (already created), but some layouts nest deeper (e.g. the
	// Rust crate's src/), so create any intermediate dirs here.
	if err = s.fs.MkdirAll(filepath.Dir(fp), dirPerm); err != nil {
		return Skipped, fmt.Errorf("creating directory for %q: %w", fp, err)
	}

	if err = afero.WriteFile(s.fs, fp, b.Bytes(), 0o600); err != nil {
		return Skipped, err
	}

	if exists {
		return Replaced, nil
	}

	return Added, nil
}
