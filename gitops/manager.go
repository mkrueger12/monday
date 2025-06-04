// Package gitops provides git repository management functionality for the monday CLI tool.
// It handles git worktree creation, repository preparation, branch management, and
// feature file creation to support isolated development environments for Linear issues.
package gitops

import (
        "fmt"
        "os"
        "os/exec"
        "path/filepath"
        "regexp"
        "strings"
)

// PrepareRepository ensures the git repository is ready for worktree creation.
// It validates the repository, checks out the base branch, and pulls the latest changes
// if a remote is configured. This ensures all worktrees start from the most recent code.
func PrepareRepository(repoPath, baseBranch string) error {
        fmt.Printf("[DEBUG] PrepareRepository called with repoPath: %s, baseBranch: %s\n", repoPath, baseBranch)
        
        // Verify that the provided path is a valid git repository
        if err := validateGitRepository(repoPath); err != nil {
                return fmt.Errorf("not a git repository: %w", err)
        }

        // Ensure the base branch exists before attempting to use it
        if !branchExists(repoPath, baseBranch) {
                return fmt.Errorf("branch %s does not exist", baseBranch)
        }

        // Switch to the base branch to ensure we're starting from the correct point
        cmd := exec.Command("git", "checkout", baseBranch)
        cmd.Dir = repoPath
        fmt.Printf("[DEBUG] Running 'git checkout %s' in directory: %s\n", baseBranch, repoPath)
        if output, err := cmd.CombinedOutput(); err != nil {
                return fmt.Errorf("failed to checkout branch %s: %s (exit code: %w)", baseBranch, string(output), err)
        }

        // Pull latest changes if the repository has a remote configured
        // This ensures worktrees are created from the most up-to-date code
        if hasRemote(repoPath) {
                cmd = exec.Command("git", "pull")
                cmd.Dir = repoPath
                if output, err := cmd.CombinedOutput(); err != nil {
                        return fmt.Errorf("failed to pull latest changes: %s (exit code: %w)", string(output), err)
                }
        }

        return nil
}

// CreateWorktreeForIssue creates an isolated git worktree for developing a specific Linear issue.
// It creates a new branch based on the base branch and sets up a separate working directory
// outside the main repository. This allows multiple issues to be worked on simultaneously
// without interfering with each other or cluttering the main repository.
func CreateWorktreeForIssue(worktreeRoot, repoPath, issueID, baseBranch string) (string, error) {
        fmt.Printf("[DEBUG] CreateWorktreeForIssue called with worktreeRoot: %s, repoPath: %s, issueID: %s, baseBranch: %s\n", 
                worktreeRoot, repoPath, issueID, baseBranch)
        
        // Validate that we're working with a git repository
        if err := validateGitRepository(repoPath); err != nil {
                return "", fmt.Errorf("not a git repository: %w", err)
        }

        // Ensure the base branch exists before creating a worktree from it
        if !branchExists(repoPath, baseBranch) {
                return "", fmt.Errorf("base branch %s does not exist", baseBranch)
        }

        // Create a git-safe branch name from the issue ID
        branchName := SanitizeBranchName(issueID)
        
        // Set up paths for the worktree structure outside the main repository
        worktreePath := filepath.Join(worktreeRoot, issueID)

        // Check if a worktree for this issue already exists
        if _, err := os.Stat(worktreePath); err == nil {
                return "", fmt.Errorf("worktree for issue %s already exists at %s", issueID, worktreePath)
        }

        // Create the worktree root directory if it doesn't exist
        if err := os.MkdirAll(worktreeRoot, 0755); err != nil {
                return "", fmt.Errorf("failed to create worktree root directory: %w", err)
        }

        // Create the git worktree with a new branch based on the base branch
        // This creates an isolated working directory with its own branch
        cmd := exec.Command("git", "worktree", "add", "-b", branchName, worktreePath, baseBranch)
        cmd.Dir = repoPath
        fmt.Printf("[DEBUG] Running 'git worktree add -b %s %s %s' in directory: %s\n", 
                branchName, worktreePath, baseBranch, repoPath)
        if output, err := cmd.CombinedOutput(); err != nil {
                return "", fmt.Errorf("failed to create worktree for issue %s: %s (exit code: %w)", issueID, string(output), err)
        }

        return worktreePath, nil
}

// SanitizeBranchName converts issue identifiers into valid git branch names.
// It removes or replaces characters that are not allowed in git branch names,
// ensuring compatibility with git's naming requirements while preserving readability.
func SanitizeBranchName(name string) string {
        // Handle empty input with a default name
        if name == "" {
                return "issue"
        }

        // Replace any characters that aren't alphanumeric, hyphens, or underscores
        // Git branch names have specific character restrictions
        reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
        sanitized := reg.ReplaceAllString(name, "-")

        // Remove leading and trailing hyphens which can cause git issues
        sanitized = strings.Trim(sanitized, "-")

        // Fallback to default if sanitization resulted in empty string
        if sanitized == "" {
                return "issue"
        }

        return sanitized
}

// validateGitRepository checks if the provided path contains a valid git repository.
// It uses git's own validation to ensure the directory is properly initialized.
func validateGitRepository(path string) error {
        cmd := exec.Command("git", "rev-parse", "--git-dir")
        cmd.Dir = path
        return cmd.Run()
}

// branchExists checks if a specific branch exists in the git repository.
// It uses git's show-ref command to verify the branch reference exists.
func branchExists(repoPath, branchName string) bool {
        cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
        cmd.Dir = repoPath
        return cmd.Run() == nil
}

// hasRemote determines if the git repository has any remote repositories configured.
// This is used to decide whether to attempt pulling latest changes.
func hasRemote(repoPath string) bool {
        cmd := exec.Command("git", "remote")
        cmd.Dir = repoPath
        output, err := cmd.Output()
        return err == nil && len(strings.TrimSpace(string(output))) > 0
}

// CreateFeatureFile creates a _feature.md file in the worktree with issue context.
// This provides developers with immediate access to the issue title and description
// right in their development environment, improving context and reducing context switching.
func CreateFeatureFile(worktreePath, issueTitle, issueDescription string) error {
        // Define the path for the feature file
        featureFilePath := filepath.Join(worktreePath, "_feature.md")
        
        // Create markdown content with issue title as header
        content := fmt.Sprintf("# %s\n\n", issueTitle)
        
        // Add description if available
        if issueDescription != "" {
                content += issueDescription + "\n"
        }
        
        // Write the file with appropriate permissions
        if err := os.WriteFile(featureFilePath, []byte(content), 0644); err != nil {
                return fmt.Errorf("failed to create _feature.md file: %w", err)
        }
        
        return nil
}
