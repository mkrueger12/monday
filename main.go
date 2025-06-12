// Package main implements the monday CLI tool for automating Linear issue development setup.
// The tool fetches Linear issues, marks them as "In Progress", creates git worktrees,
// and launches the Friday CLI in macOS Terminal for each issue.
package main

import (
        "fmt"
        "log"
        "os"
        "sync"

        "golang.org/x/sync/errgroup"

        "monday/config"
        "monday/linear"
        "monday/runner"
)

// Application represents the main monday CLI application with configuration options.
// It orchestrates the entire workflow from parsing arguments to launching development environments.
type Application struct {
        // DryRun indicates whether to simulate operations without actually launching Terminal
        DryRun bool
}

// main is the entry point for the monday CLI application.
// It creates an Application instance and runs it with command-line arguments,
// exiting with a fatal error if the application fails to run successfully.
func main() {
        app := &Application{}
        if err := app.Run(os.Args[1:]); err != nil {
                log.Fatalf("Error: %v", err)
        }
}

// Run executes the main application logic with the provided command-line arguments.
// It parses configuration, validates requirements, prepares the git repository,
// and processes all Linear issues concurrently according to the specified concurrency limit.
// Returns an error if any critical step fails or if any issues fail to process.
func (app *Application) Run(args []string) error {
        // Check for codex subcommand
        if len(args) > 0 && args[0] == "codex" {
                return app.runCodexMode(args[1:])
        }
        // Parse command-line arguments and environment variables into configuration
        cfg, err := config.ParseConfigFromArgs(args)
        if err != nil {
                return fmt.Errorf("failed to parse configuration: %w", err)
        }

        // Handle cleanup mode - clean up and exit
        if cfg.Cleanup {
                log.Printf("Running cleanup mode: removing worktrees older than %d days", cfg.CleanupDays)
                return fmt.Errorf("cleanup mode not supported in containerized workflow")
        }

        // Validate that Linear API key is provided (required for issue processing)
        if cfg.LinearAPIKey == "" {
                return fmt.Errorf("Linear API key is required (set LINEAR_API_KEY env var or use -api-key flag)")
        }

        log.Printf("Processing %d issues with concurrency %d", len(cfg.IssueIDs), cfg.Concurrency)
        log.Printf("Using repository URL: %s", cfg.RepoURL)

        // Initialize Linear API client with authentication
        linearClient := linear.NewClient(cfg.LinearAPIKey)
        if cfg.LinearEndpoint != "" {
                linearClient.SetEndpoint(cfg.LinearEndpoint) // Allow custom endpoint for testing
        }
        
        // Initialize macOS Terminal runner for launching Friday CLI
        macosRunner := runner.NewMacOSRunner()
        
        // Use config's DryRun setting if Application's DryRun is not set
        if !app.DryRun {
                app.DryRun = cfg.DryRun
        }

        // Process issues concurrently using errgroup for controlled concurrency
        g := new(errgroup.Group)
        g.SetLimit(cfg.Concurrency) // Limit concurrent goroutines to prevent overwhelming APIs

        // Thread-safe counters for tracking success/failure rates
        var mu sync.Mutex
        var successCount, errorCount int

        // Launch a goroutine for each issue to process them concurrently
        for _, issueID := range cfg.IssueIDs {
                issueID := issueID // capture loop variable to avoid closure issues
                g.Go(func() error {
                        if err := app.processIssue(issueID, cfg, linearClient, macosRunner); err != nil {
                                log.Printf("[%s] Error: %v", issueID, err)
                                mu.Lock()
                                errorCount++
                                mu.Unlock()
                                return nil // Don't fail the entire group for individual issue errors
                        }
                        
                        mu.Lock()
                        successCount++
                        mu.Unlock()
                        log.Printf("[%s] Successfully processed", issueID)
                        return nil
                })
        }

        // Wait for all goroutines to complete
        if err := g.Wait(); err != nil {
                return fmt.Errorf("failed to process issues: %w", err)
        }

        // Report final results
        log.Printf("Completed: %d successful, %d errors", successCount, errorCount)

        // Return error if any issues failed to process
        if errorCount > 0 {
                return fmt.Errorf("%d issues failed to process", errorCount)
        }

        return nil
}

// processIssue handles the complete workflow for a single Linear issue.
// It fetches issue details, marks the issue as "In Progress", creates a git worktree,
// creates a feature file with issue context, and launches Friday CLI in Terminal.
// This function is designed to be called concurrently for multiple issues.
func (app *Application) processIssue(issueID string, cfg *config.AppConfig, linearClient *linear.Client, macosRunner *runner.MacOSRunner) error {
        log.Printf("[%s] Processing issue in containerized mode...", issueID)
        
        // For containerized mode, we don't use the traditional worktree approach
        // Instead, everything happens inside containers via the codex flow
        return fmt.Errorf("use 'monday codex %s' for containerized workflow", issueID)
}

func (app *Application) runCodexMode(args []string) error {
        if len(args) == 0 {
                return fmt.Errorf("usage: monday codex <issue_id>")
        }
        
        issueID := args[0]
        
        // For codex mode, we need to parse flags without expecting issue IDs
        cfg, err := config.ParseCodexConfigFromArgs(args[1:])
        if err != nil {
                return fmt.Errorf("failed to parse configuration: %w", err)
        }
        
        if cfg.LinearAPIKey == "" {
                return fmt.Errorf("Linear API key is required (set LINEAR_API_KEY env var or use -api-key flag)")
        }
        
        if cfg.OpenAIAPIKey == "" {
                return fmt.Errorf("OpenAI API key is required (set OPENAI_API_KEY env var or use -openai-api-key flag)")
        }
        
        if cfg.RepoURL == "" {
                return fmt.Errorf("repository URL is required (use --repo-url flag)")
        }
        
        if cfg.GitHubToken == "" {
                return fmt.Errorf("GitHub token is required (set GITHUB_TOKEN env var or use --github-token flag)")
        }
        
        log.Printf("Running Codex automation for issue: %s", issueID)
        return runner.CodexFlow(cfg, issueID)
}
