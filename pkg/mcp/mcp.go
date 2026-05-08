package mcp

import (
	"context"
	"log"
	"log/slog"

	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type Client struct {
	*client.Client
	url string
}

func NewMCPConfig(url string) (*Client, error) {
	cli := &Client{}
	var err error
	cli.Client, err = client.NewSSEMCPClient(url)
	if err != nil {
		slog.Error("failed to create mcp client", "err", err)
	}
	return cli, err
}

func (c *Client) Start(ctx context.Context) error {
	err := c.Client.Start(ctx)
	if err != nil {
		slog.Error("failed to start mcp client", "err", err)
		return err
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "digtast",
		Version: "1.0.0",
	}

	_, err = c.Client.Initialize(ctx, initRequest)
	if err != nil {
		slog.Error("failed to initialize mcp client", "err", err)
	}

	return err
}

func (c *Client) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	tools, err := mcpp.GetTools(ctx, &mcpp.Config{
		Cli: c.Client,
	})
	if err != nil {
		log.Fatal(err)
	}

	return tools, nil
}
