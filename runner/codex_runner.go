package runner

import (
        "fmt"
        "os"
        "os/exec"
        "regexp"
        "strings"

        "monday/config"
        "monday/linear"
)

func CodexFlow(cfg *config.AppConfig, issueID string) error {
        if cfg.OpenAIAPIKey == "" {
                return fmt.Errorf("OpenAI API key is required")
        }
        
        if cfg.RepoURL == "" {
                return fmt.Errorf("repository URL is required")
        }
        
        if cfg.GitHubToken == "" {
                return fmt.Errorf("GitHub token is required")
        }

        client := linear.NewClient(cfg.LinearAPIKey)
        if cfg.LinearEndpoint != "" {
                client.SetEndpoint(cfg.LinearEndpoint)
        }

        issue, err := client.FetchIssueDetails(issueID)
        if err != nil {
                return fmt.Errorf("failed to fetch issue details: %w", err)
        }

        branchName := sanitizeBranchName(issueID)
        repoName := extractRepoName(cfg.RepoURL)
        
        fmt.Printf("🚀 Starting containerized workflow for issue %s\n", issueID)
        fmt.Printf("📦 Repository: %s\n", cfg.RepoURL)
        fmt.Printf("🌿 Branch: %s\n", branchName)

        if err := runContainerizedWorkflow(cfg, issue, branchName, repoName); err != nil {
                return fmt.Errorf("containerized workflow failed: %w", err)
        }

        if err := client.MarkIssueInProgress(issue); err != nil {
                return fmt.Errorf("failed to mark issue in progress: %w", err)
        }

        fmt.Printf("✅ Successfully completed workflow for issue %s\n", issueID)
        return nil
}

func buildCodexPrompt(issue *linear.IssueDetails) string {
        var parts []string
        
        parts = append(parts, issue.Title)
        
        if issue.Description != "" {
                parts = append(parts, issue.Description)
        }
        
        return strings.Join(parts, "\n\n")
}

func runContainerizedWorkflow(cfg *config.AppConfig, issue *linear.IssueDetails, branchName, repoName string) error {
        workspaceDir := "/workspace"
        prompt := buildCodexPrompt(issue)
        
        // Build the complete workflow script
        script := fmt.Sprintf(`#!/bin/bash
set -e

echo "🔄 Cloning repository..."
git clone --depth=1 %s %s
cd %s

echo "🔧 Setting up git configuration..."
git config user.name "monday-bot"
git config user.email "bot@monday.com"

echo "🌿 Creating and switching to feature branch..."
git checkout -b %s

echo "🔍 Installing dependencies..."
/app/detect-and-install.sh

echo "🤖 Running Codex code generation..."
echo '%s' > /tmp/codex_prompt.txt

echo "📝 Generating code with prompt:"
cat /tmp/codex_prompt.txt

echo "🔥 Executing Codex in full auto mode..."
codex --approval-mode full-auto --quiet "$(cat /tmp/codex_prompt.txt)"

echo "💾 Committing changes..."
git add .
if git diff --cached --quiet; then
    echo "⚠️  No changes to commit"
else
    git commit -m "feat: %s

Implemented via automated code generation

Linear issue: %s"
    
    echo "🚀 Pushing to GitHub..."
    git push origin %s
    
    echo "🔀 Creating pull request..."
    gh pr create \
        --repo %s \
        --head %s \
        --title "feat: %s" \
        --body "Automated implementation for Linear issue: %s

## Changes
- Implemented feature: %s

## Linear Issue
%s"
fi

echo "✅ Workflow completed successfully"
`, cfg.RepoURL, workspaceDir, workspaceDir, branchName, prompt, issue.Title, issue.URL, branchName, repoName, branchName, issue.Title, issue.URL, issue.Title, issue.URL)

        return runInContainer(cfg, []string{"bash", "-c", script})
}

func extractRepoName(repoURL string) string {
        parts := strings.Split(repoURL, "/")
        if len(parts) >= 2 {
                repo := parts[len(parts)-1]
                return strings.TrimSuffix(repo, ".git")
        }
        return "unknown"
}

func runInContainer(cfg *config.AppConfig, cmd []string) error {
        args := []string{
                "run", "--rm",
                "-e", "GITHUB_TOKEN=" + cfg.GitHubToken,
                "-e", "OPENAI_API_KEY=" + cfg.OpenAIAPIKey,
                cfg.CodexDockerImage,
        }
        args = append(args, cmd...)
        
        dockerCmd := exec.Command("docker", args...)
        dockerCmd.Stdout = os.Stdout
        dockerCmd.Stderr = os.Stderr
        
        return dockerCmd.Run()
}

func sanitizeBranchName(name string) string {
        if name == "" {
                return "issue"
        }

        reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
        sanitized := reg.ReplaceAllString(name, "-")

        sanitized = strings.Trim(sanitized, "-")

        if sanitized == "" {
                return "issue"
        }

        return sanitized
}
