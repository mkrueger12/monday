// Package config handles command-line argument parsing and configuration management
// for the monday CLI tool. It supports both CLI flags and environment variables
// for flexible configuration in different environments.
package config

import (
        "flag"
        "fmt"
        "os"
)

// AppConfig represents the complete configuration for the monday CLI application.
// It combines command-line arguments, environment variables, and default values
// to provide all necessary settings for processing Linear issues.
type AppConfig struct {
        // IssueIDs contains the list of Linear issue identifiers to process (e.g., ["DEL-163", "DEL-164"])
        IssueIDs        []string
        // GitRepoPath is the absolute path to the git repository where worktrees will be created
        GitRepoPath     string
        // LinearAPIKey is the authentication token for Linear API access
        LinearAPIKey    string
        // LinearEndpoint allows overriding the Linear API endpoint (primarily for testing)
        LinearEndpoint  string
        // Concurrency controls how many issues are processed simultaneously (default: 3)
        Concurrency     int
        // DryRun indicates whether to simulate operations without launching Terminal
        DryRun          bool
        // BaseBranch is the git branch to use as the base for new worktree branches (default: "develop")
        BaseBranch      string
}

// ParseConfig parses configuration from the default command-line arguments (os.Args[1:]).
// This is a convenience wrapper around ParseConfigFromArgs for standard usage.
func ParseConfig() (*AppConfig, error) {
        return ParseConfigFromArgs(os.Args[1:])
}

// ParseConfigFromArgs parses configuration from the provided command-line arguments.
// It handles flag parsing, argument validation, and environment variable fallbacks.
// Expected usage: monday [flags] <issue_id_1> [issue_id_2 ...] <git_repo_path>
// Returns a fully populated AppConfig or an error if parsing/validation fails.
func ParseConfigFromArgs(args []string) (*AppConfig, error) {
        // Initialize variables for flag parsing
        var concurrency int
        var apiKey string
        var linearEndpoint string
        var dryRun bool
        var baseBranch string

        // Create a new flag set for parsing monday CLI arguments
        fs := flag.NewFlagSet("monday", flag.ContinueOnError)
        fs.IntVar(&concurrency, "concurrency", 3, "Number of concurrent issue processors")
        fs.StringVar(&apiKey, "api-key", "", "Linear API key (overrides LINEAR_API_KEY env var)")
        fs.StringVar(&linearEndpoint, "linear-endpoint", "", "Linear API endpoint (for testing)")
        fs.BoolVar(&dryRun, "dry-run", false, "Don't actually launch Terminal")
        fs.StringVar(&baseBranch, "base-branch", "develop", "Git base branch for new worktrees")
        
        // Parse the provided arguments
        err := fs.Parse(args)
        if err != nil {
                return nil, err
        }

        // Extract non-flag arguments (issue IDs and git repo path)
        remainingArgs := fs.Args()
        if len(remainingArgs) < 2 {
                return nil, fmt.Errorf("usage: monday [flags] <issue_id_1> [issue_id_2 ...] <git_repo_path>")
        }

        // Last argument is always the git repository path
        gitRepoPath := remainingArgs[len(remainingArgs)-1]
        // All other arguments are issue IDs
        issueIDs := remainingArgs[:len(remainingArgs)-1]

        // Validate that at least one issue ID was provided
        if len(issueIDs) == 0 {
                return nil, fmt.Errorf("at least one issue ID is required")
        }

        // Use CLI flag value or fall back to environment variable
        linearAPIKey := apiKey
        if linearAPIKey == "" {
                linearAPIKey = os.Getenv("LINEAR_API_KEY")
        }

        // Return fully populated configuration
        return &AppConfig{
                IssueIDs:       issueIDs,
                GitRepoPath:    gitRepoPath,
                LinearAPIKey:   linearAPIKey,
                LinearEndpoint: linearEndpoint,
                Concurrency:    concurrency,
                DryRun:         dryRun,
                BaseBranch:     baseBranch,
        }, nil
}
