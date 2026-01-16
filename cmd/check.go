// check 命令
// 实现单 URL 健康检查功能
package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/r1ckyIn/healthcheck-cli/internal/checker"
	"github.com/r1ckyIn/healthcheck-cli/internal/output"
	"github.com/spf13/cobra"
)

// check 命令的参数
var (
	checkTimeout        time.Duration
	checkExpectedStatus int
	checkHeaders        []string
	checkInsecure       bool
	checkOutput         string
)

// checkCmd 是 check 子命令
var checkCmd = &cobra.Command{
	Use:   "check <url>",
	Short: "Check health of a single URL",
	Long: `Check the health status of a single HTTP endpoint.

The endpoint is considered healthy if:
  - Connection is established successfully
  - Response is received within timeout
  - HTTP status code matches expected value (default: 200)

Examples:
  # Basic check
  healthcheck check https://api.example.com/health

  # With custom timeout
  healthcheck check https://api.example.com/health --timeout 10s

  # With authentication header
  healthcheck check https://api.example.com/health -H "Authorization: Bearer token123"

  # Skip SSL verification (for self-signed certs)
  healthcheck check https://internal.example.com/health --insecure

  # JSON output
  healthcheck check https://api.example.com/health -o json`,
	Args: cobra.ExactArgs(1),
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// 定义参数
	checkCmd.Flags().DurationVarP(&checkTimeout, "timeout", "t", 5*time.Second,
		"Request timeout (e.g., 5s, 10s, 1m)")
	checkCmd.Flags().IntVarP(&checkExpectedStatus, "expected-status", "s", 200,
		"Expected HTTP status code")
	checkCmd.Flags().StringArrayVarP(&checkHeaders, "header", "H", nil,
		"Custom header (can be used multiple times, format: 'Key: Value')")
	checkCmd.Flags().BoolVarP(&checkInsecure, "insecure", "k", false,
		"Skip SSL certificate verification")
	checkCmd.Flags().StringVarP(&checkOutput, "output", "o", "table",
		"Output format (table/json)")
}

// runCheck 执行 check 命令
func runCheck(cmd *cobra.Command, args []string) error {
	targetURL := args[0]

	// 验证 URL 格式
	if err := validateURL(targetURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(2)
	}

	// 解析 Headers
	headers, err := parseHeaders(checkHeaders)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(2)
	}

	// 创建端点配置
	endpoint := checker.Endpoint{
		Name:            targetURL,
		URL:             targetURL,
		Timeout:         checkTimeout,
		Retries:         0,
		ExpectedStatus:  checkExpectedStatus,
		FollowRedirects: true,
		Insecure:        checkInsecure,
		Headers:         headers,
	}

	// 执行检查
	c := checker.New()
	result := c.Check(endpoint)

	// 格式化输出
	formatter := output.NewFormatter(
		output.OutputFormat(checkOutput),
		os.Stdout,
		IsNoColor(),
	)

	if err := formatter.FormatSingle(result); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// 根据结果设置退出码
	if !result.Healthy {
		os.Exit(1)
	}

	return nil
}

// validateURL 验证 URL 格式
func validateURL(rawURL string) error {
	// 检查是否有协议
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return fmt.Errorf("invalid URL '%s': must start with http:// or https://", rawURL)
	}

	// 解析 URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL '%s': %w", rawURL, err)
	}

	// 检查是否有 host
	if parsed.Host == "" {
		return fmt.Errorf("invalid URL '%s': missing host", rawURL)
	}

	return nil
}

// parseHeaders 解析 Header 参数
func parseHeaders(headerStrs []string) (map[string]string, error) {
	headers := make(map[string]string)

	for _, h := range headerStrs {
		// 查找第一个冒号
		idx := strings.Index(h, ":")
		if idx == -1 {
			return nil, fmt.Errorf("invalid header format '%s': expected 'Key: Value'", h)
		}

		key := strings.TrimSpace(h[:idx])
		value := strings.TrimSpace(h[idx+1:])

		if key == "" {
			return nil, fmt.Errorf("invalid header '%s': key cannot be empty", h)
		}

		headers[key] = value
	}

	return headers, nil
}
