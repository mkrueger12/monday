package linear

import (
        "encoding/json"
        "net/http"
        "net/http/httptest"
        "testing"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
)

func TestFetchIssueDetails_Success(t *testing.T) {
        expectedIssue := IssueDetails{
                ID:          "ISSUE-123",
                Title:       "Fix authentication bug",
                Description: "This is a detailed description of the authentication bug that needs to be fixed.",
                BranchName:  "issue-123-fix-authentication-bug",
                URL:         "https://linear.app/team/issue/ISSUE-123",
        }

        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                assert.Equal(t, "POST", r.Method)
                assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
                assert.Equal(t, "test-api-key", r.Header.Get("Authorization"))

                var req GraphQLRequest
                json.NewDecoder(r.Body).Decode(&req)
                assert.Contains(t, req.Query, "$teamKey")
                assert.Contains(t, req.Query, "$number")
                assert.Equal(t, "DEL", req.Variables["teamKey"])
                assert.Equal(t, float64(123), req.Variables["number"])

                response := GraphQLResponse{
                        Data: GraphQLData{
                                Issues: IssuesConnection{
                                        Nodes: []IssueDetails{expectedIssue},
                                },
                        },
                }
                json.NewEncoder(w).Encode(response)
        }))
        defer server.Close()

        client := NewClient("test-api-key")
        client.endpoint = server.URL

        issue, err := client.FetchIssueDetails("DEL-123")
        require.NoError(t, err)
        assert.Equal(t, expectedIssue, *issue)
}

func TestFetchIssueDetails_HTTPError(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusNotFound)
                w.Write([]byte(`{"error": "Issue not found"}`))
        }))
        defer server.Close()

        client := NewClient("test-api-key")
        client.endpoint = server.URL

        issue, err := client.FetchIssueDetails("DEL-999")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "404")
        assert.Nil(t, issue)
}

func TestFetchIssueDetails_GraphQLError(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                response := GraphQLResponse{
                        Errors: []GraphQLError{
                                {Message: "Issue not found"},
                        },
                }
                json.NewEncoder(w).Encode(response)
        }))
        defer server.Close()

        client := NewClient("test-api-key")
        client.endpoint = server.URL

        issue, err := client.FetchIssueDetails("DEL-999")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "Issue not found")
        assert.Nil(t, issue)
}

func TestFetchIssueDetails_MalformedJSON(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Write([]byte(`invalid json`))
        }))
        defer server.Close()

        client := NewClient("test-api-key")
        client.endpoint = server.URL

        issue, err := client.FetchIssueDetails("DEL-123")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "decode")
        assert.Nil(t, issue)
}

func TestFetchIssueDetails_NetworkError(t *testing.T) {
        client := NewClient("test-api-key")
        client.endpoint = "http://nonexistent-server:12345"

        issue, err := client.FetchIssueDetails("DEL-123")
        assert.Error(t, err)
        assert.Nil(t, issue)
}

func TestGraphQLQuery_Structure(t *testing.T) {
        var receivedQuery GraphQLRequest

        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                json.NewDecoder(r.Body).Decode(&receivedQuery)
                
                response := GraphQLResponse{
                        Data: GraphQLData{
                                Issues: IssuesConnection{
                                        Nodes: []IssueDetails{{
                                                ID:         "ISSUE-123",
                                                Title:      "Test Issue",
                                                BranchName: "issue-123-test-issue",
                                                URL:        "https://linear.app/team/issue/ISSUE-123",
                                        }},
                                },
                        },
                }
                json.NewEncoder(w).Encode(response)
        }))
        defer server.Close()

        client := NewClient("test-api-key")
        client.endpoint = server.URL

        _, err := client.FetchIssueDetails("DEL-123")
        require.NoError(t, err)

        assert.Contains(t, receivedQuery.Query, "query")
        assert.Contains(t, receivedQuery.Query, "issues")
        assert.Contains(t, receivedQuery.Query, "id")
        assert.Contains(t, receivedQuery.Query, "title")
        assert.Contains(t, receivedQuery.Query, "branchName")
        assert.Contains(t, receivedQuery.Query, "url")
        assert.Equal(t, "DEL", receivedQuery.Variables["teamKey"])
        assert.Equal(t, float64(123), receivedQuery.Variables["number"])
}

func TestMarkIssueInProgress_Success(t *testing.T) {
        callCount := 0
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                assert.Equal(t, "POST", r.Method)
                assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
                assert.Equal(t, "test-api-key", r.Header.Get("Authorization"))

                callCount++
                if callCount == 1 {
                        // First call: getInProgressStateID
                        response := map[string]interface{}{
                                "data": map[string]interface{}{
                                        "workflowStates": map[string]interface{}{
                                                "nodes": []map[string]interface{}{
                                                        {
                                                                "id":   "state-123",
                                                                "name": "In Progress",
                                                                "type": "started",
                                                        },
                                                },
                                        },
                                },
                        }
                        json.NewEncoder(w).Encode(response)
                } else {
                        // Second call: issueUpdate
                        response := map[string]interface{}{
                                "data": map[string]interface{}{
                                        "issueUpdate": map[string]interface{}{
                                                "success": true,
                                        },
                                },
                        }
                        json.NewEncoder(w).Encode(response)
                }
        }))
        defer server.Close()

        client := NewClient("test-api-key")
        client.endpoint = server.URL

        issue := &IssueDetails{
                ID:         "uuid-123",
                Title:      "Test Issue",
                BranchName: "test-branch",
                URL:        "https://linear.app/test",
        }
        err := client.MarkIssueInProgress(issue)
        require.NoError(t, err)
        assert.Equal(t, 2, callCount)
}

func TestMarkIssueInProgress_HTTPError(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusUnauthorized)
                w.Write([]byte(`{"error": "Unauthorized"}`))
        }))
        defer server.Close()

        client := NewClient("test-api-key")
        client.endpoint = server.URL

        issue := &IssueDetails{ID: "uuid-123"}
        err := client.MarkIssueInProgress(issue)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "401")
}

func TestMarkIssueInProgress_GraphQLError(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                response := map[string]interface{}{
                        "errors": []map[string]interface{}{
                                {"message": "Issue not found or access denied"},
                        },
                }
                json.NewEncoder(w).Encode(response)
        }))
        defer server.Close()

        client := NewClient("test-api-key")
        client.endpoint = server.URL

        issue := &IssueDetails{ID: "uuid-123"}
        err := client.MarkIssueInProgress(issue)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "Issue not found or access denied")
}

func TestMarkIssueInProgress_MutationStructure(t *testing.T) {
        var receivedQueries []GraphQLRequest
        callCount := 0

        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                var query GraphQLRequest
                json.NewDecoder(r.Body).Decode(&query)
                receivedQueries = append(receivedQueries, query)
                
                callCount++
                if callCount == 1 {
                        // First call: getInProgressStateID
                        response := map[string]interface{}{
                                "data": map[string]interface{}{
                                        "workflowStates": map[string]interface{}{
                                                "nodes": []map[string]interface{}{
                                                        {
                                                                "id":   "state-123",
                                                                "name": "In Progress",
                                                                "type": "started",
                                                        },
                                                },
                                        },
                                },
                        }
                        json.NewEncoder(w).Encode(response)
                } else {
                        // Second call: issueUpdate
                        response := map[string]interface{}{
                                "data": map[string]interface{}{
                                        "issueUpdate": map[string]interface{}{
                                                "success": true,
                                        },
                                },
                        }
                        json.NewEncoder(w).Encode(response)
                }
        }))
        defer server.Close()

        client := NewClient("test-api-key")
        client.endpoint = server.URL

        issue := &IssueDetails{
                ID:         "uuid-123",
                Title:      "Test Issue",
                BranchName: "test-branch",
                URL:        "https://linear.app/test",
        }
        err := client.MarkIssueInProgress(issue)
        require.NoError(t, err)

        require.Len(t, receivedQueries, 2)
        
        statesQuery := receivedQueries[0]
        assert.Contains(t, statesQuery.Query, "workflowStates")
        assert.Contains(t, statesQuery.Query, "nodes")
        
        updateQuery := receivedQueries[1]
        assert.Contains(t, updateQuery.Query, "mutation")
        assert.Contains(t, updateQuery.Query, "issueUpdate")
        assert.Contains(t, updateQuery.Query, "stateId")
        assert.Equal(t, "uuid-123", updateQuery.Variables["id"])
        assert.Equal(t, "state-123", updateQuery.Variables["stateId"])
}

func TestMarkIssueInProgress_StateNotFound(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                response := map[string]interface{}{
                        "data": map[string]interface{}{
                                "workflowStates": map[string]interface{}{
                                        "nodes": []map[string]interface{}{
                                                {
                                                        "id":   "state-456",
                                                        "name": "To Do",
                                                        "type": "unstarted",
                                                },
                                                {
                                                        "id":   "state-789",
                                                        "name": "Done",
                                                        "type": "completed",
                                                },
                                        },
                                },
                        },
                }
                json.NewEncoder(w).Encode(response)
        }))
        defer server.Close()

        client := NewClient("test-api-key")
        client.endpoint = server.URL

        issue := &IssueDetails{ID: "uuid-123"}
        err := client.MarkIssueInProgress(issue)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "In Progress state not found")
}

func TestFetchIssueDetails_NotFound(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                response := GraphQLResponse{
                        Data: GraphQLData{
                                Issues: IssuesConnection{
                                        Nodes: []IssueDetails{},
                                },
                        },
                }
                json.NewEncoder(w).Encode(response)
        }))
        defer server.Close()

        client := NewClient("test-api-key")
        client.endpoint = server.URL

        _, err := client.FetchIssueDetails("DEL-999")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "issue not found: DEL-999")
}

func TestParseIssueIdentifier_Success(t *testing.T) {
        tests := []struct {
                input       string
                expectedKey string
                expectedNum int
        }{
                {"DEL-123", "DEL", 123},
                {"PROJ-456", "PROJ", 456},
                {"ABC-1", "ABC", 1},
                {"del-123", "DEL", 123}, // case insensitive
        }

        for _, test := range tests {
                t.Run(test.input, func(t *testing.T) {
                        key, num, err := parseIssueIdentifier(test.input)
                        require.NoError(t, err)
                        assert.Equal(t, test.expectedKey, key)
                        assert.Equal(t, test.expectedNum, num)
                })
        }
}

func TestParseIssueIdentifier_Error(t *testing.T) {
        tests := []struct {
                input string
                error string
        }{
                {"", "issue identifier must be in format TEAM-NUMBER"},
                {"DEL", "issue identifier must be in format TEAM-NUMBER"},
                {"123", "issue identifier must be in format TEAM-NUMBER"},
                {"DEL-", "issue identifier must be in format TEAM-NUMBER"},
                {"DEL-abc", "issue identifier must be in format TEAM-NUMBER"},
                {"-123", "issue identifier must be in format TEAM-NUMBER"},
                {"DEL_123", "issue identifier must be in format TEAM-NUMBER"},
        }

        for _, test := range tests {
                t.Run(test.input, func(t *testing.T) {
                        _, _, err := parseIssueIdentifier(test.input)
                        assert.Error(t, err)
                        assert.Contains(t, err.Error(), test.error)
                })
        }
}
