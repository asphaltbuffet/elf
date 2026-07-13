package secret

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"github.com/spf13/afero"
)

// Decrypt restores plaintext from every <file>.age under exerciseDir using
// identity. The .age files are never removed. Existing plaintext is Skipped
// (Added when absent); force overwrites it as Replaced. Paths in the returned
// Report are the plaintext paths relative to exerciseDir.
func Decrypt(afs afero.Fs, identity age.Identity, exerciseDir string, force bool) (Report, error) {
	report := Report{}

	walkErr := afero.Walk(afs, exerciseDir, func(path string, info fs.FileInfo, wErr error) error {
		if wErr != nil {
			return wErr
		}
		if info.IsDir() || !strings.HasSuffix(path, ageExt) {
			return nil
		}

		entry, dErr := decryptFile(afs, identity, exerciseDir, path, force)
		if dErr != nil {
			return dErr
		}
		report = append(report, entry)
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walking %q: %w", exerciseDir, walkErr)
	}

	return report, nil
}

// decryptFile writes the plaintext for one .age file. Existing plaintext is
// Skipped unless force (then Replaced); otherwise Added.
func decryptFile(afs afero.Fs, identity age.Identity, exerciseDir, agePath string, force bool) (Entry, error) {
	plainPath := strings.TrimSuffix(agePath, ageExt)
	rel, _ := filepath.Rel(exerciseDir, plainPath)

	existed, _ := afero.Exists(afs, plainPath)
	if existed && !force {
		return Entry{Path: rel, Outcome: Skipped}, nil
	}

	in, err := afs.Open(agePath)
	if err != nil {
		return Entry{}, fmt.Errorf("opening %q: %w", agePath, err)
	}
	defer in.Close()

	r, err := age.Decrypt(in, identity)
	if err != nil {
		return Entry{}, fmt.Errorf("decrypting %q (no matching identity?): %w", agePath, err)
	}

	plaintext, err := io.ReadAll(r)
	if err != nil {
		return Entry{}, fmt.Errorf("reading decrypted %q: %w", agePath, err)
	}

	if err = afero.WriteFile(afs, plainPath, plaintext, 0o644); err != nil {
		return Entry{}, fmt.Errorf("writing %q: %w", plainPath, err)
	}

	outcome := Added
	if existed {
		outcome = Replaced
	}
	return Entry{Path: rel, Outcome: outcome}, nil
}
