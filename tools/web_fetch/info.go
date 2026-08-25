package webfetch

import (
	"context"
	"digital-labor/pkg/llm"
	toolpkg "digital-labor/pkg/tool"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
)

type FetchTool struct {
	client *http.Client
	mux    *sync.Mutex
}

func (t *FetchTool) Definition(context.Context) (llm.ToolDefinition, error) {
	return llm.ToolDefinition{
		Name:        "web_fetch",
		Description: "Fetch an HTTP URL and return its response body.",
		Parameters: toolpkg.ObjectSchema(map[string]any{
			"url":    map[string]any{"type": "string"},
			"method": map[string]any{"type": "string", "default": "GET"},
			"body":   map[string]any{"type": "string"},
		}, "url"),
	}, nil
}

type WebFetchArguments struct {
	Method string `json:"method"`
	URL    string `json:"url"`
	Body   string `json:"body"`
}

func (t *FetchTool) Execute(ctx context.Context, arguments json.RawMessage) (*toolpkg.Result, error) {
	t.mux.Lock()
	defer t.mux.Unlock()

	args := WebFetchArguments{}
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, err
	}
	if args.Method == "" {
		args.Method = http.MethodGet
	}

	req, err := http.NewRequestWithContext(ctx, args.Method, args.URL, strings.NewReader(args.Body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &toolpkg.Result{Content: string(body), Data: map[string]any{"status": resp.StatusCode}}, nil
}
