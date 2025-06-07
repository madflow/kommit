---
trigger: always_on
---

# General Rules

## General

- **Do not change the default model**: "qwen2.5-coder:7b"
- **Do not change the default prompt text**
- **Make only changes, which are crucial for the task**
- **Do not add or change non code-related text, unless special instructions are given**

## Repository Structure

- **Commands in `cmd/`**: All CLI commands go in `cmd/` directory, business logic goes in `internal/`
- **One command per file**: Each command gets its own file in `cmd/` named after the command (e.g., `cmd/serve.go`)
- **Domain-based internal packages**: Organize `internal/` by function domains like `internal/git/`, `internal/config/`
- **Keep main.go minimal**: Only application bootstrap code in `main.go`, no business logic
- **Flat cmd structure**: Don't nest command files deeply, keep `cmd/` directory flat

## Cobra CLI Patterns

- **Register commands in init()**: Add subcommands to root using `rootCmd.AddCommand()` in `init()` functions
- **Clear command descriptions**: Every command needs `Use`, `Short`, and `Long` descriptions
- **Flag binding**: Bind all flags to Viper using `viper.BindPFlag()` in command `init()` functions
- **No business logic in commands**: Command files handle CLI interface only, delegate work to `internal/` packages

## Error Handling & Code Quality

- **Always handle errors**: Never ignore errors, use `if err != nil` checks everywhere
- **Wrap errors with context**: Use `fmt.Errorf("operation failed: %w", err)` to add context
- **Early returns**: Use early returns instead of deep nesting (`if err != nil { return err }`)
- **Small focused functions**: Keep functions short and single-purpose
- **Meaningful names**: Use descriptive names for variables, functions, and packages

## Testing & Documentation

- **Test internal packages**: Write unit tests for all business logic in `internal/` packages
- **Table-driven tests**: Use table-driven tests for multiple test cases with different inputs
- **Document exported functions**: Add comments to all exported functions starting with function name
- **Do not use external 3rd-party testing frameworks**: Only use built-in `testing` package

## Git Operations & Output

- **Wrap git commands**: Use `os/exec` with proper error handling for git operations in `internal/git/`
- **Structured return data**: Return Go structs from internal functions, not raw command output strings
