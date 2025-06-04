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
        "monday/gitops"
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
        // Parse command-line arguments and environment variables into configuration
        cfg, err := config.ParseConfigFromArgs(args)
        if err != nil {
                return fmt.Errorf("failed to parse configuration: %w", err)
        }

        worktreeRoot := "/tmp/monday-worktrees"

        // Handle cleanup mode - clean up and exit
        if cfg.Cleanup {
                log.Printf("Running cleanup mode: removing worktrees older than %d days", cfg.CleanupDays)
                return gitops.CleanWorktrees(worktreeRoot, cfg.GitRepoPath, cfg.CleanupDays, cfg.DryRun)
        }

        // Validate that Linear API key is provided (required for issue processing)
        if cfg.LinearAPIKey == "" {
                return fmt.Errorf("Linear API key is required (set LINEAR_API_KEY env var or use -api-key flag)")
        }

        // Always run automatic cleanup before processing issues
        log.Printf("Running automatic cleanup: removing worktrees older than %d days", cfg.CleanupDays)
        if err := gitops.CleanWorktrees(worktreeRoot, cfg.GitRepoPath, cfg.CleanupDays, cfg.DryRun); err != nil {
                return fmt.Errorf("automatic cleanup failed: %w", err)
        }

        log.Printf("Processing %d issues with concurrency %d", len(cfg.IssueIDs), cfg.Concurrency)
        log.Printf("Using repository path: %s", cfg.GitRepoPath)

        // Ensure the git repository is ready and on the correct base branch
        if err := gitops.PrepareRepository(cfg.GitRepoPath, cfg.BaseBranch); err != nil {
                return fmt.Errorf("failed to prepare repository: %w", err)
        }

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
                        if err := app.processIssue(issueID, cfg, linearClient, macosRunner, worktreeRoot); err != nil {
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
func (app *Application) processIssue(issueID string, cfg *config.AppConfig, linearClient *linear.Client, macosRunner *runner.MacOSRunner, worktreeRoot string) error {
        // Step 1: Fetch detailed information about the Linear issue
        log.Printf("[%s] Fetching issue details...", issueID)
        issue, err := linearClient.FetchIssueDetails(issueID)
        if err != nil {
                return fmt.Errorf("failed to fetch issue details: %w", err)
        }

        // Step 2: Update the issue status to "In Progress" in Linear
        log.Printf("[%s] Marking issue as In Progress...", issueID)
        if err := linearClient.MarkIssueInProgress(issue); err != nil {
                return fmt.Errorf("failed to mark issue as in progress: %w", err)
        }

        // Step 3: Create an isolated git worktree for this issue
        log.Printf("[%s] Creating worktree at", issueID)
        worktreePath, err := gitops.CreateWorktreeForIssue(worktreeRoot, cfg.GitRepoPath, issueID, cfg.BaseBranch)
        if err != nil {
                return fmt.Errorf("failed to create worktree: %w", err)
        }

        // Step 4: Create a _feature.md file with issue context for the developer
        log.Printf("[%s] Creating feature file...", issueID)
        if err := gitops.CreateFeatureFile(worktreePath, issue.Title, issue.Description); err != nil {
                return fmt.Errorf("failed to create feature file: %w", err)
        }

        // Step 5: Launch Friday CLI in macOS Terminal (unless in dry-run mode)
        if !app.DryRun {
                log.Printf("[%s] Launching Friday in Terminal...", issueID)
                if err := macosRunner.LaunchFridayInMacOSTerminal(worktreePath, issue.Title); err != nil {
                        return fmt.Errorf("failed to launch Friday: %w", err)
                }
        } else {
                log.Printf("[%s] Dry run: would launch Friday in Terminal at %s", issueID, worktreePath)
        }

        return nil
}
