package exercise

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/asphaltbuffet/elf/pkg/utilities"
)

type Exercise struct {
	Year  int
	Day   int
	Title string
	Dir   string
	Path  string
}

func (e *Exercise) String() string {
	return fmt.Sprintf("%d - %s", e.Day, e.Title)
}

var exerciseDirRegexp = regexp.MustCompile(`(?m)^(\d{2})-([a-zA-Z-,'"]+)$`)

func ListingFromDir(sourceDir string) ([]*Exercise, error) {
	dirEntries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, err
	}

	var out []*Exercise

	for _, entry := range dirEntries {
		if entry.IsDir() && exerciseDirRegexp.MatchString(entry.Name()) {
			dir := entry.Name()

			left, right, _ := strings.Cut(dir, "-")
			dayInt, _ := strconv.Atoi(left) // error ignored because regex should have ensured this is ok
			dayTitle := utilities.CamelToTitle(right)
			out = append(out, &Exercise{
				Day:   dayInt,
				Title: dayTitle,
				Dir:   right,
				Path:  filepath.Join(sourceDir, dir),
			})
		}
	}

	return out, nil
}
