package analyze

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/asphaltbuffet/elf/pkg/advent"
	"github.com/asphaltbuffet/elf/pkg/config"
)

type Analyzer struct {
	Data   []*advent.BenchmarkData
	Dir    string
	Output string

	logger *slog.Logger
}

func NewAnalyzer(cfg config.ExerciseConfiguration, opts ...func(*Analyzer)) (*Analyzer, error) {
	analyzer := &Analyzer{
		logger: cfg.GetLogger(),
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

func WithDirectory(dir string) func(*Analyzer) {
	return func(a *Analyzer) {
		a.Dir = dir
	}
}

func WithOutput(name string) func(*Analyzer) {
	return func(a *Analyzer) {
		a.Output = name
	}
}

func (a *Analyzer) Load() error {
	files, err := getBenchmarkFiles(a.Dir)
	if err != nil {
		return fmt.Errorf("getting benchmark files: %w", err)
	}

	// load benchmark data from files
	a.logger.Debug("found benchmark files", "count", len(files))
	benchData := make([]*advent.BenchmarkData, 0, len(files))

	for _, bf := range files {
		var data []*advent.BenchmarkData

		data, err = readBenchmarkFile(bf)
		if err != nil {
			return fmt.Errorf("reading %s: %w", bf, err)
		}

		benchData = append(benchData, data...)
	}

	a.Data = benchData

	return nil
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

func readBenchmarkFile(path string) ([]*advent.BenchmarkData, error) {
	var bd []*advent.BenchmarkData

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
