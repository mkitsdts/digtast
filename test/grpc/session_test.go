package grpc

import (
	"testing"

	pb "digital-labor/proto"
)

func TestSessionMethods(t *testing.T) {
	ctx := getContext()
	agentID := cfg.AgentID

	// 确保模型存在
	client.CreateChatModel(ctx, &pb.CreateChatModelRequest{
		ModelName: cfg.ModelName,
		Key:       cfg.APIKey,
		BaseUrl:   cfg.BaseURL,
		Provider:  cfg.Provider,
	})

	// 确保 Agent 存在
	client.CreateAgent(ctx, &pb.CreateAgentRequest{
		AgentName:   agentID,
		ChatModelId: cfg.ModelName,
	})

	t.Run("GetAgentSession", func(t *testing.T) {
		resp, err := client.GetAgentSession(ctx, &pb.GetAgentSessionRequest{
			AgentId: agentID,
		})
		if err != nil {
			t.Fatalf("GetAgentSession failed: %v", err)
		}
		if resp.AgentId != agentID {
			t.Errorf("expected agent_id %s, got %s", agentID, resp.AgentId)
		}
	})

	t.Run("ClearAgentHistory", func(t *testing.T) {
		resp, err := client.ClearAgentHistory(ctx, &pb.ClearAgentHistoryRequest{
			AgentId: agentID,
		})
		if err != nil {
			t.Fatalf("ClearAgentHistory failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})

	t.Run("CompressAgentHistory", func(t *testing.T) {
		resp, err := client.CompressAgentHistory(ctx, &pb.CompressAgentHistoryRequest{
			AgentId: agentID,
		})
		if err != nil {
			t.Fatalf("CompressAgentHistory failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})
}
