// Core health check logic / 核心检查逻辑
// Implement HTTP endpoint health check functionality / 实现 HTTP 端点健康检查的核心功能
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

// Checker is the health checker / 健康检查器
type Checker struct {
	// Cached clients for different configurations / 不同配置的缓存客户端
	// Key format: "secure-follow", "secure-nofollow", "insecure-follow", "insecure-nofollow"
	clients     map[string]*http.Client
	clientMu    sync.RWMutex
	concurrency int
}

// Option is Checker configuration option / Checker 的配置选项
type Option func(*Checker)

// WithConcurrency sets maximum concurrency / 设置最大并发数
func WithConcurrency(n int) Option {
	return func(c *Checker) {
		if n > 0 {
			c.concurrency = n
		}
	}
}

// New creates a new health checker / 创建一个新的健康检查器
func New(opts ...Option) *Checker {
	c := &Checker{
		clients:     make(map[string]*http.Client),
		concurrency: 10,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// getClientKey generates cache key for client based on endpoint config / 根据端点配置生成客户端缓存键
func getClientKey(insecure, followRedirects bool) string {
	security := "secure"
	if insecure {
		security = "insecure"
	}
	redirect := "follow"
	if !followRedirects {
		redirect = "nofollow"
	}
	return security + "-" + redirect
}

// getClient returns appropriate HTTP client based on endpoint config / 根据端点配置返回合适的 HTTP 客户端
func (c *Checker) getClient(ep Endpoint) *http.Client {
	key := getClientKey(ep.Insecure, ep.FollowRedirects)

	// Try to get existing client / 尝试获取已存在的客户端
	c.clientMu.RLock()
	if client, ok := c.clients[key]; ok {
		c.clientMu.RUnlock()
		return client
	}
	c.clientMu.RUnlock()

	// Create new client / 创建新客户端
	c.clientMu.Lock()
	defer c.clientMu.Unlock()

	// Double check after acquiring write lock / 获取写锁后再次检查
	if client, ok := c.clients[key]; ok {
		return client
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: ep.Insecure,
			},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Configure redirect handling / 配置重定向处理
	if !ep.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	c.clients[key] = client
	return client
}

// Check checks single endpoint health status / 检查单个端点的健康状态
func (c *Checker) Check(ep Endpoint) Result {
	return c.CheckWithContext(context.Background(), ep)
}

// CheckWithContext checks single endpoint with context support / 带 context 支持的单个端点检查
func (c *Checker) CheckWithContext(ctx context.Context, ep Endpoint) Result {
	result := Result{
		Name: ep.Name,
		URL:  ep.URL,
	}

	// Create context with timeout / 创建带超时的 context
	ctx, cancel := context.WithTimeout(ctx, ep.Timeout)
	defer cancel()

	// Get HTTP client / 获取 HTTP 客户端
	client := c.getClient(ep)

	// Create request / 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep.URL, nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		return result
	}

	// Add custom headers / 添加自定义 Headers
	for key, value := range ep.Headers {
		req.Header.Set(key, value)
	}

	// Set User-Agent / 设置 User-Agent
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "healthcheck-cli/1.0")
	}

	// Execute request and measure time / 执行请求并计时
	start := time.Now()
	resp, err := client.Do(req)
	result.Latency = time.Since(start)

	if err != nil {
		result.Error = c.categorizeError(err)
		return result
	}
	defer resp.Body.Close()

	// Record status code / 记录状态码
	result.StatusCode = &resp.StatusCode

	// Check if status code matches expected / 检查状态码是否符合期望
	if resp.StatusCode == ep.ExpectedStatus {
		result.Healthy = true
	} else {
		result.Error = fmt.Errorf("unexpected status code: got %d, expected %d", resp.StatusCode, ep.ExpectedStatus)
	}

	return result
}

// CheckWithRetry performs health check with retry / 带重试的健康检查
func (c *Checker) CheckWithRetry(ep Endpoint) Result {
	return c.CheckWithRetryContext(context.Background(), ep)
}

// CheckWithRetryContext performs health check with retry and context / 带重试和 context 的健康检查
func (c *Checker) CheckWithRetryContext(ctx context.Context, ep Endpoint) Result {
	var result Result

	for i := 0; i <= ep.Retries; i++ {
		// Check if context is cancelled / 检查 context 是否已取消
		select {
		case <-ctx.Done():
			result.Error = ctx.Err()
			return result
		default:
		}

		result = c.CheckWithContext(ctx, ep)
		if result.Healthy {
			return result
		}

		// Wait before retry if there are more attempts / 如果还有重试机会，等待一小段时间
		if i < ep.Retries {
			select {
			case <-ctx.Done():
				result.Error = ctx.Err()
				return result
			case <-time.After(500 * time.Millisecond):
			}
		}
	}

	return result
}

// indexedResult holds result with its index / 带索引的结果
type indexedResult struct {
	idx    int
	result Result
}

// CheckAll concurrently checks multiple endpoints / 并发检查多个端点
func (c *Checker) CheckAll(endpoints []Endpoint) BatchResult {
	return c.CheckAllWithContext(context.Background(), endpoints)
}

// CheckAllWithContext concurrently checks multiple endpoints with context / 带 context 的并发检查多个端点
func (c *Checker) CheckAllWithContext(ctx context.Context, endpoints []Endpoint) BatchResult {
	startTime := time.Now()
	results := make([]Result, len(endpoints))

	if len(endpoints) == 0 {
		return BatchResult{
			Timestamp: startTime,
			Results:   results,
			Summary:   c.calculateSummary(results, time.Since(startTime)),
		}
	}

	// Use channel for collecting results safely / 使用 channel 安全地收集结果
	resultChan := make(chan indexedResult, len(endpoints))
	sem := make(chan struct{}, c.concurrency)
	var wg sync.WaitGroup

	for i, ep := range endpoints {
		wg.Add(1)
		go func(idx int, endpoint Endpoint) {
			defer wg.Done()

			// Acquire semaphore / 获取信号量
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				resultChan <- indexedResult{
					idx:    idx,
					result: Result{Name: endpoint.Name, URL: endpoint.URL, Error: ctx.Err()},
				}
				return
			}

			// Execute check with retry / 执行检查（带重试）
			resultChan <- indexedResult{
				idx:    idx,
				result: c.CheckWithRetryContext(ctx, endpoint),
			}
		}(i, ep)
	}

	// Close channel when all goroutines complete / 当所有 goroutine 完成时关闭 channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results / 收集结果
	for r := range resultChan {
		results[r.idx] = r.result
	}

	return BatchResult{
		Timestamp: startTime,
		Results:   results,
		Summary:   c.calculateSummary(results, time.Since(startTime)),
	}
}

// categorizeError categorizes error type / 分类错误类型
func (c *Checker) categorizeError(err error) error {
	errStr := err.Error()

	// Categorize based on error message / 根据错误信息分类
	switch {
	case strings.Contains(errStr, "no such host"):
		return fmt.Errorf("DNS resolution failed: %w", err)
	case strings.Contains(errStr, "connection refused"):
		return fmt.Errorf("connection refused: %w", err)
	case strings.Contains(errStr, "context deadline exceeded"):
		return fmt.Errorf("connection timeout: %w", err)
	case strings.Contains(errStr, "context canceled"):
		return fmt.Errorf("request canceled: %w", err)
	case strings.Contains(errStr, "timeout"):
		return fmt.Errorf("request timeout: %w", err)
	case strings.Contains(errStr, "certificate"):
		return fmt.Errorf("SSL certificate error: %w", err)
	default:
		return err
	}
}

// calculateSummary calculates summary info / 计算汇总信息
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
