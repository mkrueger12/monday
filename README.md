# Monday CLI

**Automated Development Tool for Linear Issues**

Monday CLI is a fully containerized automation tool that takes Linear issues and automatically implements them using OpenAI Codex. It provides end-to-end automation from Linear issue selection to GitHub pull request creation.

## 🚀 Features

- **Fully Containerized**: All operations run in isolated Alpine containers
- **Smart Issue Selection**: Filter by team, project, and tags or specify exact issue IDs
- **Multi-Language Support**: Auto-detects and installs dependencies for Node.js, Python, Ruby, Go, Rust
- **End-to-End Automation**: From Linear issue → Code generation → GitHub PR
- **Zero Local State**: No worktrees, no persistent files - everything is ephemeral

## 📋 How It Works

1. **Issue Selection**: Either specify an issue ID or use filters to auto-select
2. **Containerized Workflow**: Spins up Alpine container with all development tools
3. **Repository Setup**: Clones repository and creates feature branch
4. **Dependency Installation**: Auto-detects and installs project dependencies
5. **Code Generation**: Runs OpenAI Codex with issue title and description
6. **Git Operations**: Commits changes and pushes to GitHub
7. **PR Creation**: Automatically creates pull request with issue details
8. **Linear Update**: Marks issue as "In Progress"

## 🔧 Installation

### Build from Source

```bash
git clone <repository-url>
cd monday
go build -o monday .
```

### Install to PATH

```bash
# Copy to system PATH
sudo cp monday /usr/local/bin/monday
sudo chmod +x /usr/local/bin/monday

# Verify installation
monday --help
```

## ⚙️ Configuration

### Required Environment Variables

```bash
export LINEAR_API_KEY="your-linear-api-key"
export OPENAI_API_KEY="your-openai-api-key"  
export GITHUB_TOKEN="your-github-token"
```

### Docker Setup

The tool requires Docker to be installed and running, as all operations happen in containers.

```bash
# Build the Codex container (first time only)
docker build -f Dockerfile.codex -t monday/codex:latest .
```

## 🎯 Usage

### Specific Issue Implementation

```bash
monday DEL-163 --repo-url https://github.com/myorg/backend.git
```

### Auto-Selection with Filters

```bash
# Select first issue matching ALL criteria
monday --linear-team ENG --linear-project backend --linear-tag bug --repo-url https://github.com/myorg/backend.git
```

### CLI Flags

| Flag | Description | Required |
|------|-------------|----------|
| `--repo-url` | GitHub repository URL | ✅ |
| `--linear-team` | Filter by Linear team | ❌ |
| `--linear-project` | Filter by Linear project | ❌ |
| `--linear-tag` | Filter by Linear tag | ❌ |
| `--api-key` | Linear API key (or use env var) | ❌ |
| `--openai-api-key` | OpenAI API key (or use env var) | ❌ |
| `--github-token` | GitHub token (or use env var) | ❌ |

## 🔍 Filtering Logic

When using filters, the tool applies **strict AND conditions**:
- Issue must be on the specified team
- Issue must be in the specified project  
- Issue must have the specified tag

All specified filters must match for an issue to be selected.

## 🐳 Container Details

The tool uses an Alpine-based container (`monday/codex:latest`) that includes:
- Git and GitHub CLI
- Node.js, Python, Ruby, Go, Rust toolchains
- OpenAI Codex CLI
- Dependency auto-detection script

## 📝 Example Workflows

### Bug Fix Automation
```bash
monday --linear-team ENG --linear-project frontend --linear-tag bug --repo-url https://github.com/myorg/frontend.git
```

### Feature Implementation
```bash
monday DEL-456 --repo-url https://github.com/myorg/api.git
```

### Backend Task
```bash
monday --linear-team BACKEND --linear-project core --repo-url https://github.com/myorg/core.git
```

## 🔒 Security

- All operations run in network-disabled containers (except for API calls)
- No persistent local state or files
- API keys are passed securely as environment variables
- Git operations are scoped to the specific repository

## 🛠️ Development

### Project Structure

```
monday/
├── main.go                    # CLI entry point
├── config/                    # Configuration management
├── linear/                    # Linear API client
├── runner/                    # Containerized workflow execution
├── scripts/                   # Container scripts
├── Dockerfile.codex          # Container definition
└── README.md
```

### Building

```bash
go mod tidy
go build -o monday .
```

### Testing

```bash
go test ./...
```

## 📄 License

This project is licensed under the MIT License.
