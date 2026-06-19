// Package analyze loads benchmark data and renders run-time graphs.
package analyze

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/asphaltbuffet/elf/pkg/exercise"
)

// Analyzer loads benchmark data from a directory and renders a run-time graph.
type Analyzer struct {
	Data   []*exercise.BenchmarkData
	Dir    string
	Output string
	Scope  Scope

	logger *slog.Logger
}

// NewAnalyzer creates an Analyzer, applies options, and loads benchmark data from the configured directory.
func NewAnalyzer(logger *slog.Logger, opts ...func(*Analyzer)) (*Analyzer, error) {
	analyzer := &Analyzer{
		logger: logger,
	}

	for _, opt := range opts {
		opt(analyzer)
	}

	if analyzer.Dir == "" {
		return nil, errors.New("no directory specified")
	}

	err := analyzer.Load()
	if err != nil {
		return nil, fmt.Errorf("loading benchmark data: %w", err)
	}

	return analyzer, nil
}

// WithDirectory sets the directory from which benchmark data files are loaded.
func WithDirectory(dir string) func(*Analyzer) {
	return func(a *Analyzer) {
		a.Dir = dir
	}
}

// WithOutput sets the output file path for the generated graph.
func WithOutput(name string) func(*Analyzer) {
	return func(a *Analyzer) {
		a.Output = name
	}
}

// Load reads all benchmark JSON files from the configured directory into Data.
func (a *Analyzer) Load() error {
	scope, err := detectScope(a.Dir)
	if err != nil {
		return err
	}
	a.Scope = scope

	if a.Output == "" {
		a.Output = filepath.Join(a.Dir, "run-times.png")
	}

	files, err := getBenchmarkFiles(a.Dir)
	if err != nil {
		return fmt.Errorf("getting benchmark files: %w", err)
	}

	a.logger.Debug("found benchmark files", "count", len(files))
	benchData := make([]*exercise.BenchmarkData, 0, len(files))

	for _, bf := range files {
		var data []*exercise.BenchmarkData

		data, err = readBenchmarkFile(bf)
		if err != nil {
			return fmt.Errorf("reading %s: %w", bf, err)
		}

		benchData = append(benchData, data...)
	}

	if len(benchData) == 0 {
		return fmt.Errorf("no benchmark data found in %s; run `elf benchmark %s` first", a.Dir, a.Dir)
	}

	a.Data = benchData

	return nil
}

// Graph renders the appropriate graph for the resolved scope.
func (a *Analyzer) Graph() error {
	switch a.Scope {
	case ScopeExercise:
		return generateBoxPlot(a.Data, a.Output)
	case ScopeYear:
		return generateLineGraph(a.Data, a.Output)
	default:
		return fmt.Errorf("unknown scope: %d", a.Scope)
	}
}

func getBenchmarkFiles(dir string) ([]string, error) { //nolint:unparam // expected behavior when walking directories
	var benchFiles []string

	// get all benchmark.json files recursively
	_ = filepath.WalkDir(dir, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // expected behavior when walking directories
		}

		if filepath.Base(path) == "benchmark.json" {
			benchFiles = append(benchFiles, path)
		}

		return nil
	})

	return benchFiles, nil
}

func readBenchmarkFile(path string) ([]*exercise.BenchmarkData, error) {
	var bd []*exercise.BenchmarkData

	f, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	err = json.Unmarshal(f, &bd)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling json: %w", err)
	}

	return bd, nil
}
