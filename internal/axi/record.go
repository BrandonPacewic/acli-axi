package axi

// Record is one row in a collection. Columns on Collection determine field
// order, so record maps never affect encoded output stability.
type Record map[string]interface{}

// Collection describes a bounded page of records.
type Collection struct {
	Name    string
	Query   string
	Count   int
	Shown   int
	HasMore bool
	Columns []string
	Records []Record
	Help    []string
}

// Mutation describes the result of a state-changing operation.
type Mutation struct {
	Status  string
	Action  string
	ID      string
	Key     string
	Message string
	Data    interface{}
	Help    []string
}

// Error describes an actionable command failure.
type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
	Fix       string `json:"fix,omitempty"`
	Exit      int    `json:"exit"`
}

// ErrorEnvelope preserves the same explicit error shape in JSON mode.
type ErrorEnvelope struct {
	Error Error `json:"error"`
}
