package exercise

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_extractProblemTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		page    string
		want    string
		wantErr bool
	}{
		{
			name: "single h2 title",
			page: `<html><body><h2>Coded Triangle Numbers</h2><p>body</p></body></html>`,
			want: "Coded Triangle Numbers",
		},
		{
			name: "title with surrounding whitespace is trimmed",
			page: "<html><body><h2>\n  Multiples of 3 or 5  \n</h2></body></html>",
			want: "Multiples of 3 or 5",
		},
		{
			name:    "no h2 means bad number",
			page:    `<html><body><p>no heading here</p></body></html>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := extractProblemTitle([]byte(tt.page))
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidData)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func newTestProblemFetcher(baseURL string) *problemFetcher {
	return &problemFetcher{
		rClient: resty.New().
			SetBaseURL(baseURL).
			SetHeader("User-Agent", userAgent).
			SetRedirectPolicy(resty.NoRedirectPolicy()),
	}
}

func Test_problemFetcher_fetchTitle(t *testing.T) {
	t.Parallel()

	t.Run("returns title on 200 with h2", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`<html><body><h2>Coded Triangle Numbers</h2></body></html>`))
		}))
		defer srv.Close()

		got, err := newTestProblemFetcher(srv.URL).fetchTitle(42)

		require.NoError(t, err)
		assert.Equal(t, "Coded Triangle Numbers", got)
	})

	t.Run("bad number (200, no h2) is ErrInvalidData", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`<html><body><p>archive</p></body></html>`))
		}))
		defer srv.Close()

		_, err := newTestProblemFetcher(srv.URL).fetchTitle(99999)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidData)
	})

	t.Run("redirect (bad number) is ErrInvalidData, not transient", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/archives", http.StatusFound)
		}))
		defer srv.Close()

		title, err := newTestProblemFetcher(srv.URL).fetchTitle(99999)

		require.ErrorIs(t, err, ErrInvalidData)
		require.NotErrorIs(t, err, ErrHTTPResponse) // must be distinguishable from transient
		assert.Empty(t, title)
	})

	t.Run("non-200 is a transient HTTP response error", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		_, err := newTestProblemFetcher(srv.URL).fetchTitle(42)

		require.ErrorIs(t, err, ErrHTTPResponse)
		assert.NotErrorIs(t, err, ErrInvalidData) // must be distinguishable from bad-number
	})
}
