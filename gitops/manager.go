package gitops

import (
        "fmt"
        "os"
        "os/exec"
        "path/filepath"
        "regexp"
        "strings"
)

func PrepareRepository(repoPath, baseBranch string) error {
        if err := validateGitRepository(repoPath); err != nil {
                return fmt.Errorf("not a git repository: %w", err)
        }

        if !branchExists(repoPath, baseBranch) {
                return fmt.Errorf("branch %s does not exist", baseBranch)
        }

        cmd := exec.Command("git", "checkout", baseBranch)
        cmd.Dir = repoPath
        if err := cmd.Run(); err != nil {
                return fmt.Errorf("failed to checkout branch %s: %w", baseBranch, err)
        }

        if hasRemote(repoPath) {
                cmd = exec.Command("git", "pull")
                cmd.Dir = repoPath
                if err := cmd.Run(); err != nil {
                        return fmt.Errorf("failed to pull latest changes: %w", err)
                }
        }

        return nil
}

func CreateWorktreeForIssue(repoPath, issueID, baseBranch string) (string, error) {
        if err := validateGitRepository(repoPath); err != nil {
                return "", fmt.Errorf("not a git repository: %w", err)
        }

        if !branchExists(repoPath, baseBranch) {
                return "", fmt.Errorf("base branch %s does not exist", baseBranch)
        }

        branchName := SanitizeBranchName(issueID)
        worktreesDir := filepath.Join(repoPath, "worktrees")
        worktreePath := filepath.Join(worktreesDir, issueID)

        if _, err := os.Stat(worktreePath); err == nil {
                return "", fmt.Errorf("worktree for issue %s already exists at %s", issueID, worktreePath)
        }

        if err := os.MkdirAll(worktreesDir, 0755); err != nil {
                return "", fmt.Errorf("failed to create worktrees directory: %w", err)
        }

        cmd := exec.Command("git", "worktree", "add", "-b", branchName, worktreePath, baseBranch)
        cmd.Dir = repoPath
        if err := cmd.Run(); err != nil {
                return "", fmt.Errorf("failed to create worktree for issue %s: %w", issueID, err)
        }

        return worktreePath, nil
}

func SanitizeBranchName(name string) string {
        if name == "" {
                return "issue"
        }

        reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
        sanitized := reg.ReplaceAllString(name, "-")

        sanitized = strings.Trim(sanitized, "-")

        if sanitized == "" {
                return "issue"
        }

        return sanitized
}

func validateGitRepository(path string) error {
        cmd := exec.Command("git", "rev-parse", "--git-dir")
        cmd.Dir = path
        return cmd.Run()
}

func branchExists(repoPath, branchName string) bool {
        cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
        cmd.Dir = repoPath
        return cmd.Run() == nil
}

func hasRemote(repoPath string) bool {
        cmd := exec.Command("git", "remote")
        cmd.Dir = repoPath
        output, err := cmd.Output()
        return err == nil && len(strings.TrimSpace(string(output))) > 0
}
