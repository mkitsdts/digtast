package grpc

import (
	"testing"

	pb "digital-labor/proto"
)

func TestModelMethods(t *testing.T) {
	ctx := getContext()
	var chatModelID string

	t.Run("CreateChatModel", func(t *testing.T) {
		resp, err := client.CreateChatModel(ctx, &pb.CreateChatModelRequest{
			ModelName: cfg.ModelName,
			Key:       cfg.APIKey,
			BaseUrl:   cfg.BaseURL,
			Provider:  cfg.Provider,
		})
		if err != nil {
			t.Fatalf("CreateChatModel failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
		chatModelID = resp.ChatModelId
	})

	t.Run("RemoveChatModel", func(t *testing.T) {
		if chatModelID == "" {
			t.Skip("chatModelID is empty")
		}
		resp, err := client.RemoveChatModel(ctx, &pb.RemoveChatModelRequest{
			ChatModelId: chatModelID,
		})
		if err != nil {
			t.Fatalf("RemoveChatModel failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})
}
