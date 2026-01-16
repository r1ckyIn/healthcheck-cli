// 核心检查逻辑
// 实现 HTTP 端点健康检查的核心功能
package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Checker 是健康检查器
type Checker struct {
	client      *http.Client
	concurrency int
}

// Option 是 Checker 的配置选项
type Option func(*Checker)

// WithConcurrency 设置最大并发数
func WithConcurrency(n int) Option {
	return func(c *Checker) {
		if n > 0 {
			c.concurrency = n
		}
	}
}

// New 创建一个新的健康检查器
func New(opts ...Option) *Checker {
	c := &Checker{
		concurrency: 10,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Check 检查单个端点的健康状态
func (c *Checker) Check(ep Endpoint) Result {
	result := Result{
		Name: ep.Name,
		URL:  ep.URL,
	}

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), ep.Timeout)
	defer cancel()

	// 创建 HTTP 客户端
	client := c.createClient(ep)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep.URL, nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		return result
	}

	// 添加自定义 Headers
	for key, value := range ep.Headers {
		req.Header.Set(key, value)
	}

	// 设置 User-Agent
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "healthcheck-cli/1.0")
	}

	// 执行请求并计时
	start := time.Now()
	resp, err := client.Do(req)
	result.Latency = time.Since(start)

	if err != nil {
		result.Error = c.categorizeError(err)
		return result
	}
	defer resp.Body.Close()

	// 记录状态码
	result.StatusCode = &resp.StatusCode

	// 检查状态码是否符合期望
	if resp.StatusCode == ep.ExpectedStatus {
		result.Healthy = true
	} else {
		result.Error = fmt.Errorf("unexpected status code: got %d, expected %d", resp.StatusCode, ep.ExpectedStatus)
	}

	return result
}

// CheckWithRetry 带重试的健康检查
func (c *Checker) CheckWithRetry(ep Endpoint) Result {
	var result Result

	for i := 0; i <= ep.Retries; i++ {
		result = c.Check(ep)
		if result.Healthy {
			return result
		}

		// 如果还有重试机会，等待一小段时间
		if i < ep.Retries {
			time.Sleep(500 * time.Millisecond)
		}
	}

	return result
}

// CheckAll 并发检查多个端点
func (c *Checker) CheckAll(endpoints []Endpoint) BatchResult {
	startTime := time.Now()
	batchResult := BatchResult{
		Timestamp: startTime,
		Results:   make([]Result, len(endpoints)),
	}

	// 使用 channel 控制并发
	sem := make(chan struct{}, c.concurrency)
	var wg sync.WaitGroup

	// 使用 indexed results 保持顺序
	for i, ep := range endpoints {
		wg.Add(1)
		go func(idx int, endpoint Endpoint) {
			defer wg.Done()

			// 获取信号量
			sem <- struct{}{}
			defer func() { <-sem }()

			// 执行检查（带重试）
			batchResult.Results[idx] = c.CheckWithRetry(endpoint)
		}(i, ep)
	}

	wg.Wait()

	// 计算汇总信息
	batchResult.Summary = c.calculateSummary(batchResult.Results, time.Since(startTime))

	return batchResult
}

// createClient 根据端点配置创建 HTTP 客户端
func (c *Checker) createClient(ep Endpoint) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: ep.Insecure,
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   ep.Timeout,
	}

	// 配置重定向处理
	if !ep.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}

// categorizeError 分类错误类型
func (c *Checker) categorizeError(err error) error {
	errStr := err.Error()

	// 根据错误信息分类
	switch {
	case strings.Contains(errStr, "no such host"):
		return fmt.Errorf("DNS resolution failed: %w", err)
	case strings.Contains(errStr, "connection refused"):
		return fmt.Errorf("connection refused: %w", err)
	case strings.Contains(errStr, "context deadline exceeded"):
		return fmt.Errorf("connection timeout: %w", err)
	case strings.Contains(errStr, "timeout"):
		return fmt.Errorf("request timeout: %w", err)
	case strings.Contains(errStr, "certificate"):
		return fmt.Errorf("SSL certificate error: %w", err)
	default:
		return err
	}
}

// calculateSummary 计算汇总信息
func (c *Checker) calculateSummary(results []Result, duration time.Duration) Summary {
	summary := Summary{
		Total:    len(results),
		Duration: duration,
	}

	for _, r := range results {
		if r.Healthy {
			summary.Healthy++
		} else {
			summary.Unhealthy++
		}
	}

	return summary
}
