package gitlabClient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"glpatEye/internal/metrics"
	"glpatEye/pkg/common"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

func NewGitlabClient(token, baseUrl, pattern, cron string) (*GitlabClient, error) {
	if token == "" {
		log.Println("Gitlab token doesn't set. Check 'GITLAB_TOKEN' env variable")
		os.Exit(1)
	}

	if baseUrl == "" {
		log.Println("Gitlab base url doesn't set OR setted without PROTO (protocol should be 'https')")
		os.Exit(1)
	}

	if pattern == "" {
		log.Println("Regex pattern doesn't set. The sampling functionality won't work")
		os.Exit(1)
	}

	if cron == "" {
		log.Println("Cron expression doesn't set. It should be for looping check tokens.")
		log.Println("If you need to check tokens very frequent, you can use expression '* * * * *'.")
		os.Exit(1)
	}

	return &GitlabClient{
		Token:   token,
		BaseURL: baseUrl,
		Pattern: pattern,
	}, nil
}

func (gc GitlabClient) makeReq(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	url := fmt.Sprintf("%s%s", gc.BaseURL, endpoint)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create http req: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("PRIVATE-TOKEN", gc.Token)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send http req: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		if result != nil {
			if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}
		}

	case http.StatusBadRequest:
		return fmt.Errorf("bad request: check your request signature")

	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized: check your credentials")

	case http.StatusForbidden:
		return fmt.Errorf("forbidden: access denied to requested resource")

	case http.StatusNotFound:
		return fmt.Errorf("not found: requested resource isn't found")

	case http.StatusMethodNotAllowed:
		return fmt.Errorf("method not allowed: http method isn't allowed for this type request")

	case http.StatusTooManyRequests:
		retryAfter := req.Header.Get("Retry-After")
		return fmt.Errorf("too many requests: quantity of requests has exceeded the limit. retry after %s", retryAfter)

	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		body, _ := io.ReadAll(req.Body)
		return fmt.Errorf("server error: status code %d, %s", resp.StatusCode, string(body))

	default:
		body, _ := io.ReadAll(req.Body)
		return fmt.Errorf("unexpected status code (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}

func (gc *GitlabClient) GetIdsResource(ctx context.Context, first int, endCursor string, resourceType string) error {
	var query string
	switch resourceType {
	case "projects":
		query = `
		query ($first: Int!, $after: String!) {
			projects(first: $first, after: $after) {
				nodes {
					id
					name
					httpUrlToRepo
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
		`
	case "groups":
		query = `
		query ($first: Int!, $after: String!) {
			groups(first: $first, after: $after) {
				nodes {
					id
					name
					webUrl
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
		`
	default:
		return fmt.Errorf("unknown resource type: %v", resourceType)
	}
	reqBody := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"first": first,
			"after": endCursor,
		},
	}
	err := gc.makeReq(ctx, http.MethodPost, "/api/graphql", reqBody, &gc.Response)
	if err != nil {
		return fmt.Errorf("failed to decode GraphQL response. %w", err)
	}
	return nil
}

func (gc GitlabClient) CheckAccessTokens(ctx context.Context, ID string, projectMap map[string]ProjectNode, resourceType string) (tokens []AccessToken, err error) {
	regex, err := regexp.Compile(gc.Pattern)
	if err != nil {
		return tokens, fmt.Errorf("failed to compile regexp: %w", err)
	}
	var endpoint string
	switch resourceType {
	case "project":
		endpoint = fmt.Sprintf("/api/v4/projects/%s/access_tokens", ID)
	case "group":
		endpoint = fmt.Sprintf("/api/v4/groups/%s/access_tokens", ID)
	default:
		return tokens, fmt.Errorf("unknown resource type %s: %w", resourceType, err)
	}

	err = gc.makeReq(ctx, http.MethodGet, endpoint, nil, &tokens)
	if err != nil {
		return tokens, fmt.Errorf("failed to fetch token for project: %w", err)
	}

	var matchingTokens []AccessToken
	for _, token := range tokens {
		if regex.MatchString(token.Name) {
			token.ProjectName = projectMap[ID].Name
			if resourceType == "project" {
				token.UrlToRepo = projectMap[ID].HttpUrlToRepo
			} else if resourceType == "group" {
				token.UrlToRepo = projectMap[ID].WebUrl
			} else {
				return nil, fmt.Errorf("error matching. unknown resource type: %s", err)
			}
			token.DaysExpire, err = common.CalculateDaysUntilExpire(token.ExpiresAt)
			if err != nil || token.DaysExpire < 0 {
				continue
			}
			lastUsed := token.LastUsed
			if lastUsed == "" {
				lastUsed = "never"
			}
			metric := metrics.MetricLabels{
				Name:        token.Name,
				ProjectName: token.ProjectName,
				UrlToRepo:   token.UrlToRepo,
				Id:          strconv.Itoa(token.ID),
				LastUsed:    lastUsed,
				Root:        strconv.FormatBool(false),
				DaysExpire:  token.DaysExpire,
			}
			metric.UpdateMetric()
			matchingTokens = append(matchingTokens, token)
		}
	}
	return matchingTokens, nil
}

func (gc GitlabClient) SelfCheckMasterToken(ctx context.Context, token string) (AccessToken, error) {
	var response AccessToken
	err := gc.makeReq(ctx, http.MethodGet, "/api/v4/personal_access_tokens/self", nil, &response)
	if err != nil {
		return AccessToken{}, fmt.Errorf("failed to send http req for self-check master token: %w", err)
	}
	response.DaysExpire, err = common.CalculateDaysUntilExpire(response.ExpiresAt)
	if err != nil || response.DaysExpire < 0 {
		return AccessToken{}, fmt.Errorf("master-token is expired or has invalid expiration date")
	}
	return response, nil
}
