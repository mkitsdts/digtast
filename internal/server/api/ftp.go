package api

import (
	"context"
	"digital-labor/internal/center"
	pb "digital-labor/proto"
)

func (s *ContainerServer) StartFTPServer(ctx context.Context, req *pb.StartFTPServerRequest) (*pb.StartFTPServerResponse, error) {
	if center.FtpServer == nil {
		return &pb.StartFTPServerResponse{
			Success: false,
			Message: "FTP server not configured",
		}, nil
	}

	if err := center.FtpServer.Start(); err != nil {
		return &pb.StartFTPServerResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.StartFTPServerResponse{
		Success: true,
		FtpPort: int32(center.FtpServer.Port),
	}, nil
}

func (s *ContainerServer) StopFTPServer(ctx context.Context, req *pb.StopFTPServerRequest) (*pb.StopFTPServerResponse, error) {
	if center.FtpServer == nil {
		return &pb.StopFTPServerResponse{
			Success: false,
			Message: "FTP server not configured",
		}, nil
	}

	if err := center.FtpServer.Stop(); err != nil {
		return &pb.StopFTPServerResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.StopFTPServerResponse{
		Success: true,
	}, nil
}
