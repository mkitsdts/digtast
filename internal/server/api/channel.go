package api

import (
	"context"
	"digital-labor/pkg/gateway"
	pb "digital-labor/proto"
)

func (s *ContainerServer) GetAllChannels(ctx context.Context, req *pb.GetAllChannelsRequest) (*pb.GetAllChannelsResponse, error) {
	kinds := gateway.GetAllChannels()
	res := make([]*pb.ChannelInfo, 0, len(kinds))
	for _, kind := range kinds {
		res = append(res, &pb.ChannelInfo{
			Kind:    kind,
			Enabled: gateway.IsChannelRunning(kind),
		})
	}
	return &pb.GetAllChannelsResponse{
		Channels: res,
	}, nil
}

func (s *ContainerServer) CreateChannel(ctx context.Context, req *pb.CreateChannelRequest) (*pb.CreateChannelResponse, error) {
	params := make(map[string]any)
	for k, v := range req.Params {
		params[k] = v
	}
	err := gateway.CreateChannel(req.Kind, params)
	if err != nil {
		return &pb.CreateChannelResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.CreateChannelResponse{
		Success: true,
	}, nil
}

func (s *ContainerServer) RemoveChannel(ctx context.Context, req *pb.RemoveChannelRequest) (*pb.RemoveChannelResponse, error) {
	err := gateway.RemoveChannel(req.Kind)
	if err != nil {
		return &pb.RemoveChannelResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.RemoveChannelResponse{
		Success: true,
	}, nil
}

func (s *ContainerServer) EnableChannel(ctx context.Context, req *pb.EnableChannelRequest) (*pb.EnableChannelResponse, error) {
	err := gateway.EnableChannel(req.Kind)
	if err != nil {
		return &pb.EnableChannelResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.EnableChannelResponse{
		Success: true,
	}, nil
}

func (s *ContainerServer) DisableChannel(ctx context.Context, req *pb.DisableChannelRequest) (*pb.DisableChannelResponse, error) {
	err := gateway.DisableChannel(req.Kind)
	if err != nil {
		return &pb.DisableChannelResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.DisableChannelResponse{
		Success: true,
	}, nil
}
