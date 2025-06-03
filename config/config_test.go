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
                        args: []string{"monday", "ISSUE-123", "/path/to/repo"},
                        expected: &AppConfig{
                                IssueIDs:       []string{"ISSUE-123"},
                                GitRepoPath:    "/path/to/repo",
                                LinearEndpoint: "",
                                Concurrency:    3, // default
                                DryRun:         false,
                                BaseBranch:     "develop", // default
                        },
                },
                {
                        name: "multiple issues",
                        args: []string{"monday", "ISSUE-123", "ISSUE-456", "/path/to/repo"},
                        expected: &AppConfig{
                                IssueIDs:       []string{"ISSUE-123", "ISSUE-456"},
                                GitRepoPath:    "/path/to/repo",
                                LinearEndpoint: "",
                                Concurrency:    3, // default
                                DryRun:         false,
                                BaseBranch:     "develop", // default
                        },
                },
                {
                        name: "with concurrency flag",
                        args: []string{"monday", "-concurrency", "5", "ISSUE-123", "/path/to/repo"},
                        expected: &AppConfig{
                                IssueIDs:       []string{"ISSUE-123"},
                                GitRepoPath:    "/path/to/repo",
                                LinearEndpoint: "",
                                Concurrency:    5,
                                DryRun:         false,
                                BaseBranch:     "develop", // default
                        },
                },
                {
                        name: "with custom base branch",
                        args: []string{"monday", "-base-branch", "main", "ISSUE-123", "/path/to/repo"},
                        expected: &AppConfig{
                                IssueIDs:       []string{"ISSUE-123"},
                                GitRepoPath:    "/path/to/repo",
                                LinearEndpoint: "",
                                Concurrency:    3, // default
                                DryRun:         false,
                                BaseBranch:     "main",
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
                        name: "missing repo path",
                        args: []string{"monday", "ISSUE-123"},
                },
                {
                        name: "invalid concurrency",
                        args: []string{"monday", "-concurrency", "invalid", "ISSUE-123", "/path/to/repo"},
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

                config, err := ParseConfigFromArgs([]string{"ISSUE-123", "/path/to/repo"})
                require.NoError(t, err)
                assert.Equal(t, "test-api-key", config.LinearAPIKey)
                assert.Equal(t, []string{"ISSUE-123"}, config.IssueIDs)
        })

        t.Run("from flag overrides env", func(t *testing.T) {
                oldEnv := os.Getenv("LINEAR_API_KEY")
                defer func() {
                        os.Setenv("LINEAR_API_KEY", oldEnv)
                }()

                os.Setenv("LINEAR_API_KEY", "env-key")

                config, err := ParseConfigFromArgs([]string{"-api-key", "flag-key", "ISSUE-123", "/path/to/repo"})
                require.NoError(t, err)
                assert.Equal(t, "flag-key", config.LinearAPIKey)
                assert.Equal(t, "/path/to/repo", config.GitRepoPath)
        })
}
