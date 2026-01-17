// Core health check logic
// Implements HTTP endpoint health check functionality
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

// Checker is the health checker
type Checker struct {
	// Cached clients for different configurations
	// Key format: "secure-follow", "secure-nofollow", "insecure-follow", "insecure-nofollow"
	clients     map[string]*http.Client
	clientMu    sync.RWMutex
	concurrency int
}

// Option is Checker configuration option
type Option func(*Checker)

// WithConcurrency sets maximum concurrency
func WithConcurrency(n int) Option {
	return func(c *Checker) {
		if n > 0 {
			c.concurrency = n
		}
	}
}

// New creates a new health checker
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

// getClientKey generates cache key for client based on endpoint config
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

// getClient returns appropriate HTTP client based on endpoint config
func (c *Checker) getClient(ep Endpoint) *http.Client {
	key := getClientKey(ep.Insecure, ep.FollowRedirects)

	// Try to get existing client
	c.clientMu.RLock()
	if client, ok := c.clients[key]; ok {
		c.clientMu.RUnlock()
		return client
	}
	c.clientMu.RUnlock()

	// Create new client
	c.clientMu.Lock()
	defer c.clientMu.Unlock()

	// Double check after acquiring write lock
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

	// Configure redirect handling
	if !ep.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	c.clients[key] = client
	return client
}

// Check checks single endpoint health status
func (c *Checker) Check(ep Endpoint) Result {
	return c.CheckWithContext(context.Background(), ep)
}

// CheckWithContext checks single endpoint with context support
func (c *Checker) CheckWithContext(ctx context.Context, ep Endpoint) Result {
	result := Result{
		Name: ep.Name,
		URL:  ep.URL,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, ep.Timeout)
	defer cancel()

	// Get HTTP client
	client := c.getClient(ep)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep.URL, nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		return result
	}

	// Add custom headers
	for key, value := range ep.Headers {
		req.Header.Set(key, value)
	}

	// Set User-Agent
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "healthcheck-cli/1.0")
	}

	// Execute request and measure time
	start := time.Now()
	resp, err := client.Do(req)
	result.Latency = time.Since(start)

	if err != nil {
		result.Error = c.categorizeError(err)
		return result
	}
	defer resp.Body.Close()

	// Record status code
	result.StatusCode = &resp.StatusCode

	// Check if status code matches expected
	if resp.StatusCode == ep.ExpectedStatus {
		result.Healthy = true
	} else {
		result.Error = fmt.Errorf("unexpected status code: got %d, expected %d", resp.StatusCode, ep.ExpectedStatus)
	}

	return result
}

// CheckWithRetry performs health check with retry
func (c *Checker) CheckWithRetry(ep Endpoint) Result {
	return c.CheckWithRetryContext(context.Background(), ep)
}

// CheckWithRetryContext performs health check with retry and context
func (c *Checker) CheckWithRetryContext(ctx context.Context, ep Endpoint) Result {
	var result Result

	for i := 0; i <= ep.Retries; i++ {
		// Check if context is cancelled
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

		// Wait before retry if there are more attempts
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

// indexedResult holds result with its index
type indexedResult struct {
	idx    int
	result Result
}

// CheckAll concurrently checks multiple endpoints
func (c *Checker) CheckAll(endpoints []Endpoint) BatchResult {
	return c.CheckAllWithContext(context.Background(), endpoints)
}

// CheckAllWithContext concurrently checks multiple endpoints with context
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

	// Use channel for collecting results safely
	resultChan := make(chan indexedResult, len(endpoints))
	sem := make(chan struct{}, c.concurrency)
	var wg sync.WaitGroup

	for i, ep := range endpoints {
		wg.Add(1)
		go func(idx int, endpoint Endpoint) {
			defer wg.Done()

			// Acquire semaphore
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

			// Execute check with retry
			resultChan <- indexedResult{
				idx:    idx,
				result: c.CheckWithRetryContext(ctx, endpoint),
			}
		}(i, ep)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for r := range resultChan {
		results[r.idx] = r.result
	}

	return BatchResult{
		Timestamp: startTime,
		Results:   results,
		Summary:   c.calculateSummary(results, time.Since(startTime)),
	}
}

// categorizeError categorizes error type
func (c *Checker) categorizeError(err error) error {
	errStr := err.Error()

	// Categorize based on error message
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

// calculateSummary calculates summary info
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
