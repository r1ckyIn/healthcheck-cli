// Output formatter unit tests / 输出格式化单元测试
// Test Table and JSON formatter functionality / 测试 Table 和 JSON 格式化器功能
package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/r1ckyIn/healthcheck-cli/internal/checker"
)

// TestNewFormatter_Table 测试创建 Table 格式化器
func TestNewFormatter_Table(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(FormatTable, &buf, false)

	if _, ok := f.(*TableFormatter); !ok {
		t.Error("NewFormatter(FormatTable) did not return *TableFormatter")
	}
}

// TestNewFormatter_JSON 测试创建 JSON 格式化器
func TestNewFormatter_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(FormatJSON, &buf, false)

	if _, ok := f.(*JSONFormatter); !ok {
		t.Error("NewFormatter(FormatJSON) did not return *JSONFormatter")
	}
}

// TestNewFormatter_Default 测试默认格式化器
func TestNewFormatter_Default(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("unknown", &buf, false)

	if _, ok := f.(*TableFormatter); !ok {
		t.Error("NewFormatter with unknown format should default to TableFormatter")
	}
}

// TestTableFormatter_FormatSingle_Healthy 测试 Table 格式健康结果
func TestTableFormatter_FormatSingle_Healthy(t *testing.T) {
	var buf bytes.Buffer
	f := NewTableFormatter(&buf, true) // 禁用颜色便于测试

	statusCode := 200
	result := checker.Result{
		Name:       "Test API",
		URL:        "https://api.example.com",
		Healthy:    true,
		StatusCode: &statusCode,
		Latency:    45 * time.Millisecond,
	}

	err := f.FormatSingle(result)
	if err != nil {
		t.Fatalf("FormatSingle() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Error("output should contain '✓' for healthy result")
	}
	if !strings.Contains(output, "200") {
		t.Error("output should contain status code '200'")
	}
	if !strings.Contains(output, "45ms") {
		t.Error("output should contain latency '45ms'")
	}
}

// TestTableFormatter_FormatSingle_Unhealthy 测试 Table 格式不健康结果
func TestTableFormatter_FormatSingle_Unhealthy(t *testing.T) {
	var buf bytes.Buffer
	f := NewTableFormatter(&buf, true)

	statusCode := 500
	result := checker.Result{
		Name:       "Test API",
		URL:        "https://api.example.com",
		Healthy:    false,
		StatusCode: &statusCode,
		Latency:    100 * time.Millisecond,
		Error:      errors.New("unexpected status code"),
	}

	err := f.FormatSingle(result)
	if err != nil {
		t.Fatalf("FormatSingle() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✗") {
		t.Error("output should contain '✗' for unhealthy result")
	}
	if !strings.Contains(output, "500") {
		t.Error("output should contain status code '500'")
	}
}

// TestTableFormatter_FormatSingle_Timeout 测试 Table 格式超时结果
func TestTableFormatter_FormatSingle_Timeout(t *testing.T) {
	var buf bytes.Buffer
	f := NewTableFormatter(&buf, true)

	result := checker.Result{
		Name:    "Slow API",
		URL:     "https://slow.example.com",
		Healthy: false,
		Error:   errors.New("connection timeout"),
	}

	err := f.FormatSingle(result)
	if err != nil {
		t.Fatalf("FormatSingle() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✗") {
		t.Error("output should contain '✗'")
	}
	if !strings.Contains(output, "timeout") {
		t.Error("output should contain 'timeout'")
	}
	if !strings.Contains(output, "--") {
		t.Error("output should contain '--' for no latency")
	}
}

// TestTableFormatter_FormatBatch 测试 Table 格式批量结果
func TestTableFormatter_FormatBatch(t *testing.T) {
	var buf bytes.Buffer
	f := NewTableFormatter(&buf, true)

	statusCode200 := 200
	statusCode500 := 500
	batch := checker.BatchResult{
		Timestamp: time.Now(),
		Summary: checker.Summary{
			Total:     3,
			Healthy:   2,
			Unhealthy: 1,
			Duration:  500 * time.Millisecond,
		},
		Results: []checker.Result{
			{Name: "API 1", URL: "https://api1.com", Healthy: true, StatusCode: &statusCode200, Latency: 50 * time.Millisecond},
			{Name: "API 2", URL: "https://api2.com", Healthy: false, StatusCode: &statusCode500, Latency: 100 * time.Millisecond},
			{Name: "API 3", URL: "https://api3.com", Healthy: true, StatusCode: &statusCode200, Latency: 75 * time.Millisecond},
		},
	}

	err := f.FormatBatch(batch)
	if err != nil {
		t.Fatalf("FormatBatch() error = %v", err)
	}

	output := buf.String()

	// 验证表头
	if !strings.Contains(output, "NAME") {
		t.Error("output should contain 'NAME' header")
	}
	if !strings.Contains(output, "URL") {
		t.Error("output should contain 'URL' header")
	}
	if !strings.Contains(output, "STATUS") {
		t.Error("output should contain 'STATUS' header")
	}
	if !strings.Contains(output, "LATENCY") {
		t.Error("output should contain 'LATENCY' header")
	}

	// 验证汇总
	if !strings.Contains(output, "2/3 healthy") {
		t.Error("output should contain '2/3 healthy' summary")
	}
}

// TestTableFormatter_NoColor 测试禁用颜色
func TestTableFormatter_NoColor(t *testing.T) {
	var buf bytes.Buffer
	f := NewTableFormatter(&buf, true) // noColor = true

	statusCode := 200
	result := checker.Result{
		Name:       "Test",
		URL:        "https://test.com",
		Healthy:    true,
		StatusCode: &statusCode,
		Latency:    50 * time.Millisecond,
	}

	f.FormatSingle(result)
	output := buf.String()

	// 不应包含 ANSI 转义码
	if strings.Contains(output, "\033[") {
		t.Error("output should not contain ANSI escape codes when noColor=true")
	}
}

// TestTableFormatter_WithColor 测试启用颜色
func TestTableFormatter_WithColor(t *testing.T) {
	var buf bytes.Buffer
	f := NewTableFormatter(&buf, false) // noColor = false

	statusCode := 200
	result := checker.Result{
		Name:       "Test",
		URL:        "https://test.com",
		Healthy:    true,
		StatusCode: &statusCode,
		Latency:    50 * time.Millisecond,
	}

	f.FormatSingle(result)
	output := buf.String()

	// 应包含绿色 ANSI 转义码
	if !strings.Contains(output, colorGreen) {
		t.Error("output should contain green color code when noColor=false")
	}
}

// TestJSONFormatter_FormatSingle_Healthy 测试 JSON 格式健康结果
func TestJSONFormatter_FormatSingle_Healthy(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter(&buf)

	statusCode := 200
	result := checker.Result{
		Name:       "Test API",
		URL:        "https://api.example.com",
		Healthy:    true,
		StatusCode: &statusCode,
		Latency:    45 * time.Millisecond,
	}

	err := f.FormatSingle(result)
	if err != nil {
		t.Fatalf("FormatSingle() error = %v", err)
	}

	// 解析 JSON
	var output singleResultJSON
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if output.URL != "https://api.example.com" {
		t.Errorf("URL = %q, want %q", output.URL, "https://api.example.com")
	}
	if !output.Healthy {
		t.Error("Healthy = false, want true")
	}
	if output.StatusCode == nil || *output.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", output.StatusCode)
	}
	if output.LatencyMs == nil || *output.LatencyMs != 45 {
		t.Errorf("LatencyMs = %v, want 45", output.LatencyMs)
	}
	if output.Error != nil {
		t.Errorf("Error = %v, want nil", output.Error)
	}
}

// TestJSONFormatter_FormatSingle_Unhealthy 测试 JSON 格式不健康结果
func TestJSONFormatter_FormatSingle_Unhealthy(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter(&buf)

	result := checker.Result{
		Name:    "Test API",
		URL:     "https://api.example.com",
		Healthy: false,
		Error:   errors.New("connection timeout"),
	}

	err := f.FormatSingle(result)
	if err != nil {
		t.Fatalf("FormatSingle() error = %v", err)
	}

	var output singleResultJSON
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if output.Healthy {
		t.Error("Healthy = true, want false")
	}
	if output.Error == nil || *output.Error != "connection timeout" {
		t.Errorf("Error = %v, want 'connection timeout'", output.Error)
	}
	if output.LatencyMs != nil {
		t.Errorf("LatencyMs = %v, want nil for failed request", output.LatencyMs)
	}
}

// TestJSONFormatter_FormatBatch 测试 JSON 格式批量结果
func TestJSONFormatter_FormatBatch(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter(&buf)

	statusCode200 := 200
	batch := checker.BatchResult{
		Timestamp: time.Date(2026, 1, 17, 10, 30, 0, 0, time.UTC),
		Summary: checker.Summary{
			Total:     2,
			Healthy:   1,
			Unhealthy: 1,
			Duration:  250 * time.Millisecond,
		},
		Results: []checker.Result{
			{Name: "API 1", URL: "https://api1.com", Healthy: true, StatusCode: &statusCode200, Latency: 50 * time.Millisecond},
			{Name: "API 2", URL: "https://api2.com", Healthy: false, Error: errors.New("timeout")},
		},
	}

	err := f.FormatBatch(batch)
	if err != nil {
		t.Fatalf("FormatBatch() error = %v", err)
	}

	var output batchResultJSON
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	// 验证时间戳
	if output.Timestamp != "2026-01-17T10:30:00Z" {
		t.Errorf("Timestamp = %q, want %q", output.Timestamp, "2026-01-17T10:30:00Z")
	}

	// 验证汇总
	if output.Summary.Total != 2 {
		t.Errorf("Summary.Total = %d, want 2", output.Summary.Total)
	}
	if output.Summary.Healthy != 1 {
		t.Errorf("Summary.Healthy = %d, want 1", output.Summary.Healthy)
	}
	if output.Summary.Unhealthy != 1 {
		t.Errorf("Summary.Unhealthy = %d, want 1", output.Summary.Unhealthy)
	}
	if output.DurationMs != 250 {
		t.Errorf("DurationMs = %d, want 250", output.DurationMs)
	}

	// 验证结果数量
	if len(output.Results) != 2 {
		t.Fatalf("len(Results) = %d, want 2", len(output.Results))
	}

	// 验证第一个结果
	if output.Results[0].Name != "API 1" {
		t.Errorf("Results[0].Name = %q, want %q", output.Results[0].Name, "API 1")
	}
	if !output.Results[0].Healthy {
		t.Error("Results[0].Healthy = false, want true")
	}

	// 验证第二个结果
	if output.Results[1].Name != "API 2" {
		t.Errorf("Results[1].Name = %q, want %q", output.Results[1].Name, "API 2")
	}
	if output.Results[1].Healthy {
		t.Error("Results[1].Healthy = true, want false")
	}
}

// TestFormatLatency 测试延迟格式化
func TestFormatLatency(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"milliseconds", 45 * time.Millisecond, "45ms"},
		{"zero", 0, "0ms"},
		{"one second", 1000 * time.Millisecond, "1.0s"},
		{"seconds", 2500 * time.Millisecond, "2.5s"},
		{"sub millisecond", 500 * time.Microsecond, "0ms"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatLatency(tt.duration)
			if result != tt.expected {
				t.Errorf("formatLatency(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

// TestTruncate 测试字符串截断
func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"no truncation needed", "short", 10, "short"},
		{"exact length", "exact", 5, "exact"},
		{"needs truncation", "this is a long string", 10, "this is..."},
		{"very short max", "hello", 3, "hel"},
		{"empty string", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestGetShortError 测试错误信息简化
func TestGetShortError(t *testing.T) {
	f := &TableFormatter{noColor: true}

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"timeout error", errors.New("connection timeout"), "timeout"},
		{"connection refused", errors.New("dial: connection refused"), "refused"},
		{"DNS error", errors.New("DNS lookup failed"), "dns error"},
		{"SSL error", errors.New("x509: certificate verify failed"), "ssl error"},
		{"short error", errors.New("fail"), "fail"},
		{"long error", errors.New("this is a very long error message that should be truncated"), "this is a very ..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.getShortError(tt.err)
			if result != tt.expected {
				t.Errorf("getShortError(%q) = %q, want %q", tt.err, result, tt.expected)
			}
		})
	}
}

// TestTableFormatter_FormatBatch_AllHealthy 测试全部健康时的汇总颜色
func TestTableFormatter_FormatBatch_AllHealthy(t *testing.T) {
	var buf bytes.Buffer
	f := NewTableFormatter(&buf, false) // 启用颜色

	statusCode := 200
	batch := checker.BatchResult{
		Timestamp: time.Now(),
		Summary: checker.Summary{
			Total:     2,
			Healthy:   2,
			Unhealthy: 0,
		},
		Results: []checker.Result{
			{Name: "API 1", URL: "https://api1.com", Healthy: true, StatusCode: &statusCode},
			{Name: "API 2", URL: "https://api2.com", Healthy: true, StatusCode: &statusCode},
		},
	}

	f.FormatBatch(batch)
	output := buf.String()

	// 全部健康应使用绿色
	if !strings.Contains(output, colorGreen) {
		t.Error("all healthy summary should use green color")
	}
}

// TestTableFormatter_FormatBatch_AllUnhealthy 测试全部不健康时的汇总颜色
func TestTableFormatter_FormatBatch_AllUnhealthy(t *testing.T) {
	var buf bytes.Buffer
	f := NewTableFormatter(&buf, false)

	batch := checker.BatchResult{
		Timestamp: time.Now(),
		Summary: checker.Summary{
			Total:     2,
			Healthy:   0,
			Unhealthy: 2,
		},
		Results: []checker.Result{
			{Name: "API 1", URL: "https://api1.com", Healthy: false, Error: errors.New("error")},
			{Name: "API 2", URL: "https://api2.com", Healthy: false, Error: errors.New("error")},
		},
	}

	f.FormatBatch(batch)
	output := buf.String()

	// 全部不健康应使用红色
	if !strings.Contains(output, colorRed) {
		t.Error("all unhealthy summary should use red color")
	}
}

// TestTableFormatter_FormatBatch_PartialHealthy 测试部分健康时的汇总颜色
func TestTableFormatter_FormatBatch_PartialHealthy(t *testing.T) {
	var buf bytes.Buffer
	f := NewTableFormatter(&buf, false)

	statusCode := 200
	batch := checker.BatchResult{
		Timestamp: time.Now(),
		Summary: checker.Summary{
			Total:     2,
			Healthy:   1,
			Unhealthy: 1,
		},
		Results: []checker.Result{
			{Name: "API 1", URL: "https://api1.com", Healthy: true, StatusCode: &statusCode},
			{Name: "API 2", URL: "https://api2.com", Healthy: false, Error: errors.New("error")},
		},
	}

	f.FormatBatch(batch)
	output := buf.String()

	// 部分健康应使用黄色
	if !strings.Contains(output, colorYellow) {
		t.Error("partial healthy summary should use yellow color")
	}
}

// TestJSONFormatter_FormatBatch_Empty 测试空结果的 JSON 输出
func TestJSONFormatter_FormatBatch_Empty(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter(&buf)

	batch := checker.BatchResult{
		Timestamp: time.Now(),
		Summary: checker.Summary{
			Total:     0,
			Healthy:   0,
			Unhealthy: 0,
		},
		Results: []checker.Result{},
	}

	err := f.FormatBatch(batch)
	if err != nil {
		t.Fatalf("FormatBatch() error = %v", err)
	}

	var output batchResultJSON
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if len(output.Results) != 0 {
		t.Errorf("len(Results) = %d, want 0", len(output.Results))
	}
	if output.Summary.Total != 0 {
		t.Errorf("Summary.Total = %d, want 0", output.Summary.Total)
	}
}

// TestTableFormatter_Colorize 测试颜色函数
func TestTableFormatter_Colorize(t *testing.T) {
	tests := []struct {
		name     string
		noColor  bool
		text     string
		color    string
		expected string
	}{
		{"with color", false, "test", colorGreen, colorGreen + "test" + colorReset},
		{"no color", true, "test", colorGreen, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TableFormatter{noColor: tt.noColor}
			result := f.colorize(tt.text, tt.color)
			if result != tt.expected {
				t.Errorf("colorize() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestOutputFormat_Constants 测试输出格式常量
func TestOutputFormat_Constants(t *testing.T) {
	if FormatTable != "table" {
		t.Errorf("FormatTable = %q, want %q", FormatTable, "table")
	}
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want %q", FormatJSON, "json")
	}
}
