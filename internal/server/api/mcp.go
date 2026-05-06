package api

import (
	"context"
	"digital-labor/pkg/registry"
	pb "digital-labor/proto"
	"log"
	"log/slog"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
)

func (s *ContainerServer) AddMCP(ctx context.Context, req *pb.AddMCPRequest) (*pb.AddMCPResponse, error) {
	cli, err := client.NewSSEMCPClient(req.Url)
	if err != nil {
		slog.Error("failed to create mcp client", "err", err)
	}
	err = cli.Start(ctx)
	if err != nil {
		slog.Error("failed to start mcp client", "err", err)
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "digtast",
		Version: "1.0.0",
	}

	_, err = cli.Initialize(ctx, initRequest)
	if err != nil {
		slog.Error("failed to initialize mcp client", "err", err)
	}

	tools, err := mcpp.GetTools(ctx, &mcpp.Config{Cli: cli})
	if err != nil {
		log.Fatal(err)
	}

	for _, tool := range tools {
		registry.RegisterTool(tool)
	}

	return &pb.AddMCPResponse{
		Success: true,
		Message: "MCP client add successfully",
	}, nil
}
