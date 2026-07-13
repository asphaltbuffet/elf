package secret

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"github.com/spf13/afero"
)

const ageExt = ".age"

// infoFileName is the exercise metadata file that carries the expected answers.
const infoFileName = "info.json"

// Encrypt seals an exercise's Solution Set — info.json plus every file under the
// given language subdirectories — to per-file <file>.age siblings, encrypted to
// recipients. Plaintext is never removed (the .age files are derived artifacts;
// see ADR-0019). An existing .age is Replaced; a new one is Added. Paths in the
// returned Report are relative to exerciseDir.
func Encrypt(afs afero.Fs, recipients []age.Recipient, exerciseDir string, langKeys []string) (Report, error) {
	report := Report{}

	// info.json is mandatory.
	infoPath := filepath.Join(exerciseDir, infoFileName)
	if exists, _ := afero.Exists(afs, infoPath); !exists {
		return nil, fmt.Errorf("%s not found in %q: not an exercise directory", infoFileName, exerciseDir)
	}
	entry, err := encryptFile(afs, recipients, exerciseDir, infoPath)
	if err != nil {
		return nil, err
	}
	report = append(report, entry)

	// Each configured language subdirectory, recursively.
	for _, key := range langKeys {
		langDir := filepath.Join(exerciseDir, key)
		if exists, _ := afero.DirExists(afs, langDir); !exists {
			continue
		}

		walkErr := afero.Walk(afs, langDir, func(path string, info fs.FileInfo, wErr error) error {
			if wErr != nil {
				return wErr
			}
			if info.IsDir() || strings.HasSuffix(path, ageExt) {
				return nil
			}
			e, fErr := encryptFile(afs, recipients, exerciseDir, path)
			if fErr != nil {
				return fErr
			}
			report = append(report, e)
			return nil
		})
		if walkErr != nil {
			return nil, fmt.Errorf("walking %q: %w", langDir, walkErr)
		}
	}

	return report, nil
}

// encryptFile writes <plainPath>.age and returns its Entry. Added if the .age
// did not previously exist, Replaced if it did. Plaintext is left in place.
func encryptFile(afs afero.Fs, recipients []age.Recipient, exerciseDir, plainPath string) (Entry, error) {
	agePath := plainPath + ageExt

	existed, _ := afero.Exists(afs, agePath)

	plaintext, err := afero.ReadFile(afs, plainPath)
	if err != nil {
		return Entry{}, fmt.Errorf("reading %q: %w", plainPath, err)
	}

	out, err := afs.Create(agePath)
	if err != nil {
		return Entry{}, fmt.Errorf("creating %q: %w", agePath, err)
	}
	defer out.Close()

	w, err := age.Encrypt(out, recipients...)
	if err != nil {
		return Entry{}, fmt.Errorf("initializing encryption for %q: %w", agePath, err)
	}
	if _, err = w.Write(plaintext); err != nil {
		return Entry{}, fmt.Errorf("encrypting %q: %w", agePath, err)
	}
	if err = w.Close(); err != nil {
		return Entry{}, fmt.Errorf("finalizing %q: %w", agePath, err)
	}

	rel, _ := filepath.Rel(exerciseDir, agePath)
	outcome := Added
	if existed {
		outcome = Replaced
	}
	return Entry{Path: rel, Outcome: outcome}, nil
}
