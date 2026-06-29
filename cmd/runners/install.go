package runners

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	elfrunners "github.com/asphaltbuffet/elf/pkg/runners"
)

var installCmd *cobra.Command

func getInstallCmd() *cobra.Command {
	if installCmd == nil {
		installCmd = &cobra.Command{
			Use:               "install",
			Short:             "Install built-in runner template files",
			Long:              "Writes the built-in runner wrapper templates to ~/.config/elf/runners/ and prints the [[runner]] config blocks to add to elf.toml.",
			Args:              cobra.NoArgs,
			ValidArgsFunction: cobra.NoFileCompletions,
			RunE:              runInstallCmd,
			Example:           "elf runners install",
		}

		installCmd.Flags().BoolP("force", "f", false, "overwrite existing template files")
	}

	return installCmd
}

func runInstallCmd(cmd *cobra.Command, _ []string) error {
	force, _ := cmd.Flags().GetBool("force")

	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("getting user config directory: %w", err)
	}

	runnersDir := filepath.Join(configDir, "elf", "runners")
	if err = os.MkdirAll(runnersDir, 0o750); err != nil {
		return fmt.Errorf("creating runners directory: %w", err)
	}

	templates := []struct {
		filename string
		content  []byte
	}{
		{"python.tmpl", elfrunners.PythonTemplate},
		{"go.tmpl", elfrunners.GoTemplate},
		{"bash.tmpl", elfrunners.BashTemplate},
		{"rust.tmpl", elfrunners.RustTemplate},
		{"f77.tmpl", elfrunners.F77Template},
		{"lua.tmpl", elfrunners.LuaTemplate},
	}

	for _, tmpl := range templates {
		dest := filepath.Join(runnersDir, tmpl.filename)

		if _, statErr := os.Stat(dest); statErr == nil && !force {
			cmd.Printf("Skipping %s (already exists; use --force to overwrite)\n", dest)
			continue
		}

		if writeErr := os.WriteFile(dest, tmpl.content, 0o600); writeErr != nil {
			return fmt.Errorf("writing %s: %w", dest, writeErr)
		}

		cmd.Printf("Wrote %s\n", dest)
	}

	printRunnerConfig(cmd, runnersDir)

	return nil
}

type runnerConfigBlock struct {
	tmplFile string
	toml     func(tmplPath string) string
}

func pythonBlock(p string) string {
	return fmt.Sprintf(`[[runner]]
key = "py"
name = "Python"

[runner.prepare]
template_path = %q

[runner.open]
interpreter = "python3"
args = ["-B", "{wrapper_file}"]
# PYTHONPATH entries: the exercise dir ({lang_dir}/..) so the wrapper's
# "from py import Exercise" resolves, plus a shared library dir. The lib path
# below assumes a lib/ at the project root (exercises/<year>/<day>/py -> four
# levels up); adjust if your shared package lives elsewhere.
env = ["PYTHONPATH={lang_dir}/..:{lang_dir}/../../../../lib"]
`, p)
}

func goBlock(p string) string {
	return fmt.Sprintf(`[[runner]]
key = "go"
name = "Go"

[runner.prepare]
template_path = %q
template_vars = { import_path = "YOUR_MODULE/{year}/{day}-{title}/go" }
build_commands = [
  ["go", "mod", "tidy"],
  ["go", "build", "-tags", "runtime", "-o", "{binary_file}", "{wrapper_file}"],
]

[runner.open]
binary = "{binary_file}"
`, p)
}

func bashBlock(p string) string {
	return fmt.Sprintf(`[[runner]]
key = "bash"
name = "Bash"

[runner.prepare]
template_path = %q
wrapper_ext = ".sh"

[runner.open]
interpreter = "bash"
args = ["{wrapper_file}"]
`, p)
}

func rustBlock(p string) string {
	return fmt.Sprintf(`[[runner]]
key = "rs"
name = "Rust"

[runner.prepare]
template_path = %q
# Wrapper renders to {lang_dir}/src/runtime-wrapper.rs (the bin crate entrypoint;
# the scaffolded Cargo.toml points [[bin]] path at that file).
wrapper_subdir = "src"
wrapper_ext = ".rs"
build_commands = [
  ["cargo", "build", "--release", "--manifest-path", "{lang_dir}/Cargo.toml"],
]

[runner.open]
# Crate name is pinned to "solution" in the scaffolded Cargo.toml, so the built
# binary is always at this path regardless of year/day.
binary = "{lang_dir}/target/release/solution"
`, p)
}

func f77Block(p string) string {
	return fmt.Sprintf(`[[runner]]
key = "f77"
name = "Fortran 77"

[runner.prepare]
template_path = %q
wrapper_ext = ".c"
build_commands = [
  ["gfortran", "-O2", "-o", "{binary_file}", "{wrapper_file}", "{lang_dir}/solution.f"],
]

[runner.open]
binary = "{binary_file}"
`, p)
}

func luaBlock(p string) string {
	return fmt.Sprintf(`[[runner]]
key = "lua"
name = "Lua"

[runner.prepare]
template_path = %q
wrapper_ext = ".lua"

[runner.open]
interpreter = "lua"
args = ["{wrapper_file}"]
# Requires dkjson: lua5_2.withPackages (ps: [ps.dkjson]) in your Nix devshell,
# or via LuaRocks: luarocks install dkjson
`, p)
}

func runnerConfigBlocks() []runnerConfigBlock {
	return []runnerConfigBlock{
		{"python.tmpl", pythonBlock},
		{"go.tmpl", goBlock},
		{"bash.tmpl", bashBlock},
		{"rust.tmpl", rustBlock},
		{"f77.tmpl", f77Block},
		{"lua.tmpl", luaBlock},
	}
}

// printRunnerConfig prints the paste-able [[runner]] blocks for the built-in
// runners, with template paths resolved under runnersDir.
func printRunnerConfig(cmd *cobra.Command, runnersDir string) {
	cmd.Println()
	cmd.Println("Add the following to your elf.toml:")
	cmd.Println()

	var sb strings.Builder
	for _, block := range runnerConfigBlocks() {
		sb.WriteString(block.toml(filepath.Join(runnersDir, block.tmplFile)))
		sb.WriteByte('\n')
	}
	cmd.Print(sb.String())

	cmd.Println("Replace YOUR_MODULE with your Go module name (from go.mod, e.g. github.com/you/advent-of-code).")
}
