// Output formatter interface definition
// Defines abstract interface for output formatters
package output

import (
	"io"

	"github.com/r1ckyIn/healthcheck-cli/internal/checker"
)

// Formatter is the interface for output formatters
type Formatter interface {
	// FormatSingle formats a single check result
	FormatSingle(result checker.Result) error

	// FormatBatch formats batch check results
	FormatBatch(result checker.BatchResult) error
}

// OutputFormat is the output format type
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
)

// NewFormatter creates a formatter based on format type
func NewFormatter(format OutputFormat, w io.Writer, noColor bool) Formatter {
	switch format {
	case FormatJSON:
		return NewJSONFormatter(w)
	case FormatTable:
		fallthrough
	default:
		return NewTableFormatter(w, noColor)
	}
}
