package config

import (
        "os"
        "testing"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
)

func TestParseConfig_Success(t *testing.T) {
        tests := []struct {
                name     string
                args     []string
                expected *AppConfig
        }{
                {
                        name: "single issue",
                        args: []string{"monday", "--worktree-root", "/tmp/worktrees", "ISSUE-123", "/path/to/repo"},
                        expected: &AppConfig{
                                IssueIDs:       []string{"ISSUE-123"},
                                GitRepoPath:    "/path/to/repo",
                                WorktreeRoot:   "/tmp/worktrees",
                                LinearEndpoint: "",
                                Concurrency:    3, // default
                                DryRun:         false,
                                BaseBranch:     "develop", // default
                                Cleanup:        false, // default
                                CleanupDays:    7, // default
                        },
                },
                {
                        name: "multiple issues",
                        args: []string{"monday", "--worktree-root", "/tmp/worktrees", "ISSUE-123", "ISSUE-456", "/path/to/repo"},
                        expected: &AppConfig{
                                IssueIDs:       []string{"ISSUE-123", "ISSUE-456"},
                                GitRepoPath:    "/path/to/repo",
                                WorktreeRoot:   "/tmp/worktrees",
                                LinearEndpoint: "",
                                Concurrency:    3, // default
                                DryRun:         false,
                                BaseBranch:     "develop", // default
                                Cleanup:        false, // default
                                CleanupDays:    7, // default
                        },
                },
                {
                        name: "with concurrency flag",
                        args: []string{"monday", "--worktree-root", "/tmp/worktrees", "-concurrency", "5", "ISSUE-123", "/path/to/repo"},
                        expected: &AppConfig{
                                IssueIDs:       []string{"ISSUE-123"},
                                GitRepoPath:    "/path/to/repo",
                                WorktreeRoot:   "/tmp/worktrees",
                                LinearEndpoint: "",
                                Concurrency:    5,
                                DryRun:         false,
                                BaseBranch:     "develop", // default
                                Cleanup:        false, // default
                                CleanupDays:    7, // default
                        },
                },
                {
                        name: "with custom base branch",
                        args: []string{"monday", "--worktree-root", "/tmp/worktrees", "-base-branch", "develop", "ISSUE-123", "/path/to/repo"},
                        expected: &AppConfig{
                                IssueIDs:       []string{"ISSUE-123"},
                                GitRepoPath:    "/path/to/repo",
                                WorktreeRoot:   "/tmp/worktrees",
                                LinearEndpoint: "",
                                Concurrency:    3, // default
                                DryRun:         false,
                                BaseBranch:     "develop",
                                Cleanup:        false, // default
                                CleanupDays:    7, // default
                        },
                },
                {
                        name: "cleanup mode",
                        args: []string{"monday", "--cleanup", "--worktree-root", "/tmp/worktrees", "/path/to/repo"},
                        expected: &AppConfig{
                                IssueIDs:       []string{},
                                GitRepoPath:    "/path/to/repo",
                                WorktreeRoot:   "/tmp/worktrees",
                                LinearEndpoint: "",
                                Concurrency:    3, // default
                                DryRun:         false,
                                BaseBranch:     "develop", // default
                                Cleanup:        true,
                                CleanupDays:    7, // default
                        },
                },
                {
                        name: "cleanup mode with custom days",
                        args: []string{"monday", "--cleanup", "--worktree-root", "/tmp/worktrees", "--cleanup-days", "14", "/path/to/repo"},
                        expected: &AppConfig{
                                IssueIDs:       []string{},
                                GitRepoPath:    "/path/to/repo",
                                WorktreeRoot:   "/tmp/worktrees",
                                LinearEndpoint: "",
                                Concurrency:    3, // default
                                DryRun:         false,
                                BaseBranch:     "develop", // default
                                Cleanup:        true,
                                CleanupDays:    14,
                        },
                },
                {
                        name: "normal mode with custom cleanup days",
                        args: []string{"monday", "--worktree-root", "/tmp/worktrees", "--cleanup-days", "3", "ISSUE-123", "/path/to/repo"},
                        expected: &AppConfig{
                                IssueIDs:       []string{"ISSUE-123"},
                                GitRepoPath:    "/path/to/repo",
                                WorktreeRoot:   "/tmp/worktrees",
                                LinearEndpoint: "",
                                Concurrency:    3, // default
                                DryRun:         false,
                                BaseBranch:     "develop", // default
                                Cleanup:        false,
                                CleanupDays:    3,
                        },
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        config, err := ParseConfigFromArgs(tt.args[1:]) // skip program name
                        require.NoError(t, err)
                        assert.Equal(t, tt.expected, config)
                })
        }
}

func TestParseConfig_Errors(t *testing.T) {
        tests := []struct {
                name string
                args []string
        }{
                {
                        name: "no arguments",
                        args: []string{"monday"},
                },
                {
                        name: "missing worktree-root",
                        args: []string{"monday", "ISSUE-123", "/path/to/repo"},
                },
                {
                        name: "missing repo path",
                        args: []string{"monday", "--worktree-root", "/tmp/worktrees", "ISSUE-123"},
                },
                {
                        name: "invalid concurrency",
                        args: []string{"monday", "--worktree-root", "/tmp/worktrees", "-concurrency", "invalid", "ISSUE-123", "/path/to/repo"},
                },
                {
                        name: "cleanup mode with no repo path",
                        args: []string{"monday", "--cleanup", "--worktree-root", "/tmp/worktrees"},
                },
                {
                        name: "cleanup mode missing worktree-root",
                        args: []string{"monday", "--cleanup", "/path/to/repo"},
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        config, err := ParseConfigFromArgs(tt.args[1:]) // skip program name
                        assert.Error(t, err)
                        assert.Nil(t, config)
                })
        }
}

func TestParseConfig_LinearAPIKey(t *testing.T) {
        t.Run("from environment variable", func(t *testing.T) {
                oldEnv := os.Getenv("LINEAR_API_KEY")
                defer func() {
                        os.Setenv("LINEAR_API_KEY", oldEnv)
                }()

                os.Setenv("LINEAR_API_KEY", "test-api-key")

                config, err := ParseConfigFromArgs([]string{"--worktree-root", "/tmp/worktrees", "ISSUE-123", "/path/to/repo"})
                require.NoError(t, err)
                assert.Equal(t, "test-api-key", config.LinearAPIKey)
                assert.Equal(t, []string{"ISSUE-123"}, config.IssueIDs)
                assert.Equal(t, "/tmp/worktrees", config.WorktreeRoot)
        })

        t.Run("from flag overrides env", func(t *testing.T) {
                oldEnv := os.Getenv("LINEAR_API_KEY")
                defer func() {
                        os.Setenv("LINEAR_API_KEY", oldEnv)
                }()

                os.Setenv("LINEAR_API_KEY", "env-key")

                config, err := ParseConfigFromArgs([]string{"--worktree-root", "/tmp/worktrees", "-api-key", "flag-key", "ISSUE-123", "/path/to/repo"})
                require.NoError(t, err)
                assert.Equal(t, "flag-key", config.LinearAPIKey)
                assert.Equal(t, "/path/to/repo", config.GitRepoPath)
                assert.Equal(t, "/tmp/worktrees", config.WorktreeRoot)
        })
}

func TestParseConfig_WorktreeRoot(t *testing.T) {
        tests := []struct {
                name        string
                args        []string
                expectError bool
                wantRoot    string
        }{
                {
                        name:        "with worktree-root flag",
                        args:        []string{"--worktree-root", "/custom/worktrees", "ISSUE-123", "/path/to/repo"},
                        expectError: false,
                        wantRoot:    "/custom/worktrees",
                },
                {
                        name:        "with -w shorthand",
                        args:        []string{"-w", "/short/path", "ISSUE-123", "/path/to/repo"},
                        expectError: false,
                        wantRoot:    "/short/path",
                },
                {
                        name:        "missing worktree-root",
                        args:        []string{"ISSUE-123", "/path/to/repo"},
                        expectError: true,
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        config, err := ParseConfigFromArgs(tt.args)
                        if tt.expectError {
                                assert.Error(t, err)
                                assert.Contains(t, err.Error(), "worktree-root")
                                return
                        }
                        assert.NoError(t, err)
                        assert.Equal(t, tt.wantRoot, config.WorktreeRoot)
                })
        }
}
