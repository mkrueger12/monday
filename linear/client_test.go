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
		ID:         "ISSUE-123",
		Title:      "Fix authentication bug",
		BranchName: "issue-123-fix-authentication-bug",
		URL:        "https://linear.app/team/issue/ISSUE-123",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		response := GraphQLResponse{
			Data: GraphQLData{
				Issue: expectedIssue,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-api-key")
	client.endpoint = server.URL

	issue, err := client.FetchIssueDetails("ISSUE-123")
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

	_, err := client.FetchIssueDetails("NONEXISTENT")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
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

	_, err := client.FetchIssueDetails("NONEXISTENT")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Issue not found")
}

func TestFetchIssueDetails_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := NewClient("test-api-key")
	client.endpoint = server.URL

	_, err := client.FetchIssueDetails("ISSUE-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode")
}

func TestFetchIssueDetails_NetworkError(t *testing.T) {
	client := NewClient("test-api-key")
	client.endpoint = "http://nonexistent-server:12345"

	_, err := client.FetchIssueDetails("ISSUE-123")
	assert.Error(t, err)
}

func TestGraphQLQuery_Structure(t *testing.T) {
	var receivedQuery GraphQLRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedQuery)
		
		response := GraphQLResponse{
			Data: GraphQLData{
				Issue: IssueDetails{
					ID:         "ISSUE-123",
					Title:      "Test Issue",
					BranchName: "issue-123-test-issue",
					URL:        "https://linear.app/team/issue/ISSUE-123",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-api-key")
	client.endpoint = server.URL

	_, err := client.FetchIssueDetails("ISSUE-123")
	require.NoError(t, err)

	assert.Contains(t, receivedQuery.Query, "query")
	assert.Contains(t, receivedQuery.Query, "issue")
	assert.Contains(t, receivedQuery.Query, "id")
	assert.Contains(t, receivedQuery.Query, "title")
	assert.Contains(t, receivedQuery.Query, "branchName")
	assert.Contains(t, receivedQuery.Query, "url")
	assert.Equal(t, "ISSUE-123", receivedQuery.Variables["id"])
}
