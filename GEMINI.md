# Best Practices for Developing Golang CLIs with Cobra

This document outlines best practices for developing robust, maintainable, and user-friendly command-line interfaces (CLIs) in Go using the Cobra library.

## Project Structure

A well-organized project structure is crucial for maintainability.

*   **`cobra-cli` Tool:** Bootstrap your application using the `cobra-cli` tool for a standardized layout: `cobra-cli init --pkg-name <your-app>`
*   **`cmd/` Directory:** This is the core of your application.
    *   `cmd/root.go`: Contains the root command definition, global flags, and persistent settings.
    *   `cmd/subcommand.go`: Each subcommand should be in its own file (e.g., `cmd/add.go`). Use `cobra-cli add <command-name>` to create new commands.
*   **`main.go`:** The main entry point of the application, which should typically only call `cmd.Execute()`.
*   **`internal/` and `pkg/`:** Separate your business logic into these directories.
    *   `internal/`: For code that should not be imported by other projects.
    *   `pkg/`: For code that can be safely imported by external applications.

## Command and Flag Design

*   **Clear Naming:** Use clear and consistent names for commands.
*   **Subcommands:** Organize functionality into subcommands for logical grouping (e.g., `app config create`).
*   **Persistent vs. Local Flags:**
    *   **Persistent Flags:** Available to a command and all its subcommands (e.g., `--verbose`, `--config`).
    *   **Local Flags:** Only apply to a specific command.
*   **Required Flags:** Mark flags as required if a command cannot run without them.

## Configuration Management with Viper

Viper is recommended for handling configuration from files, environment variables, and flags.

*   **Integration:** Viper can be easily integrated to manage configuration precedence (flags > env variables > config file > defaults).
*   **Binding Flags:** Bind Cobra flags to Viper to allow command-line flags to override configuration values.
*   **Automatic Loading:** Viper can automatically read from a configuration file and environment variables.

## Error Handling

*   **Use `RunE`:** Prefer `RunE` for your command's `Run` function, as it returns an error for cleaner error handling.
*   **Meaningful Messages:** Provide clear and actionable error messages.

## Testing

*   **Unit Tests:** Write unit tests for each command.
*   **Test Helper:** Create a helper function to execute commands and capture output for assertions.
*   **Isolate Dependencies:** Separate business logic from command definitions to test core functionality independently of the Cobra framework.

## User Experience

*   **Help and Usage Text:** Write clear and comprehensive `Short` and `Long` descriptions for your commands.
*   **Command Aliases:** Provide aliases for common commands.
*   **Intelligent Suggestions:** Cobra provides "did you mean..." suggestions for mistyped commands.
