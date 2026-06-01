package cmd

import (
	"os"

	"github.com/madflow/kommit/internal/logger"
)

var exit = os.Exit

func exitWithError(format string, args ...any) {
	logger.Error(format, args...)
	exit(1)
}
