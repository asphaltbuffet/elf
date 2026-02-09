package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/krampus"
)

var initCmd *cobra.Command

func getInitCmd() *cobra.Command {
	if initCmd == nil {
		initCmd = &cobra.Command{
			Use:   "init",
			Short: "Create a new configuration file",
			Long: `Create a new elf configuration file.

By default, creates elf.toml in the current directory.
Use --global to create in the user config directory (~/.config/elf/).

The configuration file includes settings for:
  - Default solution language
  - Exercise directories
  - Advent of Code authentication token`,
			Example: `  elf config init           # Create elf.toml in current directory
  elf config init --global  # Create in ~/.config/elf/
  elf config init --force   # Overwrite existing file`,
			RunE: runInitCmd,
		}

		initCmd.Flags().BoolP("global", "g", false, "create config in user config directory")
		initCmd.Flags().BoolP("force", "f", false, "overwrite existing config file")
	}

	return initCmd
}

func runInitCmd(cmd *cobra.Command, _ []string) error {
	global, _ := cmd.Flags().GetBool("global")
	force, _ := cmd.Flags().GetBool("force")

	fs := afero.NewOsFs()

	var configPath string
	if global {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("getting user config directory: %w", err)
		}

		configPath = filepath.Join(configDir, "elf", krampus.DefaultConfigFileBase+"."+krampus.DefaultConfigExt)
	} else {
		configPath = krampus.DefaultConfigFileBase + "." + krampus.DefaultConfigExt
	}

	// Check if file exists
	if _, err := fs.Stat(configPath); err == nil {
		if !force {
			return fmt.Errorf("config file already exists: %s (use --force to overwrite)", configPath)
		}

		cmd.Printf("Overwriting existing config file: %s\n", configPath)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(configPath)
	if dir != "." {
		if err := fs.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating config directory: %w", err)
		}
	}

	// Generate and write config
	content := krampus.GenerateDefaultConfig()

	// Check if we're in interactive mode
	if isInteractive() {
		content = interactiveInit(cmd, content)
	}

	if err := afero.WriteFile(fs, configPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	cmd.Printf("Created configuration file: %s\n", configPath)

	return nil
}

// isInteractive checks if stdin is a terminal.
func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return (fi.Mode() & os.ModeCharDevice) != 0
}

// interactiveInit prompts the user for configuration values.
func interactiveInit(cmd *cobra.Command, defaultContent string) string {
	reader := bufio.NewReader(os.Stdin)

	cmd.Println("Interactive configuration setup")
	cmd.Println("Press Enter to accept defaults")
	cmd.Println()

	// Language
	cmd.Print("Default language [go]: ")

	lang, _ := reader.ReadString('\n')
	lang = strings.TrimSpace(lang)

	if lang == "" {
		lang = "go"
	}

	// Exercise directory
	cmd.Print("Exercise directory [exercises]: ")

	exerciseDir, _ := reader.ReadString('\n')
	exerciseDir = strings.TrimSpace(exerciseDir)

	if exerciseDir == "" {
		exerciseDir = "exercises"
	}

	// Token
	cmd.Println()
	cmd.Println("Advent of Code Token")
	cmd.Println("Get your token from adventofcode.com (browser cookies, 'session' cookie)")
	cmd.Print("Token (leave empty to skip): ")

	token, _ := reader.ReadString('\n')
	token = strings.TrimSpace(token)

	// Build custom content
	content := defaultContent

	// Replace language
	content = strings.Replace(content, `language = "go"`, fmt.Sprintf(`language = %q`, lang), 1)

	// Replace exercise dir
	content = strings.Replace(content, `dir = "exercises"`, fmt.Sprintf(`dir = %q`, exerciseDir), 1)

	// Replace token
	if token != "" {
		content = strings.Replace(content, `token = ""`, fmt.Sprintf(`token = %q`, token), 1)
	}

	return content
}
