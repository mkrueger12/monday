package cmd

import (
        "fmt"
        "os"
        "os/exec"
        "path/filepath"
        "strings"

        "github.com/spf13/cobra"
        "go.uber.org/zap"

        "monday/linear"
)

// runWorkflow executes the core Monday workflow logic for a given Linear issue and GitHub repository.
// This function can be called from both CLI and HTTP server contexts.
func runWorkflow(issueID, repoURL string) error {
        fmt.Printf("🚀 Starting Monday workflow for %s\n", issueID)
        logger.Info("Starting Monday workflow", 
                zap.String("issue_id", issueID),
                zap.String("repo_url", repoURL))

        linearAPIKey := os.Getenv("LINEAR_API_KEY")
        if linearAPIKey == "" {
                return fmt.Errorf("LINEAR_API_KEY environment variable is required")
        }

        githubToken := os.Getenv("GITHUB_TOKEN")
        if githubToken == "" {
                return fmt.Errorf("GITHUB_TOKEN environment variable is required")
        }

        openaiAPIKey := os.Getenv("OPENAI_API_KEY")
        if openaiAPIKey == "" {
                return fmt.Errorf("OPENAI_API_KEY environment variable is required")
        }

        linearClient := linear.NewClient(linearAPIKey)

        issueID = extractIssueID(issueID)
        logger.Info("Extracted issue ID", zap.String("issue_id", issueID))

        fmt.Printf("📋 Fetching Linear issue details...\n")
        logger.Info("Fetching Linear issue details")
        issue, err := linearClient.FetchIssueDetails(issueID)
        if err != nil {
                return fmt.Errorf("failed to fetch issue details: %w", err)
        }

        fmt.Printf("✅ Issue: %s\n", issue.Title)
        logger.Info("Issue fetched successfully", 
                zap.String("title", issue.Title),
                zap.String("branch_name", issue.BranchName))

        logger.Info("Marking issue as In Progress")
        if err := linearClient.MarkIssueInProgress(issue); err != nil {
                logger.Warn("Failed to mark issue as In Progress", zap.Error(err))
        }

        repoName := extractRepoName(repoURL)
        workDir := filepath.Join(".", repoName)

        currentDir, _ := os.Getwd()
        logger.Info("Starting repository operations", 
                zap.String("current_dir", currentDir),
                zap.String("repo_name", repoName),
                zap.String("target_work_dir", workDir))

        fmt.Printf("📦 Cloning repository...\n")
        logger.Info("Cloning repository", zap.String("repo_url", repoURL))
        if err := runGitCommand("clone", repoURL); err != nil {
                return fmt.Errorf("failed to clone repository: %w", err)
        }

        logger.Info("Changing to repository directory", zap.String("work_dir", workDir))
        if err := os.Chdir(workDir); err != nil {
                return fmt.Errorf("failed to change directory: %w", err)
        }
        
        newDir, _ := os.Getwd()
        logger.Info("Successfully changed directory", zap.String("new_dir", newDir))

        branchName := issue.BranchName
        if branchName == "" {
                branchName = fmt.Sprintf("feature/%s", strings.ToLower(strings.ReplaceAll(issueID, "-", "_")))
        }

        fmt.Printf("🌿 Creating branch: %s\n", branchName)
        logger.Info("Creating feature branch", zap.String("branch_name", branchName))
        if err := runGitCommand("checkout", "-b", branchName); err != nil {
                return fmt.Errorf("failed to create branch: %w", err)
        }

        fmt.Printf("🤖 Running Codex CLI...\n")
        logger.Info("Running Codex CLI", zap.String("description", issue.Description))
        codexPrompt := fmt.Sprintf("%s\n\n%s", issue.Title, issue.Description)
        if err := runCodex(codexPrompt, openaiAPIKey); err != nil {
                return fmt.Errorf("failed to run Codex: %w", err)
        }

        fmt.Printf("📝 Committing and pushing changes...\n")
        
        logger.Info("Checking git status before staging")
        if err := runGitCommand("status", "--porcelain"); err != nil {
                logger.Warn("Failed to check git status", zap.Error(err))
        }
        
        logger.Info("Staging changes")
        if err := runGitCommand("add", "."); err != nil {
                return fmt.Errorf("failed to stage changes: %w", err)
        }
        
        logger.Info("Checking staged changes")
        if err := runGitCommand("diff", "--cached", "--name-only"); err != nil {
                logger.Warn("Failed to check staged changes", zap.Error(err))
        }

        commitMsg := fmt.Sprintf("feat: %s\n\n%s\n\nLinear Issue: %s", issue.Title, issue.Description, issue.URL)
        logger.Info("Committing changes", zap.String("commit_message", commitMsg))
        if err := runGitCommand("commit", "-m", commitMsg); err != nil {
                return fmt.Errorf("failed to commit changes: %w", err)
        }

        logger.Info("Pushing branch to origin")
        if err := runGitCommand("push", "--set-upstream", "origin", branchName); err != nil {
                return fmt.Errorf("failed to push branch: %w", err)
        }

        fmt.Printf("💬 Adding branch link to Linear issue...\n")
        logger.Info("Adding branch link to Linear issue")
        branchURL := fmt.Sprintf("%s/tree/%s", repoURL, branchName)
        comment := fmt.Sprintf("🚀 Branch created and pushed: [%s](%s)", branchName, branchURL)
        
        if err := linearClient.AddIssueComment(issue.ID, comment); err != nil {
                logger.Warn("Failed to add comment to Linear issue", zap.Error(err))
                fmt.Printf("⚠️  Warning: Could not add branch link to Linear issue\n")
        } else {
                fmt.Printf("✅ Posted branch link to Linear issue\n")
        }

        fmt.Printf("✅ Monday workflow completed successfully!\n")
        logger.Info("Monday workflow completed successfully")
        return nil
}

// runMondayWorkflow is the CLI command handler that delegates to runWorkflow.
func runMondayWorkflow(cmd *cobra.Command, args []string) error {
        issueID := args[0]
        return runWorkflow(issueID, repoURL)
}

// extractIssueID parses the input string to extract a Linear issue ID, handling both direct IDs and Linear issue URLs.
func extractIssueID(input string) string {
        if strings.Contains(input, "linear.app") {
                parts := strings.Split(input, "/")
                for i, part := range parts {
                        if part == "issue" && i+1 < len(parts) {
                                issueID := parts[i+1]
                                if queryIndex := strings.Index(issueID, "?"); queryIndex != -1 {
                                        issueID = issueID[:queryIndex]
                                }
                                return issueID
                        }
                }
        }
        return input
}

// extractRepoName returns the repository name extracted from a repository URL, removing any ".git" suffix.
func extractRepoName(repoURL string) string {
        parts := strings.Split(repoURL, "/")
        repoName := parts[len(parts)-1]
        return strings.TrimSuffix(repoName, ".git")
}

// runGitCommand executes a git command with the specified arguments, logging its execution and output based on the verbosity setting.
// Returns an error if the git command fails.
func runGitCommand(args ...string) error {
        wd, _ := os.Getwd()
        logger.Info("Running git command", 
                zap.Strings("args", args),
                zap.String("working_dir", wd))
        
        cmd := exec.Command("git", args...)
        
        if verbose {
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
        } else {
                cmd.Stdout = nil
                cmd.Stderr = os.Stderr
        }
        
        err := cmd.Run()
        if err != nil {
                logger.Error("Git command failed", 
                        zap.Strings("args", args),
                        zap.String("working_dir", wd),
                        zap.Error(err))
        } else {
                logger.Info("Git command completed successfully", zap.Strings("args", args))
        }
        
        return err
}

// runCodex executes the Codex CLI tool with the provided prompt and OpenAI API key.
// The function sets the approval mode to "full-auto" and controls output visibility based on the verbose flag.
// Returns an error if the Codex command fails to execute.
func runCodex(prompt, apiKey string) error {
        cmd := exec.Command("codex", "--model", model, "--approval-mode", "full-auto", "-q", prompt)
        cmd.Env = append(os.Environ(), fmt.Sprintf("OPENAI_API_KEY=%s", apiKey))
        
        if verbose {
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
        } else {
                cmd.Stdout = nil
                cmd.Stderr = nil
        }
        
        logger.Debug("Running Codex", zap.String("prompt", prompt))
        return cmd.Run()
}
