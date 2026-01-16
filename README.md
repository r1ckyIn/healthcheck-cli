# Health Check CLI

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20|%20macOS%20|%20Windows-lightgrey?style=flat-square)]()

**A CLI tool for batch HTTP endpoint health checking written in Go**

[English](#english) | [中文](#中文)

</div>

---

## English

### Overview

Health Check CLI is a command-line tool written in Go for batch HTTP endpoint health checking. It supports concurrent checks, multiple output formats, and configuration file management. Perfect for deployment verification, troubleshooting, and CI/CD integration.

### Features

- **Single URL Check**: Quick health check for individual endpoints
- **Batch Check**: Concurrent checking of multiple endpoints from config file
- **Multiple Output Formats**: Table (human-readable) and JSON (machine-readable)
- **Configurable Timeout**: Customizable request timeout per endpoint
- **Retry Mechanism**: Automatic retry on failure
- **Custom Headers**: Support for authentication and custom request headers
- **SSL Verification Skip**: Support for self-signed certificates
- **Exit Codes**: CI/CD friendly exit codes for automation
- **Cross-Platform**: Supports Linux, macOS, and Windows

### Quick Start

```bash
# Using Go install
go install github.com/r1ckyIn/healthcheck-cli@latest

# Or download from releases
# https://github.com/r1ckyIn/healthcheck-cli/releases
```

### Usage

```bash
# Check single URL
healthcheck check https://api.example.com/health

# Check with custom timeout
healthcheck check https://api.example.com/health --timeout 10s

# Batch check from config file
healthcheck run -c endpoints.yaml

# JSON output for CI/CD
healthcheck run -c endpoints.yaml -o json
```

### Configuration

Create an `endpoints.yaml` file:

```yaml
# Global defaults
defaults:
  timeout: 5s
  retries: 2
  expected_status: 200

# Endpoint list
endpoints:
  - name: "API Gateway"
    url: "https://api.example.com/health"
    timeout: 10s

  - name: "Auth Service"
    url: "https://auth.example.com/ping"
    headers:
      Authorization: "Bearer ${AUTH_TOKEN}"

  - name: "Internal Service"
    url: "https://internal.local:8443/health"
    insecure: true
```

### Command Reference

| Command | Description |
|---------|-------------|
| `healthcheck check <url>` | Check single URL |
| `healthcheck run` | Batch check from config |
| `healthcheck config init` | Generate sample config |
| `healthcheck config validate` | Validate config file |
| `healthcheck completion <shell>` | Generate shell completion |
| `healthcheck version` | Show version info |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All services healthy |
| 1 | Some services unhealthy |
| 2 | Configuration error |

### Project Structure

```
healthcheck-cli/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command + global config
│   ├── check.go           # check subcommand
│   ├── run.go             # run subcommand
│   ├── config.go          # config subcommand group
│   └── version.go         # version subcommand
├── internal/
│   ├── checker/           # Core health check logic
│   ├── config/            # Configuration parsing
│   └── output/            # Output formatters
├── .github/workflows/     # CI/CD pipelines
├── main.go               # Entry point
└── .goreleaser.yaml      # Cross-platform build config
```

### Tech Stack

- **Language**: Go
- **CLI Framework**: [Cobra](https://cobra.dev/)
- **Config Management**: [Viper](https://github.com/spf13/viper)
- **Concurrency**: goroutine + channel + semaphore
- **Build**: [GoReleaser](https://goreleaser.com/)
- **CI/CD**: GitHub Actions

---

## 中文

### 项目概述

Health Check CLI 是一个用 Go 语言编写的命令行工具，用于批量检查 HTTP 端点的健康状态。支持并发检查、多种输出格式、配置文件管理，适用于部署验证、故障排查、CI/CD 集成等场景。

### 功能特性

- **单 URL 检查**：快速检查单个端点健康状态
- **批量检查**：从配置文件并发检查多个端点
- **多种输出格式**：Table（人类可读）和 JSON（机器可读）
- **超时控制**：可自定义每个端点的请求超时时间
- **重试机制**：失败时自动重试
- **自定义 Headers**：支持认证和自定义请求头
- **SSL 证书跳过**：支持自签名证书
- **退出码**：CI/CD 友好的退出码，便于自动化
- **跨平台**：支持 Linux、macOS、Windows

### 快速开始

```bash
# 使用 Go install 安装
go install github.com/r1ckyIn/healthcheck-cli@latest

# 或从 releases 下载
# https://github.com/r1ckyIn/healthcheck-cli/releases
```

### 使用方法

```bash
# 检查单个 URL
healthcheck check https://api.example.com/health

# 自定义超时时间
healthcheck check https://api.example.com/health --timeout 10s

# 从配置文件批量检查
healthcheck run -c endpoints.yaml

# JSON 输出用于 CI/CD
healthcheck run -c endpoints.yaml -o json
```

### 命令参考

| 命令 | 说明 |
|------|------|
| `healthcheck check <url>` | 检查单个 URL |
| `healthcheck run` | 从配置批量检查 |
| `healthcheck config init` | 生成示例配置 |
| `healthcheck config validate` | 校验配置文件 |
| `healthcheck completion <shell>` | 生成 Shell 补全 |
| `healthcheck version` | 显示版本信息 |

### 退出码

| 码 | 含义 |
|----|------|
| 0 | 所有服务健康 |
| 1 | 有服务不健康 |
| 2 | 配置错误 |

### 技术栈

- **语言**：Go
- **CLI 框架**：[Cobra](https://cobra.dev/)
- **配置管理**：[Viper](https://github.com/spf13/viper)
- **并发模型**：goroutine + channel + semaphore
- **构建**：[GoReleaser](https://goreleaser.com/)
- **CI/CD**：GitHub Actions

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Author

**Ricky** - CS Student @ University of Sydney

[![GitHub](https://img.shields.io/badge/GitHub-r1ckyIn-181717?style=flat-square&logo=github)](https://github.com/r1ckyIn)

Interested in Cloud Engineering & DevOps
