# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## Project Overview

**gox** (Go eXtended utilities) - `github.com/f2xme/gox` 是对常用 Go 第三方库的二次封装，提供统一的 API 风格和开箱即用的配置。专注于提升开发效率和代码一致性。每个包基于成熟的开源库进行封装。

## Build & Test Commands

```bash
# Build all packages
go build ./...

# Run all tests
go test ./...

# Run a single test by name
go test ./path/to/pkg -run TestFunctionName

# Run tests with verbose output
go test -v ./...

# Run tests with race detector
go test -race ./...

# Lint (if golangci-lint is installed)
golangci-lint run ./...

# Format code
gofmt -w .
```

## Architecture

This is a multi-package utility library. Each top-level directory is an independent package that can be imported separately:

```
github.com/f2xme/gox/<package>
```

## Code Conventions

- Exported functions and types must have doc comments.
- Errors should be returned, not panicked. Use `fmt.Errorf` with `%w` for wrapping.
- Packages should be small and focused on a single concern.
- No package should import another package from this library (keep packages independent).
- Table-driven tests are preferred.
