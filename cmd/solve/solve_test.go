package solve

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mocks "github.com/asphaltbuffet/elf/mocks/solve"
	"github.com/asphaltbuffet/elf/pkg/krampus"
)

func TestGetSolveCmd(t *testing.T) {
	t.Cleanup(func() { solveCmd = nil })

	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, GetSolveCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := GetSolveCmd()
		assert.Equal(t, cmd, GetSolveCmd())
	})
}

// resetState restores package-level variables and factory functions to defaults.
func resetState(
	t *testing.T,
	origMakeConfig func(string) (krampus.Config, error),
	origMakeChallenge func(krampus.ExerciseConfiguration, string, string, string) (Challenge, error),
) {
	t.Helper()

	t.Cleanup(func() {
		solveCmd = nil
		language = ""
		input = ""
		noTest = false
		makeConfig = origMakeConfig
		makeChallenge = origMakeChallenge
	})
}

func Test_runSolveCmd(t *testing.T) {
	origMakeConfig := makeConfig
	origMakeChallenge := makeChallenge

	t.Run("config error", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeChallenge)

		makeConfig = func(_ string) (krampus.Config, error) {
			return krampus.Config{}, errors.New("bad config")
		}

		cmd := GetSolveCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runSolveCmd(cmd, []string{"."})
		assert.ErrorContains(t, err, "bad config")
	})

	t.Run("challenge creation error", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeChallenge)

		makeConfig = func(_ string) (krampus.Config, error) {
			return krampus.NewConfig()
		}
		makeChallenge = func(_ krampus.ExerciseConfiguration, _, _, _ string) (Challenge, error) {
			return nil, errors.New("exercise not found")
		}

		cmd := GetSolveCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runSolveCmd(cmd, []string{"."})
		assert.ErrorContains(t, err, "exercise not found")
	})

	t.Run("solve error prints to stderr", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeChallenge)

		makeConfig = func(_ string) (krampus.Config, error) {
			return krampus.NewConfig()
		}

		mockCh := mocks.NewMockChallenge(t)
		mockCh.EXPECT().Solve(false).Return(nil, errors.New("solve failed"))

		makeChallenge = func(_ krampus.ExerciseConfiguration, _, _, _ string) (Challenge, error) {
			return mockCh, nil
		}

		cmd := GetSolveCmd()
		cmd.SetOut(&bytes.Buffer{})
		errBuf := &bytes.Buffer{}
		cmd.SetErr(errBuf)

		err := runSolveCmd(cmd, []string{"."})
		require.NoError(t, err, "runSolveCmd should return nil even when Solve fails")
		assert.Contains(t, errBuf.String(), "Failed to solve")
		assert.Contains(t, errBuf.String(), "solve failed")
	})

	t.Run("happy path", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeChallenge)

		makeConfig = func(_ string) (krampus.Config, error) {
			return krampus.NewConfig()
		}

		mockCh := mocks.NewMockChallenge(t)
		mockCh.EXPECT().Solve(false).Return(nil, nil)

		makeChallenge = func(_ krampus.ExerciseConfiguration, _, _, _ string) (Challenge, error) {
			return mockCh, nil
		}

		cmd := GetSolveCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runSolveCmd(cmd, []string{"."})
		assert.NoError(t, err)
	})

	t.Run("language flag used over config default", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeChallenge)

		makeConfig = func(_ string) (krampus.Config, error) {
			return krampus.NewConfig()
		}

		var gotLang string
		mockCh := mocks.NewMockChallenge(t)
		mockCh.EXPECT().Solve(false).Return(nil, nil)

		makeChallenge = func(_ krampus.ExerciseConfiguration, lang, _, _ string) (Challenge, error) {
			gotLang = lang
			return mockCh, nil
		}

		cmd := GetSolveCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		// Set language AFTER GetSolveCmd() — flag creation resets the variable.
		language = "py"

		err := runSolveCmd(cmd, []string{"."})
		require.NoError(t, err)
		assert.Equal(t, "py", gotLang)
	})

	t.Run("no-test flag passes true to Solve", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeChallenge)

		makeConfig = func(_ string) (krampus.Config, error) {
			return krampus.NewConfig()
		}

		mockCh := mocks.NewMockChallenge(t)
		mockCh.EXPECT().Solve(true).Return(nil, nil)

		makeChallenge = func(_ krampus.ExerciseConfiguration, _, _, _ string) (Challenge, error) {
			return mockCh, nil
		}

		cmd := GetSolveCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		// Set noTest AFTER GetSolveCmd() — flag creation resets the variable.
		noTest = true

		err := runSolveCmd(cmd, []string{"."})
		assert.NoError(t, err)
	})

	t.Run("input flag used when set", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeChallenge)

		makeConfig = func(_ string) (krampus.Config, error) {
			return krampus.NewConfig()
		}

		var gotInput string
		mockCh := mocks.NewMockChallenge(t)
		mockCh.EXPECT().Solve(false).Return(nil, nil)

		makeChallenge = func(_ krampus.ExerciseConfiguration, _, _, inputFile string) (Challenge, error) {
			gotInput = inputFile
			return mockCh, nil
		}

		cmd := GetSolveCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		// Set input AFTER GetSolveCmd() — flag creation resets the variable.
		input = "custom-input.txt"

		err := runSolveCmd(cmd, []string{"."})
		require.NoError(t, err)
		assert.Equal(t, "custom-input.txt", gotInput)
	})
}
