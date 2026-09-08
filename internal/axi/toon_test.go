package axi

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncodeCollectionStableAndEscaped(t *testing.T) {
	collection := Collection{
		Name:    "issues",
		Query:   `project = AXI AND summary ~ "parser"`,
		Count:   2,
		Shown:   2,
		Columns: []string{"key", "status", "summary"},
		Records: []Record{
			{"summary": "Fix commas, quotes, and\nnewlines", "status": "Open", "key": "AXI-2"},
			{"key": "AXI-1", "status": "Done", "summary": "Stable output"},
		},
		Help: []string{"Run `acli --axi jira issue get <issue-key>`"},
	}

	var output bytes.Buffer
	if err := (Encoder{}).EncodeCollection(&output, collection); err != nil {
		t.Fatalf("EncodeCollection() error = %v", err)
	}

	want := "" +
		`query: "project = AXI AND summary ~ \"parser\""` + "\n" +
		"count: 2\n" +
		"shown: 2\n" +
		"has_more: false\n" +
		"empty: false\n" +
		"issues[2]{key,status,summary}:\n" +
		`  AXI-2,Open,"Fix commas, quotes, and\nnewlines"` + "\n" +
		`  AXI-1,Done,"Stable output"` + "\n" +
		"help[1]:\n" +
		"  \"Run `acli --axi jira issue get <issue-key>`\"\n"
	if output.String() != want {
		t.Fatalf("unexpected collection output:\n%s\nwant:\n%s", output.String(), want)
	}
}

func TestEncodeEmptyCollectionIsExplicit(t *testing.T) {
	var output bytes.Buffer
	err := (Encoder{}).EncodeCollection(&output, Collection{
		Name:    "issues",
		Columns: []string{"key", "summary"},
	})
	if err != nil {
		t.Fatalf("EncodeCollection() error = %v", err)
	}

	want := "count: 0\nshown: 0\nhas_more: false\nempty: true\nissues[0]{key,summary}:\n"
	if output.String() != want {
		t.Fatalf("unexpected empty collection output %q", output.String())
	}
}

func TestEncodeMutationUsesStableDataOrder(t *testing.T) {
	var output bytes.Buffer
	err := (Encoder{}).EncodeMutation(&output, Mutation{
		Status:  "ok",
		Action:  "jira.issue.create",
		Key:     "AXI-3",
		Message: "Created issue AXI-3",
		Data: map[string]interface{}{
			"url": "https://example.atlassian.net/browse/AXI-3",
			"id":  "10003",
		},
	})
	if err != nil {
		t.Fatalf("EncodeMutation() error = %v", err)
	}

	want := "status: ok\n" +
		"action: jira.issue.create\n" +
		"key: AXI-3\n" +
		`message: "Created issue AXI-3"` + "\n" +
		"data:\n" +
		"  id: 10003\n" +
		"  url: https://example.atlassian.net/browse/AXI-3\n"
	if output.String() != want {
		t.Fatalf("unexpected mutation output:\n%s\nwant:\n%s", output.String(), want)
	}
}

func TestEncodeErrorEnvelope(t *testing.T) {
	var output bytes.Buffer
	err := (Encoder{}).EncodeError(&output, Error{
		Code:      "context_missing",
		Message:   "No Jira project is configured",
		Retryable: false,
		Fix:       "Pass --project <key>",
		Exit:      1,
	})
	if err != nil {
		t.Fatalf("EncodeError() error = %v", err)
	}

	want := "error:\n" +
		"  code: context_missing\n" +
		`  message: "No Jira project is configured"` + "\n" +
		"  retryable: false\n" +
		`  fix: "Pass --project <key>"` + "\n" +
		"  exit: 1\n"
	if output.String() != want {
		t.Fatalf("unexpected error output:\n%s\nwant:\n%s", output.String(), want)
	}
}

func TestTruncateUsesRunesAndIncludesSize(t *testing.T) {
	got := Truncate("abcde🙂界", 6)
	want := "abcde🙂... [truncated from 7 chars]"
	if got != want {
		t.Fatalf("Truncate() = %q, want %q", got, want)
	}
	if !strings.Contains(got, "from 7 chars") {
		t.Fatal("truncation result does not include original size")
	}
}

func TestEncoderTruncatesAndEscapesFields(t *testing.T) {
	var output bytes.Buffer
	err := (Encoder{StringLimit: 5}).EncodeMutation(&output, Mutation{
		Status:  "ok",
		Action:  "test",
		Message: "line 1\nline 2",
	})
	if err != nil {
		t.Fatalf("EncodeMutation() error = %v", err)
	}
	if !strings.Contains(output.String(), `message: "line ... [truncated from 13 chars]"`) {
		t.Fatalf("message was not safely truncated and escaped: %s", output.String())
	}
}
