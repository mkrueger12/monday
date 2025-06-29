# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

Build the CLI:
```bash
make build
# or
go build -o bin/monday .
```

Run tests:
```bash
make test
# or
go test ./...
```

Run a single test:
```bash
go test ./cmd -v -run TestWorkflow
go test ./linear -v -run TestClient
```

Install dependencies:
```bash
make deps
# or
go mod tidy && go mod download
```

## Architecture Overview

Monday is a DevFlow orchestrator CLI built in Go with Cobra. It automates Linear issue development workflows by:

1. **Linear Integration** (`linear/` package): GraphQL client for fetching Linear issues and updating status
2. **Workflow Engine** (`cmd/workflow.go`): Core orchestration logic that coordinates all steps
3. **HTTP Server** (`cmd/server.go`): REST API for triggering workflows remotely
4. **CLI Interface** (`cmd/root.go`): Command-line interface using Cobra framework

### Key Components

- **Main Entry**: `main.go` delegates to `cmd.Execute()`
- **Linear Client**: `linear/client.go` handles GraphQL API communication with authentication, issue fetching, and status updates
- **Workflow Logic**: `cmd/workflow.go` contains `runWorkflow()` function that orchestrates: Linear API calls → Git operations → Codex CLI execution → PR creation
- **Server Mode**: `cmd/server.go` provides HTTP endpoints (`/health`, `/trigger`) for remote workflow execution

### External Dependencies

The workflow requires these external tools to be installed:
- `git` - Repository operations
- `gh` (GitHub CLI) - Pull request creation  
- `claude` - Claude Code CLI for AI-powered development

### Environment Variables

Required for all operations:
- `LINEAR_API_KEY` - Linear API authentication
- `GITHUB_TOKEN` - GitHub API authentication  
- `ANTHROPIC_API_KEY` - Anthropic API for Claude Code

Server mode additionally requires:
- `SERVER_API_KEY` - HTTP API authentication
- `PORT` - Server port (optional, defaults to 8080)

### Workflow Steps

1. Parse Linear issue ID (supports both `DEL-163` format and full URLs)
2. Fetch issue details via Linear GraphQL API
3. Mark issue as "In Progress" 
4. Clone GitHub repository and create feature branch
5. Execute Claude Code CLI with issue title/description as prompt
6. Commit changes with structured message including Linear issue link
7. Push branch and create pull request via GitHub CLI

### Error Handling

- Comprehensive zap logging with development/production modes
- Graceful handling of missing environment variables
- API error propagation with detailed context
- Git command failure reporting with working directory context