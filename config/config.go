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
        // RepoURL is the GitHub repository URL to clone
        RepoURL         string
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
        // Cleanup indicates whether to run cleanup mode and exit (no issue processing)
        Cleanup         bool
        // CleanupDays is the number of days to retain worktrees before cleanup (default: 7)
        CleanupDays     int
        // CodexDockerImage is the Docker image to use for OpenAI Codex CLI
        CodexDockerImage string
        // CodexCLIArgs contains additional arguments to pass to the Codex CLI
        CodexCLIArgs    []string
        // AutomatedMode enables full automation without manual approval
        AutomatedMode   bool
        // OpenAIAPIKey is the authentication token for OpenAI API access
        OpenAIAPIKey    string
        // GitHubToken is the authentication token for GitHub API access
        GitHubToken     string
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
        var cleanup bool
        var cleanupDays int
        var repoURL string
        var codexDockerImage string
        var automatedMode bool
        var openaiAPIKey string
        var githubToken string

        // Create a new flag set for parsing monday CLI arguments
        fs := flag.NewFlagSet("monday", flag.ContinueOnError)
        fs.IntVar(&concurrency, "concurrency", 3, "Number of concurrent issue processors")
        fs.StringVar(&apiKey, "api-key", "", "Linear API key (overrides LINEAR_API_KEY env var)")
        fs.StringVar(&linearEndpoint, "linear-endpoint", "", "Linear API endpoint (for testing)")
        fs.BoolVar(&dryRun, "dry-run", false, "Don't actually launch Terminal")
        fs.StringVar(&baseBranch, "base-branch", "develop", "Git base branch for new worktrees")
        fs.BoolVar(&cleanup, "cleanup", false, "Run cleanup of old worktrees and exit")
        fs.IntVar(&cleanupDays, "cleanup-days", 7, "Number of days to retain worktrees")
        fs.StringVar(&repoURL, "repo-url", "", "GitHub repository URL to clone (required for codex mode)")
        fs.StringVar(&codexDockerImage, "codex-docker-image", "monday/codex:latest", "Docker image for OpenAI Codex CLI")
        fs.BoolVar(&automatedMode, "automated", false, "Enable full automation mode")
        fs.StringVar(&openaiAPIKey, "openai-api-key", "", "OpenAI API key (overrides OPENAI_API_KEY env var)")
        fs.StringVar(&githubToken, "github-token", "", "GitHub token (overrides GITHUB_TOKEN env var)")
        
        // Parse the provided arguments
        err := fs.Parse(args)
        if err != nil {
                return nil, err
        }



        // Extract non-flag arguments (issue IDs and git repo path)
        remainingArgs := fs.Args()
        
        // For cleanup mode, no additional arguments needed
        if cleanup {
                if len(remainingArgs) != 0 {
                        return nil, fmt.Errorf("usage: monday --cleanup [flags]")
                }
                issueIDs := []string{} // Empty for cleanup mode
                
                // Use CLI flag value or fall back to environment variable
                linearAPIKey := apiKey
                if linearAPIKey == "" {
                        linearAPIKey = os.Getenv("LINEAR_API_KEY")
                }
                
                // Handle OpenAI API key
                openaiKey := openaiAPIKey
                if openaiKey == "" {
                        openaiKey = os.Getenv("OPENAI_API_KEY")
                }
                
                // Handle GitHub token
                ghToken := githubToken
                if ghToken == "" {
                        ghToken = os.Getenv("GITHUB_TOKEN")
                }
                
                return &AppConfig{
                        IssueIDs:         issueIDs,
                        RepoURL:          repoURL,
                        LinearAPIKey:     linearAPIKey,
                        LinearEndpoint:   linearEndpoint,
                        Concurrency:      concurrency,
                        DryRun:           dryRun,
                        BaseBranch:       baseBranch,
                        Cleanup:          cleanup,
                        CleanupDays:      cleanupDays,
                        CodexDockerImage: codexDockerImage,
                        CodexCLIArgs:     []string{"--approval-mode", "full-auto"},
                        AutomatedMode:    automatedMode,
                        OpenAIAPIKey:     openaiKey,
                        GitHubToken:      ghToken,
                }, nil
        }
        
        // For normal mode, require at least one issue ID
        if len(remainingArgs) < 1 {
                return nil, fmt.Errorf("usage: monday [flags] <issue_id_1> [issue_id_2 ...]")
        }

        // All arguments are issue IDs
        issueIDs := remainingArgs

        // Use CLI flag value or fall back to environment variable
        linearAPIKey := apiKey
        if linearAPIKey == "" {
                linearAPIKey = os.Getenv("LINEAR_API_KEY")
        }
        
        // Handle OpenAI API key
        openaiKey := openaiAPIKey
        if openaiKey == "" {
                openaiKey = os.Getenv("OPENAI_API_KEY")
        }
        
        // Handle GitHub token
        ghToken := githubToken
        if ghToken == "" {
                ghToken = os.Getenv("GITHUB_TOKEN")
        }

        // Return fully populated configuration
        return &AppConfig{
                IssueIDs:         issueIDs,
                RepoURL:          repoURL,
                LinearAPIKey:     linearAPIKey,
                LinearEndpoint:   linearEndpoint,
                Concurrency:      concurrency,
                DryRun:           dryRun,
                BaseBranch:       baseBranch,
                Cleanup:          cleanup,
                CleanupDays:      cleanupDays,
                CodexDockerImage: codexDockerImage,
                CodexCLIArgs:     []string{"--approval-mode", "full-auto"},
                AutomatedMode:    automatedMode,
                OpenAIAPIKey:     openaiKey,
                GitHubToken:      ghToken,
        }, nil
}

// ParseCodexConfigFromArgs parses configuration for codex mode without expecting issue IDs
func ParseCodexConfigFromArgs(args []string) (*AppConfig, error) {
        // Initialize variables for flag parsing
        var concurrency int
        var apiKey string
        var linearEndpoint string
        var dryRun bool
        var baseBranch string
        var cleanup bool
        var cleanupDays int
        var repoURL string
        var codexDockerImage string
        var automatedMode bool
        var openaiAPIKey string
        var githubToken string

        // Create a new flag set for parsing monday CLI arguments
        fs := flag.NewFlagSet("monday", flag.ContinueOnError)
        fs.IntVar(&concurrency, "concurrency", 3, "Number of concurrent issue processors")
        fs.StringVar(&apiKey, "api-key", "", "Linear API key (overrides LINEAR_API_KEY env var)")
        fs.StringVar(&linearEndpoint, "linear-endpoint", "", "Linear API endpoint (for testing)")
        fs.BoolVar(&dryRun, "dry-run", false, "Don't actually launch Terminal")
        fs.StringVar(&baseBranch, "base-branch", "develop", "Git base branch for new worktrees")
        fs.BoolVar(&cleanup, "cleanup", false, "Run cleanup of old worktrees and exit")
        fs.IntVar(&cleanupDays, "cleanup-days", 7, "Number of days to retain worktrees")
        fs.StringVar(&repoURL, "repo-url", "", "GitHub repository URL to clone (required for codex mode)")
        fs.StringVar(&codexDockerImage, "codex-docker-image", "monday/codex:latest", "Docker image for OpenAI Codex CLI")
        fs.BoolVar(&automatedMode, "automated", false, "Enable full automation mode")
        fs.StringVar(&openaiAPIKey, "openai-api-key", "", "OpenAI API key (overrides OPENAI_API_KEY env var)")
        fs.StringVar(&githubToken, "github-token", "", "GitHub token (overrides GITHUB_TOKEN env var)")
        
        // Parse the provided arguments
        err := fs.Parse(args)
        if err != nil {
                return nil, err
        }

        // Use CLI flag value or fall back to environment variable
        linearAPIKey := apiKey
        if linearAPIKey == "" {
                linearAPIKey = os.Getenv("LINEAR_API_KEY")
        }
        
        // Handle OpenAI API key
        openaiKey := openaiAPIKey
        if openaiKey == "" {
                openaiKey = os.Getenv("OPENAI_API_KEY")
        }
        
        // Handle GitHub token
        ghToken := githubToken
        if ghToken == "" {
                ghToken = os.Getenv("GITHUB_TOKEN")
        }

        // Return configuration for codex mode
        return &AppConfig{
                IssueIDs:         []string{}, // No issue IDs for codex mode parsing
                RepoURL:          repoURL,
                LinearAPIKey:     linearAPIKey,
                LinearEndpoint:   linearEndpoint,
                Concurrency:      concurrency,
                DryRun:           dryRun,
                BaseBranch:       baseBranch,
                Cleanup:          cleanup,
                CleanupDays:      cleanupDays,
                CodexDockerImage: codexDockerImage,
                CodexCLIArgs:     []string{"--approval-mode", "full-auto"},
                AutomatedMode:    automatedMode,
                OpenAIAPIKey:     openaiKey,
                GitHubToken:      ghToken,
        }, nil
}
