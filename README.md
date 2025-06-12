**Product Development Document: "monday" CLI Tool (TDD Edition)**

**Version:** 1.0

**1. Overview**

The "monday" CLI tool automates the setup of development environments for one or more Linear issues. It fetches issue details, prepares a shared Git repository, creates isolated Git worktrees for each issue, and launches the "friday" CLI in new macOS Terminal tabs/windows, pre-configured for each respective worktree. The tool is designed to run on macOS and leverage Zsh/Terminal.app via `osascript`. **This project will strictly follow a Test-Driven Development (TDD) methodology.**

**NEW: Codex Integration** - The tool now supports automated code generation using OpenAI Codex CLI in Docker containers. This enables headless, fully automated development workflows from Linear issues to implemented code.

**2. Goals**

*   Automate the creation of Git worktrees for Linear issues.
*   Enable concurrent setup for multiple Linear issues.
*   Seamlessly hand off to the "friday" CLI tool within the correct context for each issue.
*   Provide clear logging and error reporting.
*   **Achieve high code quality and reliability through comprehensive testing.**

**3. Target Platform**

*   **Operating System:** macOS
*   **Shell:** Zsh
*   **Primary Terminal Application:** `Terminal.app`

**4. Core Functionality & Workflow (Same as before, reiterated for context)**

1.  **CLI Invocation:**
    `monday [FLAGS] <ISSUE_ID_1> [ISSUE_ID_2 ...] <GIT_REPO_PATH>`
2.  **Initialization & Global Git Prep**
3.  **Concurrent Issue Processing (Fetch Linear, Create Worktree, Launch "friday")**
4.  **Completion & Reporting**

**4.1. Codex Automation Mode**

For automated code generation using OpenAI Codex:

```bash
monday codex <ISSUE_ID> [FLAGS]
```

**Required Environment Variables:**
- `OPENAI_API_KEY` - OpenAI API key for Codex access
- `LINEAR_API_KEY` - Linear API key for issue access
- `GITHUB_TOKEN` - GitHub token for git operations (optional)

**Key Flags:**
- `--codex-docker-image` - Docker image for Codex CLI (default: openai/codex-cli:latest)
- `--automated` - Enable full automation mode (auto-commit and push)
- `--worktree-root` - Root directory for git worktrees (required)

**Example:**
```bash
monday codex DEL-163 \
  --worktree-root /tmp/worktrees \
  --automated \
  /path/to/repo
```

**5. Detailed Technical Specifications & TDD Approach**

*For each component/function listed below, tests (unit tests, and integration tests where appropriate) MUST be written before or in parallel with the implementation code. The development cycle will be: Write Test -> See Test Fail -> Write Code -> See Test Pass -> Refactor.*

**5.1. Project Structure (Go)**

```
monday/
├── main.go                     // Orchestration, concurrency
├── main_test.go                // Integration tests for overall CLI flow
├── go.mod
├── go.sum
├── config/
│   ├── config.go
│   └── config_test.go          // Unit tests for CLI parsing, env vars, defaults
├── linear/
│   ├── client.go
│   └── client_test.go          // Unit tests using mock HTTP server for Linear API
├── gitops/
│   ├── manager.go
│   └── manager_test.go         // Unit/Integration tests using temporary Git repos
└── runner/
    ├── macos_runner.go
    └── macos_runner_test.go    // Unit tests for osascript command generation
└── utils/                      // (Optional)
    ├── exec.go
    └── exec_test.go
```

**5.2. Configuration (`config/config.go`)**

*   **`AppConfig` Struct:** (As defined previously)
*   **TDD Approach:**
    *   **Test Cases for `Load()`:**
        *   Valid input: single issue ID, multiple issue IDs, all flags set.
        *   Default values: when optional flags are omitted.
        *   Environment variable for API key: test it's read correctly.
        *   CLI flag for API key: test it overrides environment variable.
        *   Missing required arguments (IssueIDs, GitRepoPath): test for error.
        *   Invalid flag values (if applicable).
    *   **Implementation:** Write `Load()` function to pass these tests.

**5.3. Linear API Client (`linear/client.go`)**

*   **`IssueDetails` Struct:** (As defined previously)
*   **`FetchIssueDetails(apiKey, issueID string) (*IssueDetails, error)` Function:**
*   **TDD Approach:**
    *   **Setup:** Use `net/http/httptest` to create a mock HTTP server.
    *   **Test Cases:**
        *   Successful API call: mock server returns valid JSON, verify `IssueDetails` struct is populated correctly.
        *   Correct GraphQL query and variables sent.
        *   Correct `Authorization` header sent.
        *   API error response (e.g., 401, 404, 500 from mock server): verify error is handled and returned.
        *   Malformed JSON response from mock server: verify error handling.
        *   Network error (e.g., mock server unavailable): verify error handling (harder to simulate directly, focus on server responses).
    *   **Implementation:** Write `FetchIssueDetails()` to pass these tests.

**5.4. Git Operations (`gitops/manager.go`)**

*   **`PrepareRepository(repoPath, developBranch string) error` Function:**
*   **`CreateWorktreeForIssue(repoPath, issueID, baseBranch, worktreeSubDirName string) (worktreePath string, err error)` Function:**
*   **TDD Approach:**
    *   **Setup:**
        *   Create helper functions to set up temporary Git repositories in a test directory (e.g., using `os.MkdirTemp` and `git init`, `git commit`).
        *   These helper functions should also clean up the temporary repositories after tests.
    *   **Test Cases for `PrepareRepository()`:**
        *   Valid repo: checkout specific branch, run `git pull` (mock `git pull` success, or ensure no error if repo is clean).
        *   Invalid repo path: test for error.
        *   Branch doesn't exist: test for error.
        *   `git pull` failure (e.g., if underlying command mock returns error).
    *   **Test Cases for `CreateWorktreeForIssue()`:**
        *   Successful worktree creation: verify worktree directory and branch exist.
        *   Worktree path already exists: test for error.
        *   Base branch doesn't exist: test for error.
        *   Invalid repo path for worktree creation.
    *   **Implementation:** Write `PrepareRepository()` and `CreateWorktreeForIssue()` to pass these tests. **Mock external `git` commands for unit tests if full isolation is needed, or use real `git` commands against temporary repos for more integrated tests.** For TDD, often starting with integrated tests for `gitops` is practical, then refactoring to use mocks if tests become too slow or complex.

**5.5. macOS Runner (`runner/macos_runner.go`)**

*   **`LaunchFridayInMacOSTerminal(cfg *config.AppConfig, worktreePath string, issueTitle string) error` Function:**
*   **TDD Approach:**
    *   **Focus:** Since directly testing the visual outcome of `osascript` is hard in automated tests, focus on testing the *correct generation* of the `osascript` command and the `os/exec.Cmd` struct.
    *   **Test Cases:**
        *   Correct `osascript` command string generated for various inputs (paths with/without spaces, different issue titles, different `fridayCmd`).
        *   Verify `exec.LookPath` is called if `FridayCmd` is not absolute.
        *   Verify the `Cmd.Args` are correctly set for `osascript`.
        *   Verify `Cmd.Dir` is not set (as `osascript` handles `cd` internally).
    *   **Implementation:** Write `LaunchFridayInMacOSTerminal()` to pass these tests.
    *   **Manual/Integration Test:** A separate, manually triggered test script or step will be needed to visually confirm `osascript` works as expected on a macOS machine. This falls outside pure TDD unit tests but is crucial for this feature.

**5.6. Main Application Logic (`main.go`)**

*   **TDD Approach:**
    *   **Focus:** Test the orchestration logic, concurrency, and overall flow. This will likely be more of an **integration test style within `main_test.go`**.
    *   **Setup:**
        *   Mock the `linear`, `gitops`, and `runner` package functions. This can be done using interfaces and dependency injection, or by temporarily replacing function implementations for testing (less ideal but sometimes pragmatic for `main`).
        *   For example, `linear.FetchIssueDetails` could be a variable `var LinearFetchFunc = linear.FetchIssueDetails`, which can be swapped in tests.
    *   **Test Cases:**
        *   Single issue, successful flow: verify all mocked components are called with correct arguments.
        *   Multiple issues, all successful: verify concurrent execution (e.g., using channels to track goroutine start/end times) and all mocked components called for each.
        *   One issue fails at Linear fetch: verify other issues proceed, correct error reported for the failed one.
        *   One issue fails at Git worktree creation: verify other issues proceed, correct error reported.
        *   One issue fails at `friday` launch: verify other issues proceed, correct error reported.
        *   Correct overall exit code based on success/failure of individual tasks.
        *   Correct summary report printed (capture stdout).
        *   `PrepareRepository` called once.
    *   **Implementation:** Write the `main` function and supporting concurrency logic to pass these tests.

**5.7. Error Handling and Logging**

*   **TDD Approach:**
    *   For each function, write test cases that specifically trigger error conditions and verify:
        *   The correct error type/message is returned.
        *   Errors are wrapped with context using `fmt.Errorf("...%w", err)`.
    *   For logging, capture stdout/stderr in tests (e.g., by redirecting `log.SetOutput`) and verify expected log messages (especially those prefixed with `[<issueID>]`) are present.

**6. Dependencies (Go Modules)** (Same as before)

*   Standard library.
*   Consider `stretchr/testify/assert` and `stretchr/testify/mock` for easier assertions and mocking in tests.
*   `net/http/httptest` for Linear API mocking.

**7. Build and Deployment** (Same as before)

**8. Testing Strategy Summary**

*   **Unit Tests:** For `config`, `linear` (with mock HTTP), `gitops` (potentially with mock `git` or against temp repos), `runner` (testing command generation).
*   **Integration Tests (`main_test.go`):** Test the `main` function's orchestration by mocking out the package-level functions or using interfaces.
*   **Manual E2E Tests:** Manually run the compiled `monday` binary on a macOS machine with a real (or test) Linear account and Git repository to verify the end-to-end flow, especially the `osascript` interaction with `Terminal.app`.

**9. Potential Issues & Edge Cases to Consider (Same as before)**

*   Focus on writing tests that cover these edge cases where possible (e.g., testing `gitops` with a pre-existing worktree path).

**10. Future Enhancements (Out of Scope for V1)** (Same as before)

---

This TDD-focused document guides the developer to build the application test by test, ensuring each piece of functionality is verified before moving on. This approach generally leads to more robust, maintainable, and well-understood code.


## Option 1: Direct Copy to PATH (Simplest)
```bash
# Copy the binary to a directory in your PATH
sudo cp /Users/max/code/monday/bin/monday /usr/local/bin/monday

# Make it executable (if needed)
sudo chmod +x /usr/local/bin/monday

# Verify installation
monday --help
```
