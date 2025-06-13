// Package main implements the monday CLI tool for automating Linear issue development.
// The tool fetches Linear issues and runs automated Codex workflows for development.
package main

import (
        "fmt"
        "log"
        "os"
        "strconv"
        "strings"

        "monday/config"
        "monday/linear"
        "monday/runner"
)

// Application represents the main monday CLI application.
// It orchestrates automated Codex workflows for development.
type Application struct {}

// main is the entry point for the monday CLI application.
// It runs automated Codex workflows for development.
func main() {
        app := &Application{}
        if err := app.Run(os.Args[1:]); err != nil {
                log.Fatalf("Error: %v", err)
        }
}

// Run executes the main application logic with the provided command-line arguments.
// It runs automated Codex workflows for development.
// Returns an error if any critical step fails.
func (app *Application) Run(args []string) error {
        // Parse configuration for Codex mode
        cfg, err := config.ParseCodexConfigFromArgs(args)
        if err != nil {
                return fmt.Errorf("failed to parse configuration: %w", err)
        }
        
        if cfg.LinearAPIKey == "" {
                return fmt.Errorf("Linear API key is required (set LINEAR_API_KEY env var or use --api-key flag)")
        }
        
        if cfg.OpenAIAPIKey == "" {
                return fmt.Errorf("OpenAI API key is required (set OPENAI_API_KEY env var or use --openai-api-key flag)")
        }
        
        if cfg.RepoURL == "" {
                return fmt.Errorf("repository URL is required (use --repo-url flag)")
        }
        
        if cfg.GitHubToken == "" {
                return fmt.Errorf("GitHub token is required (set GITHUB_TOKEN env var or use --github-token flag)")
        }
        
        // Initialize Linear client
        linearClient := linear.NewClient(cfg.LinearAPIKey)
        if cfg.LinearEndpoint != "" {
                linearClient.SetEndpoint(cfg.LinearEndpoint)
        }
        
        // Determine which issue to process
        var issueID string
        var issue *linear.IssueDetails
        
        // Check if a specific issue ID was provided in the parsed configuration
        if len(cfg.IssueIDs) > 0 {
                issueID = cfg.IssueIDs[0] // Use the first issue ID provided
                log.Printf("Using specified issue ID: %s", issueID)
                
                // Fetch the specific issue
                var err error
                issue, err = linearClient.FetchIssueDetails(issueID)
                if err != nil {
                        return fmt.Errorf("failed to fetch issue %s: %w", issueID, err)
                }
        } else {
                // No specific issue ID provided, find first available issue based on filters
                // All provided filters are applied as AND conditions
                log.Printf("No issue ID provided, searching for issues with filters: team=%s, project=%s, tag=%s", 
                        cfg.LinearTeam, cfg.LinearProject, cfg.LinearTag)
                
                // Fetch issues based on filters (all filters are AND conditions)
                issues, err := linearClient.FetchIssuesByFilters(cfg.LinearTeam, cfg.LinearProject, cfg.LinearTag)
                if err != nil {
                        return fmt.Errorf("failed to fetch issues: %w", err)
                }
                
                if len(issues) == 0 {
                        return fmt.Errorf("no issues found matching the specified filters (team=%s, project=%s, tag=%s)", 
                                cfg.LinearTeam, cfg.LinearProject, cfg.LinearTag)
                }
                
                // Use the first issue that matches ALL specified criteria
                issue = &issues[0]
                issueID = fmt.Sprintf("%s-%d", cfg.LinearTeam, extractIssueNumber(issue.URL))
                log.Printf("Selected first available issue matching all criteria: %s - %s", issueID, issue.Title)
        }
        
        log.Printf("Running Codex automation for issue: %s", issueID)
        return runner.CodexFlow(cfg, issueID)
}



// extractIssueNumber extracts the issue number from a Linear issue URL
func extractIssueNumber(url string) int {
        // Linear URLs typically end with the issue identifier like /issue/TEAM-123
        parts := strings.Split(url, "/")
        if len(parts) > 0 {
                lastPart := parts[len(parts)-1]
                if strings.Contains(lastPart, "-") {
                        numberPart := strings.Split(lastPart, "-")
                        if len(numberPart) > 1 {
                                if num, err := strconv.Atoi(numberPart[len(numberPart)-1]); err == nil {
                                        return num
                                }
                        }
                }
        }
        return 1 // fallback
}
