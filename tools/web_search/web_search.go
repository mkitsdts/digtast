package websearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"digital-labor/internal/agent/registry"
	"digital-labor/pkg/chatmodel"
	"digital-labor/pkg/tool"

	"github.com/PuerkitoBio/goquery"
)

type WebSearchTool struct {
	client     *http.Client
	maxResults int
}

func NewWebSearchTool() tool.Tool {
	return &WebSearchTool{client: &http.Client{Timeout: 20 * time.Second}, maxResults: 10}
}

func (t *WebSearchTool) Definition(context.Context) (chatmodel.ToolDefinition, error) {
	return chatmodel.ToolDefinition{
		Name:        "duckduckgo_search",
		Description: "Search the web using DuckDuckGo and return result titles, URLs, and snippets.",
		Parameters: tool.ObjectSchema(map[string]any{
			"query": map[string]any{"type": "string", "description": "Search query."},
		}, "query"),
	}, nil
}

func (t *WebSearchTool) Execute(ctx context.Context, arguments json.RawMessage) (*tool.Result, error) {
	var input struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(arguments, &input); err != nil {
		return nil, err
	}
	input.Query = strings.TrimSpace(input.Query)
	if input.Query == "" {
		return nil, errors.New("search query is required")
	}
	endpoint := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(input.Query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Digtast/1.0)")
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("duckduckgo returned %s", resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	var results []string
	doc.Find(".result").EachWithBreak(func(_ int, selection *goquery.Selection) bool {
		link := selection.Find(".result__a").First()
		title := strings.TrimSpace(link.Text())
		href, _ := link.Attr("href")
		snippet := strings.TrimSpace(selection.Find(".result__snippet").First().Text())
		if title != "" && href != "" {
			results = append(results, fmt.Sprintf("%d. %s\n%s\n%s", len(results)+1, title, href, snippet))
		}
		return len(results) < t.maxResults
	})
	if len(results) == 0 {
		return &tool.Result{Content: "No search results found."}, nil
	}
	return &tool.Result{Content: strings.Join(results, "\n\n")}, nil
}

func init() {
	t := NewWebSearchTool()
	registry.RegisterTool(t)
	registry.RegisterBaseTool(t)
}
