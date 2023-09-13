package aoc

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/exercise"
	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/utilities"
)

const adventBaseURL = "https://adventofcode.com"

var (
	client *AOCClient
	appFs  afero.Fs = afero.NewOsFs()
	cfgDir string

	exercises map[int]map[int]*exercise.Exercise
	info      map[int]map[int]*exercise.Info

	baseExercisesDir = "exercises"
	adventPuzzleURL  = "%d/day/%d"
	adventInputURL   = "%d/day/%d/input"
)

type RunMode int

const (
	RunModeAll RunMode = iota
	RunModeTestOnly
	RunModeNoTest
)

type AOCClient struct {
	ExercisesDir string
	Runners      map[string]runners.RunnerCreator
	Years        []int
	Days         map[int]([]int)
	RunMode      RunMode
}

// NewAOCClient returns a new AOCClient.
func NewAOCClient() (*AOCClient, error) {
	c, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("getting user config dir: %w", err)
	}

	rClient = resty.New().SetBaseURL(adventBaseURL)

	cfgDir = filepath.Join(c, "elf")

	exercises = make(map[int]map[int]*exercise.Exercise)
	info = make(map[int]map[int]*exercise.Info)

	if err := discoverExercises(); err != nil {
		return nil, fmt.Errorf("searching for exercises: %w", err)
	}

	days := make(map[int]([]int), len(exercises))
	years := make([]int, 0, len(exercises))

	for year, dayMap := range exercises {
		years = append(years, year)

		for day := range dayMap {
			days[year] = append(days[year], day)
		}
	}

	slices.Sort(years)

	for _, d := range days {
		slices.Sort(d)
	}

	client = &AOCClient{
		ExercisesDir: baseExercisesDir,
		Years:        years,
		Days:         days,
		Runners:      runners.Available,
	}

	return client, nil
}

func discoverExercises() error {
	files := []string{}
	re := regexp.MustCompile(`^(\d{2})-\w+`)

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// if a directory, check if the name is in the format "01-someText"
		if info.IsDir() && re.MatchString(info.Name()) {
			files = append(files, path)
		}

		return nil
	}

	err := afero.Walk(appFs, baseExercisesDir, walkFn)
	if err != nil {
		return err
	}

	for _, p := range files {
		year, day, name, err := parseExercisePath(p)
		if err != nil {
			return fmt.Errorf("parsing exercise path: %w", err)
		}

		if err := registerExercise(year, day, p, name); err != nil {
			return err
		}

		if err := registerInfo(year, day, p); err != nil {
			return err
		}
	}

	return nil
}

func GetClient() (*AOCClient, error) {
	if client != nil {
		return client, nil
	}

	return NewAOCClient()
}

func parseExercisePath(p string) (int, int, string, error) {
	// path is in the format "exercises/2015/01-someText"
	base, exerciseDir := filepath.Split(p)
	_, yearDir := filepath.Split(filepath.Clean(base))

	year, err := strconv.Atoi(yearDir)
	if err != nil {
		return 0, 0, "", fmt.Errorf("%q: converting year to int: %w", base, err)
	}

	d, name, _ := strings.Cut(exerciseDir, "-")

	day, err := strconv.Atoi(d)
	if err != nil {
		return 0, 0, "", fmt.Errorf("%q: converting day to int: %w", exerciseDir, err)
	}

	return year, day, name, nil
}

func registerExercise(year int, day int, path string, name string) error {
	if _, ok := exercises[year][day]; ok {
		return fmt.Errorf("duplicate exercise: year=%d day=%d", year, day)
	}

	if _, ok := exercises[year]; !ok {
		exercises[year] = make(map[int]*exercise.Exercise)
	}

	exercises[year][day] = &exercise.Exercise{
		Year:  year,
		Day:   day,
		Title: utilities.CamelToTitle(name),
		Dir:   fmt.Sprintf("%02d-%s", day, name),
		Path:  path,
	}

	return nil
}

func registerInfo(year int, day int, path string) error {
	if _, ok := info[year]; !ok {
		info[year] = make(map[int]*exercise.Info)
	}

	if _, ok := info[year][day]; ok {
		return fmt.Errorf("duplicate info: year=%d day=%d", year, day)
	}

	i, err := exercise.LoadExerciseInfo(appFs, filepath.Join(path, "info.json"))
	if err != nil {
		// skipping instead of error for now, not all exercises may have info.json files
		// return fmt.Errorf("loading exercise info: %w", err)
		// NOTE: we may want to delete the exercise from the map here
		return nil
	}

	info[year][day] = i

	return nil
}

func (ac *AOCClient) GetExercise(year int, day int) (*exercise.Exercise, error) {
	if e, ok := exercises[year][day]; ok {
		return e, nil
	}

	return nil, fmt.Errorf("no such exercise: year=%d day=%d", year, day)
}

func (ac *AOCClient) GetExerciseInfo(year int, day int) (*exercise.Info, error) {
	if i, ok := info[year][day]; ok {
		return i, nil
	}

	return nil, fmt.Errorf("no such info: year=%d day=%d", year, day)
}

func (ac *AOCClient) GetInput(year int, day int) (string, error) {
	e := exercises[year][day]
	i := info[year][day]

	// Load exercise input
	input, err := os.ReadFile(filepath.Join(e.Path, i.InputFile))
	if err != nil {
		return "", fmt.Errorf("reading input file: %w", err)
	}

	return string(input), nil
}

func (ac *AOCClient) YearDirs() ([]string, error) {
	years := make([]string, 0, len(ac.Years))

	for _, year := range ac.Years {
		yearPath := filepath.Join(ac.ExercisesDir, strconv.Itoa(year))
		years = append(years, yearPath)
	}

	slices.Sort(years)

	return years, nil
}

func (ac *AOCClient) DayDirs(year int) ([]string, error) {
	if err := isValidYear(year); err != nil {
		return nil, fmt.Errorf("getting %d path: %w", year, err)
	}

	dirs := make([]string, 0, len(exercises[year]))
	for _, e := range exercises[year] {
		dirs = append(dirs, e.Path)
	}

	// NOTE: should we return error if no dirs found?

	slices.Sort(dirs)

	return dirs, nil
}

func (ac *AOCClient) ImplementationDirs(year int, day int) (map[string]string, error) {
	if err := isValidExercise(year, day); err != nil {
		return nil, fmt.Errorf("getting %d-%d exercise: %w", year, day, err)
	}

	base := exercises[year][day].Path
	dirs := make(map[string]string, len(ac.Runners))

	// check base + runner(key) for directories and add if found
	for r := range ac.Runners {
		path := filepath.Join(base, r)
		if _, err := appFs.Stat(path); err == nil {
			dirs[r] = path
		}
	}

	if len(dirs) == 0 {
		return nil, fmt.Errorf("no implementations found in %s", base)
	}

	return dirs, nil
}

func isValidYear(year int) error {
	if _, ok := exercises[year]; !ok {
		return fmt.Errorf("year not found: %d", year)
	}

	return nil
}

func isValidExercise(year int, day int) error {
	if err := isValidYear(year); err != nil {
		return err
	}

	if _, ok := exercises[year][day]; !ok {
		return fmt.Errorf("day not found: %d", day)
	}

	return nil
}

func (ac *AOCClient) MissingDays() map[int]([]int) {
	missing := make(map[int]([]int), getMaxYear()-2015)

	for y := 2015; y <= getMaxYear(); y++ {
		missing[y] = []int{}

		for d := 1; d <= 25; d++ {
			if _, ok := exercises[y][d]; !ok {
				missing[y] = append(missing[y], d)
			}
		}
	}

	return missing
}

func getMaxYear() int {
	maxYear := time.Now().Year()

	if time.Now().Month() != time.December {
		maxYear--
	}

	return maxYear
}

func (ac *AOCClient) MissingImplementations(year, day int) []string {
	i, _ := ac.ImplementationDirs(year, day)
	if len(i) == len(ac.Runners) {
		return []string{}
	}

	missing := make([]string, 0, len(ac.Runners))

	for r := range ac.Runners {
		if _, ok := i[r]; !ok {
			missing = append(missing, r)
		}
	}

	return missing
}

func (ac *AOCClient) GetExerciseInput(year int, day int) (string, error) {
	return "", fmt.Errorf("not implemented")
}
