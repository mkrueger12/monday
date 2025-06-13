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

func runMondayWorkflow(cmd *cobra.Command, args []string) error {
        issueID := args[0]
        
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

        fmt.Printf("📦 Cloning repository...\n")
        logger.Info("Cloning repository", zap.String("repo_url", repoURL))
        if err := runGitCommand("clone", repoURL); err != nil {
                return fmt.Errorf("failed to clone repository: %w", err)
        }

        logger.Info("Changing to repository directory", zap.String("work_dir", workDir))
        if err := os.Chdir(workDir); err != nil {
                return fmt.Errorf("failed to change directory: %w", err)
        }

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
        logger.Info("Staging changes")
        if err := runGitCommand("add", "."); err != nil {
                return fmt.Errorf("failed to stage changes: %w", err)
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

        fmt.Printf("🚀 Creating pull request...\n")
        logger.Info("Creating pull request")
        if err := createPullRequest(issue, githubToken); err != nil {
                return fmt.Errorf("failed to create pull request: %w", err)
        }

        fmt.Printf("✅ Monday workflow completed successfully!\n")
        logger.Info("Monday workflow completed successfully")
        return nil
}

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

func extractRepoName(repoURL string) string {
        parts := strings.Split(repoURL, "/")
        repoName := parts[len(parts)-1]
        return strings.TrimSuffix(repoName, ".git")
}

func runCommand(name string, args ...string) error {
        cmd := exec.Command(name, args...)
        
        if verbose {
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
        } else {
                cmd.Stdout = nil
                cmd.Stderr = nil
        }
        
        logger.Debug("Running command", zap.String("command", name), zap.Strings("args", args))
        return cmd.Run()
}

func runGitCommand(args ...string) error {
        cmd := exec.Command("git", args...)
        
        if verbose {
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
        } else {
                cmd.Stdout = nil
                cmd.Stderr = os.Stderr
        }
        
        logger.Debug("Running git command", zap.Strings("args", args))
        return cmd.Run()
}

func runCodex(prompt, apiKey string) error {
        cmd := exec.Command("codex", "--approval-mode", "full-auto", "-q", prompt)
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

func createPullRequest(issue *linear.IssueDetails, token string) error {
        prTitle := fmt.Sprintf("feat: %s", issue.Title)
        prBody := fmt.Sprintf("%s\n\nLinear Issue: %s", issue.Description, issue.URL)
        
        cmd := exec.Command("gh", "pr", "create", "--title", prTitle, "--body", prBody)
        cmd.Env = append(os.Environ(), fmt.Sprintf("GITHUB_TOKEN=%s", token))
        
        if verbose {
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
        } else {
                cmd.Stdout = nil
                cmd.Stderr = nil
        }
        
        logger.Debug("Creating PR", zap.String("title", prTitle))
        return cmd.Run()
}
