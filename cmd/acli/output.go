package acli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/chinmaymk/acli/internal/axi"
	"github.com/spf13/cobra"
)

// isJSONOutput returns true if the user requested JSON output via --output json or --json.
func isJSONOutput(cmd *cobra.Command) bool {
	return outputMode(cmd) == axi.ModeJSON
}

func isAXIOutput(cmd *cobra.Command) bool {
	return outputMode(cmd) == axi.ModeAXI
}

func isAXIRequested(cmd *cobra.Command) bool {
	requested, _ := cmd.Flags().GetBool("axi")
	return requested
}

func outputMode(cmd *cobra.Command) axi.Mode {
	output, _ := cmd.Flags().GetString("output")
	requested, _ := cmd.Flags().GetBool("axi")
	jsonAlias := false
	if f := cmd.Flags().Lookup("json"); f != nil {
		jsonAlias, _ = cmd.Flags().GetBool("json")
	}
	return axi.ResolveMode(axi.ModeOptions{AXI: requested, Output: output, JSONAlias: jsonAlias})
}

// outputJSON outputs v as indented JSON to stdout. Used for data responses.
func outputJSON(v interface{}) error {
	return outputJSONTo(rootCmd.OutOrStdout(), v)
}

func outputJSONTo(w io.Writer, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

// OutputResult represents a structured result for mutation operations (create, update, delete, etc).
// Agents can parse this to confirm actions and extract identifiers.
type OutputResult struct {
	Status  string      `json:"status"`
	Action  string      `json:"action"`
	ID      string      `json:"id,omitempty"`
	Key     string      `json:"key,omitempty"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// outputResult outputs a structured result for mutation operations.
// In text mode, it prints the message. In JSON mode, it outputs a structured envelope.
func outputResult(cmd *cobra.Command, action, key, message string, data interface{}) error {
	if isJSONOutput(cmd) {
		return outputJSONTo(cmd.OutOrStdout(), OutputResult{
			Status:  "ok",
			Action:  action,
			Key:     key,
			Message: message,
			Data:    data,
		})
	}
	if isAXIOutput(cmd) {
		return (axi.Encoder{}).EncodeMutation(cmd.OutOrStdout(), axi.Mutation{
			Status:  "ok",
			Action:  axiAction(cmd, action),
			Key:     key,
			Message: message,
			Data:    data,
		})
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), message)
	return nil
}

func outputActionableError(cmd *cobra.Command, commandError axi.Error) error {
	var err error
	if outputMode(cmd) == axi.ModeJSON {
		err = outputJSONTo(cmd.OutOrStdout(), axi.ErrorEnvelope{Error: commandError})
	} else {
		err = (axi.Encoder{}).EncodeError(cmd.OutOrStdout(), commandError)
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("%s: %s", commandError.Code, commandError.Message)
}

func axiAction(cmd *cobra.Command, fallback string) string {
	parts := strings.Fields(cmd.CommandPath())
	if len(parts) > 1 {
		return strings.Join(parts[1:], ".")
	}
	return fallback
}
