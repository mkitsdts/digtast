package api

import (
	"context"
	"digital-labor/internal/agent"
	"digital-labor/pkg/ctxmanager"
	pb "proto/digital_labor"

	"github.com/google/uuid"
)

// RemoveSession 删除会话
func (s *ContainerServer) RemoveSession(ctx context.Context, req *pb.RemoveSessionRequest) (*pb.RemoveSessionResponse, error) {
	conText := context.Background()
	conText = context.WithValue(conText, "userId", req.SessionId)

	//
	agent.GetDigitalAgent().RemoveSession(conText)

	return &pb.RemoveSessionResponse{
		Success: true,
	}, nil
}

// SendMessageToSession 在会话中发起对话
func (s *ContainerServer) SendMessageToSession(req *pb.SendMessageToSessionRequest, stream pb.ContainerService_SendMessageToSessionServer) error {
	ctx := context.Background()

	if req.SessionId == "" {
		req.SessionId = uuid.New().String()
		ctx = ctxmanager.GetOrCreate(req.SessionId)
	}

	ctx = context.WithValue(ctx, "session_id", req.SessionId)

	sm, err := agent.GetDigitalAgent().SendMessageToSession(ctx, req.Message, req.IsStream)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case delta := <-sm:
			if err := stream.Send(&pb.SendMessageToSessionResponse{DeltaContent: delta}); err != nil {
				return err
			}
		}
	}
}
