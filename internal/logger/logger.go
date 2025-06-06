package logger

import (
	"fmt"
	"os"
)

// Logger provides methods for different log levels and formatted output
type Logger struct {
	// Add any configuration fields here if needed in the future
}

// New creates a new Logger instance
func New() *Logger {
	return &Logger{}
}

// Info prints an informational message to stdout
func (l *Logger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("ℹ️ %s\n", msg)
}

// Success prints a success message to stdout
func (l *Logger) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("✅ %s\n", msg)
}

// Warning prints a warning message to stderr
func (l *Logger) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "⚠️  %s\n", msg)
}

// Error prints an error message to stderr
func (l *Logger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "❌ %s\n", msg)
}

// Fatal prints an error message to stderr and exits with status code 1
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.Error(format, args...)
	os.Exit(1)
}

// Fatalf is equivalent to Error followed by a call to os.Exit(1)
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Fatal(format, args...)
}

// Printf formats according to a format specifier and writes to stdout
func (l *Logger) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Println formats using the default formats for its operands and writes to stdout
func (l *Logger) Println(args ...interface{}) {
	fmt.Println(args...)
}

// Default logger instance for package-level functions
var defaultLogger = New()

// Info prints an informational message using the default logger
func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Success prints a success message using the default logger
func Success(format string, args ...interface{}) {
	defaultLogger.Success(format, args...)
}

// Warning prints a warning message using the default logger
func Warning(format string, args ...interface{}) {
	defaultLogger.Warning(format, args...)
}

// Error prints an error message using the default logger
func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

// Fatal prints an error message and exits with status code 1 using the default logger
func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

// Fatalf is equivalent to Error followed by a call to os.Exit(1) using the default logger
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

// Printf formats according to a format specifier and writes to stdout using the default logger
func Printf(format string, args ...interface{}) {
	defaultLogger.Printf(format, args...)
}

// Println formats using the default formats for its operands and writes to stdout using the default logger
func Println(args ...interface{}) {
	defaultLogger.Println(args...)
}
