package api

import (
	"context"
	"digital-labor/internal/center"
	pb "digital-labor/proto"
	"log/slog"
)

// GetAgentSession returns the agent's single session.
func (s *ContainerServer) GetAgentSession(ctx context.Context, req *pb.GetAgentSessionRequest) (*pb.GetAgentSessionResponse, error) {
	ag, err := center.AgentManager.GetAgent(req.AgentId)
	if err != nil {
		return nil, err
	}

	sess, err := ag.GetSession()
	if err != nil {
		return nil, err
	}

	msgs := make([]*pb.Message, 0, len(sess.GetMessages()))
	for _, msg := range sess.GetMessages() {
		msgs = append(msgs, &pb.Message{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	return &pb.GetAgentSessionResponse{
		AgentId:  ag.ID,
		Messages: msgs,
	}, nil
}

// ClearAgentHistory clears the agent's conversation history.
func (s *ContainerServer) ClearAgentHistory(ctx context.Context, req *pb.ClearAgentHistoryRequest) (*pb.ClearAgentHistoryResponse, error) {
	ag, err := center.AgentManager.GetAgent(req.AgentId)
	if err != nil {
		return nil, err
	}

	if err := ag.ClearHistory(); err != nil {
		return &pb.ClearAgentHistoryResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.ClearAgentHistoryResponse{
		Success: true,
	}, nil
}

// SendMessage sends a message to the agent.
func (s *ContainerServer) SendMessage(req *pb.SendMessageRequest, stream pb.ContainerService_SendMessageServer) error {
	ctx := context.Background()

	digitalAgent, err := center.AgentManager.GetAgent(req.AgentId)
	if err != nil {
		return err
	}

	sm, err := digitalAgent.Run(ctx, req.Message, req.IsStream)
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
			if err := stream.Send(&pb.SendMessageResponse{DeltaContent: delta}); err != nil {
				return err
			}
		}
	}
}

func (s *ContainerServer) StopTask(ctx context.Context, req *pb.StopTaskRequest) (*pb.StopTaskResponse, error) {
	ag, err := center.AgentManager.GetAgent(req.AgentId)
	if err != nil {
		return nil, err
	}

	if err := ag.Cancel(); err != nil {
		slog.Error("cancel task failed", "error", err)
		return nil, err
	}

	return &pb.StopTaskResponse{Success: true}, nil
}

// CompressAgentHistory returns the status of automatic context compression.
// Compression is handled automatically by the summarization middleware during conversations.
func (s *ContainerServer) CompressAgentHistory(ctx context.Context, req *pb.CompressAgentHistoryRequest) (*pb.CompressAgentHistoryResponse, error) {
	_, err := center.AgentManager.GetAgent(req.AgentId)
	if err != nil {
		return &pb.CompressAgentHistoryResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.CompressAgentHistoryResponse{
		Success: true,
		Message: "Context compression is handled automatically during conversations.",
	}, nil
}
