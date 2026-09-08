// Package axi defines the product-neutral output contract for ACLI's agent mode.
package axi

// Mode identifies the renderer selected for a command invocation.
type Mode string

const (
	ModeText Mode = "text"
	ModeJSON Mode = "json"
	ModeAXI  Mode = "axi"
)

// ModeOptions contains the flags that influence output mode selection.
type ModeOptions struct {
	AXI       bool
	Output    string
	JSONAlias bool
}

// ResolveMode applies output precedence. An explicit JSON request always wins
// over AXI mode, including legacy per-command --json flags.
func ResolveMode(options ModeOptions) Mode {
	if options.Output == "json" || options.JSONAlias {
		return ModeJSON
	}
	if options.AXI {
		return ModeAXI
	}
	return ModeText
}
