package axi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const DefaultStringLimit = 240

// Encoder writes deterministic, compact TOON-like records.
type Encoder struct {
	StringLimit int
}

// Truncate bounds text by Unicode code points and includes the original size.
func Truncate(value string, limit int) string {
	if limit <= 0 || utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return fmt.Sprintf("%s... [truncated from %d chars]", string(runes[:limit]), len(runes))
}

// EncodeCollection writes a structured collection envelope.
func (e Encoder) EncodeCollection(w io.Writer, collection Collection) error {
	if err := validateIdentifier(collection.Name); err != nil {
		return fmt.Errorf("collection name: %w", err)
	}
	if len(collection.Columns) == 0 {
		return fmt.Errorf("collection columns cannot be empty")
	}
	for _, column := range collection.Columns {
		if err := validateIdentifier(column); err != nil {
			return fmt.Errorf("collection column %q: %w", column, err)
		}
	}

	var output bytes.Buffer
	if collection.Query != "" {
		e.writeField(&output, 0, "query", collection.Query)
	}
	e.writeField(&output, 0, "count", collection.Count)
	e.writeField(&output, 0, "shown", collection.Shown)
	e.writeField(&output, 0, "has_more", collection.HasMore)
	e.writeField(&output, 0, "empty", collection.Count == 0)
	_, _ = fmt.Fprintf(&output, "%s[%d]{%s}:\n", collection.Name, len(collection.Records), strings.Join(collection.Columns, ","))
	for _, record := range collection.Records {
		values := make([]string, len(collection.Columns))
		for i, column := range collection.Columns {
			values[i] = e.scalar(record[column])
		}
		_, _ = fmt.Fprintf(&output, "  %s\n", strings.Join(values, ","))
	}
	e.writeHints(&output, collection.Help)
	_, err := io.Copy(w, &output)
	return err
}

// EncodeMutation writes a structured mutation envelope.
func (e Encoder) EncodeMutation(w io.Writer, mutation Mutation) error {
	var output bytes.Buffer
	e.writeField(&output, 0, "status", mutation.Status)
	e.writeField(&output, 0, "action", mutation.Action)
	if mutation.ID != "" {
		e.writeField(&output, 0, "id", mutation.ID)
	}
	if mutation.Key != "" {
		e.writeField(&output, 0, "key", mutation.Key)
	}
	e.writeField(&output, 0, "message", mutation.Message)
	if mutation.Data != nil {
		_, _ = fmt.Fprintln(&output, "data:")
		if err := e.writeValue(&output, 1, mutation.Data); err != nil {
			return err
		}
	}
	e.writeHints(&output, mutation.Help)
	_, err := io.Copy(w, &output)
	return err
}

// EncodeError writes a structured error envelope.
func (e Encoder) EncodeError(w io.Writer, commandError Error) error {
	var output bytes.Buffer
	_, _ = fmt.Fprintln(&output, "error:")
	e.writeField(&output, 1, "code", commandError.Code)
	e.writeField(&output, 1, "message", commandError.Message)
	e.writeField(&output, 1, "retryable", commandError.Retryable)
	if commandError.Fix != "" {
		e.writeField(&output, 1, "fix", commandError.Fix)
	}
	e.writeField(&output, 1, "exit", commandError.Exit)
	_, err := io.Copy(w, &output)
	return err
}

func (e Encoder) writeHints(w io.Writer, hints []string) {
	hints = LimitHints(DefaultHintLimit, hints...)
	if len(hints) == 0 {
		return
	}
	_, _ = fmt.Fprintf(w, "help[%d]:\n", len(hints))
	for _, hint := range hints {
		_, _ = fmt.Fprintf(w, "  %s\n", e.scalar(hint))
	}
}

func (e Encoder) writeValue(w io.Writer, depth int, value interface{}) error {
	normalized, err := normalize(value)
	if err != nil {
		return fmt.Errorf("encode data: %w", err)
	}
	e.writeNormalized(w, depth, normalized)
	return nil
}

func (e Encoder) writeNormalized(w io.Writer, depth int, value interface{}) {
	indent := strings.Repeat("  ", depth)
	switch typed := value.(type) {
	case map[string]interface{}:
		if len(typed) == 0 {
			_, _ = fmt.Fprintf(w, "%s{}\n", indent)
			return
		}
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			child := typed[key]
			if isContainer(child) {
				_, _ = fmt.Fprintf(w, "%s%s:\n", indent, encodeKey(key))
				e.writeNormalized(w, depth+1, child)
				continue
			}
			_, _ = fmt.Fprintf(w, "%s%s: %s\n", indent, encodeKey(key), e.scalar(child))
		}
	case []interface{}:
		if len(typed) == 0 {
			_, _ = fmt.Fprintf(w, "%s[]\n", indent)
			return
		}
		for _, child := range typed {
			if isContainer(child) {
				_, _ = fmt.Fprintf(w, "%s-\n", indent)
				e.writeNormalized(w, depth+1, child)
				continue
			}
			_, _ = fmt.Fprintf(w, "%s- %s\n", indent, e.scalar(child))
		}
	default:
		_, _ = fmt.Fprintf(w, "%s%s\n", indent, e.scalar(typed))
	}
}

func normalize(value interface{}) (interface{}, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var normalized interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}

func isContainer(value interface{}) bool {
	switch value.(type) {
	case map[string]interface{}, []interface{}:
		return true
	default:
		return false
	}
}

func (e Encoder) writeField(w io.Writer, depth int, name string, value interface{}) {
	_, _ = fmt.Fprintf(w, "%s%s: %s\n", strings.Repeat("  ", depth), name, e.scalar(value))
}

func (e Encoder) scalar(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch typed := value.(type) {
	case string:
		return encodeString(Truncate(typed, e.stringLimit()))
	case bool:
		return strconv.FormatBool(typed)
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case fmt.Stringer:
		return encodeString(Truncate(typed.String(), e.stringLimit()))
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(reflected.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(reflected.Uint(), 10)
	}
	return encodeString(Truncate(fmt.Sprint(value), e.stringLimit()))
}

func (e Encoder) stringLimit() int {
	if e.StringLimit <= 0 {
		return DefaultStringLimit
	}
	return e.StringLimit
}

func encodeString(value string) string {
	if isBareString(value) {
		return value
	}
	return strconv.QuoteToGraphic(value)
}

func isBareString(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("_./:@+-", r) {
			continue
		}
		return false
	}
	return true
}

func validateIdentifier(value string) error {
	if value == "" {
		return fmt.Errorf("cannot be empty")
	}
	for i, r := range value {
		if unicode.IsLetter(r) || r == '_' || (i > 0 && unicode.IsDigit(r)) {
			continue
		}
		return fmt.Errorf("%q is not a valid identifier", value)
	}
	return nil
}

func encodeKey(value string) string {
	if validateIdentifier(value) == nil {
		return value
	}
	return strconv.QuoteToGraphic(value)
}
