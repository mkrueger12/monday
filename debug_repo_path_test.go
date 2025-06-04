package main

import (
	"os"
	"testing"
	"monday/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoPathBug_ParsesCorrectPath(t *testing.T) {
	// Test that simulates the exact command: ./bin/monday DEL-163 /Users/max/code/product-data-scraping
	args := []string{"DEL-163", "/Users/max/code/product-data-scraping"}
	
	cfg, err := config.ParseConfigFromArgs(args)
	require.NoError(t, err)
	
	// This should pass - the parsing should work correctly
	assert.Equal(t, "/Users/max/code/product-data-scraping", cfg.GitRepoPath)
	assert.Equal(t, []string{"DEL-163"}, cfg.IssueIDs)
	
	t.Logf("Parsed GitRepoPath: %s", cfg.GitRepoPath)
	t.Logf("Parsed IssueIDs: %v", cfg.IssueIDs)
}

func TestRepoPathBug_WorkingDirectoryDoesNotAffectParsing(t *testing.T) {
	// Save original working directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	
	// Change to a different directory (simulating running from /Users/max/code/monday)
	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	
	// Parse the same arguments
	args := []string{"DEL-163", "/Users/max/code/product-data-scraping"}
	cfg, err := config.ParseConfigFromArgs(args)
	require.NoError(t, err)
	
	// The parsed path should still be absolute and correct
	assert.Equal(t, "/Users/max/code/product-data-scraping", cfg.GitRepoPath)
	
	t.Logf("Current working directory: %s", tempDir)
	t.Logf("Parsed GitRepoPath: %s", cfg.GitRepoPath)
}
