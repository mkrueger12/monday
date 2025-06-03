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
        
        // Verify the repository is still valid after preparation
        gitDir := filepath.Join(repoPath, ".git")
        info, err := os.Stat(gitDir)
        require.NoError(t, err)
        assert.True(t, info.IsDir())
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
        worktreePath, err := CreateWorktreeForIssue("/nonexistent/path", "ISSUE-123", "main")
        assert.Error(t, err)
        assert.Empty(t, worktreePath)
}

func TestCreateWorktreeForIssue_InvalidBaseBranch(t *testing.T) {
        repoPath := setupTestRepo(t)

        worktreePath, err := CreateWorktreeForIssue(repoPath, "ISSUE-123", "nonexistent-branch")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "branch")
        assert.Empty(t, worktreePath)
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
                        assert.NotEmpty(t, result)
                })
        }
}

func TestCreateFeatureFile_Success(t *testing.T) {
        tempDir := t.TempDir()
        worktreePath := filepath.Join(tempDir, "test-worktree")
        err := os.MkdirAll(worktreePath, 0755)
        require.NoError(t, err)

        issueTitle := "Fix authentication bug"
        issueDescription := "This is a detailed description of the authentication bug that needs to be fixed.\n\nSteps to reproduce:\n1. Login with invalid credentials\n2. Check error handling"

        err = CreateFeatureFile(worktreePath, issueTitle, issueDescription)
        require.NoError(t, err)

        featureFilePath := filepath.Join(worktreePath, "_feature.md")
        assert.FileExists(t, featureFilePath)

        content, err := os.ReadFile(featureFilePath)
        require.NoError(t, err)

        expectedContent := "# Fix authentication bug\n\nThis is a detailed description of the authentication bug that needs to be fixed.\n\nSteps to reproduce:\n1. Login with invalid credentials\n2. Check error handling\n"
        assert.Equal(t, expectedContent, string(content))
}

func TestCreateFeatureFile_EmptyDescription(t *testing.T) {
        tempDir := t.TempDir()
        worktreePath := filepath.Join(tempDir, "test-worktree")
        err := os.MkdirAll(worktreePath, 0755)
        require.NoError(t, err)

        issueTitle := "Simple bug fix"
        issueDescription := ""

        err = CreateFeatureFile(worktreePath, issueTitle, issueDescription)
        require.NoError(t, err)

        featureFilePath := filepath.Join(worktreePath, "_feature.md")
        assert.FileExists(t, featureFilePath)

        content, err := os.ReadFile(featureFilePath)
        require.NoError(t, err)

        expectedContent := "# Simple bug fix\n\n"
        assert.Equal(t, expectedContent, string(content))
}

func TestCreateFeatureFile_InvalidPath(t *testing.T) {
        invalidPath := "/nonexistent/path/that/does/not/exist"
        
        err := CreateFeatureFile(invalidPath, "Test Title", "Test Description")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "failed to create _feature.md file")
}
