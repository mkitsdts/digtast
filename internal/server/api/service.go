package api

import (
	"context"
	"digital-labor/internal/center"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/model"
	pb "digital-labor/proto"
	"errors"
	"log/slog"
)

const (
	default_ftp_port int = 2121
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
	}

	ftp_port := -1
	if req.FtpEnabled {
		if conf.Conf.FTP.Port != 0 {
			ftp_port = conf.Conf.FTP.Port
		} else {
			ftp_port = default_ftp_port
		}
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
		}
	}

	return &pb.StartServiceResponse{
		VisualDisplayPort: int32(port),
		FtpPort:           int32(ftp_port),
		Success:           err == nil,
	}, err
}

// StopService 暂停服务
func (s *ContainerServer) StopService(ctx context.Context, req *pb.StopServiceRequest) (*pb.StopServiceResponse, error) {
	slog.Info("StopService request received", "container_id", req.ContainerId, "agent_id", req.AgentId)
	id := buildAgentKey(req.ContainerId, req.AgentId)

	var errs error
	// Stop VNC if it was running
	if err := center.GetCenter().StopVDisplay(&center.VDisplayParams{
		Key: id,
	}); err != nil {
		slog.Error("Failed to stop visual display")
		errs = errors.New(err.Error())
	}

	// Stop FTP Server (global)
	if err := center.GetCenter().StopFTPServer(); err != nil {
		slog.Error("Failed to stop FTP server", "error", err)
		errs = errors.New(errs.Error() + err.Error())
	}

	return &pb.StopServiceResponse{
		Success: errs == nil,
	}, errs
}

// BackupService 备份服务
func (s *ContainerServer) BackupService(ctx context.Context, req *pb.BackupServiceRequest) (*pb.BackupServiceResponse, error) {
	
	
	return &pb.BackupServiceResponse{
		BackupUrl: "http://backup-server/container-" + req.ContainerId + "/agent-" + req.AgentId + ".tar.gz",
	}, nil
}

func buildAgentKey(containerID, agentID string) string {
	return containerID + "/" + agentID
}
