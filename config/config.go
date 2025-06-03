package config

import (
        "flag"
        "fmt"
        "os"
)

type AppConfig struct {
        IssueIDs        []string
        GitRepoPath     string
        LinearAPIKey    string
        LinearEndpoint  string
        Concurrency     int
        DryRun          bool
}

func ParseConfig() (*AppConfig, error) {
        return ParseConfigFromArgs(os.Args[1:])
}

func ParseConfigFromArgs(args []string) (*AppConfig, error) {
        var concurrency int
        var apiKey string
        var linearEndpoint string
        var dryRun bool

        fs := flag.NewFlagSet("monday", flag.ContinueOnError)
        fs.IntVar(&concurrency, "concurrency", 3, "Number of concurrent issue processors")
        fs.StringVar(&apiKey, "api-key", "", "Linear API key (overrides LINEAR_API_KEY env var)")
        fs.StringVar(&linearEndpoint, "linear-endpoint", "", "Linear API endpoint (for testing)")
        fs.BoolVar(&dryRun, "dry-run", false, "Don't actually launch Terminal")
        
        err := fs.Parse(args)
        if err != nil {
                return nil, err
        }

        remainingArgs := fs.Args()
        if len(remainingArgs) < 2 {
                return nil, fmt.Errorf("usage: monday [flags] <issue_id_1> [issue_id_2 ...] <git_repo_path>")
        }

        gitRepoPath := remainingArgs[len(remainingArgs)-1]
        issueIDs := remainingArgs[:len(remainingArgs)-1]

        if len(issueIDs) == 0 {
                return nil, fmt.Errorf("at least one issue ID is required")
        }

        linearAPIKey := apiKey
        if linearAPIKey == "" {
                linearAPIKey = os.Getenv("LINEAR_API_KEY")
        }

        return &AppConfig{
                IssueIDs:       issueIDs,
                GitRepoPath:    gitRepoPath,
                LinearAPIKey:   linearAPIKey,
                LinearEndpoint: linearEndpoint,
                Concurrency:    concurrency,
                DryRun:         dryRun,
        }, nil
}
