// Table format output
// Implements human-readable table format output
package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/r1ckyIn/healthcheck-cli/internal/checker"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// Table column width limits
const (
	maxNameWidth = 30
	maxURLWidth  = 50
)

// TableFormatter implements table format output
type TableFormatter struct {
	writer  io.Writer
	noColor bool
}

// NewTableFormatter creates a table formatter
func NewTableFormatter(w io.Writer, noColor bool) *TableFormatter {
	return &TableFormatter{
		writer:  w,
		noColor: noColor,
	}
}

// FormatSingle formats a single check result
func (f *TableFormatter) FormatSingle(result checker.Result) error {
	var status string
	var latency string

	if result.Healthy {
		status = f.colorize("✓", colorGreen)
		if result.StatusCode != nil {
			status += fmt.Sprintf(" %d", *result.StatusCode)
		}
	} else {
		status = f.colorize("✗", colorRed)
		if result.StatusCode != nil {
			status += fmt.Sprintf(" %d", *result.StatusCode)
		} else if result.Error != nil {
			// Extract short error message
			status += " " + f.getShortError(result.Error)
		}
	}

	if result.Healthy || result.StatusCode != nil {
		latency = formatLatency(result.Latency)
	} else {
		latency = "--"
	}

	_, err := fmt.Fprintf(f.writer, "%s %s    %s\n", status, result.URL, latency)
	return err
}

// FormatBatch formats batch check results
func (f *TableFormatter) FormatBatch(batch checker.BatchResult) error {
	// Calculate column widths
	nameWidth := 4  // "NAME"
	urlWidth := 3   // "URL"

	for _, r := range batch.Results {
		if len(r.Name) > nameWidth {
			nameWidth = len(r.Name)
		}
		if len(r.URL) > urlWidth {
			urlWidth = len(r.URL)
		}
	}

	// Limit maximum width
	if nameWidth > maxNameWidth {
		nameWidth = maxNameWidth
	}
	if urlWidth > maxURLWidth {
		urlWidth = maxURLWidth
	}

	// Print header
	header := fmt.Sprintf("%-*s  %-*s  %-10s  %s\n",
		nameWidth, "NAME",
		urlWidth, "URL",
		"STATUS",
		"LATENCY")
	_, err := fmt.Fprint(f.writer, header)
	if err != nil {
		return err
	}

	// Print each row
	for _, result := range batch.Results {
		if err := f.formatRow(result, nameWidth, urlWidth); err != nil {
			return err
		}
	}

	// Print summary
	fmt.Fprintln(f.writer)
	summaryColor := colorGreen
	if batch.Summary.Unhealthy > 0 {
		summaryColor = colorYellow
	}
	if batch.Summary.Healthy == 0 && batch.Summary.Total > 0 {
		summaryColor = colorRed
	}

	summary := fmt.Sprintf("Summary: %d/%d healthy", batch.Summary.Healthy, batch.Summary.Total)
	_, err = fmt.Fprintln(f.writer, f.colorize(summary, summaryColor))
	return err
}

// formatRow formats a single row output
func (f *TableFormatter) formatRow(result checker.Result, nameWidth, urlWidth int) error {
	// Truncate long names and URLs
	name := truncate(result.Name, nameWidth)
	url := truncate(result.URL, urlWidth)

	var status string
	var latency string

	if result.Healthy {
		status = f.colorize("✓", colorGreen)
		if result.StatusCode != nil {
			status += fmt.Sprintf(" %d", *result.StatusCode)
		}
	} else {
		status = f.colorize("✗", colorRed)
		if result.StatusCode != nil {
			status += fmt.Sprintf(" %d", *result.StatusCode)
		} else if result.Error != nil {
			status += " " + f.getShortError(result.Error)
		}
	}

	if result.Healthy || result.StatusCode != nil {
		latency = formatLatency(result.Latency)
	} else {
		latency = "--"
	}

	_, err := fmt.Fprintf(f.writer, "%-*s  %-*s  %-10s  %s\n",
		nameWidth, name,
		urlWidth, url,
		status,
		latency)
	return err
}

// colorize adds color
func (f *TableFormatter) colorize(text, color string) string {
	if f.noColor {
		return text
	}
	return color + text + colorReset
}

// getShortError gets short error description
func (f *TableFormatter) getShortError(err error) string {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "connection refused"):
		return "refused"
	case strings.Contains(errStr, "DNS"):
		return "dns error"
	case strings.Contains(errStr, "certificate"):
		return "ssl error"
	default:
		// Extract first part
		if idx := strings.Index(errStr, ":"); idx > 0 && idx < 20 {
			return errStr[:idx]
		}
		if len(errStr) > 15 {
			return errStr[:15] + "..."
		}
		return errStr
	}
}

// formatLatency formats latency time
func formatLatency(d time.Duration) string {
	ms := d.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000)
}

// truncate truncates a string
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
