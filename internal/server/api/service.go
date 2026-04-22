package api

import (
	"context"
	"digital-labor/internal/agent"
	pb "proto/digital_labor"
)

// StopService 启动服务
func (s *ContainerServer) StartService(ctx context.Context, req *pb.StartServiceRequest) (*pb.StartServiceResponse, error) {

	err := agent.InitDigitalAgent(req.Provider, req.Key, req.Url, req.ModelName)
	if err != nil {
		return nil, err
	}

	return &pb.StartServiceResponse{
		Success: true,
	}, nil
}

// StopService 暂停服务
func (s *ContainerServer) StopService(ctx context.Context, req *pb.StopServiceRequest) (*pb.StopServiceResponse, error) {
	// TODO: 实现暂停容器逻辑
	return &pb.StopServiceResponse{
		Success: true,
	}, nil
}

// RestartService 重启服务
func (s *ContainerServer) RestartService(ctx context.Context, req *pb.RestartServiceRequest) (*pb.RestartServiceResponse, error) {
	// TODO: 实现重启容器逻辑
	return &pb.RestartServiceResponse{
		Success: true,
	}, nil
}

// RemoveService 移除服务
func (s *ContainerServer) RemoveService(ctx context.Context, req *pb.RemoveServiceRequest) (*pb.RemoveServiceResponse, error) {
	// TODO: 实现移除容器逻辑
	return &pb.RemoveServiceResponse{
		Success: true,
	}, nil
}

// BackupService 备份服务
func (s *ContainerServer) BackupService(ctx context.Context, req *pb.BackupServiceRequest) (*pb.BackupServiceResponse, error) {
	// TODO: 实现备份容器逻辑
	return &pb.BackupServiceResponse{
		BackupUrl: "http://backup-server/container-" + req.ServiceId + ".tar.gz",
	}, nil
}
