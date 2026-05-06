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

func (s *ContainerServer) GetAllMCPs(ctx context.Context, req *pb.GetAllMCPsRequest) (*pb.GetAllMCPsResponse, error) {
	mag := mcp.NewManager()
	clients := mag.GetAllClients()

	mcps := make([]*pb.MCPInfo, 0, len(clients))
	for _, c := range clients {
		mcps = append(mcps, &pb.MCPInfo{
			Url:     c.URL,
			Enabled: c.Enabled,
		})
	}

	return &pb.GetAllMCPsResponse{
		Mcps: mcps,
	}, nil
}

func (s *ContainerServer) DisableMCP(ctx context.Context, req *pb.DisableMCPRequest) (*pb.DisableMCPResponse, error) {
	mag := mcp.NewManager()
	mag.DisableClient(req.Url)
	return &pb.DisableMCPResponse{
		Success: true,
	}, nil
}

func (s *ContainerServer) EnableMCP(ctx context.Context, req *pb.EnableMCPRequest) (*pb.EnableMCPResponse, error) {
	mag := mcp.NewManager()
	mag.EnableClient(req.Url)
	return &pb.EnableMCPResponse{
		Success: true,
	}, nil
}
