package grpc

import (
	"testing"

	pb "digital-labor/proto"
)

func TestAgentMethods(t *testing.T) {
	ctx := getContext()
	var agentID string

	// 先创建一个模型
	modelResp, _ := client.CreateChatModel(ctx, &pb.CreateChatModelRequest{
		ModelName: cfg.ModelName,
		Key:       cfg.APIKey,
		BaseUrl:   cfg.BaseURL,
		Provider:  cfg.Provider,
	})

	t.Run("CreateAgent", func(t *testing.T) {
		resp, err := client.CreateAgent(ctx, &pb.CreateAgentRequest{
			AgentName:   cfg.AgentID,
			ChatModelId: modelResp.ChatModelId,
		})
		if err != nil {
			t.Fatalf("CreateAgent failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
		agentID = resp.AgentId
	})

	t.Run("RemoveAgent", func(t *testing.T) {
		if agentID == "" {
			t.Skip("agentID is empty")
		}
		resp, err := client.RemoveAgent(ctx, &pb.RemoveAgentRequest{
			AgentId: agentID,
		})
		if err != nil {
			t.Fatalf("RemoveAgent failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})
}
