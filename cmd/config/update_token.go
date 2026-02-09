package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/krampus"
)

var updateTokenCmd *cobra.Command

//nolint:gosec // not storing any credentials here
var updateTokenLong = `Update the Advent of Code authentication token in your configuration.

To get your token:
  1. Log in to adventofcode.com
  2. Open browser developer tools (F12)
  3. Go to Application/Storage > Cookies
  4. Copy the value of the 'session' cookie

The token is required for downloading puzzle inputs.`

func getUpdateTokenCmd() *cobra.Command {
	if updateTokenCmd == nil {
		updateTokenCmd = &cobra.Command{
			Use:               "update-token",
			Short:             "Update the Advent of Code authentication token",
			Long:              updateTokenLong,
			Args:              cobra.NoArgs,
			ValidArgsFunction: cobra.NoFileCompletions,
			Example: `elf config update-token                   # Interactive prompt
elf config update-token -t "your-token"   # Direct update`,
			RunE: runUpdateTokenCmd,
		}

		updateTokenCmd.Flags().StringP("token", "t", "", "new token value")
		updateTokenCmd.Flags().StringP("config-file", "c", "", "configuration file to update")
	}

	return updateTokenCmd
}

func runUpdateTokenCmd(cmd *cobra.Command, _ []string) error {
	tokenFlag, _ := cmd.Flags().GetString("token")
	cfgFile, _ := cmd.Flags().GetString("config-file")

	cfg, err := krampus.NewConfig(krampus.WithFile(cfgFile))
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	// Get current config file path
	configPath := cfg.GetConfigFileUsed()
	if configPath == "" {
		return errors.New("no configuration file found; run 'elf config init' first")
	}

	var token string

	if tokenFlag != "" {
		token = tokenFlag
	} else {
		// Interactive mode
		token, err = promptForToken(cmd, cfg.GetToken())
		if err != nil {
			return err
		}
	}

	if token == "" {
		return errors.New("token cannot be empty")
	}

	// Update token and write config
	cfg.SetToken(token)

	if writeErr := cfg.WriteConfig(configPath); writeErr != nil {
		return fmt.Errorf("writing configuration: %w", writeErr)
	}

	cmd.Printf("Token updated in %s\n", configPath)

	return nil
}

func promptForToken(cmd *cobra.Command, currentToken string) (string, error) {
	cmd.Println("Update Advent of Code Token")
	cmd.Println()
	cmd.Println("To get your token:")
	cmd.Println("  1. Log in to adventofcode.com")
	cmd.Println("  2. Open browser developer tools (F12)")
	cmd.Println("  3. Go to Application/Storage > Cookies")
	cmd.Println("  4. Copy the value of the 'session' cookie")
	cmd.Println()

	if currentToken != "" && currentToken != krampus.MaskToken("") {
		cmd.Printf("Current token: %s\n", krampus.MaskToken(currentToken))
	}

	cmd.Print("Enter new token: ")

	reader := bufio.NewReader(os.Stdin)

	token, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}

	return strings.TrimSpace(token), nil
}
