# monday
Okay, here's a Product Development Document (PDD) or Technical Specification that you can hand to a developer. It focuses on the technical details and implementation guidelines for the "monday" CLI tool.

---

**Product Development Document: "monday" CLI Tool**

**Version:** 1.0

**1. Overview**

The "monday" CLI tool automates the setup of development environments for one or more Linear issues. It fetches issue details, prepares a shared Git repository, creates isolated Git worktrees for each issue, and launches the "friday" CLI in new macOS Terminal tabs/windows, pre-configured for each respective worktree. The tool is designed to run on macOS and leverage Zsh/Terminal.app via `osascript`.

**2. Goals**

*   Automate the creation of Git worktrees for Linear issues.
*   Enable concurrent setup for multiple Linear issues.
*   Seamlessly hand off to the "friday" CLI tool within the correct context for each issue.
*   Provide clear logging and error reporting.

**3. Target Platform**

*   **Operating System:** macOS
*   **Shell:** Zsh (as `osascript` will interact with `Terminal.app` which typically uses the user's default shell)
*   **Primary Terminal Application:** `Terminal.app`

**4. Core Functionality & Workflow**

1.  **CLI Invocation:**
    `monday [FLAGS] <ISSUE_ID_1> [ISSUE_ID_2 ...] <GIT_REPO_PATH>`
2.  **Initialization:**
    *   Parse all CLI arguments and flags.
    *   Load Linear API Key (CLI flag overrides environment variable `LINEAR_API_KEY`).
3.  **Global Git Preparation (Once):**
    *   Change directory to `<GIT_REPO_PATH>`.
    *   Validate it's a Git repository.
    *   Checkout the `develop-branch` (default "develop").
    *   Execute `git pull` on this branch.
4.  **Concurrent Issue Processing (For each provided Issue ID):**
    *   A Go routine will be spawned for each Issue ID.
    *   **Inside each Go routine:**
        a.  **Fetch Linear Data:**
            *   API Endpoint: `https://api.linear.app/graphql`
            *   Method: POST
            *   Authentication: `Authorization: Bearer <API_KEY>`
            *   GraphQL Query: Fetch issue `id`, `title`, and `description` for the given Issue ID.
        b.  **Create Git Worktree:**
            *   Worktree Path: `<GIT_REPO_PATH>/<WORKTREE_DIR_NAME>/<LinearIssue.ID>`
            *   Worktree Branch Name: `feature/<LinearIssue.ID>` (e.g., `feature/TEAM-123`)
            *   Command: `git worktree add -b feature/<LinearIssue.ID> <WorktreePath> <develop-branch>`
        c.  **Launch "friday" in New Terminal Tab:**
            *   Use `osascript` to interact with `Terminal.app`.
            *   Command to execute via `osascript`:
                1.  Tell Terminal.app to activate.
                2.  Tell System Events to send `Cmd+T` (new tab in active Terminal window).
                3.  Tell the new tab (or frontmost tab of the frontmost window) to execute: `cd "<WorktreePath>/<FridayWorkDir>" && <FridayCmd>`
                4.  Attempt to set the title of the new tab to the Linear Issue Title.
5.  **Completion & Reporting:**
    *   The main `monday` process waits for all Go routines to complete.
    *   A summary report is printed to the console:
        *   Status (SUCCESS/FAILED) for each Issue ID.
        *   Any error messages encountered.
    *   Exit code 0 if all successful, 1 if any failures.

**5. Detailed Technical Specifications**

**5.1. Project Structure (Go)**

```
monday/
├── main.go                 // Main application logic, concurrency management
├── go.mod
├── go.sum
├── config/
│   └── config.go           // CLI flag/env var parsing, AppConfig struct
├── linear/
│   └── client.go           // Linear API interaction (GraphQL client)
├── gitops/
│   └── manager.go          // Git command execution (prepare repo, create worktree)
└── runner/
    └── macos_runner.go     // Launching "friday" via osascript for macOS Terminal
└── utils/                  // (Optional)
    └── exec.go             // Helper for os/exec.Command execution
```

**5.2. Configuration (`config/config.go`)**

*   **`AppConfig` Struct:**
    ```go
    type AppConfig struct {
        LinearAPIKey    string
        IssueIDs        []string // From positional args
        GitRepoPath     string   // Last positional arg
        DevelopBranch   string   // Flag, default "develop"
        WorktreeDirName string   // Flag, default "worktrees"
        FridayCmd       string   // Flag, default "friday"
        FridayWorkDir   string   // Flag, default "" (root of worktree)
    }
    ```
*   **Argument Parsing:**
    *   Use Go's `flag` package.
    *   Positional arguments: `flag.Args()`. Last arg is `GitRepoPath`, preceding ones are `IssueIDs`.
    *   Validation: At least one `IssueID` and `GitRepoPath` must be present.
    *   Linear API Key: Flag `--linear-api-key` overrides `os.Getenv("LINEAR_API_KEY")`.

**5.3. Linear API Client (`linear/client.go`)**

*   **`IssueDetails` Struct:** `ID string`, `Title string`, `Description string`.
*   **`FetchIssueDetails(apiKey, issueID string) (*IssueDetails, error)` Function:**
    *   Construct GraphQL query string:
        ```graphql
        query GetIssue($issueId: String!) {
          issue(id: $issueId) {
            id
            title
            description
          }
        }
        ```
    *   Use `net/http` for POST request.
    *   Set `Content-Type: application/json` and `Authorization: Bearer ...` headers.
    *   JSON request body: `{"query": "...", "variables": {"issueId": "..."}}`.
    *   Parse JSON response.
    *   Error handling for HTTP status codes and API errors.

**5.4. Git Operations (`gitops/manager.go`)**

*   **`PrepareRepository(repoPath, developBranch string) error` Function:**
    *   `os.Chdir(repoPath)`
    *   Run `git checkout <developBranch>` using `os/exec`.
    *   Run `git pull origin <developBranch>` using `os/exec`.
    *   Check command exit codes and capture stderr for errors.
*   **`CreateWorktreeForIssue(repoPath, issueID, baseBranch, worktreeSubDirName string) (worktreePath string, err error)` Function:**
    *   `worktreePath = filepath.Join(repoPath, worktreeSubDirName, issueID)`
    *   `branchName = fmt.Sprintf("feature/%s", issueID)`
    *   Check if `worktreePath` exists. If so, return an error.
    *   Run `git worktree add -b <branchName> <worktreePath> <baseBranch>` using `os/exec`.
    *   Return absolute `worktreePath`.
    *   Check command exit codes and capture stderr for errors.

**5.5. macOS Runner (`runner/macos_runner.go`)**

*   **`LaunchFridayInMacOSTerminal(cfg *config.AppConfig, worktreePath string, issueTitle string) error` Function:**
    *   Resolve `cfg.FridayCmd` to absolute path using `exec.LookPath` if not already absolute.
    *   `actualExecDir = filepath.Join(worktreePath, cfg.FridayWorkDir)` (if `FridayWorkDir` is set).
    *   Construct the shell command to be run in the new tab:
        `shellCmd := fmt.Sprintf("cd %s && %s", quotePathForShell(actualExecDir), quotePathForShell(cfg.FridayCmd))`
    *   Construct AppleScript:
        ```go
        // (Simplified, ensure proper string escaping for AppleScript)
        appleScript := fmt.Sprintf(`
        tell application "Terminal"
            activate
            tell application "System Events" to keystroke "t" using command down
            delay 0.2 -- Brief pause for tab to open
            -- The new tab should be the frontmost one of the active window
            do script "%s" in selected tab of front window
            set custom title of selected tab of front window to "%s"
        end tell`, quoteForAppleScript(shellCmd), quoteForAppleScript(issueTitle))
        ```
        *   `quotePathForShell` should enclose paths with spaces in double quotes for the shell part.
        *   `quoteForAppleScript` should escape `\` and `"` for the AppleScript string itself.
    *   Execute using `os/exec.Command("osascript", "-e", appleScript).Start()`.
    *   `Start()` is used so "monday" doesn't block on this.

**5.6. Main Application Logic (`main.go`)**

*   Initialize logging (e.g., `log.SetFlags(log.LstdFlags | log.Lshortfile)`).
*   Call `config.Load()`.
*   Call `gitops.PrepareRepository()` once.
*   Initialize `sync.WaitGroup` and a channel for results (e.g., `resultsChan := make(chan issueResult)`).
*   Loop through `cfg.IssueIDs`:
    *   `wg.Add(1)`
    *   Launch goroutine:
        *   `defer wg.Done()`
        *   Prefix log messages with `[<issueID>]`.
        *   Call `linear.FetchIssueDetails()`. If error, send to `resultsChan` and return.
        *   Call `gitops.CreateWorktreeForIssue()`. If error, send to `resultsChan` and return.
        *   Call `runner.LaunchFridayInMacOSTerminal()`. If error, send to `resultsChan` and return.
        *   Send success to `resultsChan`.
*   Start a separate goroutine to `wg.Wait()` and then `close(resultsChan)`.
*   Read from `resultsChan` in the main goroutine until it's closed. Print summary.
*   `os.Exit(0)` or `os.Exit(1)` based on overall success.

**5.7. Error Handling and Logging**

*   Each function should return an error if something goes wrong.
*   Use `fmt.Errorf("context: %w", err)` for wrapping errors to provide context.
*   Log informative messages at each step, prefixed with the relevant Issue ID when in a goroutine.
*   Main process should clearly report which issues succeeded and which failed, along with error messages.

**6. Dependencies (Go Modules)**

*   Standard library ( `flag`, `os`, `os/exec`, `fmt`, `log`, `net/http`, `encoding/json`, `sync`, `path/filepath`, `strings`).
*   No external third-party libraries are strictly required for the core functionality outlined, but consider well-vetted libraries for:
    *   GraphQL client (e.g., `machinebox/graphql` or `shurcooL/graphql`) for more robust query building and response parsing, though `net/http` is sufficient.

**7. Build and Deployment**

*   Standard Go build: `go build -o monday .`
*   The resulting `monday` binary can be placed in a directory in the user's `$PATH` (e.g., `/usr/local/bin` or `~/bin`).

**8. Testing Considerations**

*   **Unit Tests:**
    *   Test `config.Load()` with various valid and invalid inputs.
    *   Test `linear.FetchIssueDetails()` with a mock HTTP server.
    *   Test `gitops` functions by interacting with a temporary test Git repository.
    *   Test `runner.LaunchFridayInMacOSTerminal()` by checking the `osascript` command generated (actual execution is harder to unit test, might require integration tests).
*   **Integration Tests:**
    *   A script that runs `monday` against a test Linear instance (if possible, or with mocked Linear data) and a local test Git repository.
    *   Manually verify that new Terminal tabs open and `friday` launches correctly.

**9. Potential Issues & Edge Cases to Consider**

*   **`osascript` / Terminal.app Interaction:**
    *   What if Terminal.app is not running? `activate` should handle this.
    *   Reliability of `Cmd+T` and `delay`.
    *   What if the user's default shell isn't Zsh/Bash compatible enough for `cd ... && ...`? (Less likely on macOS with default Terminal settings).
    *   Are there different versions of Terminal.app that behave differently?
*   **Git:**
    *   Worktree path already exists: Currently an error for that issue.
    *   `git pull` results in merge conflicts: `PrepareRepository` will fail. `monday` should abort.
    *   Invalid Git repository path.
*   **Linear API:**
    *   Rate limiting.
    *   Network errors.
    *   Invalid issue IDs.
*   **"friday" CLI:**
    *   `friday` command not found in PATH.
    *   `friday` itself errors on launch.
*   **Concurrency:**
    *   Ensure no race conditions if any shared resources are accessed (though the design minimizes this by having each goroutine work on its own issue and worktree).
    *   Resource limits (e.g., opening too many terminal tabs simultaneously if hundreds of issues are processed – though this is unlikely for typical use).

**10. Future Enhancements (Out of Scope for V1)**

*   `--force` flag to remove and recreate existing worktrees.
*   Support for other terminal emulators on macOS (e.g., iTerm2).
*   Support for other OS (Linux, Windows) - would require different `runner` implementations.
*   Configuration file for persistent settings.
*   Option to clean up worktrees.

---

This document should provide a developer with a clear understanding of what needs to be built and how the different components should interact. Encourage the developer to ask clarifying questions.
