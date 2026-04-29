package api

import (
	"context"
	"digital-labor/internal/center"
	"digital-labor/pkg/ctxmanager"
	pb "digital-labor/proto"
	"errors"
	"log/slog"

	"github.com/google/uuid"
)

func (s *ContainerServer) GetOrCreateSession(ctx context.Context, req *pb.GetOrCreateSessionRequest) (*pb.GetOrCreateSessionResponse, error) {
	ag, err := center.AgentManager.GetAgent(req.AgentId)
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
	ag, err := center.AgentManager.GetAgent(req.AgentId)
	if err != nil {
		return nil, err
	}

	if err := ag.RemoveSession(req.SessionId); err != nil {
		return &pb.RemoveSessionResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

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
	digitalAgent, err := center.AgentManager.GetAgent(req.AgentId)
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

func (s *ContainerServer) StopTask(ctx context.Context, req *pb.StopTaskRequest) (*pb.StopTaskResponse, error) {
	ag, err := center.AgentManager.GetAgent(req.AgentId)
	if err != nil {
		return nil, errors.New("invaild container id or session id")
	}

	if err := ag.Cancel(); err != nil {
		slog.Error("cancel task failed", "error", err)
		return nil, err
	}

	return &pb.StopTaskResponse{Success: true}, nil
}

func (s *ContainerServer) CompressSession(ctx context.Context, req *pb.CompressSessionRequest) (*pb.CompressSessionResponse, error) {
	// TODO: 需要对应记忆模块的压缩
	// 目前仅作为占位符，未来可以调用 LLM 进行会话总结并替换历史记录

	return &pb.CompressSessionResponse{
		Success: true,
		Message: "Session compression is not yet implemented, but the request was received.",
	}, nil
}
