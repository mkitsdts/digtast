package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type modelCard struct {
	ID string `json:"id"`
}

type modelListResponse struct {
	Data []modelCard `json:"data"`
}

// resolveBaseURL normalizes the configured base URL and fills in the default
// OpenAI endpoint when it is empty.
func resolveBaseURL(baseURL string) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	parsed, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid openai base url: %w", err)
	}
	if strings.HasSuffix(parsed.Path, "/chat/completions") {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/chat/completions")
		baseURL = strings.TrimRight(parsed.String(), "/")
	}
	return baseURL, nil
}

// ListModels returns the model identifiers exposed by an OpenAI-compatible
// /models endpoint.
func ListModels(ctx context.Context, cfg Config) ([]string, error) {
	baseURL, err := resolveBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}
	req.Header.Set("Accept", "application/json")
	for key, values := range cfg.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var payload modelListResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode model list: %w", err)
	}

	seen := make(map[string]struct{}, len(payload.Data))
	names := make([]string, 0, len(payload.Data))
	for _, model := range payload.Data {
		id := strings.TrimSpace(model.ID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		names = append(names, id)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return nil, errors.New("provider returned no models")
	}
	return names, nil
}
