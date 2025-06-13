// Package linear provides a GraphQL client for interacting with the Linear API.
// It handles authentication, issue fetching, status updates, and workflow management
// for automating Linear issue processing in development workflows.
package linear

import (
        "bytes"
        "encoding/json"
        "fmt"
        "io"
        "net/http"
        "regexp"
        "strconv"
        "strings"
        "time"
)

// DefaultLinearEndpoint is the standard Linear API GraphQL endpoint
const DefaultLinearEndpoint = "https://api.linear.app/graphql"

// IssueDetails represents the essential information about a Linear issue
// that is needed for creating development environments and tracking progress.
type IssueDetails struct {
        // ID is the internal UUID used by Linear for API operations
        ID          string `json:"id"`
        // Title is the human-readable issue title
        Title       string `json:"title"`
        // Description contains the detailed issue description/requirements
        Description string `json:"description"`
        // BranchName is the suggested git branch name for this issue
        BranchName  string `json:"branchName"`
        // URL is the direct link to view the issue in Linear's web interface
        URL         string `json:"url"`
}

// GraphQLRequest represents a standard GraphQL request structure
// with query string and variables for parameterized queries.
type GraphQLRequest struct {
        Query     string                 `json:"query"`
        Variables map[string]interface{} `json:"variables"`
}

// GraphQLResponse represents the standard GraphQL response structure
// containing either data or errors from the Linear API.
type GraphQLResponse struct {
        Data   GraphQLData    `json:"data"`
        Errors []GraphQLError `json:"errors"`
}

// GraphQLData contains the actual data returned from Linear API queries.
// The structure varies based on the specific query being executed.
type GraphQLData struct {
        Issues IssuesConnection `json:"issues"`
        Teams  TeamsConnection  `json:"teams"`
}

// IssuesConnection represents a paginated collection of issues
// following GraphQL connection patterns used by Linear.
type IssuesConnection struct {
        Nodes []IssueDetails `json:"nodes"`
}

// TeamsConnection represents a paginated collection of teams
type TeamsConnection struct {
        Nodes []Team `json:"nodes"`
}

// Team represents a Linear team with projects
type Team struct {
        ID       string    `json:"id"`
        Key      string    `json:"key"`
        Name     string    `json:"name"`
        Projects ProjectsConnection `json:"projects"`
}

// ProjectsConnection represents a paginated collection of projects
type ProjectsConnection struct {
        Nodes []Project `json:"nodes"`
}

// Project represents a Linear project
type Project struct {
        ID   string `json:"id"`
        Name string `json:"name"`
        Key  string `json:"key"`
}

// GraphQLError represents an error returned by the Linear GraphQL API
// with a human-readable error message.
type GraphQLError struct {
        Message string `json:"message"`
}

// IssueUpdateResponse represents the response from issue mutation operations
// such as changing issue status or updating properties.
type IssueUpdateResponse struct {
        Data   IssueUpdateData `json:"data"`
        Errors []GraphQLError  `json:"errors"`
}

// IssueUpdateData contains the result of an issue update mutation.
type IssueUpdateData struct {
        IssueUpdate IssueUpdateResult `json:"issueUpdate"`
}

// IssueUpdateResult indicates whether an issue update operation succeeded.
type IssueUpdateResult struct {
        Success bool `json:"success"`
}

// Client provides authenticated access to the Linear API with configurable endpoints
// and timeout settings for reliable API communication.
type Client struct {
        // apiKey is the Linear API authentication token
        apiKey   string
        // endpoint is the GraphQL API URL (configurable for testing)
        endpoint string
        // client is the HTTP client with configured timeouts
        client   *http.Client
}

// NewClient creates a new Linear API client with the provided API key.
// It initializes the client with the default Linear endpoint and a 30-second timeout
// for reliable API communication even under network latency.
func NewClient(apiKey string) *Client {
        return &Client{
                apiKey:   apiKey,
                endpoint: DefaultLinearEndpoint,
                client: &http.Client{
                        Timeout: 30 * time.Second,
                },
        }
}

// SetEndpoint allows overriding the Linear API endpoint URL.
// This is primarily used for testing with mock servers or custom Linear instances.
func (c *Client) SetEndpoint(endpoint string) {
        c.endpoint = endpoint
}

// FetchIssueDetails retrieves comprehensive information about a Linear issue by its identifier.
// It accepts issue identifiers in the format "TEAM-123" (e.g., "DEL-163") and returns
// all necessary details for creating development environments and tracking progress.
func (c *Client) FetchIssueDetails(issueID string) (*IssueDetails, error) {
        // Parse the issue identifier into team key and issue number
        teamKey, number, err := parseIssueIdentifier(issueID)
        if err != nil {
                return nil, fmt.Errorf("invalid issue identifier format: %w", err)
        }

        // GraphQL query to fetch issue details using team key and number filtering
        // This approach works with human-readable identifiers like "DEL-163"
        query := `
                query GetIssue($teamKey: String!, $number: Float!) {
                        issues(filter: {
                                team: { key: { eq: $teamKey } },
                                number: { eq: $number }
                        }, first: 1) {
                                nodes {
                                        id
                                        title
                                        description
                                        branchName
                                        url
                                }
                        }
                }
        `

        // Prepare the GraphQL request with variables
        request := GraphQLRequest{
                Query: query,
                Variables: map[string]interface{}{
                        "teamKey": teamKey,
                        "number":  float64(number), // Linear expects Float for number field
                },
        }

        // Marshal the request to JSON for HTTP transmission
        jsonData, err := json.Marshal(request)
        if err != nil {
                return nil, fmt.Errorf("failed to marshal GraphQL request: %w", err)
        }

        // Create HTTP POST request to Linear's GraphQL endpoint
        req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonData))
        if err != nil {
                return nil, fmt.Errorf("failed to create HTTP request: %w", err)
        }

        // Set required headers for Linear API authentication and content type
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Authorization", c.apiKey) // Linear expects API key directly, not Bearer token

        // Execute the HTTP request
        resp, err := c.client.Do(req)
        if err != nil {
                return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
        }
        defer resp.Body.Close()

        // Check for HTTP-level errors and include response body for debugging
        if resp.StatusCode != http.StatusOK {
                body, _ := io.ReadAll(resp.Body)
                return nil, fmt.Errorf("Linear API returned status %d: %s", resp.StatusCode, string(body))
        }

        // Parse the GraphQL response
        var response GraphQLResponse
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
                return nil, fmt.Errorf("failed to decode GraphQL response: %w", err)
        }

        // Check for GraphQL-level errors
        if len(response.Errors) > 0 {
                return nil, fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
        }

        // Verify that the issue was found
        if len(response.Data.Issues.Nodes) == 0 {
                return nil, fmt.Errorf("issue not found: %s", issueID)
        }

        // Return the first (and only) issue from the results
        return &response.Data.Issues.Nodes[0], nil
}

// MarkIssueInProgress updates the status of a Linear issue to "In Progress".
// This automatically moves the issue through the workflow to indicate active development.
// It first looks up the "In Progress" state ID for the issue's team, then updates the issue.
func (c *Client) MarkIssueInProgress(issue *IssueDetails) error {
        // First, find the "In Progress" state ID for this team's workflow
        stateID, err := c.getInProgressStateID()
        if err != nil {
                return fmt.Errorf("failed to get In Progress state ID: %w", err)
        }

        // GraphQL mutation to update the issue's state
        mutation := `
                mutation UpdateIssue($id: String!, $stateId: String!) {
                        issueUpdate(id: $id, input: { stateId: $stateId }) {
                                success
                        }
                }
        `

        // Prepare the mutation request with issue ID and target state ID
        request := GraphQLRequest{
                Query: mutation,
                Variables: map[string]interface{}{
                        "id":      issue.ID,      // Internal UUID of the issue
                        "stateId": stateID,       // UUID of the "In Progress" state
                },
        }

        // Marshal the request to JSON
        jsonData, err := json.Marshal(request)
        if err != nil {
                return fmt.Errorf("failed to marshal GraphQL request: %w", err)
        }

        // Create HTTP POST request
        req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonData))
        if err != nil {
                return fmt.Errorf("failed to create HTTP request: %w", err)
        }

        // Set authentication and content type headers
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Authorization", c.apiKey)

        // Execute the mutation
        resp, err := c.client.Do(req)
        if err != nil {
                return fmt.Errorf("failed to execute HTTP request: %w", err)
        }
        defer resp.Body.Close()

        // Check for HTTP-level errors
        if resp.StatusCode != http.StatusOK {
                body, _ := io.ReadAll(resp.Body)
                return fmt.Errorf("Linear API returned status %d: %s", resp.StatusCode, string(body))
        }

        // Parse the mutation response
        var response IssueUpdateResponse
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
                return fmt.Errorf("failed to decode GraphQL response: %w", err)
        }

        // Check for GraphQL-level errors
        if len(response.Errors) > 0 {
                return fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
        }

        // Verify that the update operation succeeded
        if !response.Data.IssueUpdate.Success {
                return fmt.Errorf("failed to update issue status")
        }

        return nil
}

// getInProgressStateID dynamically looks up the "In Progress" workflow state ID.
// Different Linear workspaces may have different state configurations, so we query
// all available workflow states and find the one that matches "In Progress" criteria.
func (c *Client) getInProgressStateID() (string, error) {
        // GraphQL query to fetch all workflow states across the workspace
        query := `
                query GetWorkflowStates {
                        workflowStates {
                                nodes {
                                        id
                                        name
                                        type
                                }
                        }
                }
        `

        // Prepare the query request (no variables needed)
        request := GraphQLRequest{
                Query:     query,
                Variables: map[string]interface{}{},
        }

        // Marshal request to JSON
        jsonData, err := json.Marshal(request)
        if err != nil {
                return "", fmt.Errorf("failed to marshal GraphQL request: %w", err)
        }

        // Create HTTP POST request
        req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonData))
        if err != nil {
                return "", fmt.Errorf("failed to create HTTP request: %w", err)
        }

        // Set authentication headers
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Authorization", c.apiKey)

        // Execute the request
        resp, err := c.client.Do(req)
        if err != nil {
                return "", fmt.Errorf("failed to execute HTTP request: %w", err)
        }
        defer resp.Body.Close()

        // Check for HTTP errors
        if resp.StatusCode != http.StatusOK {
                body, _ := io.ReadAll(resp.Body)
                return "", fmt.Errorf("Linear API returned status %d: %s", resp.StatusCode, string(body))
        }

        // Define response structure for workflow states query
        var response struct {
                Data struct {
                        WorkflowStates struct {
                                Nodes []struct {
                                        ID   string `json:"id"`
                                        Name string `json:"name"`
                                        Type string `json:"type"`
                                } `json:"nodes"`
                        } `json:"workflowStates"`
                } `json:"data"`
                Errors []GraphQLError `json:"errors"`
        }

        // Parse the response
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
                return "", fmt.Errorf("failed to decode GraphQL response: %w", err)
        }

        // Check for GraphQL errors
        if len(response.Errors) > 0 {
                return "", fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
        }

        // Search for the "In Progress" state with type "started"
        // Linear uses "started" type for active development states
        for _, state := range response.Data.WorkflowStates.Nodes {
                if state.Name == "In Progress" && state.Type == "started" {
                        return state.ID, nil
                }
        }

        return "", fmt.Errorf("In Progress state not found")
}

// parseIssueIdentifier extracts team key and issue number from Linear issue identifiers.
// It accepts identifiers in the format "TEAM-123" (e.g., "DEL-163", "ENG-42") and
// parseIssueIdentifier parses a Linear issue identifier of the form "TEAM-123" into its team key and numeric issue number.
// Returns an error if the identifier does not match the expected format or if the issue number is invalid.
func parseIssueIdentifier(identifier string) (string, int, error) {
        // Regular expression to match Linear issue format: letters-digits
        re := regexp.MustCompile(`^([A-Z]+)-(\d+)$`)
        matches := re.FindStringSubmatch(strings.ToUpper(identifier))
        
        // Validate that we have exactly 3 matches (full match + 2 capture groups)
        if len(matches) != 3 {
                return "", 0, fmt.Errorf("issue identifier must be in format TEAM-NUMBER (e.g., DEL-163)")
        }

        // Extract team key (letters before the dash)
        teamKey := matches[1]
        
        // Parse issue number (digits after the dash)
        number, err := strconv.Atoi(matches[2])
        if err != nil {
                return "", 0, fmt.Errorf("invalid issue number: %s", matches[2])
        }

        return teamKey, number, nil
}

// FetchIssuesByFilters retrieves issues based on team, project, and tag filters
func (c *Client) FetchIssuesByFilters(teamKey, projectKey, tag string) ([]IssueDetails, error) {
        var filters []string
        var variables = make(map[string]interface{})
        
        if teamKey != "" {
                filters = append(filters, "team: { key: { eq: $teamKey } }")
                variables["teamKey"] = teamKey
        }
        
        if projectKey != "" {
                filters = append(filters, "project: { key: { eq: $projectKey } }")
                variables["projectKey"] = projectKey
        }
        
        if tag != "" {
                filters = append(filters, "labels: { name: { eq: $tag } }")
                variables["tag"] = tag
        }
        
        filterStr := ""
        if len(filters) > 0 {
                filterStr = fmt.Sprintf("filter: { %s }", strings.Join(filters, ", "))
        }
        
        query := fmt.Sprintf(`
                query GetIssues($teamKey: String, $projectKey: String, $tag: String) {
                        issues(%s, first: 50, orderBy: createdAt) {
                                nodes {
                                        id
                                        title
                                        description
                                        branchName
                                        url
                                }
                        }
                }
        `, filterStr)
        
        request := GraphQLRequest{
                Query:     query,
                Variables: variables,
        }
        
        jsonData, err := json.Marshal(request)
        if err != nil {
                return nil, fmt.Errorf("failed to marshal GraphQL request: %w", err)
        }
        
        req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonData))
        if err != nil {
                return nil, fmt.Errorf("failed to create HTTP request: %w", err)
        }
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Authorization", c.apiKey)
        
        resp, err := c.client.Do(req)
        if err != nil {
                return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusOK {
                body, _ := io.ReadAll(resp.Body)
                return nil, fmt.Errorf("Linear API returned status %d: %s", resp.StatusCode, string(body))
        }
        
        var response GraphQLResponse
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
                return nil, fmt.Errorf("failed to decode GraphQL response: %w", err)
        }
        
        if len(response.Errors) > 0 {
                return nil, fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
        }
        
        return response.Data.Issues.Nodes, nil
}

// FetchTeams retrieves all teams available to the authenticated user
func (c *Client) FetchTeams() ([]Team, error) {
        query := `
                query GetTeams {
                        teams {
                                nodes {
                                        id
                                        key
                                        name
                                        projects {
                                                nodes {
                                                        id
                                                        name
                                                        key
                                                }
                                        }
                                }
                        }
                }
        `
        
        request := GraphQLRequest{
                Query:     query,
                Variables: map[string]interface{}{},
        }
        
        jsonData, err := json.Marshal(request)
        if err != nil {
                return nil, fmt.Errorf("failed to marshal GraphQL request: %w", err)
        }
        
        req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonData))
        if err != nil {
                return nil, fmt.Errorf("failed to create HTTP request: %w", err)
        }
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Authorization", c.apiKey)
        
        resp, err := c.client.Do(req)
        if err != nil {
                return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusOK {
                body, _ := io.ReadAll(resp.Body)
                return nil, fmt.Errorf("Linear API returned status %d: %s", resp.StatusCode, string(body))
        }
        
        var response GraphQLResponse
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
                return nil, fmt.Errorf("failed to decode GraphQL response: %w", err)
        }
        
        if len(response.Errors) > 0 {
                return nil, fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
        }
        
        return response.Data.Teams.Nodes, nil
}
