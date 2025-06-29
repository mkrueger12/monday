# Monday - DevFlow Orchestrator CLI

Monday is a command-line interface (CLI) tool designed to streamline the initial development setup process. It automates fetching Linear issue details, cloning a GitHub repository, creating a feature branch, and then executing Codex CLI for automated development, followed by committing changes and creating a pull request.

## Features

- 🔗 **Linear Integration**: Fetch issue details and mark issues as "In Progress"
- 🚀 **GitHub Automation**: Clone repositories, create feature branches, and open PRs
- 🤖 **AI-Powered Development**: Integrate with Claude Code for automated development
- 📝 **Structured Logging**: Comprehensive logging with Zap for debugging and monitoring
- 🔐 **Secure Credentials**: Environment variable-based authentication
- 🌐 **HTTP Server**: REST API endpoints for triggering workflows remotely

## Installation

### Prerequisites

Ensure you have the following tools installed:
- [Git](https://git-scm.com/)
- [GitHub CLI (gh)](https://cli.github.com/)
- [Claude Code CLI](https://claude.ai/code) 

### Build from Source

```bash
git clone <this-repo>
cd monday
go build -o monday .
```

## Configuration

Set the following environment variables:

```bash
export LINEAR_API_KEY="your-linear-api-key"
export GITHUB_TOKEN="your-github-token"
export ANTHROPIC_API_KEY="your-anthropic-api-key"
```

### Getting API Keys

1. **Linear API Key**: Go to Linear Settings → API → Create new API key
2. **GitHub Token**: Go to GitHub Settings → Developer settings → Personal access tokens → Generate new token
   - Required scopes: `repo`, `workflow`
3. **Anthropic API Key**: Get from [Anthropic Console](https://console.anthropic.com/)

## Usage

### CLI Usage

```bash
monday <linear_issue_id> --repo-url <github_repo_url>
```

#### Examples

```bash
# Using Linear issue ID
monday DEL-163 --repo-url https://github.com/username/repo

# Using full Linear issue URL
monday https://linear.app/company/issue/DEL-163 --repo-url https://github.com/username/repo

# With verbose logging
monday DEL-163 --repo-url https://github.com/username/repo --verbose
```

### HTTP Server Usage

Start the HTTP server to trigger workflows via REST API:

```bash
# Set server API key
export SERVER_API_KEY="your-secure-api-key"

# Start server (default port 8080)
monday server

# Start server on custom port
monday server --port 9090
```

#### API Endpoints

**Health Check**
```bash
GET /health
```
Returns: `OK` (200 status)

**Trigger Workflow**
```bash
POST /trigger
Content-Type: application/json
X-API-Key: your-secure-api-key

{
  "linear_id": "DEL-163",
  "github_url": "https://github.com/username/repo"
}
```
Returns: `{"status":"started","message":"Workflow started for Linear issue DEL-163"}` (202 status)

#### API Examples

```bash
# Health check
curl http://localhost:8080/health

# Trigger workflow
curl -X POST http://localhost:8080/trigger \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secure-api-key" \
  -d '{
    "linear_id": "DEL-163",
    "github_url": "https://github.com/username/repo"
  }'
```

## Workflow

When you run Monday, it performs the following steps:

1. **Fetch Linear Issue**: Retrieves issue details using the Linear API
2. **Mark In Progress**: Updates the issue status to "In Progress"
3. **Clone Repository**: Clones the specified GitHub repository
4. **Create Branch**: Creates a feature branch using Linear's suggested branch name
5. **Run Claude Code**: Executes Claude Code CLI with the issue description for automated development
6. **Commit Changes**: Stages and commits all changes with a structured commit message
7. **Push Branch**: Pushes the feature branch to origin
8. **Create PR**: Opens a pull request with issue details

## Command Line Options

| Flag | Description | Required |
|------|-------------|----------|
| `--repo-url` | GitHub repository URL | ✅ |
| `--verbose`, `-v` | Enable verbose logging | ❌ |
| `--help`, `-h` | Show help message | ❌ |

## Environment Variables

| Variable | Description | Required | Used By |
|----------|-------------|----------|---------|
| `LINEAR_API_KEY` | Linear API authentication token | ✅ | CLI & Server |
| `GITHUB_TOKEN` | GitHub personal access token | ✅ | CLI & Server |
| `ANTHROPIC_API_KEY` | Anthropic API key for Claude Code | ✅ | CLI & Server |
| `SERVER_API_KEY` | API key for HTTP server authentication | ✅ (Server only) | Server |
| `PORT` | HTTP server port (fallback if --port not specified) | ❌ | Server |

## Error Handling

Monday provides comprehensive error handling and logging:

- **Missing Environment Variables**: Clear error messages for missing API keys
- **API Failures**: Detailed error information for Linear/GitHub API issues
- **Git Operations**: Informative messages for repository and branch operations
- **Claude Code Integration**: Error handling for AI-powered development steps

## Troubleshooting

### Common Issues

1. **"LINEAR_API_KEY environment variable is required"**
   - Ensure you've set the Linear API key in your environment

2. **"failed to clone repository"**
   - Verify the repository URL is correct and accessible
   - Check your GitHub token has appropriate permissions

3. **"failed to run Claude Code"**
   - Ensure Claude Code CLI is installed and in your PATH
   - Verify your Anthropic API key is valid

### Debug Mode

Use the `--verbose` flag to enable detailed logging:

```bash
monday DEL-163 --repo-url https://github.com/username/repo --verbose
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

[Add your license information here]
