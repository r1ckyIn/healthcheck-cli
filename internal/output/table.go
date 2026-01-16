// Table 格式输出
// 实现人类可读的表格格式输出
package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/r1ckyIn/healthcheck-cli/internal/checker"
)

// ANSI 颜色代码
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// TableFormatter 实现表格格式输出
type TableFormatter struct {
	writer  io.Writer
	noColor bool
}

// NewTableFormatter 创建表格格式化器
func NewTableFormatter(w io.Writer, noColor bool) *TableFormatter {
	return &TableFormatter{
		writer:  w,
		noColor: noColor,
	}
}

// FormatSingle 格式化单个检查结果
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
			// 提取简短错误信息
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

// FormatBatch 格式化批量检查结果
func (f *TableFormatter) FormatBatch(batch checker.BatchResult) error {
	// 计算列宽
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

	// 限制最大宽度
	if nameWidth > 30 {
		nameWidth = 30
	}
	if urlWidth > 50 {
		urlWidth = 50
	}

	// 打印表头
	header := fmt.Sprintf("%-*s  %-*s  %-10s  %s\n",
		nameWidth, "NAME",
		urlWidth, "URL",
		"STATUS",
		"LATENCY")
	_, err := fmt.Fprint(f.writer, header)
	if err != nil {
		return err
	}

	// 打印每一行
	for _, result := range batch.Results {
		if err := f.formatRow(result, nameWidth, urlWidth); err != nil {
			return err
		}
	}

	// 打印汇总
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

// formatRow 格式化单行输出
func (f *TableFormatter) formatRow(result checker.Result, nameWidth, urlWidth int) error {
	// 截断过长的名称和 URL
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

// colorize 添加颜色
func (f *TableFormatter) colorize(text, color string) string {
	if f.noColor {
		return text
	}
	return color + text + colorReset
}

// getShortError 获取简短的错误描述
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
		// 截取第一部分
		if idx := strings.Index(errStr, ":"); idx > 0 && idx < 20 {
			return errStr[:idx]
		}
		if len(errStr) > 15 {
			return errStr[:15] + "..."
		}
		return errStr
	}
}

// formatLatency 格式化延迟时间
func formatLatency(d time.Duration) string {
	ms := d.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000)
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
