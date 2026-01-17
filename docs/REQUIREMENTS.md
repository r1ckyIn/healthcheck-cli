# Health Check CLI - Requirements Document

> Version: 1.0
> Date: 2026-01-17
> Author: Ricky

[English](#english) | [中文](#中文)

---

## English

### 1. Project Overview

#### 1.1 Introduction

Health Check CLI is a command-line tool written in Go for batch HTTP endpoint health checking. It supports concurrent checks, multiple output formats, and configuration file management. Ideal for deployment verification, troubleshooting, and CI/CD integration.

#### 1.2 Project Goals

- Eliminate repetitive manual health check work
- Provide fast, reliable service status checking
- Seamless integration into CI/CD pipelines
- Cross-platform support (Linux, macOS, Windows)

#### 1.3 Tech Stack

| Technology | Purpose |
|------------|---------|
| Go | Primary programming language |
| Cobra | CLI framework |
| Viper | Configuration management |
| goroutine + channel | Concurrency model |
| GoReleaser | Cross-platform build and release |
| GitHub Actions | CI/CD |

### 2. Target Users & Use Cases

#### 2.1 Target Users

- DevOps Engineers
- SRE Engineers
- Backend Developers
- QA Engineers

#### 2.2 Use Cases

| Scenario | Description |
|----------|-------------|
| Post-deployment verification | Quick check of all services after manual deployment |
| Troubleshooting | Quickly locate problematic services after receiving alerts |
| CI/CD integration | Automated smoke test in deployment pipeline |
| Local development | Check dependency services before starting project |
| Cross-environment check | Simultaneously check service status across multiple environments |

### 3. Functional Requirements

#### 3.1 Core Features

| Feature | Description | Priority |
|---------|-------------|----------|
| Single URL check | Check health status of a single HTTP endpoint | P0 |
| Batch check | Read multiple endpoints from config file and check concurrently | P0 |
| Timeout control | Configurable request timeout | P0 |
| Multiple output formats | Support Table and JSON format output | P0 |
| Exit codes | Return different exit codes based on check results | P0 |
| Concurrency control | Configurable maximum concurrency | P1 |
| Retry mechanism | Auto retry N times on failure | P1 |
| Custom headers | Support adding authentication and other headers | P1 |
| SSL verification skip | Support skipping self-signed certificate verification | P1 |
| Config validation | Validate config file format and content | P2 |
| Shell completion | Generate Bash/Zsh/Fish completion scripts | P2 |

#### 3.2 Health Check Criteria

A service is considered "healthy" when:

1. Connection can be established (no DNS errors, no connection timeout)
2. Response received within timeout
3. Returns expected HTTP status code (default 200)

#### 3.3 Error Classification

| Error Type | Cause | Is Service Issue? | Handling |
|------------|-------|-------------------|----------|
| Config file not found | User error | No | Exit with code 2 |
| Config file format error | User error | No | Exit with code 2 |
| Invalid URL format | Config error | No | Exit with code 2 |
| DNS resolution failed | Service issue | Yes | Mark as unhealthy |
| Connection timeout | Service issue | Yes | Mark as unhealthy |
| Connection refused | Service issue | Yes | Mark as unhealthy |
| HTTP status mismatch | Service issue | Yes | Mark as unhealthy |

### 4. CLI Interface Design

#### 4.1 Command Structure

```
healthcheck
├── check <url>              # Check single URL
├── run                      # Batch check from config
├── config
│   ├── init                 # Generate sample config
│   └── validate             # Validate config file
├── completion <shell>       # Generate shell completion
└── version                  # Show version info
```

#### 4.2 Command Details

##### `healthcheck check <url>`

Check health status of a single URL.

```bash
healthcheck check https://api.example.com/health [flags]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--timeout` | `-t` | duration | 5s | Request timeout |
| `--expected-status` | `-s` | int | 200 | Expected HTTP status code |
| `--header` | `-H` | string | - | Custom header (can be used multiple times) |
| `--insecure` | `-k` | bool | false | Skip SSL certificate verification |
| `--output` | `-o` | string | table | Output format (table/json) |

**Examples:**

```bash
# Basic usage
healthcheck check https://api.example.com/health

# Custom timeout and authentication
healthcheck check https://api.example.com/health \
    --timeout 10s \
    --header "Authorization: Bearer token123"

# JSON output
healthcheck check https://api.example.com/health -o json
```

##### `healthcheck run`

Batch check multiple endpoints from config file.

```bash
healthcheck run [flags]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--config` | `-c` | string | endpoints.yaml | Config file path |
| `--timeout` | `-t` | duration | - | Override config timeout |
| `--concurrency` | `-n` | int | 10 | Maximum concurrency |
| `--output` | `-o` | string | table | Output format (table/json) |
| `--quiet` | `-q` | bool | false | Quiet mode |
| `--insecure` | `-k` | bool | false | Global skip SSL verification |

**Examples:**

```bash
# Basic usage
healthcheck run -c endpoints.yaml

# Adjust concurrency and timeout
healthcheck run -c endpoints.yaml --concurrency 20 --timeout 10s

# CI/CD usage (JSON output)
healthcheck run -c endpoints.yaml -o json

# Quiet mode (only exit code)
healthcheck run -c endpoints.yaml -q
```

##### `healthcheck config init`

Generate sample configuration file.

```bash
healthcheck config init [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--full` | bool | false | Output complete example with all optional configs |

**Examples:**

```bash
# Output to stdout
healthcheck config init

# Save to file
healthcheck config init > endpoints.yaml

# Generate full example
healthcheck config init --full > endpoints.yaml
```

##### `healthcheck config validate`

Validate configuration file format and content.

```bash
healthcheck config validate [flags]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--config` | `-c` | string | endpoints.yaml | Config file path |

##### `healthcheck completion <shell>`

Generate shell completion scripts.

```bash
healthcheck completion bash|zsh|fish|powershell
```

**Examples:**

```bash
# Bash
healthcheck completion bash > /etc/bash_completion.d/healthcheck

# Zsh
healthcheck completion zsh > "${fpath[1]}/_healthcheck"

# Fish
healthcheck completion fish > ~/.config/fish/completions/healthcheck.fish
```

##### `healthcheck version`

Display version information.

```bash
healthcheck version
```

**Output Example:**

```
healthcheck v1.0.0
  Built:    2026-01-17T10:30:00Z
  Commit:   abc1234
  Go:       go1.21.5
```

#### 4.3 Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Show help |
| `--no-color` | - | Disable colored output |

#### 4.4 Exit Codes

| Code | Meaning | Trigger |
|------|---------|---------|
| 0 | Success | All services healthy |
| 1 | Check failed | Some services unhealthy |
| 2 | Config error | Config file not found, format error, invalid params |

### 5. Configuration File Format

#### 5.1 File Format

Uses YAML format, default filename is `endpoints.yaml`.

#### 5.2 Complete Structure

```yaml
# Global defaults (optional)
defaults:
  timeout: 5s
  retries: 2
  expected_status: 200
  follow_redirects: true

# Endpoint list
endpoints:
  # Minimal config
  - name: "Public Website"
    url: "https://www.example.com"

  # Full config
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

  # Internal service (self-signed cert)
  - name: "Internal Service"
    url: "https://internal.local:8443/ping"
    insecure: true

  # Expect non-200 status
  - name: "Redirect Check"
    url: "https://old.example.com"
    expected_status: 301
    follow_redirects: false
```

#### 5.3 Configuration Options

##### Global Config (defaults)

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| timeout | duration | 5s | Request timeout |
| retries | int | 0 | Retry count on failure |
| expected_status | int | 200 | Expected HTTP status code |
| follow_redirects | bool | true | Follow redirects |
| insecure | bool | false | Skip SSL verification |

##### Endpoint Config (endpoints[])

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| name | string | Yes | Endpoint name (for display) |
| url | string | Yes | URL to check |
| timeout | duration | No | Override global timeout |
| retries | int | No | Override global retry count |
| expected_status | int | No | Override global expected status |
| follow_redirects | bool | No | Override global redirect setting |
| insecure | bool | No | Override global SSL setting |
| headers | map | No | Custom request headers |

#### 5.4 Environment Variable Support

```yaml
# Basic usage
headers:
  Authorization: "Bearer ${ADMIN_TOKEN}"

# With default value
url: "${API_URL:-https://localhost:8080}/health"
```

#### 5.5 Configuration Priority

From highest to lowest:

1. Command line flags
2. Individual endpoint config
3. defaults config
4. Program built-in defaults

### 6. Output Formats

#### 6.1 Table Format (Default)

Human-readable table format for terminal viewing.

**Normal Output:**

```
NAME              URL                                 STATUS    LATENCY
API Gateway       https://api.example.com/health      ✓ 200     45ms
Auth Service      https://auth.example.com/ping       ✓ 200     128ms
User Service      https://user.example.com/health     ✓ 200     67ms
Payment Service   https://payment.example.com/ping    ✗ 503     --
Database Proxy    https://db.internal:8080/status     ✗ timeout --

Summary: 3/5 healthy
```

**Design Rules:**

| Element | Design |
|---------|--------|
| Success marker | `✓` green |
| Failure marker | `✗` red |
| Timeout/no response | Show error type (timeout/connection refused) |
| Latency display | In milliseconds, show `--` for failures |
| Column alignment | Left-aligned, auto-calculate width |
| Summary | Last line shows healthy/total |

**Color Control:**

- Auto-detect terminal, enable colors if supported
- `--no-color` forces colors off
- Auto-disable colors when output redirected to file

#### 6.2 JSON Format

Machine-readable format for CI/CD integration and script processing.

```json
{
  "timestamp": "2026-01-17T10:30:00Z",
  "duration_ms": 1250,
  "summary": {
    "total": 5,
    "healthy": 3,
    "unhealthy": 2
  },
  "results": [
    {
      "name": "API Gateway",
      "url": "https://api.example.com/health",
      "healthy": true,
      "status_code": 200,
      "latency_ms": 45,
      "error": null
    },
    {
      "name": "Payment Service",
      "url": "https://payment.example.com/ping",
      "healthy": false,
      "status_code": 503,
      "latency_ms": 230,
      "error": null
    },
    {
      "name": "Database Proxy",
      "url": "https://db.internal:8080/status",
      "healthy": false,
      "status_code": null,
      "latency_ms": null,
      "error": "connection timeout after 5s"
    }
  ]
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| timestamp | string | Check start time (ISO 8601) |
| duration_ms | int | Total check duration (ms) |
| summary.total | int | Total endpoint count |
| summary.healthy | int | Healthy count |
| summary.unhealthy | int | Unhealthy count |
| results[].name | string | Endpoint name |
| results[].url | string | Checked URL |
| results[].healthy | bool | Is healthy |
| results[].status_code | int/null | HTTP status code |
| results[].latency_ms | int/null | Response time (ms) |
| results[].error | string/null | Error message |

#### 6.3 Single URL Check Output

**Table Format:**

```
✓ https://api.example.com/health    200    45ms
```

**JSON Format:**

```json
{
  "url": "https://api.example.com/health",
  "healthy": true,
  "status_code": 200,
  "latency_ms": 45,
  "error": null
}
```

#### 6.4 Quiet Mode

`--quiet` flag enables quiet mode, no output, only exit code indicates result.

```bash
healthcheck run -c endpoints.yaml -q
echo $?  # 0 or 1
```

### 7. Technical Implementation

#### 7.1 Project Structure

```
healthcheck/
├── cmd/
│   ├── root.go           # Root command + global config
│   ├── check.go          # check subcommand
│   ├── run.go            # run subcommand
│   ├── config.go         # config subcommand group
│   └── version.go        # version subcommand
├── internal/
│   ├── checker/
│   │   ├── checker.go    # Core check logic
│   │   └── checker_test.go
│   ├── config/
│   │   ├── config.go     # Config parsing
│   │   ├── validate.go   # Config validation
│   │   └── config_test.go
│   └── output/
│       ├── formatter.go  # Output interface definition
│       ├── table.go      # Table format implementation
│       └── json.go       # JSON format implementation
├── main.go               # Entry point
├── go.mod
├── go.sum
├── .goreleaser.yaml      # Cross-platform build config
├── .github/
│   └── workflows/
│       ├── ci.yml        # CI pipeline
│       └── release.yml   # Release pipeline
├── README.md
└── LICENSE
```

#### 7.2 Concurrency Model

Using goroutine + channel + semaphore pattern:

```go
func (c *Checker) CheckAll(endpoints []Endpoint) []Result {
    results := make(chan Result, len(endpoints))
    sem := make(chan struct{}, c.concurrency)

    var wg sync.WaitGroup
    for _, ep := range endpoints {
        wg.Add(1)
        go func(ep Endpoint) {
            defer wg.Done()
            sem <- struct{}{}        // Acquire semaphore
            results <- c.Check(ep)
            <-sem                    // Release semaphore
        }(ep)
    }

    wg.Wait()
    close(results)

    // Collect results in original order
    // ...
}
```

#### 7.3 Timeout Control

Using context for timeout:

```go
func (c *Checker) Check(ep Endpoint) Result {
    ctx, cancel := context.WithTimeout(context.Background(), ep.Timeout)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "GET", ep.URL, nil)
    // ...
}
```

#### 7.4 Retry Mechanism

```go
func (c *Checker) CheckWithRetry(ep Endpoint) Result {
    var result Result
    for i := 0; i <= ep.Retries; i++ {
        result = c.Check(ep)
        if result.Healthy {
            return result
        }
        // Optional: retry interval
    }
    return result
}
```

### 8. Development Plan

| Day | Task | Output |
|-----|------|--------|
| Day 1 | Project init, Cobra command framework, `check` command | Can check single URL |
| Day 2 | Config file parsing (Viper), `run` command, concurrency | Can batch check |
| Day 3 | Timeout control, retry mechanism, JSON/Table output | Complete check functionality |
| Day 4 | Shell completion, GoReleaser config, unit tests | Release-ready state |
| Day 5 | GitHub Actions CI, README docs, release v1.0 | Official release |

### 9. Resume Highlights

Skills demonstrated by completing this project:

| Skill | Demonstration |
|-------|---------------|
| Go Concurrency | goroutine + channel + semaphore for controlled concurrency |
| CLI Development | Cobra + Viper framework best practices |
| Interface Abstraction | Output format interface design, supports extension |
| DevOps Mindset | Exit code design, JSON output for CI/CD integration |
| Cross-platform Release | GoReleaser auto-build multi-platform binaries |
| CI/CD Practice | GitHub Actions automated testing and release |

### 10. References

- [Cobra Documentation](https://cobra.dev/)
- [Viper Documentation](https://github.com/spf13/viper)
- [GoReleaser Documentation](https://goreleaser.com/)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)

---

## 中文

### 1. 项目概述

#### 1.1 项目简介

Health Check CLI 是一个用 Go 语言编写的命令行工具，用于批量检查 HTTP 端点的健康状态。支持并发检查、多种输出格式、配置文件管理，适用于部署验证、故障排查、CI/CD 集成等场景。

#### 1.2 项目目标

- 消除重复性的手动健康检查工作
- 提供快速、可靠的服务状态检查能力
- 无缝集成到 CI/CD 流程中
- 跨平台支持（Linux, macOS, Windows）

#### 1.3 技术栈

| 技术 | 用途 |
|------|------|
| Go | 主要编程语言 |
| Cobra | CLI 框架 |
| Viper | 配置管理 |
| goroutine + channel | 并发模型 |
| GoReleaser | 跨平台构建和发布 |
| GitHub Actions | CI/CD |

### 2. 目标用户与使用场景

#### 2.1 目标用户

- DevOps 工程师
- SRE 工程师
- 后端开发者
- QA 工程师

#### 2.2 使用场景

| 场景 | 描述 |
|------|------|
| 部署后验证 | 手动部署新版本后快速检查所有服务 |
| 故障排查 | 收到告警后快速定位问题服务 |
| CI/CD 集成 | 自动化部署流程中的 Smoke Test |
| 本地开发 | 启动项目前检查依赖服务 |
| 跨环境检查 | 同时检查多个环境的服务状态 |

### 3. 功能需求

#### 3.1 核心功能

| 功能 | 描述 | 优先级 |
|------|------|--------|
| 单 URL 检查 | 检查单个 HTTP 端点的健康状态 | P0 |
| 批量检查 | 从配置文件读取多个端点并发检查 | P0 |
| 超时控制 | 可配置的请求超时时间 | P0 |
| 多种输出格式 | 支持 Table 和 JSON 格式输出 | P0 |
| 退出码 | 根据检查结果返回不同退出码 | P0 |
| 并发控制 | 可配置的最大并发数 | P1 |
| 重试机制 | 失败时自动重试 N 次 | P1 |
| 自定义 Headers | 支持添加认证等请求头 | P1 |
| SSL 证书验证 | 支持跳过自签名证书验证 | P1 |
| 配置校验 | 校验配置文件格式和内容 | P2 |
| Shell 补全 | 生成 Bash/Zsh/Fish 补全脚本 | P2 |

#### 3.2 健康判断标准

服务被判定为"健康"需满足：

1. 能够建立连接（无 DNS 错误、无连接超时）
2. 在超时时间内响应
3. 返回期望的 HTTP 状态码（默认 200）

#### 3.3 错误分类

| 错误类型 | 原因 | 是服务问题？ | 处理方式 |
|----------|------|--------------|----------|
| 配置文件不存在 | 用户错误 | 否 | 报错退出，退出码 2 |
| 配置文件格式错误 | 用户错误 | 否 | 报错退出，退出码 2 |
| URL 格式无效 | 配置错误 | 否 | 报错退出，退出码 2 |
| DNS 解析失败 | 服务问题 | 是 | 标记为不健康 |
| 连接超时 | 服务问题 | 是 | 标记为不健康 |
| 连接被拒绝 | 服务问题 | 是 | 标记为不健康 |
| HTTP 状态码不匹配 | 服务问题 | 是 | 标记为不健康 |

### 4. 命令行接口设计

#### 4.1 命令结构

```
healthcheck
├── check <url>              # 检查单个 URL
├── run                      # 从配置文件批量检查
├── config
│   ├── init                 # 生成示例配置文件
│   └── validate             # 校验配置文件
├── completion <shell>       # 生成 Shell 补全脚本
└── version                  # 显示版本信息
```

#### 4.2 命令详情

##### `healthcheck check <url>`

检查单个 URL 的健康状态。

```bash
healthcheck check https://api.example.com/health [flags]
```

**参数：**

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--timeout` | `-t` | duration | 5s | 请求超时时间 |
| `--expected-status` | `-s` | int | 200 | 期望的 HTTP 状态码 |
| `--header` | `-H` | string | - | 自定义 Header（可多次使用） |
| `--insecure` | `-k` | bool | false | 跳过 SSL 证书验证 |
| `--output` | `-o` | string | table | 输出格式 |

**示例：**

```bash
# 基本用法
healthcheck check https://api.example.com/health

# 自定义超时和认证
healthcheck check https://api.example.com/health \
    --timeout 10s \
    --header "Authorization: Bearer token123"

# JSON 输出
healthcheck check https://api.example.com/health -o json
```

##### `healthcheck run`

从配置文件批量检查多个端点。

```bash
healthcheck run [flags]
```

**参数：**

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--config` | `-c` | string | endpoints.yaml | 配置文件路径 |
| `--timeout` | `-t` | duration | - | 覆盖配置的超时时间 |
| `--concurrency` | `-n` | int | 10 | 最大并发数 |
| `--output` | `-o` | string | table | 输出格式 |
| `--quiet` | `-q` | bool | false | 静默模式 |
| `--insecure` | `-k` | bool | false | 全局跳过 SSL 验证 |

**示例：**

```bash
# 基本用法
healthcheck run -c endpoints.yaml

# 调整并发和超时
healthcheck run -c endpoints.yaml --concurrency 20 --timeout 10s

# CI/CD 用法（JSON 输出）
healthcheck run -c endpoints.yaml -o json

# 静默模式（只看退出码）
healthcheck run -c endpoints.yaml -q
```

##### `healthcheck config init`

生成示例配置文件。

```bash
healthcheck config init [flags]
```

**参数：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--full` | bool | false | 输出包含所有可选配置的完整示例 |

**示例：**

```bash
# 输出到标准输出
healthcheck config init

# 保存到文件
healthcheck config init > endpoints.yaml

# 生成完整示例
healthcheck config init --full > endpoints.yaml
```

##### `healthcheck config validate`

校验配置文件格式和内容。

```bash
healthcheck config validate [flags]
```

**参数：**

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--config` | `-c` | string | endpoints.yaml | 配置文件路径 |

##### `healthcheck completion <shell>`

生成 Shell 补全脚本。

```bash
healthcheck completion bash|zsh|fish|powershell
```

**示例：**

```bash
# Bash
healthcheck completion bash > /etc/bash_completion.d/healthcheck

# Zsh
healthcheck completion zsh > "${fpath[1]}/_healthcheck"

# Fish
healthcheck completion fish > ~/.config/fish/completions/healthcheck.fish
```

##### `healthcheck version`

显示版本信息。

```bash
healthcheck version
```

**输出示例：**

```
healthcheck v1.0.0
  Built:    2026-01-17T10:30:00Z
  Commit:   abc1234
  Go:       go1.21.5
```

#### 4.3 全局参数

| 参数 | 简写 | 说明 |
|------|------|------|
| `--help` | `-h` | 显示帮助信息 |
| `--no-color` | - | 禁用颜色输出 |

#### 4.4 退出码

| 退出码 | 含义 | 触发场景 |
|--------|------|----------|
| 0 | 成功 | 所有服务健康 |
| 1 | 检查失败 | 有服务不健康 |
| 2 | 配置错误 | 配置文件不存在、格式错误、参数无效 |

### 5. 配置文件格式

#### 5.1 文件格式

使用 YAML 格式，文件名默认为 `endpoints.yaml`。

#### 5.2 完整结构

```yaml
# 全局默认配置（可选）
defaults:
  timeout: 5s
  retries: 2
  expected_status: 200
  follow_redirects: true

# 端点列表
endpoints:
  # 最简配置
  - name: "Public Website"
    url: "https://www.example.com"

  # 完整配置
  - name: "API Gateway"
    url: "https://api.example.com/health"
    timeout: 10s
    retries: 3
    expected_status: 200
    follow_redirects: true

  # 带认证
  - name: "Admin API"
    url: "https://admin.example.com/health"
    headers:
      Authorization: "Bearer ${ADMIN_TOKEN}"
      X-Request-ID: "healthcheck"

  # 内网服务（自签名证书）
  - name: "Internal Service"
    url: "https://internal.local:8443/ping"
    insecure: true

  # 期望非 200 状态码
  - name: "Redirect Check"
    url: "https://old.example.com"
    expected_status: 301
    follow_redirects: false
```

#### 5.3 配置项说明

##### 全局配置 (defaults)

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| timeout | duration | 5s | 请求超时时间 |
| retries | int | 0 | 失败重试次数 |
| expected_status | int | 200 | 期望的 HTTP 状态码 |
| follow_redirects | bool | true | 是否跟随重定向 |
| insecure | bool | false | 是否跳过 SSL 验证 |

##### 端点配置 (endpoints[])

| 配置项 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 是 | 端点名称（用于显示） |
| url | string | 是 | 检查的 URL |
| timeout | duration | 否 | 覆盖全局超时 |
| retries | int | 否 | 覆盖全局重试次数 |
| expected_status | int | 否 | 覆盖全局期望状态码 |
| follow_redirects | bool | 否 | 覆盖全局重定向设置 |
| insecure | bool | 否 | 覆盖全局 SSL 设置 |
| headers | map | 否 | 自定义请求头 |

#### 5.4 环境变量支持

```yaml
# 基本用法
headers:
  Authorization: "Bearer ${ADMIN_TOKEN}"

# 带默认值
url: "${API_URL:-https://localhost:8080}/health"
```

#### 5.5 配置优先级

从高到低：

1. 命令行参数
2. 单个 endpoint 配置
3. defaults 配置
4. 程序内置默认值

### 6. 输出格式

#### 6.1 Table 格式（默认）

人类可读的表格格式，用于终端查看。

**正常输出：**

```
NAME              URL                                 STATUS    LATENCY
API Gateway       https://api.example.com/health      ✓ 200     45ms
Auth Service      https://auth.example.com/ping       ✓ 200     128ms
User Service      https://user.example.com/health     ✓ 200     67ms
Payment Service   https://payment.example.com/ping    ✗ 503     --
Database Proxy    https://db.internal:8080/status     ✗ timeout --

Summary: 3/5 healthy
```

**设计规则：**

| 元素 | 设计 |
|------|------|
| 成功标记 | `✓` 绿色 |
| 失败标记 | `✗` 红色 |
| 超时/无响应 | 显示错误类型 |
| 延迟显示 | 毫秒为单位，失败显示 `--` |
| 列对齐 | 左对齐，自动计算宽度 |
| Summary | 最后一行显示 健康数/总数 |

**颜色控制：**

- 自动检测终端，支持颜色则启用
- `--no-color` 强制关闭颜色
- 输出重定向到文件时自动关闭颜色

#### 6.2 JSON 格式

机器可读格式，用于 CI/CD 集成和脚本处理。

```json
{
  "timestamp": "2026-01-17T10:30:00Z",
  "duration_ms": 1250,
  "summary": {
    "total": 5,
    "healthy": 3,
    "unhealthy": 2
  },
  "results": [
    {
      "name": "API Gateway",
      "url": "https://api.example.com/health",
      "healthy": true,
      "status_code": 200,
      "latency_ms": 45,
      "error": null
    }
  ]
}
```

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| timestamp | string | 检查开始时间（ISO 8601） |
| duration_ms | int | 整体检查耗时（毫秒） |
| summary.total | int | 端点总数 |
| summary.healthy | int | 健康数量 |
| summary.unhealthy | int | 不健康数量 |
| results[].name | string | 端点名称 |
| results[].url | string | 检查的 URL |
| results[].healthy | bool | 是否健康 |
| results[].status_code | int/null | HTTP 状态码 |
| results[].latency_ms | int/null | 响应时间（毫秒） |
| results[].error | string/null | 错误信息 |

#### 6.3 单 URL 检查输出

**Table 格式：**

```
✓ https://api.example.com/health    200    45ms
```

**JSON 格式：**

```json
{
  "url": "https://api.example.com/health",
  "healthy": true,
  "status_code": 200,
  "latency_ms": 45,
  "error": null
}
```

#### 6.4 静默模式

`--quiet` 参数启用静默模式，不输出任何内容，仅通过退出码表示结果。

```bash
healthcheck run -c endpoints.yaml -q
echo $?  # 0 or 1
```

### 7. 技术实现要点

#### 7.1 项目结构

```
healthcheck/
├── cmd/
│   ├── root.go           # 根命令 + 全局配置
│   ├── check.go          # check 子命令
│   ├── run.go            # run 子命令
│   ├── config.go         # config 子命令组
│   └── version.go        # version 子命令
├── internal/
│   ├── checker/
│   │   ├── checker.go    # 核心检查逻辑
│   │   └── checker_test.go
│   ├── config/
│   │   ├── config.go     # 配置解析
│   │   ├── validate.go   # 配置校验
│   │   └── config_test.go
│   └── output/
│       ├── formatter.go  # 输出接口定义
│       ├── table.go      # Table 格式实现
│       └── json.go       # JSON 格式实现
├── main.go               # 入口
├── go.mod
├── go.sum
├── .goreleaser.yaml      # 跨平台构建配置
├── .github/
│   └── workflows/
│       ├── ci.yml        # CI 流程
│       └── release.yml   # 发布流程
├── README.md
└── LICENSE
```

#### 7.2 并发模型

使用 goroutine + channel + semaphore 模式：

```go
func (c *Checker) CheckAll(endpoints []Endpoint) []Result {
    results := make(chan Result, len(endpoints))
    sem := make(chan struct{}, c.concurrency)

    var wg sync.WaitGroup
    for _, ep := range endpoints {
        wg.Add(1)
        go func(ep Endpoint) {
            defer wg.Done()
            sem <- struct{}{}        // 获取信号量
            results <- c.Check(ep)
            <-sem                    // 释放信号量
        }(ep)
    }

    wg.Wait()
    close(results)

    // 按原顺序收集结果
    // ...
}
```

#### 7.3 超时控制

使用 context 实现超时：

```go
func (c *Checker) Check(ep Endpoint) Result {
    ctx, cancel := context.WithTimeout(context.Background(), ep.Timeout)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "GET", ep.URL, nil)
    // ...
}
```

#### 7.4 重试机制

```go
func (c *Checker) CheckWithRetry(ep Endpoint) Result {
    var result Result
    for i := 0; i <= ep.Retries; i++ {
        result = c.Check(ep)
        if result.Healthy {
            return result
        }
        // 可选：重试间隔
    }
    return result
}
```

### 8. 开发计划

| 天数 | 任务 | 产出 |
|------|------|------|
| Day 1 | 项目初始化、Cobra 命令框架、`check` 命令 | 能检查单个 URL |
| Day 2 | 配置文件解析(Viper)、`run` 命令、并发实现 | 能批量检查 |
| Day 3 | 超时控制、重试机制、JSON/Table 输出 | 完整的检查功能 |
| Day 4 | Shell 补全、GoReleaser 配置、单元测试 | 可发布状态 |
| Day 5 | GitHub Actions CI、README 文档、发布 v1.0 | 正式发布 |

### 9. 简历亮点

完成此项目后，可在简历中体现以下技能：

| 技能点 | 体现 |
|--------|------|
| Go 并发编程 | goroutine + channel + semaphore 可控并发实现 |
| CLI 开发 | Cobra + Viper 框架最佳实践 |
| 接口抽象 | 输出格式的接口设计，支持扩展 |
| DevOps 思维 | 退出码设计、JSON 输出便于 CI/CD 集成 |
| 跨平台发布 | GoReleaser 自动构建多平台二进制 |
| CI/CD 实践 | GitHub Actions 自动化测试和发布 |

### 10. 参考资料

- [Cobra Documentation](https://cobra.dev/)
- [Viper Documentation](https://github.com/spf13/viper)
- [GoReleaser Documentation](https://goreleaser.com/)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
