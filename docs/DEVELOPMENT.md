# Development Guide

## Prerequisites

- Go 1.21+
- [buf](https://buf.build/) (for proto linting)
- [golangci-lint](https://golangci-lint.run/) (for Go linting)

## Getting Started

```bash
git clone https://github.com/balcieren/connect-go-errors.git
cd connect-go-errors
go mod download
make test
```

## Project Structure

- `error.go` - Main API: `New()`, `NewWithMessage()`, `FromCode()`, `Wrap()`
- `registry.go` - Error definitions and constants
- `template.go` - Template parsing and substitution
- `proto/connect/error.proto` - Proto extension definition
- `cmd/protoc-gen-connect-errors/` - Protoc plugin
- `internal/codegen/` - Code generation templates
- `internal/parser/` - Proto parsing utilities
- `examples/` - Usage examples

## Running Tests

```bash
make test          # Run all tests with race detection
go test -cover ./... # With coverage
```

## Building the Plugin

```bash
make build         # Build to bin/
make install       # Install to GOPATH/bin
```

## Code Style

- Follow standard Go conventions
- Run `gofmt -s -w .` before committing
- Run `golangci-lint run` to check for issues
