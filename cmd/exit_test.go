package cmd

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestExitWithErrorLogsAndExits(t *testing.T) {
	oldExit := exit
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() returned error: %v", err)
	}
	oldStderr := os.Stderr

	t.Cleanup(func() {
		exit = oldExit
		os.Stderr = oldStderr
	})

	os.Stderr = w

	var code int
	exit = func(got int) {
		code = got
	}

	exitWithError("Command failed: %s", "boom")

	if err := w.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}

	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll() returned error: %v", err)
	}

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(string(output), "Command failed: boom") {
		t.Fatalf("expected error output to contain command failure, got %q", string(output))
	}
	if !strings.Contains(string(output), "❌") {
		t.Fatalf("expected error output prefix, got %q", string(output))
	}
}
