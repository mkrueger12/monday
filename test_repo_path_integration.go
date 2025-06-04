package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"monday/gitops"
	"github.com/stretchr/testify/require"
)

func TestRepoPathIntegration_GitOperationsUseCorrectRepo(t *testing.T) {
	// Create two temporary git repositories
	repo1 := setupTestRepo(t, "repo1")
	repo2 := setupTestRepo(t, "repo2")
	
	// Change working directory to repo1
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	
	err = os.Chdir(repo1)
	require.NoError(t, err)
	
	t.Logf("Current working directory: %s", repo1)
	t.Logf("Target repository: %s", repo2)
	
	// Test PrepareRepository on repo2 while working in repo1
	err = gitops.PrepareRepository(repo2, "main")
	require.NoError(t, err, "PrepareRepository should work on target repo, not current directory")
	
	// Test CreateWorktreeForIssue on repo2 while working in repo1
	worktreeRoot := t.TempDir()
	worktreePath, err := gitops.CreateWorktreeForIssue(worktreeRoot, repo2, "TEST-123", "main")
	require.NoError(t, err, "CreateWorktreeForIssue should work on target repo, not current directory")
	
	// Verify the worktree was created
	require.DirExists(t, worktreePath)
	
	// Verify the worktree is linked to repo2, not repo1
	// Check that the worktree has the commit from repo2
	cmd := exec.Command("git", "log", "--oneline", "-1")
	cmd.Dir = worktreePath
	output, err := cmd.Output()
	require.NoError(t, err)
	
	t.Logf("Worktree commit: %s", string(output))
	
	// The worktree should contain the commit message from repo2
	require.Contains(t, string(output), "Initial commit repo2")
}

func setupTestRepo(t *testing.T, name string) string {
	tempDir := filepath.Join(t.TempDir(), name)
	err := os.MkdirAll(tempDir, 0755)
	require.NoError(t, err)
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	require.NoError(t, cmd.Run())
	
	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	require.NoError(t, cmd.Run())
	
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	require.NoError(t, cmd.Run())
	
	// Create a unique file for this repo
	readmeContent := fmt.Sprintf("# Test Repository %s\n", name)
	readmePath := filepath.Join(tempDir, "README.md")
	err = os.WriteFile(readmePath, []byte(readmeContent), 0644)
	require.NoError(t, err)
	
	// Add and commit
	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = tempDir
	require.NoError(t, cmd.Run())
	
	commitMessage := fmt.Sprintf("Initial commit %s", name)
	cmd = exec.Command("git", "commit", "-m", commitMessage)
	cmd.Dir = tempDir
	require.NoError(t, cmd.Run())
	
	// Ensure we're on main branch
	cmd = exec.Command("git", "branch", "-M", "main")
	cmd.Dir = tempDir
	require.NoError(t, cmd.Run())
	
	return tempDir
}
