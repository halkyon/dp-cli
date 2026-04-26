package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrMissingAPIKey   = errors.New("API key is required; set DATAPACKET_API_KEY or api_key in ~/.config/dp/credentials")
	ErrGraphQLResponse = errors.New("GraphQL errors")
)

const maxResponseSize = 10 * 1024 * 1024 // 10 MB

type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
		baseURL:    "https://api.datapacket.com/v0/graphql",
	}, nil
}

func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

func (c *Client) Query(ctx context.Context, query string, variables map[string]any, result any) error {
	bodyBytes, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // standard defer pattern

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Truncate very large error bodies to avoid flooding terminals
		msg := string(body)
		if len(msg) > 1024 {
			msg = msg[:1024] + "... (truncated)"
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
	}

	var gqlResp struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		errMsgs := make([]string, 0, len(gqlResp.Errors))
		for _, e := range gqlResp.Errors {
			errMsgs = append(errMsgs, e.Message)
		}
		return fmt.Errorf("%w: %s", ErrGraphQLResponse, strings.Join(errMsgs, "; "))
	}

	if err := json.Unmarshal(gqlResp.Data, result); err != nil {
		return fmt.Errorf("unmarshaling data: %w", err)
	}

	return nil
}
