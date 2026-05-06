package api

import (
	"context"
	"digital-labor/internal/center"
	"digital-labor/pkg/model"
	pb "digital-labor/proto"
)

func (s *ContainerServer) CreateAgent(ctx context.Context, req *pb.CreateAgentRequest) (*pb.CreateAgentResponse, error) {
	ag, err := center.AgentManager.CreateAgent(req.AgentName, &model.DigitalAgentConfig{
		Name:  req.AgentName,
		Model: req.ChatModelId,
	})
	if err != nil {
		return &pb.CreateAgentResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.CreateAgentResponse{
		AgentId: ag.ID,
		Success: true,
	}, nil
}

func (s *ContainerServer) RemoveAgent(ctx context.Context, req *pb.RemoveAgentRequest) (*pb.RemoveAgentResponse, error) {
	err := center.AgentManager.RemoveAgent(req.AgentId)
	if err != nil {
		return &pb.RemoveAgentResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.RemoveAgentResponse{
		Success: true,
	}, nil
}
