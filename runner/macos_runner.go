// Package runner provides platform-specific functionality for launching development tools.
// Currently focused on macOS Terminal integration using AppleScript to automate
// the process of opening new Terminal tabs and launching the Friday CLI.
package runner

import (
        "fmt"
        "os/exec"
        "strings"
)

// MacOSRunner handles launching development tools in macOS Terminal.
// It uses AppleScript to automate Terminal operations and provide a seamless
// developer experience when starting work on Linear issues.
type MacOSRunner struct{}

// NewMacOSRunner creates a new instance of the macOS Terminal runner.
// This runner is specifically designed for macOS environments and uses
// AppleScript for Terminal automation.
func NewMacOSRunner() *MacOSRunner {
        return &MacOSRunner{}
}

// LaunchFridayInMacOSTerminal opens a new Terminal tab, navigates to the worktree,
// and launches the Friday CLI. It also sets a descriptive tab title that includes
// the issue ID and title for easy identification when working on multiple issues.
func (r *MacOSRunner) LaunchFridayInMacOSTerminal(worktreePath, issueTitle string) error {
        // Generate the AppleScript that will automate Terminal operations
        script := GenerateAppleScript(worktreePath, issueTitle)
        
        // Execute the AppleScript using macOS's osascript command
        cmd := exec.Command("osascript", "-e", script)
        if err := cmd.Run(); err != nil {
                return fmt.Errorf("failed to execute osascript: %w", err)
        }
        
        return nil
}

// GenerateAppleScript creates an AppleScript that opens a new Terminal tab,
// changes to the worktree directory, launches Friday CLI, and sets a descriptive tab title.
// The script is carefully formatted to handle paths and titles with special characters.
func GenerateAppleScript(worktreePath, issueTitle string) string {
        // Escape the worktree path for safe use in AppleScript
        escapedPath := escapeAppleScriptString(worktreePath)
        
        // Extract issue ID from worktree path for tab title
        // The worktree path typically ends with the issue ID (e.g., /path/to/worktrees/DEL-163)
        pathParts := strings.Split(worktreePath, "/")
        issueID := pathParts[len(pathParts)-1]
        
        // Create a descriptive tab title combining issue ID and title
        tabTitle := escapeAppleScriptString(fmt.Sprintf("%s: %s", issueID, issueTitle))
        
        // Generate the complete AppleScript
        // This script tells Terminal to:
        // 1. Execute a command that changes directory and runs Friday
        // 2. Set a custom title for the new tab
        script := fmt.Sprintf(`tell application "Terminal"
        do script "cd '%s' && friday"
        set custom title of tab 1 of front window to "%s"
end tell`, escapedPath, tabTitle)
        
        return script
}

// escapeAppleScriptString properly escapes strings for safe use in AppleScript.
// It handles backslashes and quotes which are common in file paths and titles
// and could break AppleScript execution if not properly escaped.
func escapeAppleScriptString(s string) string {
        // Escape backslashes first (must be done before quotes to avoid double-escaping)
        s = strings.ReplaceAll(s, "\\", "\\\\")
        // Escape double quotes to prevent breaking the AppleScript string literals
        s = strings.ReplaceAll(s, "\"", "\\\"")
        return s
}
