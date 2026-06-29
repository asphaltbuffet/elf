package runners_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/cmd/runners"
)

func TestRunListCmd_NoRunners(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a minimal elf.toml with no runners
	cfgFile := filepath.Join(tmpDir, "elf.toml")
	require.NoError(t, os.WriteFile(cfgFile, []byte(`language = "go"`+"\n"), 0o644))

	runners.ResetForTest()
	cmd := runners.GetRunnersCmd()
	var out bytes.Buffer

	listCmd, _, findErr := cmd.Find([]string{"list"})
	require.NoError(t, findErr)
	listCmd.SetOut(&out)
	require.NoError(t, listCmd.Flags().Set("config-file", cfgFile))

	err := listCmd.RunE(listCmd, nil)
	require.NoError(t, err)

	assert.Contains(t, out.String(), "No runners configured")
}

func TestRunListCmd_WithRunners(t *testing.T) {
	tmpDir := t.TempDir()

	tmplPath := filepath.Join(tmpDir, "python.tmpl")
	require.NoError(t, os.WriteFile(tmplPath, []byte("# template"), 0o644))

	cfgContent := "language = \"go\"\n\n[[runner]]\nkey = \"py\"\nname = \"Python\"\n\n[runner.prepare]\ntemplate_path = \"" + tmplPath + "\"\n\n[runner.open]\ninterpreter = \"python3\"\nargs = [\"-B\", \"{wrapper_file}\"]\n"

	cfgFile := filepath.Join(tmpDir, "elf.toml")
	require.NoError(t, os.WriteFile(cfgFile, []byte(cfgContent), 0o644))

	runners.ResetForTest()
	cmd := runners.GetRunnersCmd()
	var out bytes.Buffer

	listCmd, _, findErr := cmd.Find([]string{"list"})
	require.NoError(t, findErr)
	listCmd.SetOut(&out)

	require.NoError(t, listCmd.Flags().Set("config-file", cfgFile))

	runErr := listCmd.RunE(listCmd, nil)
	require.NoError(t, runErr)

	output := out.String()
	assert.Contains(t, output, "py")
	assert.Contains(t, output, "Python")
	assert.Contains(t, output, "ok")
}
