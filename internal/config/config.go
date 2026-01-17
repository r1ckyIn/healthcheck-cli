// Config file parsing
// Implements YAML config file parsing and management
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

// Config represents complete config structure
type Config struct {
	Defaults  Defaults   `mapstructure:"defaults"`
	Endpoints []Endpoint `mapstructure:"endpoints"`
}

// Defaults is global default config
type Defaults struct {
	Timeout         string `mapstructure:"timeout"`
	Retries         int    `mapstructure:"retries"`
	ExpectedStatus  int    `mapstructure:"expected_status"`
	FollowRedirects *bool  `mapstructure:"follow_redirects"`
	Insecure        bool   `mapstructure:"insecure"`
}

// Endpoint is single endpoint config
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

// Load loads config from file
func Load(path string) (*Config, error) {
	// Check if file exists
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

// ToCheckerEndpoints converts config to checker.Endpoint list
func (c *Config) ToCheckerEndpoints() ([]checker.Endpoint, error) {
	endpoints := make([]checker.Endpoint, 0, len(c.Endpoints))

	// Parse defaults
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

	// Convert each endpoint
	for i, ep := range c.Endpoints {
		if ep.URL == "" {
			return nil, fmt.Errorf("endpoint #%d: missing url", i+1)
		}

		// Expand environment variables
		url := expandEnvVars(ep.URL)
		name := ep.Name
		if name == "" {
			name = url
		}

		// Parse timeout
		timeout := defaultTimeout
		if ep.Timeout != "" {
			t, err := time.ParseDuration(ep.Timeout)
			if err != nil {
				return nil, fmt.Errorf("endpoint '%s': invalid timeout '%s': %w", name, ep.Timeout, err)
			}
			timeout = t
		}

		// Retry count
		retries := defaultRetries
		if ep.Retries != nil {
			retries = *ep.Retries
		}

		// Expected status code
		expectedStatus := defaultExpectedStatus
		if ep.ExpectedStatus != nil {
			expectedStatus = *ep.ExpectedStatus
		}

		// Follow redirects
		followRedirects := defaultFollowRedirects
		if ep.FollowRedirects != nil {
			followRedirects = *ep.FollowRedirects
		}

		// SSL verification
		insecure := defaultInsecure
		if ep.Insecure != nil {
			insecure = *ep.Insecure
		}

		// Expand environment variables in headers
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

// envVarPattern matches ${VAR} or ${VAR:-default}
var envVarPattern = regexp.MustCompile(`\$\{([^}:]+)(:-([^}]*))?\}`)

// expandEnvVars expands environment variables
// Supports ${VAR} and ${VAR:-default} format
func expandEnvVars(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		// Parse variable name and default value
		parts := envVarPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		varName := parts[1]
		defaultValue := ""
		if len(parts) >= 4 && parts[3] != "" {
			defaultValue = parts[3]
		}

		// Get environment variable
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return defaultValue
	})
}

// findEnvVars finds all environment variables in a string
func findEnvVars(s string) []string {
	matches := envVarPattern.FindAllStringSubmatch(s, -1)
	vars := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) >= 2 {
			vars = append(vars, m[1])
		}
	}
	return vars
}

// GenerateSampleConfig generates sample config
func GenerateSampleConfig(full bool) string {
	if full {
		return `# Health Check CLI Configuration

# Global default settings
defaults:
  timeout: 5s
  retries: 2
  expected_status: 200
  follow_redirects: true
  insecure: false

# Endpoint list
endpoints:
  # Basic configuration
  - name: "Public Website"
    url: "https://www.example.com"

  # Full configuration
  - name: "API Gateway"
    url: "https://api.example.com/health"
    timeout: 10s
    retries: 3
    expected_status: 200
    follow_redirects: true

  # With authentication
  - name: "Admin API"
    url: "https://admin.example.com/health"
    headers:
      Authorization: "Bearer ${ADMIN_TOKEN}"
      X-Request-ID: "healthcheck"

  # Internal service (self-signed certificate)
  - name: "Internal Service"
    url: "https://internal.local:8443/ping"
    insecure: true

  # Expect non-200 status
  - name: "Redirect Check"
    url: "https://old.example.com"
    expected_status: 301
    follow_redirects: false
`
	}

	return `# Health Check CLI Configuration

# Global defaults (optional)
defaults:
  timeout: 5s
  retries: 2

# Endpoints to check
endpoints:
  - name: "Example API"
    url: "https://api.example.com/health"

  - name: "Website"
    url: "https://www.example.com"
`
}

// ValidationResult contains errors and warnings
type ValidationResult struct {
	Errors   []string
	Warnings []string
}

// ValidateConfig validates config file
func ValidateConfig(cfg *Config) []string {
	result := ValidateConfigWithWarnings(cfg)
	return result.Errors
}

// ValidateConfigWithWarnings validates config and returns both errors and warnings
func ValidateConfigWithWarnings(cfg *Config) ValidationResult {
	result := ValidationResult{
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// Check if endpoints exist
	if len(cfg.Endpoints) == 0 {
		result.Errors = append(result.Errors, "no endpoints defined")
	}

	// Track unset environment variables
	unsetEnvVars := make(map[string]bool)

	// Validate each endpoint
	for i, ep := range cfg.Endpoints {
		prefix := fmt.Sprintf("endpoint #%d", i+1)
		if ep.Name != "" {
			prefix = fmt.Sprintf("endpoint '%s'", ep.Name)
		}

		// URL is required
		if ep.URL == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: missing url", prefix))
			continue
		}

		// URL format check
		if !strings.HasPrefix(ep.URL, "http://") && !strings.HasPrefix(ep.URL, "https://") &&
			!strings.HasPrefix(ep.URL, "${") {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: url must start with http:// or https://", prefix))
		}

		// Check for unset environment variables in URL
		for _, varName := range findEnvVars(ep.URL) {
			if os.Getenv(varName) == "" && !unsetEnvVars[varName] {
				// Check if has default value
				if !strings.Contains(ep.URL, "${"+varName+":-") {
					unsetEnvVars[varName] = true
					result.Warnings = append(result.Warnings, fmt.Sprintf("%s: environment variable '%s' is not set and has no default value", prefix, varName))
				}
			}
		}

		// Check for unset environment variables in headers
		for headerName, headerValue := range ep.Headers {
			for _, varName := range findEnvVars(headerValue) {
				if os.Getenv(varName) == "" && !unsetEnvVars[varName] {
					if !strings.Contains(headerValue, "${"+varName+":-") {
						unsetEnvVars[varName] = true
						result.Warnings = append(result.Warnings, fmt.Sprintf("%s: header '%s' uses environment variable '%s' which is not set and has no default value", prefix, headerName, varName))
					}
				}
			}
		}

		// Timeout format check
		if ep.Timeout != "" {
			if _, err := time.ParseDuration(ep.Timeout); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("%s: invalid timeout format '%s'", prefix, ep.Timeout))
			}
		}

		// Status code range check
		if ep.ExpectedStatus != nil && (*ep.ExpectedStatus < 100 || *ep.ExpectedStatus > 599) {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: expected_status must be between 100 and 599", prefix))
		}
	}

	// Validate defaults
	if cfg.Defaults.Timeout != "" {
		if _, err := time.ParseDuration(cfg.Defaults.Timeout); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("defaults: invalid timeout format '%s'", cfg.Defaults.Timeout))
		}
	}

	if cfg.Defaults.ExpectedStatus != 0 && (cfg.Defaults.ExpectedStatus < 100 || cfg.Defaults.ExpectedStatus > 599) {
		result.Errors = append(result.Errors, "defaults: expected_status must be between 100 and 599")
	}

	return result
}
