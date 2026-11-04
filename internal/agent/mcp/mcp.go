package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	toolpkg "digital-labor/component/tool"
	"digital-labor/schema"

	"github.com/mark3labs/mcp-go/client"
	mcpsdk "github.com/mark3labs/mcp-go/mcp"
)

type Client struct {
	*client.Client
	url string
}

func NewMCPConfig(url string) (*Client, error) {
	if strings.TrimSpace(url) == "" {
		return nil, errors.New("mcp url is required")
	}
	var cli *client.Client
	var err error
	if strings.HasSuffix(strings.TrimRight(url, "/"), "/sse") {
		cli, err = client.NewSSEMCPClient(url)
	} else {
		cli, err = client.NewStreamableHttpClient(url)
	}
	if err != nil {
		return nil, fmt.Errorf("create mcp client: %w", err)
	}
	return &Client{Client: cli, url: url}, nil
}

func (c *Client) Start(ctx context.Context) error {
	if err := c.Client.Start(ctx); err != nil {
		return fmt.Errorf("start mcp client: %w", err)
	}
	request := mcpsdk.InitializeRequest{}
	request.Params.ProtocolVersion = mcpsdk.LATEST_PROTOCOL_VERSION
	request.Params.ClientInfo = mcpsdk.Implementation{Name: "digtast", Version: "1.0.0"}
	if _, err := c.Client.Initialize(ctx, request); err != nil {
		return fmt.Errorf("initialize mcp client: %w", err)
	}
	return nil
}

func (c *Client) GetTools(ctx context.Context) ([]toolpkg.Tool, error) {
	var tools []toolpkg.Tool
	var cursor mcpsdk.Cursor
	for {
		request := mcpsdk.ListToolsRequest{}
		request.Params.Cursor = cursor
		result, err := c.Client.ListTools(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("list mcp tools: %w", err)
		}
		for _, definition := range result.Tools {
			tools = append(tools, &SDKTool{client: c.Client, definition: definition})
		}
		if result.NextCursor == "" {
			break
		}
		cursor = result.NextCursor
	}
	return tools, nil
}

type SDKTool struct {
	client     *client.Client
	definition mcpsdk.Tool
}

func (t *SDKTool) Definition(context.Context) (schema.ToolDefinition, error) {
	parameters := make(map[string]any)
	data := t.definition.RawInputSchema
	if len(data) == 0 {
		var err error
		data, err = json.Marshal(t.definition.InputSchema)
		if err != nil {
			return schema.ToolDefinition{}, err
		}
	}
	if err := json.Unmarshal(data, &parameters); err != nil {
		return schema.ToolDefinition{}, fmt.Errorf("decode schema for mcp tool %q: %w", t.definition.Name, err)
	}
	return schema.ToolDefinition{Name: t.definition.Name, Description: t.definition.Description, Parameters: parameters}, nil
}

func (t *SDKTool) Execute(ctx context.Context, arguments json.RawMessage) (*toolpkg.Result, error) {
	var args map[string]any
	if len(arguments) > 0 {
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, fmt.Errorf("decode mcp tool arguments: %w", err)
		}
	}
	request := mcpsdk.CallToolRequest{}
	request.Params = mcpsdk.CallToolParams{Name: t.definition.Name, Arguments: args}
	result, err := t.client.CallTool(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("call mcp tool %q: %w", t.definition.Name, err)
	}
	content, err := formatResult(result)
	if err != nil {
		return nil, err
	}
	return &toolpkg.Result{Content: content, IsError: result.IsError}, nil
}

func formatResult(result *mcpsdk.CallToolResult) (string, error) {
	if result == nil {
		return "", errors.New("mcp tool returned no result")
	}
	var parts []string
	for _, content := range result.Content {
		if text, ok := mcpsdk.AsTextContent(content); ok {
			parts = append(parts, text.Text)
			continue
		}
		data, err := json.Marshal(content)
		if err != nil {
			return "", err
		}
		parts = append(parts, string(data))
	}
	if result.StructuredContent != nil {
		data, err := json.Marshal(result.StructuredContent)
		if err != nil {
			return "", err
		}
		parts = append(parts, string(data))
	}
	return strings.Join(parts, "\n"), nil
}
