package gitops

import (
        "os"
        "os/exec"
        "path/filepath"
        "testing"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) string {
        tempDir, err := os.MkdirTemp("", "monday-test-repo-*")
        require.NoError(t, err)

        t.Cleanup(func() {
                os.RemoveAll(tempDir)
        })

        cmd := exec.Command("git", "init")
        cmd.Dir = tempDir
        require.NoError(t, cmd.Run())

        cmd = exec.Command("git", "config", "user.email", "test@example.com")
        cmd.Dir = tempDir
        require.NoError(t, cmd.Run())

        cmd = exec.Command("git", "config", "user.name", "Test User")
        cmd.Dir = tempDir
        require.NoError(t, cmd.Run())

        readmeFile := filepath.Join(tempDir, "README.md")
        err = os.WriteFile(readmeFile, []byte("# Test Repo"), 0644)
        require.NoError(t, err)

        cmd = exec.Command("git", "add", "README.md")
        cmd.Dir = tempDir
        require.NoError(t, cmd.Run())

        cmd = exec.Command("git", "commit", "-m", "Initial commit")
        cmd.Dir = tempDir
        require.NoError(t, cmd.Run())

        cmd = exec.Command("git", "branch", "-M", "main")
        cmd.Dir = tempDir
        require.NoError(t, cmd.Run())

        return tempDir
}

func TestPrepareRepository_Success(t *testing.T) {
        repoPath := setupTestRepo(t)

        err := PrepareRepository(repoPath, "main")
        assert.NoError(t, err)
}

func TestPrepareRepository_InvalidPath(t *testing.T) {
        err := PrepareRepository("/nonexistent/path", "main")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "not a git repository")
}

func TestPrepareRepository_InvalidBranch(t *testing.T) {
        repoPath := setupTestRepo(t)

        err := PrepareRepository(repoPath, "nonexistent-branch")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "branch")
}

func TestCreateWorktreeForIssue_Success(t *testing.T) {
        repoPath := setupTestRepo(t)

        worktreePath, err := CreateWorktreeForIssue(repoPath, "ISSUE-123", "main")
        require.NoError(t, err)

        expectedPath := filepath.Join(repoPath, "worktrees", "ISSUE-123")
        assert.Equal(t, expectedPath, worktreePath)

        info, err := os.Stat(worktreePath)
        require.NoError(t, err)
        assert.True(t, info.IsDir())

        gitDir := filepath.Join(worktreePath, ".git")
        _, err = os.Stat(gitDir)
        assert.NoError(t, err)
}

func TestCreateWorktreeForIssue_ExistingWorktree(t *testing.T) {
        repoPath := setupTestRepo(t)

        _, err := CreateWorktreeForIssue(repoPath, "ISSUE-123", "main")
        require.NoError(t, err)

        _, err = CreateWorktreeForIssue(repoPath, "ISSUE-123", "main")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "already exists")
}

func TestCreateWorktreeForIssue_InvalidRepo(t *testing.T) {
        _, err := CreateWorktreeForIssue("/nonexistent/path", "ISSUE-123", "main")
        assert.Error(t, err)
}

func TestCreateWorktreeForIssue_InvalidBaseBranch(t *testing.T) {
        repoPath := setupTestRepo(t)

        _, err := CreateWorktreeForIssue(repoPath, "ISSUE-123", "nonexistent-branch")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "branch")
}

func TestSanitizeBranchName(t *testing.T) {
        tests := []struct {
                input    string
                expected string
        }{
                {"ISSUE-123", "ISSUE-123"},
                {"ISSUE-123-fix-bug", "ISSUE-123-fix-bug"},
                {"ISSUE 123", "ISSUE-123"},
                {"ISSUE/123", "ISSUE-123"},
                {"ISSUE\\123", "ISSUE-123"},
                {"ISSUE@123", "ISSUE-123"},
                {"issue-123", "issue-123"},
                {"", "issue"},
        }

        for _, tt := range tests {
                t.Run(tt.input, func(t *testing.T) {
                        result := SanitizeBranchName(tt.input)
                        assert.Equal(t, tt.expected, result)
                })
        }
}
