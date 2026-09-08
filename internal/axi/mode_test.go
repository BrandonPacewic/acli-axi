package axi

import "testing"

func TestResolveMode(t *testing.T) {
	tests := []struct {
		name    string
		options ModeOptions
		want    Mode
	}{
		{name: "legacy text", options: ModeOptions{Output: "text"}, want: ModeText},
		{name: "axi", options: ModeOptions{AXI: true, Output: "text"}, want: ModeAXI},
		{name: "explicit json", options: ModeOptions{Output: "json"}, want: ModeJSON},
		{name: "json overrides axi", options: ModeOptions{AXI: true, Output: "json"}, want: ModeJSON},
		{name: "legacy json alias overrides axi", options: ModeOptions{AXI: true, JSONAlias: true}, want: ModeJSON},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ResolveMode(test.options); got != test.want {
				t.Fatalf("ResolveMode(%+v) = %q, want %q", test.options, got, test.want)
			}
		})
	}
}
