package acli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/chinmaymk/acli/internal/axi"
	"github.com/spf13/cobra"
)

func TestIsJSONOutputDefault(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("output", "o", "text", "output format")

	if isJSONOutput(cmd) {
		t.Error("expected text output by default")
	}
}

func TestIsJSONOutputFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("output", "o", "text", "output format")
	_ = cmd.Flags().Set("output", "json")

	if !isJSONOutput(cmd) {
		t.Error("expected JSON output when --output=json")
	}
}

func TestIsJSONOutputBoolFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("output", "o", "text", "output format")
	cmd.Flags().Bool("json", false, "json output")
	_ = cmd.Flags().Set("json", "true")

	if !isJSONOutput(cmd) {
		t.Error("expected JSON output when --json flag is set")
	}
}

func TestOutputModeAXI(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("output", "o", "text", "output format")
	cmd.Flags().Bool("axi", false, "AXI output")
	_ = cmd.Flags().Set("axi", "true")

	if !isAXIOutput(cmd) {
		t.Fatal("expected AXI output when --axi is set")
	}
}

func TestExplicitJSONOverridesAXI(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("output", "o", "text", "output format")
	cmd.Flags().Bool("axi", false, "AXI output")
	_ = cmd.Flags().Set("axi", "true")
	_ = cmd.Flags().Set("output", "json")

	if !isJSONOutput(cmd) {
		t.Fatal("expected explicit JSON output to override AXI")
	}
	if isAXIOutput(cmd) {
		t.Fatal("did not expect AXI rendering when JSON was explicitly requested")
	}
}

func TestOutputJSON(t *testing.T) {
	data := map[string]string{"key": "TEST-1", "status": "open"}
	var output bytes.Buffer
	err := outputJSONTo(&output, data)
	if err != nil {
		t.Fatalf("outputJSON failed: %v", err)
	}
	want := "{\n  \"key\": \"TEST-1\",\n  \"status\": \"open\"\n}\n"
	if output.String() != want {
		t.Fatalf("unexpected JSON output %q, want %q", output.String(), want)
	}
}

func TestOutputResultText(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("output", "o", "text", "output format")
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := outputResult(cmd, "create", "TEST-1", "Issue TEST-1 created", nil)
	if err != nil {
		t.Fatalf("outputResult failed: %v", err)
	}
	if buf.String() != "Issue TEST-1 created\n" {
		t.Errorf("unexpected text output: %q", buf.String())
	}
}

func TestOutputResultJSON(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("output", "o", "text", "output format")
	_ = cmd.Flags().Set("output", "json")
	var output bytes.Buffer
	cmd.SetOut(&output)

	err := outputResult(cmd, "delete", "TEST-1", "Issue TEST-1 deleted", map[string]string{"id": "123"})
	if err != nil {
		t.Fatalf("outputResult JSON failed: %v", err)
	}
	want := "{\n  \"status\": \"ok\",\n  \"action\": \"delete\",\n  \"key\": \"TEST-1\",\n  \"message\": \"Issue TEST-1 deleted\",\n  \"data\": {\n    \"id\": \"123\"\n  }\n}\n"
	if output.String() != want {
		t.Fatalf("unexpected JSON result %q, want %q", output.String(), want)
	}
}

func TestOutputResultAXI(t *testing.T) {
	root := &cobra.Command{Use: "acli"}
	root.PersistentFlags().StringP("output", "o", "text", "output format")
	root.PersistentFlags().Bool("axi", false, "AXI output")
	cmd := &cobra.Command{
		Use: "create",
		RunE: func(cmd *cobra.Command, args []string) error {
			return outputResult(cmd, "created", "TEST-1", "Issue TEST-1 created", nil)
		},
	}
	root.AddCommand(cmd)

	var output bytes.Buffer
	root.SetOut(&output)
	root.SetArgs([]string{"--axi", "create"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	want := "status: ok\naction: create\nkey: TEST-1\nmessage: \"Issue TEST-1 created\"\n"
	if output.String() != want {
		t.Fatalf("unexpected AXI output %q, want %q", output.String(), want)
	}
}

func TestOutputActionableErrorJSONPrecedence(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("output", "o", "text", "output format")
	cmd.Flags().Bool("axi", false, "AXI output")
	_ = cmd.Flags().Set("axi", "true")
	_ = cmd.Flags().Set("output", "json")

	var output bytes.Buffer
	cmd.SetOut(&output)
	err := outputActionableError(cmd, axi.Error{
		Code:    "interaction_required",
		Message: "Interactive command",
		Exit:    1,
	})
	if err == nil {
		t.Fatal("expected command error")
	}

	want := "{\n  \"error\": {\n    \"code\": \"interaction_required\",\n    \"message\": \"Interactive command\",\n    \"retryable\": false,\n    \"exit\": 1\n  }\n}\n"
	if output.String() != want {
		t.Fatalf("unexpected JSON error %q, want %q", output.String(), want)
	}
}

func TestOutputResultStruct(t *testing.T) {
	r := OutputResult{
		Status:  "ok",
		Action:  "create",
		Key:     "TEST-1",
		Message: "Created",
		Data:    map[string]string{"url": "https://example.com"},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded OutputResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Status != "ok" {
		t.Errorf("unexpected status: %s", decoded.Status)
	}
	if decoded.Action != "create" {
		t.Errorf("unexpected action: %s", decoded.Action)
	}
	if decoded.Key != "TEST-1" {
		t.Errorf("unexpected key: %s", decoded.Key)
	}
}
