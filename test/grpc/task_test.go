package grpc

import (
	"testing"

	pb "digital-labor/proto"
)

func TestTaskMethods(t *testing.T) {
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

	t.Run("GetTaskStatus", func(t *testing.T) {
		resp, err := client.GetTaskStatus(ctx, &pb.GetTaskStatusRequest{
			AgentId: agentID,
		})
		if err != nil {
			t.Fatalf("GetTaskStatus failed: %v", err)
		}
		if resp.AgentId != agentID {
			t.Errorf("expected agent_id %s, got %s", agentID, resp.AgentId)
		}
	})

	t.Run("StopTask", func(t *testing.T) {
		resp, err := client.StopTask(ctx, &pb.StopTaskRequest{
			AgentId: agentID,
		})
		if err != nil {
			t.Fatalf("StopTask failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false")
		}
	})
}
