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

type Application struct {
        DryRun bool
}

func main() {
        app := &Application{}
        if err := app.Run(os.Args[1:]); err != nil {
                log.Fatalf("Error: %v", err)
        }
}

func (app *Application) Run(args []string) error {
        cfg, err := config.ParseConfigFromArgs(args)
        if err != nil {
                return fmt.Errorf("failed to parse configuration: %w", err)
        }

        if cfg.LinearAPIKey == "" {
                return fmt.Errorf("Linear API key is required (set LINEAR_API_KEY env var or use -api-key flag)")
        }

        log.Printf("Processing %d issues with concurrency %d", len(cfg.IssueIDs), cfg.Concurrency)

        if err := gitops.PrepareRepository(cfg.GitRepoPath, "main"); err != nil {
                return fmt.Errorf("failed to prepare repository: %w", err)
        }

        linearClient := linear.NewClient(cfg.LinearAPIKey)
        if cfg.LinearEndpoint != "" {
                linearClient.SetEndpoint(cfg.LinearEndpoint)
        }
        
        macosRunner := runner.NewMacOSRunner()
        
        // Use config's DryRun setting if Application's DryRun is not set
        if !app.DryRun {
                app.DryRun = cfg.DryRun
        }

        // Process issues concurrently
        g := new(errgroup.Group)
        g.SetLimit(cfg.Concurrency)

        var mu sync.Mutex
        var successCount, errorCount int

        for _, issueID := range cfg.IssueIDs {
                issueID := issueID // capture loop variable
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

        if err := g.Wait(); err != nil {
                return fmt.Errorf("failed to process issues: %w", err)
        }

        log.Printf("Completed: %d successful, %d errors", successCount, errorCount)

        if errorCount > 0 {
                return fmt.Errorf("%d issues failed to process", errorCount)
        }

        return nil
}

func (app *Application) processIssue(issueID string, cfg *config.AppConfig, linearClient *linear.Client, macosRunner *runner.MacOSRunner) error {
        log.Printf("[%s] Fetching issue details...", issueID)
        issue, err := linearClient.FetchIssueDetails(issueID)
        if err != nil {
                return fmt.Errorf("failed to fetch issue details: %w", err)
        }

        log.Printf("[%s] Marking issue as In Progress...", issueID)
        if err := linearClient.MarkIssueInProgress(issueID); err != nil {
                return fmt.Errorf("failed to mark issue as in progress: %w", err)
        }

        log.Printf("[%s] Creating worktree...", issueID)
        worktreePath, err := gitops.CreateWorktreeForIssue(cfg.GitRepoPath, issueID, "main")
        if err != nil {
                return fmt.Errorf("failed to create worktree: %w", err)
        }

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
