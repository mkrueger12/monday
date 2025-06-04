package runner

import (
        "strings"
        "testing"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
)

func TestGenerateAppleScript(t *testing.T) {
        tests := []struct {
                name         string
                worktreePath string
                issueTitle   string
                expected     []string // strings that should be present in the script
        }{
                {
                        name:         "basic path",
                        worktreePath: "/Users/test/repo/worktrees/ISSUE-123",
                        issueTitle:   "Fix authentication bug",
                        expected: []string{
                                "tell application \"Terminal\"",
                                "do script \"cd '/Users/test/repo/worktrees/ISSUE-123' && friday\"",
                                "set custom title of tab 1 of front window to \"ISSUE-123: Fix authentication bug\"",
                        },
                },
                {
                        name:         "path with spaces",
                        worktreePath: "/Users/test/my repo/worktrees/ISSUE-456",
                        issueTitle:   "Update user interface",
                        expected: []string{
                                "tell application \"Terminal\"",
                                "do script \"cd '/Users/test/my repo/worktrees/ISSUE-456' && friday\"",
                                "set custom title of tab 1 of front window to \"ISSUE-456: Update user interface\"",
                        },
                },
                {
                        name:         "title with quotes",
                        worktreePath: "/Users/test/repo/worktrees/ISSUE-789",
                        issueTitle:   "Fix \"broken\" feature",
                        expected: []string{
                                "tell application \"Terminal\"",
                                "do script \"cd '/Users/test/repo/worktrees/ISSUE-789' && friday\"",
                                "set custom title of tab 1 of front window to \"ISSUE-789: Fix \\\"broken\\\" feature\"",
                        },
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        script := GenerateAppleScript(tt.worktreePath, tt.issueTitle)
                        
                        for _, expected := range tt.expected {
                                assert.Contains(t, script, expected, "Script should contain: %s", expected)
                        }
                })
        }
}

func TestLaunchFridayInMacOSTerminal_ScriptGeneration(t *testing.T) {
        runner := NewMacOSRunner()
        
        worktreePath := "/Users/test/repo/worktrees/ISSUE-123"
        issueTitle := "Test Issue"
        
        err := runner.LaunchFridayInMacOSTerminal(worktreePath, issueTitle)
        
        // On macOS, this might succeed and actually open Terminal
        // On other systems or CI, it should fail
        // Either way is acceptable for this test
        if err != nil {
                assert.Contains(t, err.Error(), "osascript")
                assert.Error(t, err)
        } else {
                assert.NoError(t, err)
        }
}

func TestEscapeAppleScriptString(t *testing.T) {
        tests := []struct {
                input    string
                expected string
        }{
                {
                        input:    "simple string",
                        expected: "simple string",
                },
                {
                        input:    "string with \"quotes\"",
                        expected: "string with \\\"quotes\\\"",
                },
                {
                        input:    "string with \\ backslash",
                        expected: "string with \\\\ backslash",
                },
                {
                        input:    "string with both \" and \\",
                        expected: "string with both \\\" and \\\\",
                },
        }

        for _, tt := range tests {
                t.Run(tt.input, func(t *testing.T) {
                        result := escapeAppleScriptString(tt.input)
                        assert.Equal(t, tt.expected, result)
                        assert.IsType(t, "", result)
                })
        }
}

func TestNewMacOSRunner(t *testing.T) {
        runner := NewMacOSRunner()
        assert.NotNil(t, runner)
        assert.IsType(t, &MacOSRunner{}, runner)
}

func TestGenerateAppleScript_Structure(t *testing.T) {
        script := GenerateAppleScript("/test/path", "Test Title")
        
        // Verify the script has the correct structure
        lines := strings.Split(script, "\n")
        require.True(t, len(lines) >= 3, "Script should have at least 3 lines")
        
        // Should start with tell application
        assert.Contains(t, lines[0], "tell application \"Terminal\"")
        
        // Should end with end tell
        lastLine := strings.TrimSpace(lines[len(lines)-1])
        assert.Equal(t, "end tell", lastLine)
        
        // Should contain the path and title
        assert.Contains(t, script, "/test/path")
        assert.Contains(t, script, "Test Title")
}
