package acli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRejectAXIInteractionReturnsStructuredError(t *testing.T) {
	cmd := &cobra.Command{Use: "setup"}
	cmd.Flags().StringP("output", "o", "text", "output format")
	cmd.Flags().Bool("axi", false, "AXI output")
	_ = cmd.Flags().Set("axi", "true")
	var output bytes.Buffer
	cmd.SetOut(&output)

	err := rejectAXIInteraction(cmd, "Setup needs input", "Run setup without --axi")
	if err == nil {
		t.Fatal("expected interactive AXI command to fail")
	}
	for _, expected := range []string{
		"error:\n",
		"code: interaction_required",
		"message: \"Setup needs input\"",
		"fix: \"Run setup without --axi\"",
		"exit: 1",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("structured error missing %q:\n%s", expected, output.String())
		}
	}
}

func TestRejectAXIInteractionLeavesLegacyModeAlone(t *testing.T) {
	cmd := &cobra.Command{Use: "setup"}
	cmd.Flags().StringP("output", "o", "text", "output format")
	cmd.Flags().Bool("axi", false, "AXI output")
	var output bytes.Buffer
	cmd.SetOut(&output)

	if err := rejectAXIInteraction(cmd, "Setup needs input", "Run setup without --axi"); err != nil {
		t.Fatalf("legacy mode unexpectedly rejected: %v", err)
	}
	if output.Len() != 0 {
		t.Fatalf("legacy mode unexpectedly wrote output: %q", output.String())
	}
}
