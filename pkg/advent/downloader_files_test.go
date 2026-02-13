package advent

import (
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloader_writeInputFile(t *testing.T) {
	type fields struct {
		Exercise   *Exercise
		overwrites *Overwrites
	}

	tests := []struct {
		name      string
		fields    fields
		assertion require.ErrorAssertionFunc
	}{
		{
			name: "file exists",
			fields: fields{
				Exercise: &Exercise{
					ID:       "",
					Title:    "",
					Language: "",
					Year:     0,
					Day:      0,
					URL:      "",
					Data: &Data{
						InputData:     "",
						InputFileName: "fakeInput.txt",
						TestCases:     TestCase{},
						Answers:       Answer{},
					},
					Path:   "",
					runner: nil,
					appFs:  nil,
					logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				},
				overwrites: &Overwrites{
					Input: false,
				},
			},
			assertion: require.NoError,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			mockDownloader := new(Downloader)
			mockDownloader.Exercise = tt.fields.Exercise
			mockDownloader.exerciseBaseDir = "test_exercises"
			mockDownloader.cacheDir = "testCacheDir"
			mockDownloader.cfgDir = "testCfgDir"
			mockDownloader.rClient = nil
			mockDownloader.token = "fakeToken"
			mockDownloader.overwrites = tt.fields.overwrites
			mockDownloader.appFs = testFs

			tt.assertion(t, mockDownloader.writeInputFile())
		})
	}
}
