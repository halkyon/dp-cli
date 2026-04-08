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
)

var (
	ErrMissingAPIKey   = errors.New("DATAPACKET_API_KEY environment variable is not set")
	ErrGraphQLResponse = errors.New("GraphQL errors")
)

type Client struct {
	httpClient *http.Client
	apiKey     string
}

func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}
	return &Client{
		httpClient: &http.Client{},
		apiKey:     apiKey,
	}, nil
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
		"https://api.datapacket.com/v0/graphql",
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("closing response body: %w", err)
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
