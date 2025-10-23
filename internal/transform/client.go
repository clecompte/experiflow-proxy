package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communication with the ExperiFlow API
type Client struct {
	baseURL    string
	edgeToken  string
	httpClient *http.Client
	timeout    time.Duration
}

// NewClient creates a new ExperiFlow API client
func NewClient(baseURL, edgeToken string, timeout time.Duration) *Client {
	return &Client{
		baseURL:   baseURL,
		edgeToken: edgeToken,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// GetVariants fetches all variants for an experiment
func (c *Client) GetVariants(ctx context.Context, experimentID string) ([]Variant, error) {
	url := fmt.Sprintf("%s/behavior/experiments/%s/public/variants", c.baseURL, experimentID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch variants: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var variants []Variant
	if err := json.NewDecoder(resp.Body).Decode(&variants); err != nil {
		return nil, fmt.Errorf("decode variants: %w", err)
	}

	return variants, nil
}

// GetTransformSpec fetches the transform specification for a variant
func (c *Client) GetTransformSpec(ctx context.Context, experimentID, variantID string) (*TransformSpec, error) {
	url := fmt.Sprintf("%s/v1/experiments/%s/transform-spec", c.baseURL, experimentID)

	reqBody := map[string]string{
		"variant_id": variantID,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.edgeToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.edgeToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch transform spec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var spec TransformSpec
	if err := json.NewDecoder(resp.Body).Decode(&spec); err != nil {
		return nil, fmt.Errorf("decode transform spec: %w", err)
	}

	return &spec, nil
}
