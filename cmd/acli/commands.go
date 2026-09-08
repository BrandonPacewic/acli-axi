package acli

import (
	"github.com/chinmaymk/acli/internal/axi"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandInfo describes a command in a machine-readable format for agent discovery.
type CommandInfo struct {
	Name        string        `json:"name"`
	FullCommand string        `json:"full_command"`
	Description string        `json:"description"`
	Aliases     []string      `json:"aliases,omitempty"`
	Usage       string        `json:"usage"`
	Args        string        `json:"args,omitempty"`
	Flags       []FlagInfo    `json:"flags,omitempty"`
	Subcommands []CommandInfo `json:"subcommands,omitempty"`
}

// FlagInfo describes a flag in a machine-readable format.
type FlagInfo struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Type      string `json:"type"`
	Default   string `json:"default,omitempty"`
	Required  bool   `json:"required,omitempty"`
	Usage     string `json:"usage"`
}

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "List all commands in machine-readable JSON (for agents and scripts)",
	Long: `Outputs the full command tree as JSON, including all subcommands, flags,
aliases, and descriptions. Designed for programmatic consumption by LLM coding
agents and automation scripts that need to discover available operations.`,
	RunE: runCommands,
}

func runCommands(cmd *cobra.Command, _ []string) error {
	if isAXIOutput(cmd) {
		return outputAXICommands(cmd)
	}
	return outputJSONTo(cmd.OutOrStdout(), buildCommandTree(rootCmd))
}

func outputAXICommands(cmd *cobra.Command) error {
	products := []axi.Record{
		{"name": "jira", "aliases": "j", "status": "foundation", "command": "acli --axi jira", "description": jiraCmd.Short},
		{"name": "confluence", "aliases": "conf,c", "status": "planned", "command": "acli --axi confluence", "description": confluenceCmd.Short},
		{"name": "bitbucket", "aliases": "bb", "status": "planned", "command": "acli --axi bitbucket", "description": bitbucketCmd.Short},
	}
	return (axi.Encoder{}).EncodeCollection(cmd.OutOrStdout(), axi.Collection{
		Name:    "products",
		Count:   len(products),
		Shown:   len(products),
		Columns: []string{"name", "aliases", "status", "command", "description"},
		Records: products,
		Help: axi.LimitHints(2,
			"Run `acli --axi jira issue list`",
			"Run `acli --axi jira search --jql <jql>`",
		),
	})
}

func buildCommandTree(cmd *cobra.Command) CommandInfo {
	info := CommandInfo{
		Name:        cmd.Name(),
		FullCommand: cmd.CommandPath(),
		Description: cmd.Short,
		Aliases:     cmd.Aliases,
		Usage:       cmd.UseLine(),
	}

	// Determine expected args from Use string
	if cmd.Args != nil {
		info.Args = extractArgsFromUse(cmd.Use)
	}

	// Collect local flags (not inherited)
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		fi := FlagInfo{
			Name:      f.Name,
			Shorthand: f.Shorthand,
			Type:      f.Value.Type(),
			Default:   f.DefValue,
			Usage:     f.Usage,
		}
		// Check if flag is required via annotations
		if ann, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok {
			for _, v := range ann {
				if v == "true" {
					fi.Required = true
				}
			}
		}
		info.Flags = append(info.Flags, fi)
	})

	// Recurse into subcommands
	for _, sub := range cmd.Commands() {
		if sub.Hidden || sub.Name() == "help" || sub.Name() == "completion" {
			continue
		}
		info.Subcommands = append(info.Subcommands, buildCommandTree(sub))
	}

	return info
}

// extractArgsFromUse extracts the argument portion from a cobra Use string.
// e.g. "get <issue-key>" -> "<issue-key>"
func extractArgsFromUse(use string) string {
	for i, c := range use {
		if c == ' ' {
			return use[i+1:]
		}
	}
	return ""
}
