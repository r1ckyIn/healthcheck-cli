// Checker unit tests / 检查器单元测试
// Test core health check functionality / 测试健康检查核心功能
package checker

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestDefaultEndpoint 测试默认端点配置
func TestDefaultEndpoint(t *testing.T) {
	url := "https://example.com/health"
	ep := DefaultEndpoint(url)

	if ep.URL != url {
		t.Errorf("URL = %q, want %q", ep.URL, url)
	}
	if ep.Name != url {
		t.Errorf("Name = %q, want %q", ep.Name, url)
	}
	if ep.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want %v", ep.Timeout, 5*time.Second)
	}
	if ep.Retries != 0 {
		t.Errorf("Retries = %d, want 0", ep.Retries)
	}
	if ep.ExpectedStatus != 200 {
		t.Errorf("ExpectedStatus = %d, want 200", ep.ExpectedStatus)
	}
	if !ep.FollowRedirects {
		t.Error("FollowRedirects = false, want true")
	}
	if ep.Insecure {
		t.Error("Insecure = true, want false")
	}
	if ep.Headers == nil {
		t.Error("Headers is nil, want empty map")
	}
}

// TestNew 测试检查器创建
func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.concurrency != 10 {
		t.Errorf("concurrency = %d, want 10", c.concurrency)
	}
}

// TestWithConcurrency 测试并发数配置选项
func TestWithConcurrency(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{"valid concurrency", 5, 5},
		{"zero concurrency", 0, 10},      // 无效值保持默认
		{"negative concurrency", -1, 10}, // 无效值保持默认
		{"large concurrency", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(WithConcurrency(tt.value))
			if c.concurrency != tt.expected {
				t.Errorf("concurrency = %d, want %d", c.concurrency, tt.expected)
			}
		})
	}
}

// TestCheck_Success 测试成功的健康检查
func TestCheck_Success(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	c := New()
	ep := Endpoint{
		Name:           "test-server",
		URL:            server.URL,
		Timeout:        5 * time.Second,
		ExpectedStatus: 200,
	}

	result := c.Check(ep)

	if !result.Healthy {
		t.Error("Healthy = false, want true")
	}
	if result.StatusCode == nil || *result.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", result.StatusCode)
	}
	if result.Error != nil {
		t.Errorf("Error = %v, want nil", result.Error)
	}
	if result.Latency <= 0 {
		t.Errorf("Latency = %v, want > 0", result.Latency)
	}
}

// TestCheck_UnexpectedStatus 测试状态码不匹配
func TestCheck_UnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := New()
	ep := Endpoint{
		Name:           "test-server",
		URL:            server.URL,
		Timeout:        5 * time.Second,
		ExpectedStatus: 200,
	}

	result := c.Check(ep)

	if result.Healthy {
		t.Error("Healthy = true, want false")
	}
	if result.StatusCode == nil || *result.StatusCode != 500 {
		t.Errorf("StatusCode = %v, want 500", result.StatusCode)
	}
	if result.Error == nil {
		t.Error("Error = nil, want error")
	}
}

// TestCheck_CustomExpectedStatus 测试自定义期望状态码
func TestCheck_CustomExpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := New()
	ep := Endpoint{
		Name:           "test-server",
		URL:            server.URL,
		Timeout:        5 * time.Second,
		ExpectedStatus: 201, // 期望 201
	}

	result := c.Check(ep)

	if !result.Healthy {
		t.Error("Healthy = false, want true")
	}
}

// TestCheck_Timeout 测试请求超时
func TestCheck_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := New()
	ep := Endpoint{
		Name:           "slow-server",
		URL:            server.URL,
		Timeout:        50 * time.Millisecond, // 超时时间短于服务器响应时间
		ExpectedStatus: 200,
	}

	result := c.Check(ep)

	if result.Healthy {
		t.Error("Healthy = true, want false (timeout)")
	}
	if result.Error == nil {
		t.Error("Error = nil, want timeout error")
	}
	if !strings.Contains(result.Error.Error(), "timeout") {
		t.Errorf("Error = %q, want to contain 'timeout'", result.Error.Error())
	}
}

// TestCheck_CustomHeaders 测试自定义请求头
func TestCheck_CustomHeaders(t *testing.T) {
	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := New()
	ep := Endpoint{
		Name:           "test-server",
		URL:            server.URL,
		Timeout:        5 * time.Second,
		ExpectedStatus: 200,
		Headers: map[string]string{
			"Authorization": "Bearer test-token",
			"X-Custom":      "custom-value",
		},
	}

	c.Check(ep)

	if receivedHeaders.Get("Authorization") != "Bearer test-token" {
		t.Errorf("Authorization = %q, want %q", receivedHeaders.Get("Authorization"), "Bearer test-token")
	}
	if receivedHeaders.Get("X-Custom") != "custom-value" {
		t.Errorf("X-Custom = %q, want %q", receivedHeaders.Get("X-Custom"), "custom-value")
	}
}

// TestCheck_UserAgent 测试默认 User-Agent
func TestCheck_UserAgent(t *testing.T) {
	var receivedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := New()
	ep := Endpoint{
		Name:           "test-server",
		URL:            server.URL,
		Timeout:        5 * time.Second,
		ExpectedStatus: 200,
	}

	c.Check(ep)

	if receivedUA != "healthcheck-cli/1.0" {
		t.Errorf("User-Agent = %q, want %q", receivedUA, "healthcheck-cli/1.0")
	}
}

// TestCheck_NoFollowRedirects 测试不跟随重定向
func TestCheck_NoFollowRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirected", http.StatusMovedPermanently)
	}))
	defer server.Close()

	c := New()
	ep := Endpoint{
		Name:            "redirect-server",
		URL:             server.URL,
		Timeout:         5 * time.Second,
		ExpectedStatus:  301, // 期望重定向状态码
		FollowRedirects: false,
	}

	result := c.Check(ep)

	if !result.Healthy {
		t.Error("Healthy = false, want true")
	}
	if result.StatusCode == nil || *result.StatusCode != 301 {
		t.Errorf("StatusCode = %v, want 301", result.StatusCode)
	}
}

// TestCheckWithRetry_SuccessOnFirstTry 测试首次成功无需重试
func TestCheckWithRetry_SuccessOnFirstTry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := New()
	ep := Endpoint{
		Name:           "test-server",
		URL:            server.URL,
		Timeout:        5 * time.Second,
		ExpectedStatus: 200,
		Retries:        3,
	}

	result := c.CheckWithRetry(ep)

	if !result.Healthy {
		t.Error("Healthy = false, want true")
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (no retry needed)", callCount)
	}
}

// TestCheckWithRetry_SuccessAfterRetry 测试重试后成功
func TestCheckWithRetry_SuccessAfterRetry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	c := New()
	ep := Endpoint{
		Name:           "flaky-server",
		URL:            server.URL,
		Timeout:        5 * time.Second,
		ExpectedStatus: 200,
		Retries:        3,
	}

	result := c.CheckWithRetry(ep)

	if !result.Healthy {
		t.Error("Healthy = false, want true")
	}
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}
}

// TestCheckWithRetry_AllFailed 测试所有重试都失败
func TestCheckWithRetry_AllFailed(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := New()
	ep := Endpoint{
		Name:           "bad-server",
		URL:            server.URL,
		Timeout:        5 * time.Second,
		ExpectedStatus: 200,
		Retries:        2,
	}

	result := c.CheckWithRetry(ep)

	if result.Healthy {
		t.Error("Healthy = true, want false")
	}
	// 初始尝试 + 2次重试 = 3次
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}
}

// TestCheckAll 测试并发批量检查
func TestCheckAll(t *testing.T) {
	// 创建多个模拟服务器
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server2.Close()

	server3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server3.Close()

	c := New(WithConcurrency(2))
	endpoints := []Endpoint{
		{Name: "server1", URL: server1.URL, Timeout: 5 * time.Second, ExpectedStatus: 200},
		{Name: "server2", URL: server2.URL, Timeout: 5 * time.Second, ExpectedStatus: 200},
		{Name: "server3", URL: server3.URL, Timeout: 5 * time.Second, ExpectedStatus: 200},
	}

	batch := c.CheckAll(endpoints)

	// 验证结果数量
	if len(batch.Results) != 3 {
		t.Errorf("len(Results) = %d, want 3", len(batch.Results))
	}

	// 验证结果顺序保持
	if batch.Results[0].Name != "server1" {
		t.Errorf("Results[0].Name = %q, want %q", batch.Results[0].Name, "server1")
	}
	if batch.Results[1].Name != "server2" {
		t.Errorf("Results[1].Name = %q, want %q", batch.Results[1].Name, "server2")
	}
	if batch.Results[2].Name != "server3" {
		t.Errorf("Results[2].Name = %q, want %q", batch.Results[2].Name, "server3")
	}

	// 验证汇总
	if batch.Summary.Total != 3 {
		t.Errorf("Summary.Total = %d, want 3", batch.Summary.Total)
	}
	if batch.Summary.Healthy != 2 {
		t.Errorf("Summary.Healthy = %d, want 2", batch.Summary.Healthy)
	}
	if batch.Summary.Unhealthy != 1 {
		t.Errorf("Summary.Unhealthy = %d, want 1", batch.Summary.Unhealthy)
	}
}

// TestCheckAll_EmptyEndpoints 测试空端点列表
func TestCheckAll_EmptyEndpoints(t *testing.T) {
	c := New()
	batch := c.CheckAll([]Endpoint{})

	if len(batch.Results) != 0 {
		t.Errorf("len(Results) = %d, want 0", len(batch.Results))
	}
	if batch.Summary.Total != 0 {
		t.Errorf("Summary.Total = %d, want 0", batch.Summary.Total)
	}
}

// TestCategorizeError 测试错误分类
func TestCategorizeError(t *testing.T) {
	c := New()

	tests := []struct {
		name     string
		err      error
		contains string
	}{
		{"DNS error", errors.New("no such host"), "DNS resolution failed"},
		{"Connection refused", errors.New("connection refused"), "connection refused"},
		{"Context deadline", errors.New("context deadline exceeded"), "connection timeout"},
		{"Timeout", errors.New("request timeout"), "timeout"},
		{"Certificate error", errors.New("certificate verify failed"), "SSL certificate error"},
		{"Unknown error", errors.New("some random error"), "some random error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.categorizeError(tt.err)
			if !strings.Contains(result.Error(), tt.contains) {
				t.Errorf("categorizeError(%q) = %q, want to contain %q", tt.err, result.Error(), tt.contains)
			}
		})
	}
}

// TestCalculateSummary 测试汇总计算
func TestCalculateSummary(t *testing.T) {
	c := New()

	results := []Result{
		{Healthy: true},
		{Healthy: false},
		{Healthy: true},
		{Healthy: true},
		{Healthy: false},
	}

	summary := c.calculateSummary(results, 100*time.Millisecond)

	if summary.Total != 5 {
		t.Errorf("Total = %d, want 5", summary.Total)
	}
	if summary.Healthy != 3 {
		t.Errorf("Healthy = %d, want 3", summary.Healthy)
	}
	if summary.Unhealthy != 2 {
		t.Errorf("Unhealthy = %d, want 2", summary.Unhealthy)
	}
	if summary.Duration != 100*time.Millisecond {
		t.Errorf("Duration = %v, want %v", summary.Duration, 100*time.Millisecond)
	}
}

// TestCalculateSummary_Empty 测试空结果汇总
func TestCalculateSummary_Empty(t *testing.T) {
	c := New()

	summary := c.calculateSummary([]Result{}, 0)

	if summary.Total != 0 {
		t.Errorf("Total = %d, want 0", summary.Total)
	}
	if summary.Healthy != 0 {
		t.Errorf("Healthy = %d, want 0", summary.Healthy)
	}
	if summary.Unhealthy != 0 {
		t.Errorf("Unhealthy = %d, want 0", summary.Unhealthy)
	}
}

// TestCheck_InvalidURL 测试无效 URL
func TestCheck_InvalidURL(t *testing.T) {
	c := New()
	ep := Endpoint{
		Name:           "invalid",
		URL:            "not-a-valid-url",
		Timeout:        1 * time.Second,
		ExpectedStatus: 200,
	}

	result := c.Check(ep)

	if result.Healthy {
		t.Error("Healthy = true, want false")
	}
	if result.Error == nil {
		t.Error("Error = nil, want error")
	}
}
