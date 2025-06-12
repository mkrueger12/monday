# Monday CLI - Enhanced Usage Guide

## Overview
The Monday CLI now supports flexible Linear issue selection through team, project, and tag filters, with full containerized automation.

## New CLI Interface

### Basic Usage with Specific Issue ID
```bash
monday codex DEL-123 --repo-url https://github.com/user/repo.git
```

### Usage with Linear Filters (Auto-select First Available Issue)
```bash
# Filter by team only
monday codex --linear-team ENG --repo-url https://github.com/user/repo.git

# Filter by team and project
monday codex --linear-team ENG --linear-project backend --repo-url https://github.com/user/repo.git

# Filter by team, project, and tag
monday codex --linear-team ENG --linear-project backend --linear-tag bug --repo-url https://github.com/user/repo.git

# No filters (uses first available team/project)
monday codex --repo-url https://github.com/user/repo.git
```

## Required Environment Variables
```bash
export LINEAR_API_KEY="your_linear_api_key"
export OPENAI_API_KEY="your_openai_api_key"  # Passed to container
export GITHUB_TOKEN="your_github_token"
```

## CLI Flags
- `--linear-team`: Linear team key to filter issues by
- `--linear-project`: Linear project key to filter issues by  
- `--linear-tag`: Linear tag/label to filter issues by
- `--repo-url`: GitHub repository URL to clone (required)
- `--api-key`: Linear API key (overrides env var)
- `--openai-api-key`: OpenAI API key (overrides env var)
- `--github-token`: GitHub token (overrides env var)

## Workflow
1. **Issue Selection**: Either use provided issue ID or auto-select first available issue based on filters
2. **Repository Setup**: Clone repository in container
3. **Dependency Detection**: Auto-detect and install dependencies (Node.js, Python, Ruby, Go, Rust)
4. **Code Generation**: Run OpenAI Codex with issue context
5. **Git Operations**: Create branch, commit changes, push to GitHub
6. **Pull Request**: Automatically create PR with issue details
7. **Linear Update**: Mark issue as "In Progress"

## Examples

### Specific Issue
```bash
monday codex DEL-163 \
  --repo-url https://github.com/myorg/myrepo.git
```

### Auto-select from Engineering Team
```bash
monday codex \
  --linear-team ENG \
  --repo-url https://github.com/myorg/myrepo.git
```

### Auto-select Bug from Backend Project
```bash
monday codex \
  --linear-team ENG \
  --linear-project backend \
  --linear-tag bug \
  --repo-url https://github.com/myorg/myrepo.git
```

All operations are ephemeral and containerized - no local state is maintained.
