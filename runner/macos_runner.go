package runner

import (
        "fmt"
        "os/exec"
        "strings"
)

type MacOSRunner struct{}

func NewMacOSRunner() *MacOSRunner {
        return &MacOSRunner{}
}

func (r *MacOSRunner) LaunchFridayInMacOSTerminal(worktreePath, issueTitle string) error {
        script := GenerateAppleScript(worktreePath, issueTitle)
        
        cmd := exec.Command("osascript", "-e", script)
        if err := cmd.Run(); err != nil {
                return fmt.Errorf("failed to execute osascript: %w", err)
        }
        
        return nil
}

func GenerateAppleScript(worktreePath, issueTitle string) string {
        escapedPath := escapeAppleScriptString(worktreePath)
        
        // Extract issue ID from worktree path for tab title
        pathParts := strings.Split(worktreePath, "/")
        issueID := pathParts[len(pathParts)-1]
        tabTitle := escapeAppleScriptString(fmt.Sprintf("%s: %s", issueID, issueTitle))
        
        script := fmt.Sprintf(`tell application "Terminal"
        do script "cd '%s' && friday"
        set custom title of tab 1 of front window to "%s"
end tell`, escapedPath, tabTitle)
        
        return script
}

func escapeAppleScriptString(s string) string {
        // Escape backslashes first, then quotes
        s = strings.ReplaceAll(s, "\\", "\\\\")
        s = strings.ReplaceAll(s, "\"", "\\\"")
        return s
}
