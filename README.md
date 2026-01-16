# Health Check CLI

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)
![Platform](https://img.shields.io/badge/Platform-Linux%20|%20macOS%20|%20Windows-lightgrey?style=flat-square)

## Overview / 项目概述

A command-line tool written in Go for batch HTTP endpoint health checking. Supports concurrent checks, multiple output formats, and configuration file management. Perfect for deployment verification, troubleshooting, and CI/CD integration.

用 Go 语言编写的命令行工具，用于批量检查 HTTP 端点的健康状态。支持并发检查、多种输出格式、配置文件管理，适用于部署验证、故障排查、CI/CD 集成等场景。

## Features / 功能特性

- **Single URL Check / 单 URL 检查**: Quick health check for individual endpoints
- **Batch Check / 批量检查**: Concurrent checking of multiple endpoints from config file
- **Multiple Output Formats / 多种输出格式**: Table (human-readable) and JSON (machine-readable)
- **Configurable Timeout / 超时控制**: Customizable request timeout per endpoint
- **Retry Mechanism / 重试机制**: Automatic retry on failure
- **Custom Headers / 自定义 Headers**: Support for authentication and custom request headers
- **SSL Verification Skip / SSL 证书跳过**: Support for self-signed certificates
- **Exit Codes / 退出码**: CI/CD friendly exit codes for automation
- **Cross-Platform / 跨平台**: Supports Linux, macOS, and Windows

## Quick Start / 快速开始

### Installation / 安装

```bash
# Using Go install
go install github.com/r1ckyIn/healthcheck-cli@latest

# Or download from releases
# https://github.com/r1ckyIn/healthcheck-cli/releases
```

### Basic Usage / 基本用法

```bash
# Check single URL / 检查单个 URL
healthcheck check https://api.example.com/health

# Check with custom timeout / 自定义超时
healthcheck check https://api.example.com/health --timeout 10s

# Batch check from config file / 从配置文件批量检查
healthcheck run -c endpoints.yaml

# JSON output for CI/CD / JSON 输出用于 CI/CD
healthcheck run -c endpoints.yaml -o json
```

## Project Structure / 项目结构

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
├── docs/                  # Documentation
├── tests/                 # Test files
├── main.go               # Entry point
├── go.mod
├── .goreleaser.yaml      # Cross-platform build config
└── README.md
```

## Configuration / 配置文件

Create an `endpoints.yaml` file:

```yaml
# Global defaults / 全局默认配置
defaults:
  timeout: 5s
  retries: 2
  expected_status: 200

# Endpoint list / 端点列表
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

## Command Reference / 命令参考

| Command / 命令 | Description / 描述 |
|---------------|-------------------|
| `healthcheck check <url>` | Check single URL / 检查单个 URL |
| `healthcheck run` | Batch check from config / 从配置批量检查 |
| `healthcheck config init` | Generate sample config / 生成示例配置 |
| `healthcheck config validate` | Validate config file / 校验配置文件 |
| `healthcheck completion <shell>` | Generate shell completion / 生成 Shell 补全 |
| `healthcheck version` | Show version info / 显示版本信息 |

## Exit Codes / 退出码

| Code / 码 | Meaning / 含义 |
|-----------|---------------|
| 0 | All services healthy / 所有服务健康 |
| 1 | Some services unhealthy / 有服务不健康 |
| 2 | Configuration error / 配置错误 |

## Tech Stack / 技术栈

- **Language / 语言**: Go
- **CLI Framework / CLI 框架**: [Cobra](https://cobra.dev/)
- **Config Management / 配置管理**: [Viper](https://github.com/spf13/viper)
- **Concurrency / 并发模型**: goroutine + channel + semaphore
- **Build / 构建**: [GoReleaser](https://goreleaser.com/)
- **CI/CD**: GitHub Actions

## Contributing / 贡献指南

Contributions are welcome! Please feel free to submit a Pull Request.

欢迎贡献！请随时提交 Pull Request。

1. Fork the repository / Fork 本仓库
2. Create your feature branch / 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. Commit your changes / 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch / 推送到分支 (`git push origin feature/amazing-feature`)
5. Open a Pull Request / 开启 Pull Request

## License / 许可证

MIT License - see [LICENSE](LICENSE) file for details.

MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## Author / 作者

**Ricky** - CS Student @ University of Sydney

- GitHub: [@r1ckyIn](https://github.com/r1ckyIn)
- Email: rickyqin919@gmail.com
