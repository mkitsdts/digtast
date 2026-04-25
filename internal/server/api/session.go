package api

import (
	"context"
	"digital-labor/internal/center"
	"digital-labor/pkg/ctxmanager"
	pb "digital-labor/proto"

	"github.com/google/uuid"
)

func (s *ContainerServer) GetOrCreateSession(ctx context.Context, req *pb.GetOrCreateSessionRequest) (*pb.GetOrCreateSessionResponse, error) {
	id := buildAgentKey(req.ContainerId, req.AgentId)
	ag, err := center.GetCenter().GetAgent(id)
	if err != nil {
		return nil, err
	}

	sessionId, sess := ag.GetOrCreateSession(req.SessionId)

	msgs := make([]*pb.Message, 0, len(sess.GetMessages()))
	for _, msg := range sess.GetMessages() {
		msgs = append(msgs, &pb.Message{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	return &pb.GetOrCreateSessionResponse{
		SessionId: sessionId,
		Messages:  msgs,
	}, nil
}

// RemoveSession 删除会话
func (s *ContainerServer) RemoveSession(ctx context.Context, req *pb.RemoveSessionRequest) (*pb.RemoveSessionResponse, error) {
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
	id := buildAgentKey(req.ContainerId, req.AgentId)
	digitalAgent, err := center.GetCenter().GetAgent(id)
	if err != nil {
		return err
	}

	sm, err := digitalAgent.Run(ctx, req.Message, req.IsStream, req.SessionId)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case delta, ok := <-sm:
			if !ok {
				return nil
			}
			if err := stream.Send(&pb.SendMessageToSessionResponse{DeltaContent: delta}); err != nil {
				return err
			}
		}
	}
}
