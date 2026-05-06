package api

import (
	"context"
	"digital-labor/pkg/mcp"
	pb "digital-labor/proto"
)

func (s *ContainerServer) AddMCP(ctx context.Context, req *pb.AddMCPRequest) (*pb.AddMCPResponse, error) {
	mag := mcp.NewManager()

	err := mag.AddClient(req.Url)
	if err != nil {
		return &pb.AddMCPResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.AddMCPResponse{
		Success: true,
		Message: "MCP client add successfully",
	}, nil
}
