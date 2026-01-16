# Health Check CLI - Requirements Document / 需求文档

> Version: 1.0
> Date: 2026-01-17
> Author: Ricky

---

## 1. Project Overview / 项目概述

### 1.1 Introduction / 项目简介

Health Check CLI is a command-line tool written in Go for batch HTTP endpoint health checking. It supports concurrent checks, multiple output formats, and configuration file management. Ideal for deployment verification, troubleshooting, and CI/CD integration.

Health Check CLI 是一个用 Go 语言编写的命令行工具，用于批量检查 HTTP 端点的健康状态。支持并发检查、多种输出格式、配置文件管理，适用于部署验证、故障排查、CI/CD 集成等场景。

### 1.2 Project Goals / 项目目标

- Eliminate repetitive manual health check work / 消除重复性的手动健康检查工作
- Provide fast, reliable service status checking / 提供快速、可靠的服务状态检查能力
- Seamless integration into CI/CD pipelines / 无缝集成到 CI/CD 流程中
- Cross-platform support (Linux, macOS, Windows) / 跨平台支持

### 1.3 Tech Stack / 技术栈

| Technology | Purpose / 用途 |
|------------|----------------|
| Go | Primary programming language / 主要编程语言 |
| Cobra | CLI framework / CLI 框架 |
| Viper | Configuration management / 配置管理 |
| goroutine + channel | Concurrency model / 并发模型 |
| GoReleaser | Cross-platform build and release / 跨平台构建和发布 |
| GitHub Actions | CI/CD |

---

## 2. Target Users & Use Cases / 目标用户与使用场景

### 2.1 Target Users / 目标用户

- DevOps Engineers / DevOps 工程师
- SRE Engineers / SRE 工程师
- Backend Developers / 后端开发者
- QA Engineers / QA 工程师

### 2.2 Use Cases / 使用场景

| Scenario / 场景 | Description / 描述 |
|-----------------|-------------------|
| Post-deployment verification / 部署后验证 | Quick check of all services after manual deployment / 手动部署新版本后快速检查所有服务 |
| Troubleshooting / 故障排查 | Quickly locate problematic services after receiving alerts / 收到告警后快速定位问题服务 |
| CI/CD integration / CI/CD 集成 | Automated smoke test in deployment pipeline / 自动化部署流程中的 Smoke Test |
| Local development / 本地开发 | Check dependency services before starting project / 启动项目前检查依赖服务 |
| Cross-environment check / 跨环境检查 | Simultaneously check service status across multiple environments / 同时检查多个环境的服务状态 |

---

## 3. Functional Requirements / 功能需求

### 3.1 Core Features / 核心功能

| Feature / 功能 | Description / 描述 | Priority / 优先级 |
|----------------|-------------------|------------------|
| Single URL check / 单 URL 检查 | Check health status of a single HTTP endpoint / 检查单个 HTTP 端点的健康状态 | P0 |
| Batch check / 批量检查 | Read multiple endpoints from config file and check concurrently / 从配置文件读取多个端点并发检查 | P0 |
| Timeout control / 超时控制 | Configurable request timeout / 可配置的请求超时时间 | P0 |
| Multiple output formats / 多种输出格式 | Support Table and JSON format output / 支持 Table 和 JSON 格式输出 | P0 |
| Exit codes / 退出码 | Return different exit codes based on check results / 根据检查结果返回不同退出码 | P0 |
| Concurrency control / 并发控制 | Configurable maximum concurrency / 可配置的最大并发数 | P1 |
| Retry mechanism / 重试机制 | Auto retry N times on failure / 失败时自动重试 N 次 | P1 |
| Custom headers / 自定义 Headers | Support adding authentication and other headers / 支持添加认证等请求头 | P1 |
| SSL verification skip / SSL 证书验证 | Support skipping self-signed certificate verification / 支持跳过自签名证书验证 | P1 |
| Config validation / 配置校验 | Validate config file format and content / 校验配置文件格式和内容 | P2 |
| Shell completion / Shell 补全 | Generate Bash/Zsh/Fish completion scripts / 生成 Bash/Zsh/Fish 补全脚本 | P2 |

### 3.2 Health Check Criteria / 健康判断标准

A service is considered "healthy" when:

服务被判定为"健康"需满足：

1. Connection can be established (no DNS errors, no connection timeout) / 能够建立连接（无 DNS 错误、无连接超时）
2. Response received within timeout / 在超时时间内响应
3. Returns expected HTTP status code (default 200) / 返回期望的 HTTP 状态码（默认 200）

### 3.3 Error Classification / 错误分类

| Error Type / 错误类型 | Cause / 原因 | Is Service Issue? / 是服务问题？ | Handling / 处理方式 |
|----------------------|--------------|--------------------------------|---------------------|
| Config file not found / 配置文件不存在 | User error | No / 否 | Exit with code 2 / 报错退出，退出码 2 |
| Config file format error / 配置文件格式错误 | User error | No / 否 | Exit with code 2 / 报错退出，退出码 2 |
| Invalid URL format / URL 格式无效 | Config error | No / 否 | Exit with code 2 / 报错退出，退出码 2 |
| DNS resolution failed / DNS 解析失败 | Service issue | Yes / 是 | Mark as unhealthy / 标记为不健康 |
| Connection timeout / 连接超时 | Service issue | Yes / 是 | Mark as unhealthy / 标记为不健康 |
| Connection refused / 连接被拒绝 | Service issue | Yes / 是 | Mark as unhealthy / 标记为不健康 |
| HTTP status mismatch / HTTP 状态码不匹配 | Service issue | Yes / 是 | Mark as unhealthy / 标记为不健康 |

---

## 4. CLI Interface Design / 命令行接口设计

### 4.1 Command Structure / 命令结构

```
healthcheck
├── check <url>              # Check single URL / 检查单个 URL
├── run                      # Batch check from config / 从配置文件批量检查
├── config
│   ├── init                 # Generate sample config / 生成示例配置文件
│   └── validate             # Validate config file / 校验配置文件
├── completion <shell>       # Generate shell completion / 生成 Shell 补全脚本
└── version                  # Show version info / 显示版本信息
```

### 4.2 Command Details / 命令详情

#### `healthcheck check <url>`

Check health status of a single URL.

检查单个 URL 的健康状态。

```bash
healthcheck check https://api.example.com/health [flags]
```

**Flags / 参数：**

| Flag | Short | Type | Default | Description / 说明 |
|------|-------|------|---------|-------------------|
| `--timeout` | `-t` | duration | 5s | Request timeout / 请求超时时间 |
| `--expected-status` | `-s` | int | 200 | Expected HTTP status code / 期望的 HTTP 状态码 |
| `--header` | `-H` | string | - | Custom header (can be used multiple times) / 自定义 Header（可多次使用） |
| `--insecure` | `-k` | bool | false | Skip SSL certificate verification / 跳过 SSL 证书验证 |
| `--output` | `-o` | string | table | Output format (table/json) / 输出格式 |

**Examples / 示例：**

```bash
# Basic usage / 基本用法
healthcheck check https://api.example.com/health

# Custom timeout and authentication / 自定义超时和认证
healthcheck check https://api.example.com/health \
    --timeout 10s \
    --header "Authorization: Bearer token123"

# JSON output / JSON 输出
healthcheck check https://api.example.com/health -o json
```

#### `healthcheck run`

Batch check multiple endpoints from config file.

从配置文件批量检查多个端点。

```bash
healthcheck run [flags]
```

**Flags / 参数：**

| Flag | Short | Type | Default | Description / 说明 |
|------|-------|------|---------|-------------------|
| `--config` | `-c` | string | endpoints.yaml | Config file path / 配置文件路径 |
| `--timeout` | `-t` | duration | - | Override config timeout / 覆盖配置的超时时间 |
| `--concurrency` | `-n` | int | 10 | Maximum concurrency / 最大并发数 |
| `--output` | `-o` | string | table | Output format (table/json) / 输出格式 |
| `--quiet` | `-q` | bool | false | Quiet mode / 静默模式 |
| `--insecure` | `-k` | bool | false | Global skip SSL verification / 全局跳过 SSL 验证 |

**Examples / 示例：**

```bash
# Basic usage / 基本用法
healthcheck run -c endpoints.yaml

# Adjust concurrency and timeout / 调整并发和超时
healthcheck run -c endpoints.yaml --concurrency 20 --timeout 10s

# CI/CD usage (JSON output) / CI/CD 用法（JSON 输出）
healthcheck run -c endpoints.yaml -o json

# Quiet mode (only exit code) / 静默模式（只看退出码）
healthcheck run -c endpoints.yaml -q
```

#### `healthcheck config init`

Generate sample configuration file.

生成示例配置文件。

```bash
healthcheck config init [flags]
```

**Flags / 参数：**

| Flag | Type | Default | Description / 说明 |
|------|------|---------|-------------------|
| `--full` | bool | false | Output complete example with all optional configs / 输出包含所有可选配置的完整示例 |

**Examples / 示例：**

```bash
# Output to stdout / 输出到标准输出
healthcheck config init

# Save to file / 保存到文件
healthcheck config init > endpoints.yaml

# Generate full example / 生成完整示例
healthcheck config init --full > endpoints.yaml
```

#### `healthcheck config validate`

Validate configuration file format and content.

校验配置文件格式和内容。

```bash
healthcheck config validate [flags]
```

**Flags / 参数：**

| Flag | Short | Type | Default | Description / 说明 |
|------|-------|------|---------|-------------------|
| `--config` | `-c` | string | endpoints.yaml | Config file path / 配置文件路径 |

#### `healthcheck completion <shell>`

Generate shell completion scripts.

生成 Shell 补全脚本。

```bash
healthcheck completion bash|zsh|fish|powershell
```

**Examples / 示例：**

```bash
# Bash
healthcheck completion bash > /etc/bash_completion.d/healthcheck

# Zsh
healthcheck completion zsh > "${fpath[1]}/_healthcheck"

# Fish
healthcheck completion fish > ~/.config/fish/completions/healthcheck.fish
```

#### `healthcheck version`

Display version information.

显示版本信息。

```bash
healthcheck version
```

**Output Example / 输出示例：**

```
healthcheck v1.0.0
  Built:    2026-01-17T10:30:00Z
  Commit:   abc1234
  Go:       go1.21.5
```

### 4.3 Global Flags / 全局参数

| Flag | Short | Description / 说明 |
|------|-------|-------------------|
| `--help` | `-h` | Show help / 显示帮助信息 |
| `--no-color` | - | Disable colored output / 禁用颜色输出 |

### 4.4 Exit Codes / 退出码

| Code | Meaning / 含义 | Trigger / 触发场景 |
|------|---------------|-------------------|
| 0 | Success / 成功 | All services healthy / 所有服务健康 |
| 1 | Check failed / 检查失败 | Some services unhealthy / 有服务不健康 |
| 2 | Config error / 配置错误 | Config file not found, format error, invalid params / 配置文件不存在、格式错误、参数无效 |

---

## 5. Configuration File Format / 配置文件格式

### 5.1 File Format / 文件格式

Uses YAML format, default filename is `endpoints.yaml`.

使用 YAML 格式，文件名默认为 `endpoints.yaml`。

### 5.2 Complete Structure / 完整结构

```yaml
# Global defaults (optional) / 全局默认配置（可选）
defaults:
  timeout: 5s
  retries: 2
  expected_status: 200
  follow_redirects: true

# Endpoint list / 端点列表
endpoints:
  # Minimal config / 最简配置
  - name: "Public Website"
    url: "https://www.example.com"

  # Full config / 完整配置
  - name: "API Gateway"
    url: "https://api.example.com/health"
    timeout: 10s
    retries: 3
    expected_status: 200
    follow_redirects: true

  # With authentication / 带认证
  - name: "Admin API"
    url: "https://admin.example.com/health"
    headers:
      Authorization: "Bearer ${ADMIN_TOKEN}"
      X-Request-ID: "healthcheck"

  # Internal service (self-signed cert) / 内网服务（自签名证书）
  - name: "Internal Service"
    url: "https://internal.local:8443/ping"
    insecure: true

  # Expect non-200 status / 期望非 200 状态码
  - name: "Redirect Check"
    url: "https://old.example.com"
    expected_status: 301
    follow_redirects: false
```

### 5.3 Configuration Options / 配置项说明

#### Global Config (defaults) / 全局配置

| Option | Type | Default | Description / 说明 |
|--------|------|---------|-------------------|
| timeout | duration | 5s | Request timeout / 请求超时时间 |
| retries | int | 0 | Retry count on failure / 失败重试次数 |
| expected_status | int | 200 | Expected HTTP status code / 期望的 HTTP 状态码 |
| follow_redirects | bool | true | Follow redirects / 是否跟随重定向 |
| insecure | bool | false | Skip SSL verification / 是否跳过 SSL 验证 |

#### Endpoint Config (endpoints[]) / 端点配置

| Option | Type | Required | Description / 说明 |
|--------|------|----------|-------------------|
| name | string | Yes / 是 | Endpoint name (for display) / 端点名称（用于显示） |
| url | string | Yes / 是 | URL to check / 检查的 URL |
| timeout | duration | No / 否 | Override global timeout / 覆盖全局超时 |
| retries | int | No / 否 | Override global retry count / 覆盖全局重试次数 |
| expected_status | int | No / 否 | Override global expected status / 覆盖全局期望状态码 |
| follow_redirects | bool | No / 否 | Override global redirect setting / 覆盖全局重定向设置 |
| insecure | bool | No / 否 | Override global SSL setting / 覆盖全局 SSL 设置 |
| headers | map | No / 否 | Custom request headers / 自定义请求头 |

### 5.4 Environment Variable Support / 环境变量支持

```yaml
# Basic usage / 基本用法
headers:
  Authorization: "Bearer ${ADMIN_TOKEN}"

# With default value / 带默认值
url: "${API_URL:-https://localhost:8080}/health"
```

### 5.5 Configuration Priority / 配置优先级

From highest to lowest:

从高到低：

1. Command line flags / 命令行参数
2. Individual endpoint config / 单个 endpoint 配置
3. defaults config / defaults 配置
4. Program built-in defaults / 程序内置默认值

---

## 6. Output Formats / 输出格式

### 6.1 Table Format (Default) / Table 格式（默认）

Human-readable table format for terminal viewing.

人类可读的表格格式，用于终端查看。

**Normal Output / 正常输出：**

```
NAME              URL                                 STATUS    LATENCY
API Gateway       https://api.example.com/health      ✓ 200     45ms
Auth Service      https://auth.example.com/ping       ✓ 200     128ms
User Service      https://user.example.com/health     ✓ 200     67ms
Payment Service   https://payment.example.com/ping    ✗ 503     --
Database Proxy    https://db.internal:8080/status     ✗ timeout --

Summary: 3/5 healthy
```

**Design Rules / 设计规则：**

| Element / 元素 | Design / 设计 |
|---------------|---------------|
| Success marker / 成功标记 | `✓` green / 绿色 |
| Failure marker / 失败标记 | `✗` red / 红色 |
| Timeout/no response / 超时/无响应 | Show error type (timeout/connection refused) / 显示错误类型 |
| Latency display / 延迟显示 | In milliseconds, show `--` for failures / 毫秒为单位，失败显示 `--` |
| Column alignment / 列对齐 | Left-aligned, auto-calculate width / 左对齐，自动计算宽度 |
| Summary | Last line shows healthy/total / 最后一行显示 健康数/总数 |

**Color Control / 颜色控制：**

- Auto-detect terminal, enable colors if supported / 自动检测终端，支持颜色则启用
- `--no-color` forces colors off / 强制关闭颜色
- Auto-disable colors when output redirected to file / 输出重定向到文件时自动关闭颜色

### 6.2 JSON Format / JSON 格式

Machine-readable format for CI/CD integration and script processing.

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

**Field Descriptions / 字段说明：**

| Field | Type | Description / 说明 |
|-------|------|-------------------|
| timestamp | string | Check start time (ISO 8601) / 检查开始时间 |
| duration_ms | int | Total check duration (ms) / 整体检查耗时（毫秒） |
| summary.total | int | Total endpoint count / 端点总数 |
| summary.healthy | int | Healthy count / 健康数量 |
| summary.unhealthy | int | Unhealthy count / 不健康数量 |
| results[].name | string | Endpoint name / 端点名称 |
| results[].url | string | Checked URL / 检查的 URL |
| results[].healthy | bool | Is healthy / 是否健康 |
| results[].status_code | int/null | HTTP status code / HTTP 状态码 |
| results[].latency_ms | int/null | Response time (ms) / 响应时间（毫秒） |
| results[].error | string/null | Error message / 错误信息 |

### 6.3 Single URL Check Output / 单 URL 检查输出

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

### 6.4 Quiet Mode / 静默模式

`--quiet` flag enables quiet mode, no output, only exit code indicates result.

`--quiet` 参数启用静默模式，不输出任何内容，仅通过退出码表示结果。

```bash
healthcheck run -c endpoints.yaml -q
echo $?  # 0 or 1
```

---

## 7. Technical Implementation / 技术实现要点

### 7.1 Project Structure / 项目结构

```
healthcheck/
├── cmd/
│   ├── root.go           # Root command + global config / 根命令 + 全局配置
│   ├── check.go          # check subcommand / check 子命令
│   ├── run.go            # run subcommand / run 子命令
│   ├── config.go         # config subcommand group / config 子命令组
│   └── version.go        # version subcommand / version 子命令
├── internal/
│   ├── checker/
│   │   ├── checker.go    # Core check logic / 核心检查逻辑
│   │   └── checker_test.go
│   ├── config/
│   │   ├── config.go     # Config parsing / 配置解析
│   │   ├── validate.go   # Config validation / 配置校验
│   │   └── config_test.go
│   └── output/
│       ├── formatter.go  # Output interface definition / 输出接口定义
│       ├── table.go      # Table format implementation / Table 格式实现
│       └── json.go       # JSON format implementation / JSON 格式实现
├── main.go               # Entry point / 入口
├── go.mod
├── go.sum
├── .goreleaser.yaml      # Cross-platform build config / 跨平台构建配置
├── .github/
│   └── workflows/
│       ├── ci.yml        # CI pipeline / CI 流程
│       └── release.yml   # Release pipeline / 发布流程
├── README.md
└── LICENSE
```

### 7.2 Concurrency Model / 并发模型

Using goroutine + channel + semaphore pattern:

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
            sem <- struct{}{}        // Acquire semaphore / 获取信号量
            results <- c.Check(ep)
            <-sem                    // Release semaphore / 释放信号量
        }(ep)
    }

    wg.Wait()
    close(results)

    // Collect results in original order / 按原顺序收集结果
    // ...
}
```

### 7.3 Timeout Control / 超时控制

Using context for timeout:

使用 context 实现超时：

```go
func (c *Checker) Check(ep Endpoint) Result {
    ctx, cancel := context.WithTimeout(context.Background(), ep.Timeout)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "GET", ep.URL, nil)
    // ...
}
```

### 7.4 Retry Mechanism / 重试机制

```go
func (c *Checker) CheckWithRetry(ep Endpoint) Result {
    var result Result
    for i := 0; i <= ep.Retries; i++ {
        result = c.Check(ep)
        if result.Healthy {
            return result
        }
        // Optional: retry interval / 可选：重试间隔
    }
    return result
}
```

---

## 8. Development Plan / 开发计划

| Day | Task / 任务 | Output / 产出 |
|-----|------------|--------------|
| Day 1 | Project init, Cobra command framework, `check` command / 项目初始化、Cobra 命令框架、`check` 命令 | Can check single URL / 能检查单个 URL |
| Day 2 | Config file parsing (Viper), `run` command, concurrency / 配置文件解析(Viper)、`run` 命令、并发实现 | Can batch check / 能批量检查 |
| Day 3 | Timeout control, retry mechanism, JSON/Table output / 超时控制、重试机制、JSON/Table 输出 | Complete check functionality / 完整的检查功能 |
| Day 4 | Shell completion, GoReleaser config, unit tests / Shell 补全、GoReleaser 配置、单元测试 | Release-ready state / 可发布状态 |
| Day 5 | GitHub Actions CI, README docs, release v1.0 / GitHub Actions CI、README 文档、发布 v1.0 | Official release / 正式发布 |

---

## 9. Resume Highlights / 简历亮点

Skills demonstrated by completing this project:

完成此项目后，可在简历中体现以下技能：

| Skill / 技能点 | Demonstration / 体现 |
|---------------|---------------------|
| Go Concurrency / Go 并发编程 | goroutine + channel + semaphore for controlled concurrency / 可控并发实现 |
| CLI Development / CLI 开发 | Cobra + Viper framework best practices / 框架最佳实践 |
| Interface Abstraction / 接口抽象 | Output format interface design, supports extension / 输出格式的接口设计，支持扩展 |
| DevOps Mindset / DevOps 思维 | Exit code design, JSON output for CI/CD integration / 退出码设计、JSON 输出便于 CI/CD 集成 |
| Cross-platform Release / 跨平台发布 | GoReleaser auto-build multi-platform binaries / 自动构建多平台二进制 |
| CI/CD Practice / CI/CD 实践 | GitHub Actions automated testing and release / 自动化测试和发布 |

---

## 10. References / 参考资料

- [Cobra Documentation](https://cobra.dev/)
- [Viper Documentation](https://github.com/spf13/viper)
- [GoReleaser Documentation](https://goreleaser.com/)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
