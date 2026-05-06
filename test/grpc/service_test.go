package grpc

import (
	"testing"

	pb "digital-labor/proto"
)

func TestServiceMethods(t *testing.T) {
	ctx := getContext()

	t.Run("StartService", func(t *testing.T) {
		resp, err := client.StartService(ctx, &pb.StartServiceRequest{
			ContainerId: cfg.ContainerID,
			AgentId:     cfg.AgentID,
		})
		if err != nil {
			t.Fatalf("StartService failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})

	t.Run("RestartService", func(t *testing.T) {
		resp, err := client.RestartService(ctx, &pb.RestartServiceRequest{})
		if err != nil {
			t.Fatalf("RestartService failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})

	t.Run("StopService", func(t *testing.T) {
		resp, err := client.StopService(ctx, &pb.StopServiceRequest{})
		if err != nil {
			t.Fatalf("StopService failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})

	t.Run("RemoveService", func(t *testing.T) {
		resp, err := client.RemoveService(ctx, &pb.RemoveServiceRequest{})
		if err != nil {
			t.Fatalf("RemoveService failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})
}
