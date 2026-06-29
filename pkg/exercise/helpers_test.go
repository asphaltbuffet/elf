package exercise

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

var (
	roBase    afero.Fs
	testFs    afero.Fs
	testAdder *Adder
)

// FileExists checks whether a file exists in the given path. It also fails if
// the path points to a directory or there is an error when trying to check the file.
func FileExists(t *testing.T, afs afero.Fs, path string, msgAndArgs ...any) bool {
	t.Helper()

	info, err := afs.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return assert.Fail(t, fmt.Sprintf("unable to find file %q", path), msgAndArgs...)
		}
		return assert.Fail(t, fmt.Sprintf("error when running Fs.Stat(%q): %s", path, err), msgAndArgs...)
	}

	if info.IsDir() {
		return assert.Fail(t, fmt.Sprintf("%q is a directory", path), msgAndArgs...)
	}

	return true
}

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Helper()

	base := afero.NewBasePathFs(afero.NewOsFs(), "testdata")
	roBase = afero.NewReadOnlyFs(base)

	// Seed the runner registry with stub entries so registry-based checks
	// ("is this language supported?") pass without the real runner binaries.
	restore := runners.ResetRegistry(map[string]runners.RunnerCreator{
		"go": func(_ runners.ExerciseMeta) runners.Runner { return nil },
		"py": func(_ runners.ExerciseMeta) runners.Runner { return nil },
	})

	return func(t *testing.T) {
		t.Helper()

		httpmock.DeactivateAndReset()
		restore()
	}
}

func setupSubTest(t *testing.T) func(t *testing.T) {
	t.Helper()

	testFs = afero.NewCopyOnWriteFs(roBase, afero.NewMemMapFs())
	require.NoError(t, testFs.MkdirAll("testCache", 0o755))

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	fetcher, err := newPageFetcher("fakeToken", "testCache", testFs, logger)
	require.NoError(t, err)

	testAdder = &Adder{
		cfgDir:          "./",
		exerciseBaseDir: "exercises",
		appFs:           testFs,
		logger:          logger,
		fetcher:         fetcher,
		scaffold: &exerciseScaffold{
			fs:            testFs,
			inputFileName: "input.txt",
			overwrites:    &Overwrites{},
			logger:        logger,
		},
	}

	httpmock.ActivateNonDefault(testAdder.fetcher.rClient.GetClient())

	httpmock.Reset()

	return func(t *testing.T) {
		t.Helper()

		// t.Log("teardown sub-test")
	}
}

func goldenValue(t *testing.T, goldenFile string) []byte {
	t.Helper()

	content, err := os.ReadFile(goldenFile)
	require.NoError(t, err)

	return content
}
