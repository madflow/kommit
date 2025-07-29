package cmd

import (
	"testing"
)

func TestSanitizeBranchName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid characters",
			input:    "feat/test-branch_123",
			expected: "feat/test-branch_123",
		},
		{
			name:     "invalid characters",
			input:    "feat/test branch with spaces and invalid chars!@#$",
			expected: "feat/testbranchwithspacesandinvalidchars",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only invalid characters",
			input:    "!@#$%^&*()",
			expected: "",
		},
		{
			name:     "branch name too long",
			input:    "feat/this-is-a-very-long-branch-name-that-should-be-truncated",
			expected: "feat/this-is-a-very-long-branch-name-that-should-b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := sanitizeBranchName(tt.input)

			if actual != tt.expected {
				t.Errorf("expected %q, but got %q", tt.expected, actual)
			}
		})
	}
}
