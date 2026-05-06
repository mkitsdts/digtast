package grpc

import (
	"io"
	"testing"

	pb "digital-labor/proto"
)

func TestChatMethods(t *testing.T) {
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

	t.Run("SendMessage", func(t *testing.T) {
		stream, err := client.SendMessage(ctx, &pb.SendMessageRequest{
			AgentId:  agentID,
			Message:  cfg.Message,
			IsStream: true,
		})
		if err != nil {
			t.Fatalf("SendMessage failed: %v", err)
		}

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("stream.Recv failed: %v", err)
			}
			t.Logf("received delta: %s", resp.DeltaContent)
			if resp.IsFinal {
				break
			}
		}
	})
}
