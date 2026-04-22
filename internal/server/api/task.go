package api

import (
	"context"
	"digital-labor/internal/agent"
	pb "digital-labor/proto"
)

// StopTask 暂停任务
func (s *ContainerServer) StopTask(ctx context.Context, req *pb.StopTaskRequest) (*pb.StopTaskResponse, error) {
	if err := agent.GetManager().Pause(req.ContainerId, req.AgentId, req.SessionId); err != nil {
		return nil, err
	}
	return &pb.StopTaskResponse{
		Success: true,
	}, nil
}

// RestartTask 恢复/重启任务
func (s *ContainerServer) RestartTask(ctx context.Context, req *pb.RestartTaskRequest) (*pb.RestartTaskResponse, error) {
	// TODO: 实现恢复任务逻辑
	return &pb.RestartTaskResponse{
		Success: true,
	}, nil
}

// RemoveTask 移除任务
func (s *ContainerServer) RemoveTask(ctx context.Context, req *pb.RemoveTaskRequest) (*pb.RemoveTaskResponse, error) {
	// TODO: 实现移除任务逻辑
	return &pb.RemoveTaskResponse{
		Success: true,
	}, nil
}
