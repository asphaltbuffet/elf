package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRunners_ParsesDescriptors(t *testing.T) {
	tomlContent := `
[[runner]]
key = "py"
name = "Python"

[runner.prepare]
template_path = "/home/user/.config/elf/runners/python.tmpl"

[runner.open]
interpreter = "python3"
args = ["-B", "{wrapper_file}"]
env = ["PYTHONPATH={lang_dir}/../../../lib:{lang_dir}"]

[[runner]]
key = "go"
name = "Go"

[runner.prepare]
template_path = "/home/user/.config/elf/runners/go.tmpl"
template_vars = { import_path = "github.com/me/aoc/{year}/{day}-{title}/go" }
build_commands = [["go", "mod", "tidy"], ["go", "build", "-o", "{binary_file}", "{wrapper_file}"]]

[runner.open]
binary = "{binary_file}"
`
	// Viper resolves config paths against the OS working directory, so the file
	// must be written at the absolute CWD path within the in-memory filesystem.
	cwd, err := os.Getwd()
	require.NoError(t, err)

	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, filepath.Join(cwd, "elf.toml"), []byte(tomlContent), 0o644))

	cfg, err := NewConfig(WithFs(fs), WithFile("elf.toml"))
	require.NoError(t, err)

	runners := cfg.GetRunners()
	require.Len(t, runners, 2)

	assert.Equal(t, "py", runners[0].Key)
	assert.Equal(t, "Python", runners[0].Name)
	assert.Equal(t, "/home/user/.config/elf/runners/python.tmpl", runners[0].Prepare.TemplatePath)
	assert.Equal(t, "python3", runners[0].Open.Interpreter)
	assert.Equal(t, []string{"-B", "{wrapper_file}"}, runners[0].Open.Args)

	assert.Equal(t, "go", runners[1].Key)
	assert.Equal(t, "{binary_file}", runners[1].Open.Binary)
	assert.Equal(t, "github.com/me/aoc/{year}/{day}-{title}/go", runners[1].Prepare.TemplateVars["import_path"])
}

func TestGetRunners_Empty(t *testing.T) {
	cfg, err := NewConfig()
	require.NoError(t, err)
	assert.Empty(t, cfg.GetRunners())
}
