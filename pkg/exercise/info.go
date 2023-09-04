package exercise

import (
	"encoding/json"
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
	data, err := afero.ReadFile(fs, path.Clean(fname))
	if err != nil {
		return nil, err
	}

	c := new(Info)

	err = json.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
