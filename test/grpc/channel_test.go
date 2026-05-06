package grpc

import (
	"testing"

	pb "digital-labor/proto"
)

func TestChannelMethods(t *testing.T) {
	ctx := getContext()
	channelKind := "qq"

	t.Run("CreateChannel", func(t *testing.T) {
		resp, err := client.CreateChannel(ctx, &pb.CreateChannelRequest{
			Kind: channelKind,
			Params: map[string]string{
				"account": "123456",
			},
		})
		if err != nil {
			t.Fatalf("CreateChannel failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})

	t.Run("GetAllChannels", func(t *testing.T) {
		resp, err := client.GetAllChannels(ctx, &pb.GetAllChannelsRequest{})
		if err != nil {
			t.Fatalf("GetAllChannels failed: %v", err)
		}
		found := false
		for _, c := range resp.Channels {
			if c.Kind == channelKind {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("channel kind %s not found in list", channelKind)
		}
	})

	t.Run("DisableChannel", func(t *testing.T) {
		resp, err := client.DisableChannel(ctx, &pb.DisableChannelRequest{
			Kind: channelKind,
		})
		if err != nil {
			t.Fatalf("DisableChannel failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})

	t.Run("EnableChannel", func(t *testing.T) {
		resp, err := client.EnableChannel(ctx, &pb.EnableChannelRequest{
			Kind: channelKind,
		})
		if err != nil {
			t.Fatalf("EnableChannel failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})

	t.Run("RemoveChannel", func(t *testing.T) {
		resp, err := client.RemoveChannel(ctx, &pb.RemoveChannelRequest{
			Kind: channelKind,
		})
		if err != nil {
			t.Fatalf("RemoveChannel failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})
}
