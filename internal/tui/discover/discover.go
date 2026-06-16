// Package discover scans the exercise directory tree and groups exercises by year.
package discover

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

// ExerciseInfo holds metadata about a discovered exercise directory.
type ExerciseInfo struct {
	Year  int
	Day   int
	Title string
	Path  string
	Langs []string
	HasP1 bool
	HasP2 bool
}

// infoFile is a minimal representation of info.json for scanning purposes.
type infoFile struct {
	Year  int      `json:"year"`
	Day   int      `json:"day"`
	Title string   `json:"title"`
	Data  infoData `json:"data"`
}

type infoData struct {
	Answers infoAnswers `json:"answers"`
}

type infoAnswers struct {
	One string `json:"a"`
	Two string `json:"b"`
}

// Scan walks root looking for exercise directories containing info.json files.
// It returns exercises grouped by year, sorted by day within each year.
func Scan(fs afero.Fs, root string) (map[int][]ExerciseInfo, error) {
	result := make(map[int][]ExerciseInfo)

	err := afero.Walk(fs, root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || info.Name() != "info.json" {
			return nil
		}

		data, readErr := afero.ReadFile(fs, path)
		if readErr != nil {
			return readErr
		}

		var inf infoFile

		_ = json.Unmarshal(data, &inf)

		if inf.Year == 0 || inf.Day == 0 {
			return nil // skip malformed or incomplete entries
		}

		exerciseDir := filepath.Dir(path)

		langs := detectLanguages(fs, exerciseDir)

		ei := ExerciseInfo{
			Year:  inf.Year,
			Day:   inf.Day,
			Title: inf.Title,
			Path:  exerciseDir,
			Langs: langs,
			HasP1: inf.Data.Answers.One != "",
			HasP2: inf.Data.Answers.Two != "",
		}

		result[inf.Year] = append(result[inf.Year], ei)

		return nil
	})
	if err != nil {
		return nil, err
	}

	// sort exercises by day within each year
	for year := range result {
		sort.Slice(result[year], func(i, j int) bool {
			return result[year][i].Day < result[year][j].Day
		})
	}

	return result, nil
}

// detectLanguages checks for subdirectories matching known runner names.
func detectLanguages(fs afero.Fs, dir string) []string {
	entries, err := afero.ReadDir(fs, dir)
	if err != nil {
		return nil
	}

	var langs []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := strings.ToLower(entry.Name())

		if _, ok := runners.Available[name]; ok {
			langs = append(langs, name)
		}
	}

	sort.Strings(langs)

	return langs
}
