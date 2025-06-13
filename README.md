# Monday - DevFlow Orchestrator CLI

Monday is a command-line interface (CLI) tool designed to streamline the initial development setup process. It automates fetching Linear issue details, cloning a GitHub repository, creating a feature branch, and then executing Codex CLI for automated development, followed by committing changes and creating a pull request.

## Features

- 🔗 **Linear Integration**: Fetch issue details and mark issues as "In Progress"
- 🚀 **GitHub Automation**: Clone repositories, create feature branches, and open PRs
- 🤖 **AI-Powered Development**: Integrate with OpenAI Codex for automated code generation
- 📝 **Structured Logging**: Comprehensive logging with Zap for debugging and monitoring
- 🔐 **Secure Credentials**: Environment variable-based authentication

## Installation

### Prerequisites

Ensure you have the following tools installed:
- [Git](https://git-scm.com/)
- [GitHub CLI (gh)](https://cli.github.com/)
- [Codex CLI](https://github.com/your-codex-cli-repo) 

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
export OPENAI_API_KEY="your-openai-api-key"
```

### Getting API Keys

1. **Linear API Key**: Go to Linear Settings → API → Create new API key
2. **GitHub Token**: Go to GitHub Settings → Developer settings → Personal access tokens → Generate new token
   - Required scopes: `repo`, `workflow`
3. **OpenAI API Key**: Get from [OpenAI Platform](https://platform.openai.com/api-keys)

## Usage

### Basic Usage

```bash
monday <linear_issue_id> --repo-url <github_repo_url>
```

### Examples

```bash
# Using Linear issue ID
monday DEL-163 --repo-url https://github.com/username/repo

# Using full Linear issue URL
monday https://linear.app/company/issue/DEL-163 --repo-url https://github.com/username/repo

# With verbose logging
monday DEL-163 --repo-url https://github.com/username/repo --verbose
```

## Workflow

When you run Monday, it performs the following steps:

1. **Fetch Linear Issue**: Retrieves issue details using the Linear API
2. **Mark In Progress**: Updates the issue status to "In Progress"
3. **Clone Repository**: Clones the specified GitHub repository
4. **Create Branch**: Creates a feature branch using Linear's suggested branch name
5. **Run Codex**: Executes Codex CLI with the issue description for automated development
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

| Variable | Description | Required |
|----------|-------------|----------|
| `LINEAR_API_KEY` | Linear API authentication token | ✅ |
| `GITHUB_TOKEN` | GitHub personal access token | ✅ |
| `OPENAI_API_KEY` | OpenAI API key for Codex | ✅ |

## Error Handling

Monday provides comprehensive error handling and logging:

- **Missing Environment Variables**: Clear error messages for missing API keys
- **API Failures**: Detailed error information for Linear/GitHub API issues
- **Git Operations**: Informative messages for repository and branch operations
- **Codex Integration**: Error handling for AI-powered development steps

## Troubleshooting

### Common Issues

1. **"LINEAR_API_KEY environment variable is required"**
   - Ensure you've set the Linear API key in your environment

2. **"failed to clone repository"**
   - Verify the repository URL is correct and accessible
   - Check your GitHub token has appropriate permissions

3. **"failed to run Codex"**
   - Ensure Codex CLI is installed and in your PATH
   - Verify your OpenAI API key is valid

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
