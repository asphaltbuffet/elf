package advent

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	roBase  afero.Fs
	testFs  afero.Fs
	mockDlr *Downloader
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

	return func(t *testing.T) {
		t.Helper()

		httpmock.DeactivateAndReset()
	}
}

func setupSubTest(t *testing.T) func(t *testing.T) {
	t.Helper()

	testFs = afero.NewCopyOnWriteFs(roBase, afero.NewMemMapFs())
	require.NoError(t, testFs.MkdirAll("testCache", 0o755))

	mockDlr = &Downloader{
		Exercise: &Exercise{
			ID:       "",
			Title:    "",
			Language: "",
			Year:     0,
			Day:      0,
			URL:      "",
			Data:     &Data{},
			Path:     "",
			runner:   nil,
			appFs:    testFs,
			logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
		cacheDir:        "testCache",
		cfgDir:          "./",
		exerciseBaseDir: "exercises",
		rClient:         resty.New().SetBaseURL("https://test.fake"),
		token:           "fakeToken",
	}

	httpmock.ActivateNonDefault(mockDlr.rClient.GetClient())

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
