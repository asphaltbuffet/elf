package test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mocks "github.com/asphaltbuffet/elf/mocks/test"
	"github.com/asphaltbuffet/elf/pkg/krampus"
)

func TestGetTestCmd(t *testing.T) {
	t.Cleanup(func() { testCmd = nil })

	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, GetTestCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := GetTestCmd()
		assert.Equal(t, cmd, GetTestCmd())
	})
}

// resetState restores package-level variables and factory functions to defaults.
func resetState(
	t *testing.T,
	origMakeConfig func(string) (krampus.Config, error),
	origMakeChallengeTester func(krampus.ExerciseConfiguration, string, string) (ChallengeTester, error),
) {
	t.Helper()

	t.Cleanup(func() {
		testCmd = nil
		language = ""
		makeConfig = origMakeConfig
		makeChallengeTester = origMakeChallengeTester
	})
}

func Test_runTestCmd(t *testing.T) {
	origMakeConfig := makeConfig
	origMakeChallengeTester := makeChallengeTester

	t.Run("config error", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeChallengeTester)

		makeConfig = func(_ string) (krampus.Config, error) {
			return krampus.Config{}, errors.New("bad config")
		}

		cmd := GetTestCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runTestCmd(cmd, []string{"."})
		assert.ErrorContains(t, err, "bad config")
	})

	t.Run("challenge tester creation error", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeChallengeTester)

		makeConfig = func(_ string) (krampus.Config, error) {
			return krampus.NewConfig()
		}
		makeChallengeTester = func(_ krampus.ExerciseConfiguration, _, _ string) (ChallengeTester, error) {
			return nil, errors.New("exercise not found")
		}

		cmd := GetTestCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runTestCmd(cmd, []string{"."})
		assert.ErrorContains(t, err, "exercise not found")
	})

	t.Run("test error prints to stdout", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeChallengeTester)

		makeConfig = func(_ string) (krampus.Config, error) {
			return krampus.NewConfig()
		}

		mockCh := mocks.NewMockChallengeTester(t)
		mockCh.EXPECT().Test().Return(nil, errors.New("test failed"))

		makeChallengeTester = func(_ krampus.ExerciseConfiguration, _, _ string) (ChallengeTester, error) {
			return mockCh, nil
		}

		cmd := GetTestCmd()
		outBuf := &bytes.Buffer{}
		cmd.SetOut(outBuf)
		cmd.SetErr(&bytes.Buffer{})

		err := runTestCmd(cmd, []string{"."})
		require.NoError(t, err, "runTestCmd should return nil even when Test fails")
		assert.Contains(t, outBuf.String(), "Failed to run tests")
		assert.Contains(t, outBuf.String(), "test failed")
	})

	t.Run("happy path", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeChallengeTester)

		makeConfig = func(_ string) (krampus.Config, error) {
			return krampus.NewConfig()
		}

		mockCh := mocks.NewMockChallengeTester(t)
		mockCh.EXPECT().Test().Return(nil, nil)

		makeChallengeTester = func(_ krampus.ExerciseConfiguration, _, _ string) (ChallengeTester, error) {
			return mockCh, nil
		}

		cmd := GetTestCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runTestCmd(cmd, []string{"."})
		assert.NoError(t, err)
	})

	t.Run("language flag used over config default", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeChallengeTester)

		makeConfig = func(_ string) (krampus.Config, error) {
			return krampus.NewConfig()
		}

		var gotLang string
		mockCh := mocks.NewMockChallengeTester(t)
		mockCh.EXPECT().Test().Return(nil, nil)

		makeChallengeTester = func(_ krampus.ExerciseConfiguration, lang, _ string) (ChallengeTester, error) {
			gotLang = lang
			return mockCh, nil
		}

		cmd := GetTestCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		// Set language AFTER GetTestCmd() — flag creation resets the variable.
		language = "py"

		err := runTestCmd(cmd, []string{"."})
		require.NoError(t, err)
		assert.Equal(t, "py", gotLang)
	})
}
