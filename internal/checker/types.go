// Data type definitions / 数据类型定义
// Define core data structures for health check / 定义健康检查相关的核心数据结构
package checker

import (
	"time"
)

// Endpoint represents an endpoint to check / 表示一个需要检查的端点
type Endpoint struct {
	Name            string            // Endpoint name for display / 端点名称（用于显示）
	URL             string            // URL to check / 检查的 URL
	Timeout         time.Duration     // Request timeout / 请求超时时间
	Retries         int               // Retry count on failure / 失败重试次数
	ExpectedStatus  int               // Expected HTTP status code / 期望的 HTTP 状态码
	FollowRedirects bool              // Whether to follow redirects / 是否跟随重定向
	Insecure        bool              // Whether to skip SSL verification / 是否跳过 SSL 验证
	Headers         map[string]string // Custom request headers / 自定义请求头
}

// Result represents health check result / 表示健康检查的结果
type Result struct {
	Name       string        // Endpoint name / 端点名称
	URL        string        // Checked URL / 检查的 URL
	Healthy    bool          // Whether healthy / 是否健康
	StatusCode *int          // HTTP status code (nil if connection failed) / HTTP 状态码（nil 表示无法连接）
	Latency    time.Duration // Response latency / 响应延迟
	Error      error         // Error message / 错误信息
}

// Summary represents batch check summary / 表示批量检查的汇总信息
type Summary struct {
	Total     int           // Total endpoints / 端点总数
	Healthy   int           // Healthy count / 健康数量
	Unhealthy int           // Unhealthy count / 不健康数量
	Duration  time.Duration // Total duration / 总耗时
}

// BatchResult represents complete batch check result / 表示批量检查的完整结果
type BatchResult struct {
	Timestamp time.Time // Check start time / 检查开始时间
	Summary   Summary   // Summary info / 汇总信息
	Results   []Result  // Detailed results / 详细结果列表
}

// DefaultEndpoint creates an endpoint with default config / 创建一个带有默认配置的端点
func DefaultEndpoint(url string) Endpoint {
	return Endpoint{
		Name:            url,
		URL:             url,
		Timeout:         5 * time.Second,
		Retries:         0,
		ExpectedStatus:  200,
		FollowRedirects: true,
		Insecure:        false,
		Headers:         make(map[string]string),
	}
}
