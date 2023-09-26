package exercise

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/spf13/afero"
)

// Info contains the relative path to exercise input and the specific test case data for an exercise.
type Info struct {
	InputFile string   `json:"inputFile"`
	TestCases TestCase `json:"testCases"`
}

// TestCase contains the test case for each part of an exercise.
type TestCase struct {
	One []*Test `json:"one"`
	Two []*Test `json:"two"`
}

// Test contains the input and expected output for a test case.
type Test struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

// LoadExerciseInfo loads the input and test cases for an exercise from the given json file.
func LoadExerciseInfo(fs afero.Fs, fname string) (*Info, error) {
	infoFile := path.Clean(fname)

	data, err := afero.ReadFile(fs, infoFile)
	if err != nil {
		return nil, fmt.Errorf("read info file %q: %w", infoFile, err)
	}

	c := &Info{}

	err = json.Unmarshal(data, c)
	if err != nil {
		return nil, fmt.Errorf("unmarshal info file %s: %w", infoFile, err)
	}

	return c, nil
}
