// JSON format output
// Implements machine-readable JSON format output
package output

import (
	"encoding/json"
	"io"

	"github.com/r1ckyIn/healthcheck-cli/internal/checker"
)

// JSONFormatter implements JSON format output
type JSONFormatter struct {
	writer io.Writer
}

// NewJSONFormatter creates a JSON formatter
func NewJSONFormatter(w io.Writer) *JSONFormatter {
	return &JSONFormatter{
		writer: w,
	}
}

// singleResultJSON is the JSON structure for single result
type singleResultJSON struct {
	URL        string  `json:"url"`
	Healthy    bool    `json:"healthy"`
	StatusCode *int    `json:"status_code"`
	LatencyMs  *int64  `json:"latency_ms"`
	Error      *string `json:"error"`
}

// batchResultJSON is the JSON structure for batch results
type batchResultJSON struct {
	Timestamp  string            `json:"timestamp"`
	DurationMs int64             `json:"duration_ms"`
	Summary    summaryJSON       `json:"summary"`
	Results    []resultItemJSON  `json:"results"`
}

// summaryJSON is the JSON structure for summary information
type summaryJSON struct {
	Total     int `json:"total"`
	Healthy   int `json:"healthy"`
	Unhealthy int `json:"unhealthy"`
}

// resultItemJSON is the JSON structure for result item
type resultItemJSON struct {
	Name       string  `json:"name"`
	URL        string  `json:"url"`
	Healthy    bool    `json:"healthy"`
	StatusCode *int    `json:"status_code"`
	LatencyMs  *int64  `json:"latency_ms"`
	Error      *string `json:"error"`
}

// FormatSingle formats a single check result
func (f *JSONFormatter) FormatSingle(result checker.Result) error {
	output := singleResultJSON{
		URL:        result.URL,
		Healthy:    result.Healthy,
		StatusCode: result.StatusCode,
	}

	// Calculate latency (milliseconds)
	if result.Healthy || result.StatusCode != nil {
		latencyMs := result.Latency.Milliseconds()
		output.LatencyMs = &latencyMs
	}

	// Error message
	if result.Error != nil {
		errStr := result.Error.Error()
		output.Error = &errStr
	}

	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// FormatBatch formats batch check results
func (f *JSONFormatter) FormatBatch(batch checker.BatchResult) error {
	output := batchResultJSON{
		Timestamp:  batch.Timestamp.Format("2006-01-02T15:04:05Z"),
		DurationMs: batch.Summary.Duration.Milliseconds(),
		Summary: summaryJSON{
			Total:     batch.Summary.Total,
			Healthy:   batch.Summary.Healthy,
			Unhealthy: batch.Summary.Unhealthy,
		},
		Results: make([]resultItemJSON, len(batch.Results)),
	}

	// Convert each result
	for i, result := range batch.Results {
		item := resultItemJSON{
			Name:       result.Name,
			URL:        result.URL,
			Healthy:    result.Healthy,
			StatusCode: result.StatusCode,
		}

		// Latency time
		if result.Healthy || result.StatusCode != nil {
			latencyMs := result.Latency.Milliseconds()
			item.LatencyMs = &latencyMs
		}

		// Error message
		if result.Error != nil {
			errStr := result.Error.Error()
			item.Error = &errStr
		}

		output.Results[i] = item
	}

	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
