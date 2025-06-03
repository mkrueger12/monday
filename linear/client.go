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

const DefaultLinearEndpoint = "https://api.linear.app/graphql"

type IssueDetails struct {
        ID         string `json:"id"`
        Title      string `json:"title"`
        BranchName string `json:"branchName"`
        URL        string `json:"url"`
}

type GraphQLRequest struct {
        Query     string                 `json:"query"`
        Variables map[string]interface{} `json:"variables"`
}

type GraphQLResponse struct {
        Data   GraphQLData    `json:"data"`
        Errors []GraphQLError `json:"errors"`
}

type GraphQLData struct {
        Issues IssuesConnection `json:"issues"`
}

type IssuesConnection struct {
        Nodes []IssueDetails `json:"nodes"`
}

type GraphQLError struct {
        Message string `json:"message"`
}

type IssueUpdateResponse struct {
        Data   IssueUpdateData `json:"data"`
        Errors []GraphQLError  `json:"errors"`
}

type IssueUpdateData struct {
        IssueUpdate IssueUpdateResult `json:"issueUpdate"`
}

type IssueUpdateResult struct {
        Success bool `json:"success"`
}

type Client struct {
        apiKey   string
        endpoint string
        client   *http.Client
}

func NewClient(apiKey string) *Client {
        return &Client{
                apiKey:   apiKey,
                endpoint: DefaultLinearEndpoint,
                client: &http.Client{
                        Timeout: 30 * time.Second,
                },
        }
}

func (c *Client) SetEndpoint(endpoint string) {
        c.endpoint = endpoint
}

func (c *Client) FetchIssueDetails(issueID string) (*IssueDetails, error) {
        teamKey, number, err := parseIssueIdentifier(issueID)
        if err != nil {
                return nil, fmt.Errorf("invalid issue identifier format: %w", err)
        }

        query := `
                query GetIssue($teamKey: String!, $number: Float!) {
                        issues(filter: {
                                team: { key: { eq: $teamKey } },
                                number: { eq: $number }
                        }, first: 1) {
                                nodes {
                                        id
                                        title
                                        branchName
                                        url
                                }
                        }
                }
        `

        request := GraphQLRequest{
                Query: query,
                Variables: map[string]interface{}{
                        "teamKey": teamKey,
                        "number":  float64(number),
                },
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

        if len(response.Data.Issues.Nodes) == 0 {
                return nil, fmt.Errorf("issue not found: %s", issueID)
        }

        return &response.Data.Issues.Nodes[0], nil
}

func (c *Client) MarkIssueInProgress(issue *IssueDetails) error {
        stateID, err := c.getInProgressStateID()
        if err != nil {
                return fmt.Errorf("failed to get In Progress state ID: %w", err)
        }

        mutation := `
                mutation UpdateIssue($id: String!, $stateId: String!) {
                        issueUpdate(id: $id, input: { stateId: $stateId }) {
                                success
                        }
                }
        `

        request := GraphQLRequest{
                Query: mutation,
                Variables: map[string]interface{}{
                        "id":      issue.ID,
                        "stateId": stateID,
                },
        }

        jsonData, err := json.Marshal(request)
        if err != nil {
                return fmt.Errorf("failed to marshal GraphQL request: %w", err)
        }

        req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonData))
        if err != nil {
                return fmt.Errorf("failed to create HTTP request: %w", err)
        }

        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Authorization", c.apiKey)

        resp, err := c.client.Do(req)
        if err != nil {
                return fmt.Errorf("failed to execute HTTP request: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                body, _ := io.ReadAll(resp.Body)
                return fmt.Errorf("Linear API returned status %d: %s", resp.StatusCode, string(body))
        }

        var response IssueUpdateResponse
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
                return fmt.Errorf("failed to decode GraphQL response: %w", err)
        }

        if len(response.Errors) > 0 {
                return fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
        }

        if !response.Data.IssueUpdate.Success {
                return fmt.Errorf("failed to update issue status")
        }

        return nil
}

func (c *Client) getInProgressStateID() (string, error) {
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

        request := GraphQLRequest{
                Query:     query,
                Variables: map[string]interface{}{},
        }

        jsonData, err := json.Marshal(request)
        if err != nil {
                return "", fmt.Errorf("failed to marshal GraphQL request: %w", err)
        }

        req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonData))
        if err != nil {
                return "", fmt.Errorf("failed to create HTTP request: %w", err)
        }

        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Authorization", c.apiKey)

        resp, err := c.client.Do(req)
        if err != nil {
                return "", fmt.Errorf("failed to execute HTTP request: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                body, _ := io.ReadAll(resp.Body)
                return "", fmt.Errorf("Linear API returned status %d: %s", resp.StatusCode, string(body))
        }

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

        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
                return "", fmt.Errorf("failed to decode GraphQL response: %w", err)
        }

        if len(response.Errors) > 0 {
                return "", fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
        }

        for _, state := range response.Data.WorkflowStates.Nodes {
                if state.Name == "In Progress" && state.Type == "started" {
                        return state.ID, nil
                }
        }

        return "", fmt.Errorf("In Progress state not found")
}

func parseIssueIdentifier(identifier string) (string, int, error) {
        re := regexp.MustCompile(`^([A-Z]+)-(\d+)$`)
        matches := re.FindStringSubmatch(strings.ToUpper(identifier))
        if len(matches) != 3 {
                return "", 0, fmt.Errorf("issue identifier must be in format TEAM-NUMBER (e.g., DEL-163)")
        }

        teamKey := matches[1]
        number, err := strconv.Atoi(matches[2])
        if err != nil {
                return "", 0, fmt.Errorf("invalid issue number: %s", matches[2])
        }

        return teamKey, number, nil
}
