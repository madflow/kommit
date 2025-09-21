# Agent Instructions for Kommit

## Build/Test Commands
- **Build**: `make build` or `go build`
- **Test all**: `make test` or `go test -v ./...`
- **Test single package**: `go test -v ./cmd` or `go test -v ./internal/config`
- **Lint**: `golangci-lint run` (uses config from `.golangci.yml`)
- **Format**: Uses `gofumpt` formatter

## Code Style Guidelines
- **Imports**: Group standard library, third-party, then local packages with blank lines between groups
- **Naming**: Use camelCase for variables/functions, PascalCase for exported types/functions
- **Error handling**: Always check and handle errors explicitly, use `logger.Fatal()` for critical errors
- **Types**: Use struct tags for config binding (`mapstructure:"field_name"`)
- **Comments**: Add package comments and exported function/type comments following Go conventions
- **Variables**: Use descriptive names, avoid abbreviations except for common ones (err, ctx, cfg)

## Architecture
- **CLI**: Uses Cobra framework with Viper for configuration
- **Config**: Supports multiple config file locations (`.kommit.yaml` in project/git dir, XDG dirs)
- **Logging**: Use `internal/logger` package with structured levels (Info, Success, Error, Fatal)
- **Git operations**: Centralized in `internal/git` package
- **Testing**: Table-driven tests following Go conventions

## Key Patterns
- Initialize with `cobra.OnInitialize(initConfig)` pattern
- Use `viper` for configuration with defaults and environment variable support
- Error messages should be user-friendly and actionable