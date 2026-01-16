// 配置文件解析
// 实现 YAML 配置文件的解析和管理
package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/r1ckyIn/healthcheck-cli/internal/checker"
	"github.com/spf13/viper"
)

// Config 表示完整的配置结构
type Config struct {
	Defaults  Defaults   `mapstructure:"defaults"`
	Endpoints []Endpoint `mapstructure:"endpoints"`
}

// Defaults 全局默认配置
type Defaults struct {
	Timeout         string `mapstructure:"timeout"`
	Retries         int    `mapstructure:"retries"`
	ExpectedStatus  int    `mapstructure:"expected_status"`
	FollowRedirects *bool  `mapstructure:"follow_redirects"`
	Insecure        bool   `mapstructure:"insecure"`
}

// Endpoint 单个端点配置
type Endpoint struct {
	Name            string            `mapstructure:"name"`
	URL             string            `mapstructure:"url"`
	Timeout         string            `mapstructure:"timeout"`
	Retries         *int              `mapstructure:"retries"`
	ExpectedStatus  *int              `mapstructure:"expected_status"`
	FollowRedirects *bool             `mapstructure:"follow_redirects"`
	Insecure        *bool             `mapstructure:"insecure"`
	Headers         map[string]string `mapstructure:"headers"`
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// ToCheckerEndpoints 将配置转换为 checker.Endpoint 列表
func (c *Config) ToCheckerEndpoints() ([]checker.Endpoint, error) {
	endpoints := make([]checker.Endpoint, 0, len(c.Endpoints))

	// 解析默认值
	defaultTimeout := 5 * time.Second
	if c.Defaults.Timeout != "" {
		t, err := time.ParseDuration(c.Defaults.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid default timeout '%s': %w", c.Defaults.Timeout, err)
		}
		defaultTimeout = t
	}

	defaultRetries := c.Defaults.Retries
	defaultExpectedStatus := 200
	if c.Defaults.ExpectedStatus != 0 {
		defaultExpectedStatus = c.Defaults.ExpectedStatus
	}

	defaultFollowRedirects := true
	if c.Defaults.FollowRedirects != nil {
		defaultFollowRedirects = *c.Defaults.FollowRedirects
	}

	defaultInsecure := c.Defaults.Insecure

	// 转换每个端点
	for i, ep := range c.Endpoints {
		if ep.URL == "" {
			return nil, fmt.Errorf("endpoint #%d: missing url", i+1)
		}

		// 处理环境变量替换
		url := expandEnvVars(ep.URL)
		name := ep.Name
		if name == "" {
			name = url
		}

		// 解析超时时间
		timeout := defaultTimeout
		if ep.Timeout != "" {
			t, err := time.ParseDuration(ep.Timeout)
			if err != nil {
				return nil, fmt.Errorf("endpoint '%s': invalid timeout '%s': %w", name, ep.Timeout, err)
			}
			timeout = t
		}

		// 重试次数
		retries := defaultRetries
		if ep.Retries != nil {
			retries = *ep.Retries
		}

		// 期望状态码
		expectedStatus := defaultExpectedStatus
		if ep.ExpectedStatus != nil {
			expectedStatus = *ep.ExpectedStatus
		}

		// 跟随重定向
		followRedirects := defaultFollowRedirects
		if ep.FollowRedirects != nil {
			followRedirects = *ep.FollowRedirects
		}

		// SSL 验证
		insecure := defaultInsecure
		if ep.Insecure != nil {
			insecure = *ep.Insecure
		}

		// 处理 Headers 中的环境变量
		headers := make(map[string]string)
		for k, v := range ep.Headers {
			headers[k] = expandEnvVars(v)
		}

		endpoints = append(endpoints, checker.Endpoint{
			Name:            name,
			URL:             url,
			Timeout:         timeout,
			Retries:         retries,
			ExpectedStatus:  expectedStatus,
			FollowRedirects: followRedirects,
			Insecure:        insecure,
			Headers:         headers,
		})
	}

	return endpoints, nil
}

// expandEnvVars 扩展环境变量
// 支持 ${VAR} 和 ${VAR:-default} 格式
func expandEnvVars(s string) string {
	// 匹配 ${VAR} 或 ${VAR:-default}
	re := regexp.MustCompile(`\$\{([^}:]+)(:-([^}]*))?\}`)

	return re.ReplaceAllStringFunc(s, func(match string) string {
		// 解析变量名和默认值
		parts := re.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		varName := parts[1]
		defaultValue := ""
		if len(parts) >= 4 && parts[3] != "" {
			defaultValue = parts[3]
		}

		// 获取环境变量
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return defaultValue
	})
}

// GenerateSampleConfig 生成示例配置
func GenerateSampleConfig(full bool) string {
	if full {
		return `# Health Check CLI Configuration
# 健康检查 CLI 配置文件

# Global default settings
# 全局默认配置
defaults:
  timeout: 5s
  retries: 2
  expected_status: 200
  follow_redirects: true
  insecure: false

# Endpoint list
# 端点列表
endpoints:
  # Basic configuration
  # 基本配置
  - name: "Public Website"
    url: "https://www.example.com"

  # Full configuration
  # 完整配置
  - name: "API Gateway"
    url: "https://api.example.com/health"
    timeout: 10s
    retries: 3
    expected_status: 200
    follow_redirects: true

  # With authentication
  # 带认证
  - name: "Admin API"
    url: "https://admin.example.com/health"
    headers:
      Authorization: "Bearer ${ADMIN_TOKEN}"
      X-Request-ID: "healthcheck"

  # Internal service (self-signed certificate)
  # 内网服务（自签名证书）
  - name: "Internal Service"
    url: "https://internal.local:8443/ping"
    insecure: true

  # Expect non-200 status
  # 期望非 200 状态码
  - name: "Redirect Check"
    url: "https://old.example.com"
    expected_status: 301
    follow_redirects: false
`
	}

	return `# Health Check CLI Configuration
# 健康检查 CLI 配置文件

# Global defaults (optional)
# 全局默认配置（可选）
defaults:
  timeout: 5s
  retries: 2

# Endpoints to check
# 待检查的端点
endpoints:
  - name: "Example API"
    url: "https://api.example.com/health"

  - name: "Website"
    url: "https://www.example.com"
`
}

// ValidateConfig 验证配置文件
func ValidateConfig(cfg *Config) []string {
	var errors []string

	// 检查是否有端点
	if len(cfg.Endpoints) == 0 {
		errors = append(errors, "no endpoints defined")
	}

	// 验证每个端点
	for i, ep := range cfg.Endpoints {
		prefix := fmt.Sprintf("endpoint #%d", i+1)
		if ep.Name != "" {
			prefix = fmt.Sprintf("endpoint '%s'", ep.Name)
		}

		// URL 必填
		if ep.URL == "" {
			errors = append(errors, fmt.Sprintf("%s: missing url", prefix))
			continue
		}

		// URL 格式检查
		if !strings.HasPrefix(ep.URL, "http://") && !strings.HasPrefix(ep.URL, "https://") &&
			!strings.HasPrefix(ep.URL, "${") {
			errors = append(errors, fmt.Sprintf("%s: url must start with http:// or https://", prefix))
		}

		// 超时格式检查
		if ep.Timeout != "" {
			if _, err := time.ParseDuration(ep.Timeout); err != nil {
				errors = append(errors, fmt.Sprintf("%s: invalid timeout format '%s'", prefix, ep.Timeout))
			}
		}

		// 状态码范围检查
		if ep.ExpectedStatus != nil && (*ep.ExpectedStatus < 100 || *ep.ExpectedStatus > 599) {
			errors = append(errors, fmt.Sprintf("%s: expected_status must be between 100 and 599", prefix))
		}
	}

	// 验证默认值
	if cfg.Defaults.Timeout != "" {
		if _, err := time.ParseDuration(cfg.Defaults.Timeout); err != nil {
			errors = append(errors, fmt.Sprintf("defaults: invalid timeout format '%s'", cfg.Defaults.Timeout))
		}
	}

	if cfg.Defaults.ExpectedStatus != 0 && (cfg.Defaults.ExpectedStatus < 100 || cfg.Defaults.ExpectedStatus > 599) {
		errors = append(errors, "defaults: expected_status must be between 100 and 599")
	}

	return errors
}
