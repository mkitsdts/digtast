package api

import (
	"context"
	"digital-labor/internal/center"
	"digital-labor/pkg/model"
	pb "digital-labor/proto"
	"log/slog"
)

// StartService 启动服务
func (s *ContainerServer) StartService(ctx context.Context, req *pb.StartServiceRequest) (*pb.StartServiceResponse, error) {
	slog.Info("StartService request received", "container_id", req.ContainerId, "agent_id", req.AgentId)

	id := buildAgentKey(req.ContainerId, req.AgentId)
	_, err := center.GetCenter().CreateAgent(&model.DigitalAgentConfig{
		ID:       id,
		Key:      req.Key,
		Name:     req.ModelName,
		Model:    req.ModelName,
		URL:      req.Url,
		Provider: req.Provider,
	})
	if err != nil {
		slog.Error("Failed to create agent", "error", err)
		return nil, err
	}

	ftp_port := 2121
	if req.FtpEnabled {
		err := center.GetCenter().StartFTPServer(&center.FTPServerParams{
			Port: ftp_port,
		})
		if err != nil {
			slog.Error("Failed to start FTP server", "error", err)
			return nil, err
		}
	}

	port := -1
	if req.VncEnabled {
		port, err = center.GetCenter().StartVDisplay(&center.VDisplayParams{
			Key: id,
		})
		if err != nil {
			slog.Error("Failed to start VNC server", "error", err)
			return nil, err
		}
	}

	slog.Info("Agent created successfully", "container_id", req.ContainerId, "agent_id", req.AgentId)
	return &pb.StartServiceResponse{
		VisualDisplayPort: int32(port),
		FtpPort:           int32(ftp_port),
		Success:           true,
	}, nil
}

// StopService 暂停服务
func (s *ContainerServer) StopService(ctx context.Context, req *pb.StopServiceRequest) (*pb.StopServiceResponse, error) {
	slog.Info("StopService request received", "container_id", req.ContainerId, "agent_id", req.AgentId)
	id := buildAgentKey(req.ContainerId, req.AgentId)

	// Stop VNC if it was running
	_ = center.GetCenter().StopVDisplay(&center.VDisplayParams{
		Key: id,
	})

	// Stop FTP Server (global)
	err := center.GetCenter().StopFTPServer()

	return &pb.StopServiceResponse{
		Success: err == nil,
	}, err
}

// RestartService 重启服务
func (s *ContainerServer) RestartService(ctx context.Context, req *pb.RestartServiceRequest) (*pb.RestartServiceResponse, error) {
	// TODO: 实现实现 FTP 和 远程桌面的重启
	return &pb.RestartServiceResponse{
		Success: true,
	}, nil
}

// RemoveService 移除服务
func (s *ContainerServer) RemoveService(ctx context.Context, req *pb.RemoveServiceRequest) (*pb.RemoveServiceResponse, error) {
	slog.Info("RemoveService request received", "container_id", req.ContainerId, "agent_id", req.AgentId)

	id := buildAgentKey(req.ContainerId, req.AgentId)

	// Stop VNC first
	_ = center.GetCenter().StopVDisplay(&center.VDisplayParams{
		Key: id,
	})

	err := center.GetCenter().RemoveAgent(id)
	if err != nil {
		slog.Error("Failed to remove agent", "error", err)
		return nil, err
	}

	slog.Info("Agent removed successfully", "container_id", req.ContainerId, "agent_id", req.AgentId)
	return &pb.RemoveServiceResponse{
		Success: true,
	}, nil
}

// BackupService 备份服务
func (s *ContainerServer) BackupService(ctx context.Context, req *pb.BackupServiceRequest) (*pb.BackupServiceResponse, error) {
	// TODO: 实现备份容器逻辑
	return &pb.BackupServiceResponse{
		BackupUrl: "http://backup-server/container-" + req.ContainerId + "/agent-" + req.AgentId + ".tar.gz",
	}, nil
}

func buildAgentKey(containerID, agentID string) string {
	return containerID + "/" + agentID
}
