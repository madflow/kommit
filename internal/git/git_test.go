package git

import (
	"os/exec"
	"testing"
)

// TestHasStagedChanges tests the HasStagedChanges function with various scenarios
func TestHasStagedChanges(t *testing.T) {
	// Save original execCommand and restore it after the test
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	tests := []struct {
		name     string
		setup    func()
		expected bool
		hasError bool
	}{
		{
			name: "has staged changes",
			setup: func() {
				execCommand = func(name string, arg ...string) *exec.Cmd {
					cmd := exec.Command("echo", "M\tfile1.go\nA\tfile2.go")
					return cmd
				}
			},
			expected: true,
			hasError: false,
		},
		{
			name: "no staged changes",
			setup: func() {
				execCommand = func(name string, arg ...string) *exec.Cmd {
					return exec.Command("echo", "")
				}
			},
			expected: false,
			hasError: false,
		},
		{
			name: "git error",
			setup: func() {
				execCommand = func(name string, arg ...string) *exec.Cmd {
					cmd := exec.Command("false") // This will return non-zero exit code
					return cmd
				}
			},
			expected: false,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the mock
			tt.setup()

			// Run the function under test
			result, err := HasStagedChanges()

			// Check the error
			if (err != nil) != tt.hasError {
				t.Errorf("HasStagedChanges() error = %v, hasError %v", err, tt.hasError)
				return
			}

			// Check the result
			if result != tt.expected {
				t.Errorf("HasStagedChanges() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsGitRepo tests the IsGitRepo function
func TestIsGitRepo(t *testing.T) {
	// Save original execCommand and restore it after the test
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	tests := []struct {
		name     string
		setup    func()
		expected bool
	}{
		{
			name: "is git repo",
			setup: func() {
				execCommand = func(name string, arg ...string) *exec.Cmd {
					// Simulate successful git rev-parse --git-dir
					return exec.Command("true")
				}
			},
			expected: true,
		},
		{
			name: "not a git repo",
			setup: func() {
				execCommand = func(name string, arg ...string) *exec.Cmd {
					// Simulate git rev-parse --git-dir failing with non-zero exit code
					return exec.Command("false")
				}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the mock
			tt.setup()

			// Run the function under test
			result := IsGitRepo()

			// Check the result
			if result != tt.expected {
				t.Errorf("IsGitRepo() = %v, want %v", result, tt.expected)
			}
		})
	}
}
