package runners

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	elfrunners "github.com/asphaltbuffet/elf/pkg/runners"
)

var installCmd *cobra.Command

func getInstallCmd() *cobra.Command {
	if installCmd == nil {
		installCmd = &cobra.Command{
			Use:               "install",
			Short:             "Install built-in runner template files",
			Long:              "Writes the built-in Go and Python runner wrapper templates to ~/.config/elf/runners/ and prints the [[runner]] config blocks to add to elf.toml.",
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
		{"python.templ", elfrunners.PythonTemplate},
		{"go.tmpl", elfrunners.GoTemplate},
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

	cmd.Println()
	cmd.Println("Add the following to your elf.toml:")
	cmd.Println()
	cmd.Printf(`[[runner]]
key = "py"
name = "Python"

[runner.prepare]
template_path = %q

[runner.open]
interpreter = "python3"
args = ["-B", "{wrapper_file}"]
env = ["PYTHONPATH={lang_dir}/../../../lib:{lang_dir}"]

[[runner]]
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
`,
		filepath.Join(runnersDir, "python.templ"),
		filepath.Join(runnersDir, "go.tmpl"),
	)

	cmd.Println("Replace YOUR_MODULE with your Go module name (from go.mod, e.g. github.com/you/advent-of-code).")

	return nil
}
