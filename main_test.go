package main

import (
        "encoding/json"
        "fmt"
        "net/http"
        "net/http/httptest"
        "os"
        "os/exec"
        "path/filepath"
        "testing"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"

        "monday/linear"
)

func setupTestRepoForMain(t *testing.T) string {
        tempDir, err := os.MkdirTemp("", "monday-main-test-*")
        require.NoError(t, err)

        t.Cleanup(func() {
                os.RemoveAll(tempDir)
        })

        cmd := exec.Command("git", "init")
        cmd.Dir = tempDir
        require.NoError(t, cmd.Run())

        cmd = exec.Command("git", "config", "user.email", "test@example.com")
        cmd.Dir = tempDir
        require.NoError(t, cmd.Run())

        cmd = exec.Command("git", "config", "user.name", "Test User")
        cmd.Dir = tempDir
        require.NoError(t, cmd.Run())

        readmeFile := filepath.Join(tempDir, "README.md")
        err = os.WriteFile(readmeFile, []byte("# Test Repo"), 0644)
        require.NoError(t, err)

        cmd = exec.Command("git", "add", "README.md")
        cmd.Dir = tempDir
        require.NoError(t, cmd.Run())

        cmd = exec.Command("git", "commit", "-m", "Initial commit")
        cmd.Dir = tempDir
        require.NoError(t, cmd.Run())

        cmd = exec.Command("git", "branch", "-M", "main")
        cmd.Dir = tempDir
        require.NoError(t, cmd.Run())

        return tempDir
}

func setupMockLinearServer(t *testing.T) *httptest.Server {
        return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                var req linear.GraphQLRequest
                json.NewDecoder(r.Body).Decode(&req)

                if req.Variables["teamKey"] != nil && req.Variables["number"] != nil {
                        teamKey := req.Variables["teamKey"].(string)
                        number := req.Variables["number"].(float64)
                        issueID := teamKey + "-" + fmt.Sprintf("%.0f", number)
                        
                        // Handle issue details query
                        response := linear.GraphQLResponse{
                                Data: linear.GraphQLData{
                                        Issues: linear.IssuesConnection{
                                                Nodes: []linear.IssueDetails{{
                                                        ID:          issueID,
                                                        Title:       "Fix authentication bug",
                                                        Description: "This is a detailed description of the authentication bug that needs to be fixed.",
                                                        BranchName:  issueID + "-test-issue",
                                                        URL:         "https://linear.app/team/issue/" + issueID,
                                                }},
                                        },
                                },
                        }
                        json.NewEncoder(w).Encode(response)
                } else if req.Variables["id"] != nil {
                        // Handle issue update mutation
                        response := map[string]interface{}{
                                "data": map[string]interface{}{
                                        "issueUpdate": map[string]interface{}{
                                                "success": true,
                                        },
                                },
                        }
                        json.NewEncoder(w).Encode(response)
                } else {
                        // Handle workflow states query
                        response := map[string]interface{}{
                                "data": map[string]interface{}{
                                        "workflowStates": map[string]interface{}{
                                                "nodes": []map[string]interface{}{
                                                        {
                                                                "id":   "state-in-progress",
                                                                "name": "In Progress",
                                                                "type": "started",
                                                        },
                                                },
                                        },
                                },
                        }
                        json.NewEncoder(w).Encode(response)
                }
        }))
}

func TestRunApplication_SingleIssue_Success(t *testing.T) {
        repoPath := setupTestRepoForMain(t)
        server := setupMockLinearServer(t)
        defer server.Close()

        app := &Application{
                DryRun: true, // Don't actually launch Terminal
        }

        err := app.Run([]string{
                "-api-key", "test-key",
                "-linear-endpoint", server.URL,
                "DEL-123",
                repoPath,
        })

        assert.NoError(t, err)

        // Verify worktree was created
        worktreePath := filepath.Join(repoPath, "worktrees", "DEL-123")
        info, err := os.Stat(worktreePath)
        require.NoError(t, err)
        assert.True(t, info.IsDir())

        // Verify _feature.md file was created with issue description
        featureFilePath := filepath.Join(worktreePath, "_feature.md")
        assert.FileExists(t, featureFilePath)

        content, err := os.ReadFile(featureFilePath)
        require.NoError(t, err)

        expectedContent := "# Fix authentication bug\n\nThis is a detailed description of the authentication bug that needs to be fixed.\n"
        assert.Equal(t, expectedContent, string(content))
}

func TestRunApplication_MultipleIssues_Success(t *testing.T) {
        repoPath := setupTestRepoForMain(t)
        server := setupMockLinearServer(t)
        defer server.Close()

        app := &Application{
                DryRun: true,
        }

        err := app.Run([]string{
                "-api-key", "test-key",
                "-linear-endpoint", server.URL,
                "DEL-123", "DEL-456",
                repoPath,
        })

        assert.NoError(t, err)

        // Verify both worktrees were created
        for _, issueID := range []string{"DEL-123", "DEL-456"} {
                worktreePath := filepath.Join(repoPath, "worktrees", issueID)
                info, err := os.Stat(worktreePath)
                require.NoError(t, err)
                assert.True(t, info.IsDir())
        }
}

func TestRunApplication_LinearAPIError(t *testing.T) {
        repoPath := setupTestRepoForMain(t)
        
        // Server that returns errors
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusNotFound)
                w.Write([]byte(`{"error": "Issue not found"}`))
        }))
        defer server.Close()

        app := &Application{
                DryRun: true,
        }

        err := app.Run([]string{
                "-api-key", "test-key",
                "-linear-endpoint", server.URL,
                "NONEXISTENT",
                repoPath,
        })

        assert.Error(t, err)
        assert.Contains(t, err.Error(), "failed to process")
}

func TestRunApplication_InvalidRepo(t *testing.T) {
        server := setupMockLinearServer(t)
        defer server.Close()

        app := &Application{
                DryRun: true,
        }

        err := app.Run([]string{
                "-api-key", "test-key",
                "-linear-endpoint", server.URL,
                "DEL-123",
                "/nonexistent/repo",
        })

        assert.Error(t, err)
        assert.Contains(t, err.Error(), "not a git repository")
}

func TestRunApplication_MissingAPIKey(t *testing.T) {
        repoPath := setupTestRepoForMain(t)

        app := &Application{
                DryRun: true,
        }

        err := app.Run([]string{
                "DEL-123",
                repoPath,
        })

        assert.Error(t, err)
        assert.Contains(t, err.Error(), "API key")
}

func TestRunApplication_ConcurrentProcessing(t *testing.T) {
        repoPath := setupTestRepoForMain(t)
        server := setupMockLinearServer(t)
        defer server.Close()

        app := &Application{
                DryRun: true,
        }

        // Test with multiple issues to verify concurrent processing
        issues := []string{"DEL-1", "DEL-2", "DEL-3", "DEL-4", "DEL-5"}
        args := []string{
                "-api-key", "test-key",
                "-linear-endpoint", server.URL,
                "-concurrency", "2",
        }
        args = append(args, issues...)
        args = append(args, repoPath)

        err := app.Run(args)
        assert.NoError(t, err)

        // Verify all worktrees were created
        for _, issueID := range issues {
                worktreePath := filepath.Join(repoPath, "worktrees", issueID)
                info, err := os.Stat(worktreePath)
                require.NoError(t, err)
                assert.True(t, info.IsDir())
        }
}

func TestRunApplication_MarkIssueInProgress(t *testing.T) {
        repoPath := setupTestRepoForMain(t)
        
        var receivedQueries []linear.GraphQLRequest
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                var req linear.GraphQLRequest
                json.NewDecoder(r.Body).Decode(&req)
                receivedQueries = append(receivedQueries, req)

                if req.Variables["teamKey"] != nil && req.Variables["number"] != nil {
                        teamKey := req.Variables["teamKey"].(string)
                        number := req.Variables["number"].(float64)
                        issueID := teamKey + "-" + fmt.Sprintf("%.0f", number)
                        
                        // Handle issue details query
                        response := linear.GraphQLResponse{
                                Data: linear.GraphQLData{
                                        Issues: linear.IssuesConnection{
                                                Nodes: []linear.IssueDetails{{
                                                        ID:         "uuid-" + issueID,
                                                        Title:      "Test Issue " + issueID,
                                                        BranchName: issueID + "-test-issue",
                                                        URL:        "https://linear.app/team/issue/" + issueID,
                                                }},
                                        },
                                },
                        }
                        json.NewEncoder(w).Encode(response)
                } else if req.Variables["id"] != nil {
                        // Handle issue update mutation
                        response := map[string]interface{}{
                                "data": map[string]interface{}{
                                        "issueUpdate": map[string]interface{}{
                                                "success": true,
                                        },
                                },
                        }
                        json.NewEncoder(w).Encode(response)
                } else {
                        // Handle workflow states query
                        response := map[string]interface{}{
                                "data": map[string]interface{}{
                                        "workflowStates": map[string]interface{}{
                                                "nodes": []map[string]interface{}{
                                                        {
                                                                "id":   "state-in-progress",
                                                                "name": "In Progress",
                                                                "type": "started",
                                                        },
                                                },
                                        },
                                },
                        }
                        json.NewEncoder(w).Encode(response)
                }
        }))
        defer server.Close()

        app := &Application{
                DryRun: true,
        }

        err := app.Run([]string{
                "-api-key", "test-key",
                "-linear-endpoint", server.URL,
                "DEL-123",
                repoPath,
        })

        assert.NoError(t, err)
        
        // Verify that we made the expected API calls:
        // 1. Get issue details
        // 2. Get workflow states 
        // 3. Update issue status
        require.Len(t, receivedQueries, 3)
        
        // First call should be issue details
        assert.Contains(t, receivedQueries[0].Query, "issues")
        assert.Equal(t, "DEL", receivedQueries[0].Variables["teamKey"])
        assert.Equal(t, float64(123), receivedQueries[0].Variables["number"])
        
        // Second call should be workflow states
        assert.Contains(t, receivedQueries[1].Query, "workflowStates")
        
        // Third call should be issue update
        assert.Contains(t, receivedQueries[2].Query, "issueUpdate")
        assert.Equal(t, "uuid-DEL-123", receivedQueries[2].Variables["id"])
        assert.Equal(t, "state-in-progress", receivedQueries[2].Variables["stateId"])
}
