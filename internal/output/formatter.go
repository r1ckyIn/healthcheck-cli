// 输出格式化接口定义
// 定义输出格式化器的抽象接口
package output

import (
	"io"

	"github.com/r1ckyIn/healthcheck-cli/internal/checker"
)

// Formatter 是输出格式化器的接口
type Formatter interface {
	// FormatSingle 格式化单个检查结果
	FormatSingle(result checker.Result) error

	// FormatBatch 格式化批量检查结果
	FormatBatch(result checker.BatchResult) error
}

// OutputFormat 输出格式类型
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
)

// NewFormatter 根据格式类型创建对应的格式化器
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
