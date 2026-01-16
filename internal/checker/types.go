// 数据类型定义
// 定义健康检查相关的核心数据结构
package checker

import (
	"time"
)

// Endpoint 表示一个需要检查的端点
type Endpoint struct {
	Name            string            // 端点名称（用于显示）
	URL             string            // 检查的 URL
	Timeout         time.Duration     // 请求超时时间
	Retries         int               // 失败重试次数
	ExpectedStatus  int               // 期望的 HTTP 状态码
	FollowRedirects bool              // 是否跟随重定向
	Insecure        bool              // 是否跳过 SSL 验证
	Headers         map[string]string // 自定义请求头
}

// Result 表示健康检查的结果
type Result struct {
	Name       string        // 端点名称
	URL        string        // 检查的 URL
	Healthy    bool          // 是否健康
	StatusCode *int          // HTTP 状态码（nil 表示无法连接）
	Latency    time.Duration // 响应延迟
	Error      error         // 错误信息
}

// Summary 表示批量检查的汇总信息
type Summary struct {
	Total     int           // 端点总数
	Healthy   int           // 健康数量
	Unhealthy int           // 不健康数量
	Duration  time.Duration // 总耗时
}

// BatchResult 表示批量检查的完整结果
type BatchResult struct {
	Timestamp time.Time // 检查开始时间
	Summary   Summary   // 汇总信息
	Results   []Result  // 详细结果列表
}

// DefaultEndpoint 创建一个带有默认配置的端点
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
