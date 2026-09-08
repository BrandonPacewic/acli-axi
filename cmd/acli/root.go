package acli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "acli",
	Short: "Atlassian CLI - manage Jira, Confluence, and Bitbucket from the terminal",
	Long: `acli is a command-line interface for Atlassian Cloud products.

Manage your Jira issues, Confluence pages, and Bitbucket repositories
without leaving the terminal.`,
}

func init() {
	rootCmd.AddCommand(jiraCmd)
	rootCmd.AddCommand(confluenceCmd)
	rootCmd.AddCommand(bitbucketCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(commandsCmd)

	addPersistentFlags(rootCmd)
}

func addPersistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("profile", "p", "", "configuration profile to use (defaults to the default profile)")
	cmd.PersistentFlags().StringP("output", "o", "text", "output format: text or json (json is recommended for programmatic/agent use)")
	cmd.PersistentFlags().Bool("axi", false, "enable compact, non-interactive agent output")
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("acli %s (commit: %s, built: %s)\n", version, commit, date)
	},
}

func Execute() error {
	args, err := resolveCLIArgs(os.Args[1:], os.Stdin)
	if err != nil {
		return err
	}

	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func resolveCLIArgs(args []string, in io.Reader) ([]string, error) {
	if len(args) == 1 && args[0] == "-" {
		return readArgsFromStdin(in)
	}

	return args, nil
}

func readArgsFromStdin(in io.Reader) ([]string, error) {
	raw, err := io.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdin arguments: %w", err)
	}

	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil, fmt.Errorf("stdin argument mode expects a JSON array, got empty input")
	}

	var args []string
	if err := json.Unmarshal([]byte(trimmed), &args); err != nil {
		return nil, fmt.Errorf("invalid stdin argument JSON array: %w", err)
	}

	return args, nil
}
