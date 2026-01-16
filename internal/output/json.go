// JSON 格式输出
// 实现机器可读的 JSON 格式输出
package output

import (
	"encoding/json"
	"io"

	"github.com/r1ckyIn/healthcheck-cli/internal/checker"
)

// JSONFormatter 实现 JSON 格式输出
type JSONFormatter struct {
	writer io.Writer
}

// NewJSONFormatter 创建 JSON 格式化器
func NewJSONFormatter(w io.Writer) *JSONFormatter {
	return &JSONFormatter{
		writer: w,
	}
}

// singleResultJSON 单个结果的 JSON 结构
type singleResultJSON struct {
	URL        string  `json:"url"`
	Healthy    bool    `json:"healthy"`
	StatusCode *int    `json:"status_code"`
	LatencyMs  *int64  `json:"latency_ms"`
	Error      *string `json:"error"`
}

// batchResultJSON 批量结果的 JSON 结构
type batchResultJSON struct {
	Timestamp  string            `json:"timestamp"`
	DurationMs int64             `json:"duration_ms"`
	Summary    summaryJSON       `json:"summary"`
	Results    []resultItemJSON  `json:"results"`
}

// summaryJSON 汇总信息的 JSON 结构
type summaryJSON struct {
	Total     int `json:"total"`
	Healthy   int `json:"healthy"`
	Unhealthy int `json:"unhealthy"`
}

// resultItemJSON 结果项的 JSON 结构
type resultItemJSON struct {
	Name       string  `json:"name"`
	URL        string  `json:"url"`
	Healthy    bool    `json:"healthy"`
	StatusCode *int    `json:"status_code"`
	LatencyMs  *int64  `json:"latency_ms"`
	Error      *string `json:"error"`
}

// FormatSingle 格式化单个检查结果
func (f *JSONFormatter) FormatSingle(result checker.Result) error {
	output := singleResultJSON{
		URL:        result.URL,
		Healthy:    result.Healthy,
		StatusCode: result.StatusCode,
	}

	// 计算延迟（毫秒）
	if result.Healthy || result.StatusCode != nil {
		latencyMs := result.Latency.Milliseconds()
		output.LatencyMs = &latencyMs
	}

	// 错误信息
	if result.Error != nil {
		errStr := result.Error.Error()
		output.Error = &errStr
	}

	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// FormatBatch 格式化批量检查结果
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

	// 转换每个结果
	for i, result := range batch.Results {
		item := resultItemJSON{
			Name:       result.Name,
			URL:        result.URL,
			Healthy:    result.Healthy,
			StatusCode: result.StatusCode,
		}

		// 延迟时间
		if result.Healthy || result.StatusCode != nil {
			latencyMs := result.Latency.Milliseconds()
			item.LatencyMs = &latencyMs
		}

		// 错误信息
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
