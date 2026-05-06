package api

import (
	"context"
	"digital-labor/pkg/chatmodel"
	"digital-labor/pkg/conf"
	pb "digital-labor/proto"
)

func (s *ContainerServer) CreateChatModel(ctx context.Context, req *pb.CreateChatModelRequest) (*pb.CreateChatModelResponse, error) {
	err := chatmodel.ModelManager.CreateModel(req.ModelName, conf.ModelConfig{
		ModelNames: []string{req.ModelName},
		Provider:   req.Provider,
		URL:        req.BaseUrl,
		Key:        req.Key,
	})

	if err != nil {
		return &pb.CreateChatModelResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.CreateChatModelResponse{
		Success:     true,
		ChatModelId: req.ModelName,
	}, nil
}

func (s *ContainerServer) RemoveChatModel(ctx context.Context, req *pb.RemoveChatModelRequest) (*pb.RemoveChatModelResponse, error) {
	err := chatmodel.ModelManager.RemoveModel(req.ChatModelId)

	if err != nil {
		return &pb.RemoveChatModelResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.RemoveChatModelResponse{
		Success: true,
	}, nil
}
