package runner

import (
        "fmt"
        "os"
        "os/exec"
        "path/filepath"
        "strings"

        "monday/config"
        "monday/linear"
        "monday/gitops"
)

func CodexFlow(cfg *config.AppConfig, issueID string) error {
        if cfg.OpenAIAPIKey == "" {
                return fmt.Errorf("OpenAI API key is required")
        }

        client := linear.NewClient(cfg.LinearAPIKey)
        if cfg.LinearEndpoint != "" {
                client.SetEndpoint(cfg.LinearEndpoint)
        }

        issue, err := client.FetchIssueDetails(issueID)
        if err != nil {
                return fmt.Errorf("failed to fetch issue details: %w", err)
        }

        worktreePath, err := gitops.CreateWorktreeForIssue(
                cfg.WorktreeRoot,
                cfg.GitRepoPath,
                issueID,
                cfg.BaseBranch,
        )
        if err != nil {
                return fmt.Errorf("failed to create worktree: %w", err)
        }

        prompt := buildCodexPrompt(issue)
        
        promptFile := filepath.Join(worktreePath, "codex_prompt.txt")
        if err := os.WriteFile(promptFile, []byte(prompt), 0644); err != nil {
                return fmt.Errorf("failed to write prompt file: %w", err)
        }

        env := map[string]string{
                "OPENAI_API_KEY": cfg.OpenAIAPIKey,
        }
        
        cliArgs := append(cfg.CodexCLIArgs, "-q", prompt)
        
        if err := RunCodexInDocker(cfg.CodexDockerImage, cliArgs, worktreePath, env); err != nil {
                return fmt.Errorf("codex execution failed: %w", err)
        }

        if cfg.AutomatedMode {
                if err := commitAndPushChanges(worktreePath, issue); err != nil {
                        return fmt.Errorf("failed to commit changes: %w", err)
                }
                
                if err := client.MarkIssueInProgress(issue); err != nil {
                        return fmt.Errorf("failed to mark issue in progress: %w", err)
                }
        }

        return nil
}

func buildCodexPrompt(issue *linear.IssueDetails) string {
        var parts []string
        
        parts = append(parts, fmt.Sprintf("Implement feature: %s", issue.Title))
        
        if issue.Description != "" {
                parts = append(parts, fmt.Sprintf("Description: %s", issue.Description))
        }
        
        parts = append(parts, "Please implement this feature following best practices and include appropriate tests.")
        
        return strings.Join(parts, "\n\n")
}

func commitAndPushChanges(worktreePath string, issue *linear.IssueDetails) error {
        commands := [][]string{
                {"git", "add", "."},
                {"git", "commit", "-m", fmt.Sprintf("feat: %s\n\nImplemented via automated code generation\n\nLinear issue: %s", issue.Title, issue.URL)},
                {"git", "push", "origin", "HEAD"},
        }
        
        for _, cmdArgs := range commands {
                cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
                cmd.Dir = worktreePath
                if output, err := cmd.CombinedOutput(); err != nil {
                        return fmt.Errorf("failed to run %s: %s (exit code: %w)", strings.Join(cmdArgs, " "), string(output), err)
                }
        }
        
        return nil
}
