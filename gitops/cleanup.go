package gitops

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// CleanWorktrees removes worktree directories older than the specified threshold
// and runs git worktree prune to clean up git metadata.
// It scans the worktreeRoot directory for subdirectories and removes those
// whose modification time is older than thresholdDays.
func CleanWorktrees(worktreeRoot, repoPath string, thresholdDays int, dryRun bool) error {
	if thresholdDays < 0 {
		return fmt.Errorf("cleanup days must be non-negative, got %d", thresholdDays)
	}

	// Check if worktree root exists
	if _, err := os.Stat(worktreeRoot); os.IsNotExist(err) {
		log.Printf("Worktree root %s does not exist, nothing to clean", worktreeRoot)
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -thresholdDays)
	log.Printf("Cleaning worktrees older than %s (threshold: %d days)", cutoff.Format("2006-01-02 15:04:05"), thresholdDays)

	entries, err := ioutil.ReadDir(worktreeRoot)
	if err != nil {
		return fmt.Errorf("reading worktree root %s: %w", worktreeRoot, err)
	}

	removedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(worktreeRoot, entry.Name())
		if entry.ModTime().Before(cutoff) {
			if dryRun {
				log.Printf("[dry-run] would remove worktree: %s (age: %s)", dirPath, entry.ModTime().Format("2006-01-02 15:04:05"))
			} else {
				log.Printf("Removing worktree: %s (age: %s)", dirPath, entry.ModTime().Format("2006-01-02 15:04:05"))
				if err := os.RemoveAll(dirPath); err != nil {
					return fmt.Errorf("removing worktree directory %s: %w", dirPath, err)
				}
			}
			removedCount++
		}
	}

	if removedCount == 0 {
		log.Printf("No worktrees found older than %d days", thresholdDays)
	} else if dryRun {
		log.Printf("[dry-run] would remove %d worktrees", removedCount)
	} else {
		log.Printf("Removed %d worktrees", removedCount)
	}

	// Run git worktree prune to clean up git metadata
	if err := pruneGitWorktrees(repoPath, dryRun); err != nil {
		return fmt.Errorf("pruning git worktrees: %w", err)
	}

	return nil
}

// pruneGitWorktrees runs 'git worktree prune' to clean up stale worktree references
// in the git repository metadata.
func pruneGitWorktrees(repoPath string, dryRun bool) error {
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("git directory not found at %s", gitDir)
	}

	cmd := exec.Command("git", "worktree", "prune")
	cmd.Dir = repoPath

	if dryRun {
		log.Printf("[dry-run] would run: git worktree prune in %s", repoPath)
		return nil
	}

	log.Printf("Running git worktree prune in %s", repoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree prune failed: %w (output: %s)", err, string(output))
	}

	if len(output) > 0 {
		log.Printf("Git worktree prune output: %s", string(output))
	}

	return nil
}
