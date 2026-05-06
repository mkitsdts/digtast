package grpc

import (
	"testing"

	pb "digital-labor/proto"
)

func TestMCPMethods(t *testing.T) {
	ctx := getContext()
	mcpURL := "http://localhost:8080/mcp"

	t.Run("AddMCP", func(t *testing.T) {
		resp, err := client.AddMCP(ctx, &pb.AddMCPRequest{
			Url:      mcpURL,
			Provider: "test-provider",
		})
		if err != nil {
			t.Fatalf("AddMCP failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})

	t.Run("GetAllMCPs", func(t *testing.T) {
		resp, err := client.GetAllMCPs(ctx, &pb.GetAllMCPsRequest{})
		if err != nil {
			t.Fatalf("GetAllMCPs failed: %v", err)
		}
		found := false
		for _, m := range resp.Mcps {
			if m.Url == mcpURL {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("MCP with URL %s not found in list", mcpURL)
		}
	})

	t.Run("DisableMCP", func(t *testing.T) {
		resp, err := client.DisableMCP(ctx, &pb.DisableMCPRequest{
			Url: mcpURL,
		})
		if err != nil {
			t.Fatalf("DisableMCP failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})

	t.Run("EnableMCP", func(t *testing.T) {
		resp, err := client.EnableMCP(ctx, &pb.EnableMCPRequest{
			Url: mcpURL,
		})
		if err != nil {
			t.Fatalf("EnableMCP failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})
}
