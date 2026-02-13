package advent

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

	if err = d.appFs.MkdirAll(implPath, dirPerm); err != nil {
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
		logger.Info("info file already exists, overwrite by using --force",
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

		return nil
	}

	t := template.Must(template.New(templateFile.Name).Parse(string(templateFile.Data)))
	b := new(bytes.Buffer)

	if err = t.Execute(b, d); err != nil {
		return fmt.Errorf("template %q: %w", templateFile.Name, err)
	}

	return afero.WriteFile(d.appFs, fp, b.Bytes(), 0o600)
}
