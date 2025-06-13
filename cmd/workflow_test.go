package cmd

import "testing"

func TestExtractIssueID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bare issue ID",
			input:    "DEL-163",
			expected: "DEL-163",
		},
		{
			name:     "full Linear URL",
			input:    "https://linear.app/company/issue/DEL-163",
			expected: "DEL-163",
		},
		{
			name:     "Linear URL with query params",
			input:    "https://linear.app/company/issue/ENG-42?tab=comments",
			expected: "ENG-42",
		},
		{
			name:     "different team prefix",
			input:    "BACKEND-123",
			expected: "BACKEND-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractIssueID(tt.input)
			if result != tt.expected {
				t.Errorf("extractIssueID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTTPS URL",
			input:    "https://github.com/username/repo",
			expected: "repo",
		},
		{
			name:     "HTTPS URL with .git",
			input:    "https://github.com/username/repo.git",
			expected: "repo",
		},
		{
			name:     "SSH URL",
			input:    "git@github.com:username/repo.git",
			expected: "repo",
		},
		{
			name:     "complex repo name",
			input:    "https://github.com/org/my-awesome-project",
			expected: "my-awesome-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRepoName(tt.input)
			if result != tt.expected {
				t.Errorf("extractRepoName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
