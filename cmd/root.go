package cmd

import (
        "fmt"
        "os"

        "github.com/spf13/cobra"
        "go.uber.org/zap"
)

var (
        logger   *zap.Logger
        repoURL  string
        verbose  bool
)

var rootCmd = &cobra.Command{
        Use:   "monday <linear_issue_id>",
        Short: "DevFlow Orchestrator - Automate Linear issue development workflow",
        Long: `Monday CLI automates the development workflow by:
1. Fetching Linear issue details
2. Cloning GitHub repository and creating feature branch
3. Running Codex CLI for automated development
4. Committing changes and creating pull request`,
        Args: cobra.ExactArgs(1),
        PersistentPreRun: func(cmd *cobra.Command, args []string) {
                initLogger()
        },
        RunE: runMondayWorkflow,
}

// Execute runs the root CLI command and handles any execution errors by logging or printing them, then exits with a non-zero status on failure.
func Execute() {
        if err := rootCmd.Execute(); err != nil {
                if logger != nil {
                        logger.Error("Command execution failed", zap.Error(err))
                } else {
                        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                }
                os.Exit(1)
        }
}

// init configures persistent and required flags for the CLI, including verbose logging and the GitHub repository URL.
func init() {
        rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
        rootCmd.Flags().StringVar(&repoURL, "repo-url", "", "GitHub repository URL (required)")
        rootCmd.MarkFlagRequired("repo-url")
}

// initLogger initializes the global logger with either development or production settings based on the verbose flag.
// Exits the program if logger initialization fails.
func initLogger() {
        var err error
        if verbose {
                logger, err = zap.NewDevelopment()
        } else {
                config := zap.NewProductionConfig()
                config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
                logger, err = config.Build()
        }
        if err != nil {
                fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
                os.Exit(1)
        }
}
