# Monday CLI - Usage Guide

## Overview
The Monday CLI provides automated Linear issue development through containerized Codex workflows with precise issue selection.

## CLI Interface

### Basic Usage with Specific Issue ID
```bash
monday DEL-123 --repo-url https://github.com/user/repo.git
```

### Usage with Linear Filters (Auto-select First Matching Issue)
All provided filters are applied as AND conditions - an issue must match ALL specified criteria.

```bash
# Filter by team only
monday --linear-team ENG --repo-url https://github.com/user/repo.git

# Filter by team AND project (both required)
monday --linear-team ENG --linear-project backend --repo-url https://github.com/user/repo.git

# Filter by team AND project AND tag (all three required)
monday --linear-team ENG --linear-project backend --linear-tag bug --repo-url https://github.com/user/repo.git

# No filters (searches all available issues)
monday --repo-url https://github.com/user/repo.git
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
1. **Issue Selection**: Either use provided issue ID or auto-select first issue matching ALL specified filter criteria
2. **Repository Setup**: Clone repository in container
3. **Dependency Detection**: Auto-detect and install dependencies (Node.js, Python, Ruby, Go, Rust)
4. **Code Generation**: Run OpenAI Codex with issue context
5. **Git Operations**: Create branch, commit changes, push to GitHub
6. **Pull Request**: Automatically create PR with issue details
7. **Linear Update**: Mark issue as "In Progress"

## Examples

### Specific Issue
```bash
monday DEL-163 \
  --repo-url https://github.com/myorg/myrepo.git
```

### Auto-select from Engineering Team
```bash
monday \
  --linear-team ENG \
  --repo-url https://github.com/myorg/myrepo.git
```

### Auto-select Bug from Backend Project (All Criteria Required)
```bash
monday \
  --linear-team ENG \
  --linear-project backend \
  --linear-tag bug \
  --repo-url https://github.com/myorg/myrepo.git
```

**Note**: This will only select issues that are on the ENG team AND in the backend project AND have the bug tag.

All operations are ephemeral and containerized - no local state is maintained.
