package gitops

import (
        "io/ioutil"
        "os"
        "os/exec"
        "path/filepath"
        "testing"
        "time"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
)

func setupTestRepoForCleanup(t *testing.T, tempDir string) string {
        repoPath := filepath.Join(tempDir, "repo")
        err := os.MkdirAll(repoPath, 0755)
        require.NoError(t, err)
        
        // Initialize git repository
        cmd := exec.Command("git", "init")
        cmd.Dir = repoPath
        require.NoError(t, cmd.Run())
        
        return repoPath
}

func TestCleanWorktrees_Success(t *testing.T) {
        // Create temporary directories for testing
        tempDir, err := ioutil.TempDir("", "monday-cleanup-test")
        require.NoError(t, err)
        defer os.RemoveAll(tempDir)

        worktreeRoot := filepath.Join(tempDir, "worktrees")
        repoPath := setupTestRepoForCleanup(t, tempDir)

        // Create worktree root directory
        err = os.MkdirAll(worktreeRoot, 0755)
        require.NoError(t, err)

        // Create test worktree directories with different ages
        oldDir := filepath.Join(worktreeRoot, "old-issue")
        recentDir := filepath.Join(worktreeRoot, "recent-issue")
        
        err = os.MkdirAll(oldDir, 0755)
        require.NoError(t, err)
        err = os.MkdirAll(recentDir, 0755)
        require.NoError(t, err)

        // Set modification times
        oldTime := time.Now().AddDate(0, 0, -10) // 10 days ago
        recentTime := time.Now().AddDate(0, 0, -3) // 3 days ago

        err = os.Chtimes(oldDir, oldTime, oldTime)
        require.NoError(t, err)
        err = os.Chtimes(recentDir, recentTime, recentTime)
        require.NoError(t, err)

        // Test cleanup with 7-day threshold
        err = CleanWorktrees(worktreeRoot, repoPath, 7, false)
        require.NoError(t, err)

        // Verify old directory was removed
        _, err = os.Stat(oldDir)
        assert.True(t, os.IsNotExist(err))

        // Verify recent directory still exists
        _, err = os.Stat(recentDir)
        assert.NoError(t, err)
}

func TestCleanWorktrees_DryRun(t *testing.T) {
        // Create temporary directories for testing
        tempDir, err := ioutil.TempDir("", "monday-cleanup-test")
        require.NoError(t, err)
        defer os.RemoveAll(tempDir)

        worktreeRoot := filepath.Join(tempDir, "worktrees")
        repoPath := setupTestRepoForCleanup(t, tempDir)

        // Create worktree root directory
        err = os.MkdirAll(worktreeRoot, 0755)
        require.NoError(t, err)

        // Create old worktree directory
        oldDir := filepath.Join(worktreeRoot, "old-issue")
        err = os.MkdirAll(oldDir, 0755)
        require.NoError(t, err)

        // Set old modification time
        oldTime := time.Now().AddDate(0, 0, -10) // 10 days ago
        err = os.Chtimes(oldDir, oldTime, oldTime)
        require.NoError(t, err)

        // Test dry run cleanup
        err = CleanWorktrees(worktreeRoot, repoPath, 7, true)
        require.NoError(t, err)

        // Verify directory still exists (dry run shouldn't remove)
        _, err = os.Stat(oldDir)
        assert.NoError(t, err)
}

func TestCleanWorktrees_NonexistentWorktreeRoot(t *testing.T) {
        // Create temporary directory for repo
        tempDir, err := ioutil.TempDir("", "monday-cleanup-test")
        require.NoError(t, err)
        defer os.RemoveAll(tempDir)

        repoPath := setupTestRepoForCleanup(t, tempDir)
        worktreeRoot := filepath.Join(tempDir, "nonexistent")

        // Test cleanup with nonexistent worktree root
        err = CleanWorktrees(worktreeRoot, repoPath, 7, false)
        assert.NoError(t, err) // Should not error, just log and return
}

func TestCleanWorktrees_InvalidThreshold(t *testing.T) {
        // Test with negative threshold
        err := CleanWorktrees("/tmp", "/tmp", -1, false)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "cleanup days must be non-negative")
}

func TestCleanWorktrees_InvalidRepoPath(t *testing.T) {
        // Create temporary directories for testing
        tempDir, err := ioutil.TempDir("", "monday-cleanup-test")
        require.NoError(t, err)
        defer os.RemoveAll(tempDir)

        worktreeRoot := filepath.Join(tempDir, "worktrees")
        repoPath := filepath.Join(tempDir, "nonexistent-repo")

        // Create worktree root directory
        err = os.MkdirAll(worktreeRoot, 0755)
        require.NoError(t, err)

        // Test cleanup with invalid repo path
        err = CleanWorktrees(worktreeRoot, repoPath, 7, false)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "git directory not found")
}

func TestCleanWorktrees_EmptyWorktreeRoot(t *testing.T) {
        // Create temporary directories for testing
        tempDir, err := ioutil.TempDir("", "monday-cleanup-test")
        require.NoError(t, err)
        defer os.RemoveAll(tempDir)

        worktreeRoot := filepath.Join(tempDir, "worktrees")
        repoPath := setupTestRepoForCleanup(t, tempDir)

        // Create empty worktree root directory
        err = os.MkdirAll(worktreeRoot, 0755)
        require.NoError(t, err)

        // Test cleanup with empty worktree root
        err = CleanWorktrees(worktreeRoot, repoPath, 7, false)
        assert.NoError(t, err)
}

func TestCleanWorktrees_FilesInWorktreeRoot(t *testing.T) {
        // Create temporary directories for testing
        tempDir, err := ioutil.TempDir("", "monday-cleanup-test")
        require.NoError(t, err)
        defer os.RemoveAll(tempDir)

        worktreeRoot := filepath.Join(tempDir, "worktrees")
        repoPath := setupTestRepoForCleanup(t, tempDir)

        // Create worktree root directory
        err = os.MkdirAll(worktreeRoot, 0755)
        require.NoError(t, err)

        // Create a file (not directory) in worktree root
        filePath := filepath.Join(worktreeRoot, "some-file.txt")
        err = ioutil.WriteFile(filePath, []byte("test"), 0644)
        require.NoError(t, err)

        // Set old modification time on file
        oldTime := time.Now().AddDate(0, 0, -10)
        err = os.Chtimes(filePath, oldTime, oldTime)
        require.NoError(t, err)

        // Test cleanup - should ignore files, only process directories
        err = CleanWorktrees(worktreeRoot, repoPath, 7, false)
        assert.NoError(t, err)

        // Verify file still exists (should be ignored)
        _, err = os.Stat(filePath)
        assert.NoError(t, err)
}
