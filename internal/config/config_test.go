// Config module unit tests / 配置模块单元测试
// Test config file parsing, validation and conversion / 测试配置文件解析、验证和转换功能
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestLoad_Success 测试成功加载配置文件
func TestLoad_Success(t *testing.T) {
	// 创建临时配置文件
	content := `
defaults:
  timeout: 10s
  retries: 2
  expected_status: 200

endpoints:
  - name: "Test API"
    url: "https://api.example.com/health"
  - name: "Website"
    url: "https://www.example.com"
    timeout: 5s
`
	tmpFile := createTempFile(t, "config-*.yaml", content)
	defer os.Remove(tmpFile)

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Defaults.Timeout != "10s" {
		t.Errorf("Defaults.Timeout = %q, want %q", cfg.Defaults.Timeout, "10s")
	}
	if cfg.Defaults.Retries != 2 {
		t.Errorf("Defaults.Retries = %d, want 2", cfg.Defaults.Retries)
	}
	if len(cfg.Endpoints) != 2 {
		t.Errorf("len(Endpoints) = %d, want 2", len(cfg.Endpoints))
	}
	if cfg.Endpoints[0].Name != "Test API" {
		t.Errorf("Endpoints[0].Name = %q, want %q", cfg.Endpoints[0].Name, "Test API")
	}
}

// TestLoad_FileNotFound 测试文件不存在
func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want to contain 'not found'", err.Error())
	}
}

// TestLoad_InvalidYAML 测试无效 YAML 格式
func TestLoad_InvalidYAML(t *testing.T) {
	content := `
endpoints:
  - name: "Test"
    url: "https://example.com"
  invalid yaml here
    not properly indented
`
	tmpFile := createTempFile(t, "invalid-*.yaml", content)
	defer os.Remove(tmpFile)

	_, err := Load(tmpFile)
	if err == nil {
		t.Error("Load() error = nil, want error for invalid YAML")
	}
}

// TestToCheckerEndpoints_Basic 测试基本配置转换
func TestToCheckerEndpoints_Basic(t *testing.T) {
	cfg := &Config{
		Endpoints: []Endpoint{
			{
				Name: "Test API",
				URL:  "https://api.example.com/health",
			},
		},
	}

	endpoints, err := cfg.ToCheckerEndpoints()
	if err != nil {
		t.Fatalf("ToCheckerEndpoints() error = %v", err)
	}

	if len(endpoints) != 1 {
		t.Fatalf("len(endpoints) = %d, want 1", len(endpoints))
	}

	ep := endpoints[0]
	if ep.Name != "Test API" {
		t.Errorf("Name = %q, want %q", ep.Name, "Test API")
	}
	if ep.URL != "https://api.example.com/health" {
		t.Errorf("URL = %q, want %q", ep.URL, "https://api.example.com/health")
	}
	// 默认值
	if ep.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want %v", ep.Timeout, 5*time.Second)
	}
	if ep.ExpectedStatus != 200 {
		t.Errorf("ExpectedStatus = %d, want 200", ep.ExpectedStatus)
	}
	if !ep.FollowRedirects {
		t.Error("FollowRedirects = false, want true")
	}
}

// TestToCheckerEndpoints_WithDefaults 测试带默认值的配置转换
func TestToCheckerEndpoints_WithDefaults(t *testing.T) {
	followRedirects := false
	cfg := &Config{
		Defaults: Defaults{
			Timeout:         "10s",
			Retries:         3,
			ExpectedStatus:  201,
			FollowRedirects: &followRedirects,
			Insecure:        true,
		},
		Endpoints: []Endpoint{
			{
				Name: "Test",
				URL:  "https://example.com",
			},
		},
	}

	endpoints, err := cfg.ToCheckerEndpoints()
	if err != nil {
		t.Fatalf("ToCheckerEndpoints() error = %v", err)
	}

	ep := endpoints[0]
	if ep.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want %v", ep.Timeout, 10*time.Second)
	}
	if ep.Retries != 3 {
		t.Errorf("Retries = %d, want 3", ep.Retries)
	}
	if ep.ExpectedStatus != 201 {
		t.Errorf("ExpectedStatus = %d, want 201", ep.ExpectedStatus)
	}
	if ep.FollowRedirects {
		t.Error("FollowRedirects = true, want false")
	}
	if !ep.Insecure {
		t.Error("Insecure = false, want true")
	}
}

// TestToCheckerEndpoints_EndpointOverridesDefaults 测试端点配置覆盖默认值
func TestToCheckerEndpoints_EndpointOverridesDefaults(t *testing.T) {
	retries := 5
	expectedStatus := 204
	insecure := true
	cfg := &Config{
		Defaults: Defaults{
			Timeout:        "10s",
			Retries:        2,
			ExpectedStatus: 200,
		},
		Endpoints: []Endpoint{
			{
				Name:           "Override",
				URL:            "https://example.com",
				Timeout:        "30s",
				Retries:        &retries,
				ExpectedStatus: &expectedStatus,
				Insecure:       &insecure,
			},
		},
	}

	endpoints, err := cfg.ToCheckerEndpoints()
	if err != nil {
		t.Fatalf("ToCheckerEndpoints() error = %v", err)
	}

	ep := endpoints[0]
	if ep.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", ep.Timeout, 30*time.Second)
	}
	if ep.Retries != 5 {
		t.Errorf("Retries = %d, want 5", ep.Retries)
	}
	if ep.ExpectedStatus != 204 {
		t.Errorf("ExpectedStatus = %d, want 204", ep.ExpectedStatus)
	}
	if !ep.Insecure {
		t.Error("Insecure = false, want true")
	}
}

// TestToCheckerEndpoints_MissingURL 测试缺少 URL 的错误
func TestToCheckerEndpoints_MissingURL(t *testing.T) {
	cfg := &Config{
		Endpoints: []Endpoint{
			{Name: "No URL"},
		},
	}

	_, err := cfg.ToCheckerEndpoints()
	if err == nil {
		t.Error("ToCheckerEndpoints() error = nil, want error for missing URL")
	}
	if !strings.Contains(err.Error(), "missing url") {
		t.Errorf("error = %q, want to contain 'missing url'", err.Error())
	}
}

// TestToCheckerEndpoints_InvalidTimeout 测试无效超时格式
func TestToCheckerEndpoints_InvalidTimeout(t *testing.T) {
	cfg := &Config{
		Defaults: Defaults{
			Timeout: "invalid",
		},
		Endpoints: []Endpoint{
			{URL: "https://example.com"},
		},
	}

	_, err := cfg.ToCheckerEndpoints()
	if err == nil {
		t.Error("ToCheckerEndpoints() error = nil, want error for invalid timeout")
	}
}

// TestToCheckerEndpoints_DefaultName 测试默认使用 URL 作为名称
func TestToCheckerEndpoints_DefaultName(t *testing.T) {
	cfg := &Config{
		Endpoints: []Endpoint{
			{URL: "https://example.com/health"},
		},
	}

	endpoints, err := cfg.ToCheckerEndpoints()
	if err != nil {
		t.Fatalf("ToCheckerEndpoints() error = %v", err)
	}

	if endpoints[0].Name != "https://example.com/health" {
		t.Errorf("Name = %q, want URL as default name", endpoints[0].Name)
	}
}

// TestToCheckerEndpoints_Headers 测试请求头处理
func TestToCheckerEndpoints_Headers(t *testing.T) {
	cfg := &Config{
		Endpoints: []Endpoint{
			{
				URL: "https://example.com",
				Headers: map[string]string{
					"Authorization": "Bearer token123",
					"X-Custom":      "value",
				},
			},
		},
	}

	endpoints, err := cfg.ToCheckerEndpoints()
	if err != nil {
		t.Fatalf("ToCheckerEndpoints() error = %v", err)
	}

	if endpoints[0].Headers["Authorization"] != "Bearer token123" {
		t.Errorf("Authorization = %q, want %q", endpoints[0].Headers["Authorization"], "Bearer token123")
	}
	if endpoints[0].Headers["X-Custom"] != "value" {
		t.Errorf("X-Custom = %q, want %q", endpoints[0].Headers["X-Custom"], "value")
	}
}

// TestExpandEnvVars_Basic 测试基本环境变量扩展
func TestExpandEnvVars_Basic(t *testing.T) {
	t.Setenv("TEST_VAR", "test-value")

	result := expandEnvVars("prefix-${TEST_VAR}-suffix")
	expected := "prefix-test-value-suffix"

	if result != expected {
		t.Errorf("expandEnvVars() = %q, want %q", result, expected)
	}
}

// TestExpandEnvVars_WithDefault 测试带默认值的环境变量扩展
func TestExpandEnvVars_WithDefault(t *testing.T) {
	// 确保变量未设置
	os.Unsetenv("UNSET_VAR")

	result := expandEnvVars("value-${UNSET_VAR:-default-value}")
	expected := "value-default-value"

	if result != expected {
		t.Errorf("expandEnvVars() = %q, want %q", result, expected)
	}
}

// TestExpandEnvVars_SetVarIgnoresDefault 测试已设置变量忽略默认值
func TestExpandEnvVars_SetVarIgnoresDefault(t *testing.T) {
	t.Setenv("SET_VAR", "actual-value")

	result := expandEnvVars("${SET_VAR:-default}")
	expected := "actual-value"

	if result != expected {
		t.Errorf("expandEnvVars() = %q, want %q", result, expected)
	}
}

// TestExpandEnvVars_MultipleVars 测试多个环境变量
func TestExpandEnvVars_MultipleVars(t *testing.T) {
	t.Setenv("VAR1", "value1")
	t.Setenv("VAR2", "value2")

	result := expandEnvVars("${VAR1} and ${VAR2}")
	expected := "value1 and value2"

	if result != expected {
		t.Errorf("expandEnvVars() = %q, want %q", result, expected)
	}
}

// TestExpandEnvVars_NoVars 测试无环境变量
func TestExpandEnvVars_NoVars(t *testing.T) {
	result := expandEnvVars("plain text without vars")
	expected := "plain text without vars"

	if result != expected {
		t.Errorf("expandEnvVars() = %q, want %q", result, expected)
	}
}

// TestExpandEnvVars_UnsetVarEmptyResult 测试未设置变量且无默认值
func TestExpandEnvVars_UnsetVarEmptyResult(t *testing.T) {
	os.Unsetenv("NONEXISTENT_VAR")

	result := expandEnvVars("prefix-${NONEXISTENT_VAR}-suffix")
	expected := "prefix--suffix"

	if result != expected {
		t.Errorf("expandEnvVars() = %q, want %q", result, expected)
	}
}

// TestValidateConfig_Valid 测试有效配置
func TestValidateConfig_Valid(t *testing.T) {
	cfg := &Config{
		Defaults: Defaults{
			Timeout:        "5s",
			ExpectedStatus: 200,
		},
		Endpoints: []Endpoint{
			{Name: "API", URL: "https://api.example.com"},
			{Name: "Web", URL: "http://www.example.com"},
		},
	}

	errors := ValidateConfig(cfg)
	if len(errors) != 0 {
		t.Errorf("ValidateConfig() returned errors: %v", errors)
	}
}

// TestValidateConfig_NoEndpoints 测试无端点
func TestValidateConfig_NoEndpoints(t *testing.T) {
	cfg := &Config{
		Endpoints: []Endpoint{},
	}

	errors := ValidateConfig(cfg)
	if len(errors) == 0 {
		t.Error("ValidateConfig() returned no errors, want 'no endpoints defined'")
	}

	found := false
	for _, e := range errors {
		if strings.Contains(e, "no endpoints defined") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("errors = %v, want to contain 'no endpoints defined'", errors)
	}
}

// TestValidateConfig_MissingURL 测试缺少 URL
func TestValidateConfig_MissingURL(t *testing.T) {
	cfg := &Config{
		Endpoints: []Endpoint{
			{Name: "No URL Endpoint"},
		},
	}

	errors := ValidateConfig(cfg)
	found := false
	for _, e := range errors {
		if strings.Contains(e, "missing url") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("errors = %v, want to contain 'missing url'", errors)
	}
}

// TestValidateConfig_InvalidURLScheme 测试无效 URL 协议
func TestValidateConfig_InvalidURLScheme(t *testing.T) {
	cfg := &Config{
		Endpoints: []Endpoint{
			{Name: "Invalid", URL: "ftp://example.com"},
		},
	}

	errors := ValidateConfig(cfg)
	found := false
	for _, e := range errors {
		if strings.Contains(e, "must start with http") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("errors = %v, want to contain 'must start with http'", errors)
	}
}

// TestValidateConfig_InvalidTimeout 测试无效超时格式
func TestValidateConfig_InvalidTimeout(t *testing.T) {
	cfg := &Config{
		Endpoints: []Endpoint{
			{Name: "Test", URL: "https://example.com", Timeout: "invalid"},
		},
	}

	errors := ValidateConfig(cfg)
	found := false
	for _, e := range errors {
		if strings.Contains(e, "invalid timeout format") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("errors = %v, want to contain 'invalid timeout format'", errors)
	}
}

// TestValidateConfig_InvalidStatusCode 测试无效状态码
func TestValidateConfig_InvalidStatusCode(t *testing.T) {
	invalidStatus := 999
	cfg := &Config{
		Endpoints: []Endpoint{
			{Name: "Test", URL: "https://example.com", ExpectedStatus: &invalidStatus},
		},
	}

	errors := ValidateConfig(cfg)
	found := false
	for _, e := range errors {
		if strings.Contains(e, "expected_status must be between") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("errors = %v, want to contain 'expected_status must be between'", errors)
	}
}

// TestValidateConfig_InvalidDefaultTimeout 测试无效默认超时
func TestValidateConfig_InvalidDefaultTimeout(t *testing.T) {
	cfg := &Config{
		Defaults: Defaults{
			Timeout: "not-a-duration",
		},
		Endpoints: []Endpoint{
			{URL: "https://example.com"},
		},
	}

	errors := ValidateConfig(cfg)
	found := false
	for _, e := range errors {
		if strings.Contains(e, "defaults: invalid timeout") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("errors = %v, want to contain 'defaults: invalid timeout'", errors)
	}
}

// TestValidateConfig_InvalidDefaultStatus 测试无效默认状态码
func TestValidateConfig_InvalidDefaultStatus(t *testing.T) {
	cfg := &Config{
		Defaults: Defaults{
			ExpectedStatus: 50, // 小于 100
		},
		Endpoints: []Endpoint{
			{URL: "https://example.com"},
		},
	}

	errors := ValidateConfig(cfg)
	found := false
	for _, e := range errors {
		if strings.Contains(e, "defaults: expected_status must be between") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("errors = %v, want to contain 'defaults: expected_status'", errors)
	}
}

// TestValidateConfig_EnvVarURL 测试环境变量 URL 不报错
func TestValidateConfig_EnvVarURL(t *testing.T) {
	cfg := &Config{
		Endpoints: []Endpoint{
			{Name: "EnvURL", URL: "${API_URL}"},
		},
	}

	errors := ValidateConfig(cfg)
	// 环境变量 URL 不应触发协议错误
	for _, e := range errors {
		if strings.Contains(e, "must start with http") {
			t.Errorf("EnvVar URL should not trigger protocol error: %v", errors)
		}
	}
}

// TestGenerateSampleConfig_Basic 测试生成基本示例配置
func TestGenerateSampleConfig_Basic(t *testing.T) {
	sample := GenerateSampleConfig(false)

	if !strings.Contains(sample, "defaults:") {
		t.Error("sample config should contain 'defaults:'")
	}
	if !strings.Contains(sample, "endpoints:") {
		t.Error("sample config should contain 'endpoints:'")
	}
	if !strings.Contains(sample, "timeout:") {
		t.Error("sample config should contain 'timeout:'")
	}
}

// TestGenerateSampleConfig_Full 测试生成完整示例配置
func TestGenerateSampleConfig_Full(t *testing.T) {
	sample := GenerateSampleConfig(true)

	// 完整配置应包含更多内容
	if !strings.Contains(sample, "headers:") {
		t.Error("full sample config should contain 'headers:'")
	}
	if !strings.Contains(sample, "Authorization:") {
		t.Error("full sample config should contain 'Authorization:'")
	}
	if !strings.Contains(sample, "insecure:") {
		t.Error("full sample config should contain 'insecure:'")
	}
	if !strings.Contains(sample, "follow_redirects:") {
		t.Error("full sample config should contain 'follow_redirects:'")
	}
}

// TestToCheckerEndpoints_EnvVarInHeaders 测试请求头中的环境变量
func TestToCheckerEndpoints_EnvVarInHeaders(t *testing.T) {
	t.Setenv("AUTH_TOKEN", "secret-token-123")

	cfg := &Config{
		Endpoints: []Endpoint{
			{
				URL: "https://example.com",
				Headers: map[string]string{
					"Authorization": "Bearer ${AUTH_TOKEN}",
				},
			},
		},
	}

	endpoints, err := cfg.ToCheckerEndpoints()
	if err != nil {
		t.Fatalf("ToCheckerEndpoints() error = %v", err)
	}

	if endpoints[0].Headers["Authorization"] != "Bearer secret-token-123" {
		t.Errorf("Authorization = %q, want %q", endpoints[0].Headers["Authorization"], "Bearer secret-token-123")
	}
}

// TestToCheckerEndpoints_EnvVarInURL 测试 URL 中的环境变量
func TestToCheckerEndpoints_EnvVarInURL(t *testing.T) {
	t.Setenv("API_HOST", "api.example.com")

	cfg := &Config{
		Endpoints: []Endpoint{
			{
				URL: "https://${API_HOST}/health",
			},
		},
	}

	endpoints, err := cfg.ToCheckerEndpoints()
	if err != nil {
		t.Fatalf("ToCheckerEndpoints() error = %v", err)
	}

	if endpoints[0].URL != "https://api.example.com/health" {
		t.Errorf("URL = %q, want %q", endpoints[0].URL, "https://api.example.com/health")
	}
}

// TestValidateConfigWithWarnings_EnvVarWarning 测试环境变量警告
func TestValidateConfigWithWarnings_EnvVarWarning(t *testing.T) {
	os.Unsetenv("UNSET_TOKEN")

	cfg := &Config{
		Endpoints: []Endpoint{
			{
				Name: "API",
				URL:  "https://api.example.com",
				Headers: map[string]string{
					"Authorization": "Bearer ${UNSET_TOKEN}",
				},
			},
		},
	}

	result := ValidateConfigWithWarnings(cfg)

	if len(result.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", result.Errors)
	}
	if len(result.Warnings) == 0 {
		t.Error("Warnings is empty, want warning about unset env var")
	}

	found := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "UNSET_TOKEN") && strings.Contains(w, "not set") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Warnings = %v, want to contain warning about UNSET_TOKEN", result.Warnings)
	}
}

// TestValidateConfigWithWarnings_EnvVarWithDefault 测试带默认值的环境变量不警告
func TestValidateConfigWithWarnings_EnvVarWithDefault(t *testing.T) {
	os.Unsetenv("OPTIONAL_VAR")

	cfg := &Config{
		Endpoints: []Endpoint{
			{
				Name: "API",
				URL:  "https://api.example.com/${OPTIONAL_VAR:-default}",
			},
		},
	}

	result := ValidateConfigWithWarnings(cfg)

	// 有默认值的环境变量不应该产生警告
	for _, w := range result.Warnings {
		if strings.Contains(w, "OPTIONAL_VAR") {
			t.Errorf("Should not warn about env var with default value: %v", result.Warnings)
		}
	}
}

// TestValidateConfigWithWarnings_EnvVarSet 测试已设置的环境变量不警告
func TestValidateConfigWithWarnings_EnvVarSet(t *testing.T) {
	t.Setenv("SET_TOKEN", "my-token")

	cfg := &Config{
		Endpoints: []Endpoint{
			{
				Name: "API",
				URL:  "https://api.example.com",
				Headers: map[string]string{
					"Authorization": "Bearer ${SET_TOKEN}",
				},
			},
		},
	}

	result := ValidateConfigWithWarnings(cfg)

	for _, w := range result.Warnings {
		if strings.Contains(w, "SET_TOKEN") {
			t.Errorf("Should not warn about set env var: %v", result.Warnings)
		}
	}
}

// TestFindEnvVars 测试查找环境变量
func TestFindEnvVars(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"no vars", []string{}},
		{"${VAR1}", []string{"VAR1"}},
		{"${VAR1} and ${VAR2}", []string{"VAR1", "VAR2"}},
		{"${VAR:-default}", []string{"VAR"}},
		{"prefix${VAR}suffix", []string{"VAR"}},
	}

	for _, tt := range tests {
		result := findEnvVars(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("findEnvVars(%q) = %v, want %v", tt.input, result, tt.expected)
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("findEnvVars(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
			}
		}
	}
}

// createTempFile 创建临时文件
func createTempFile(t *testing.T, pattern, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, pattern)

	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return tmpFile
}
