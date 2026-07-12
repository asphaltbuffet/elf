package secret_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/secret"
)

func TestOutcomeString(t *testing.T) {
	t.Parallel()

	cases := map[secret.Outcome]string{
		secret.Added:    "added",
		secret.Skipped:  "skipped",
		secret.Replaced: "replaced",
	}

	for outcome, want := range cases {
		got := outcome.String()
		require.Equal(t, want, got)
	}
}
