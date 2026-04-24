package server

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"digital-labor/internal/center"
	"digital-labor/internal/server/api"
	pb "digital-labor/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// serverImpl 包装了 internal/server/api 里的 ContainerServer
// 从而补充 grpc 接口的实现
type serverImpl struct {
	*api.ContainerServer
}

func newServer() *serverImpl {
	return &serverImpl{
		ContainerServer: &api.ContainerServer{},
	}
}

// StartService 启动服务 (补充实现)
func (s *serverImpl) StartService(ctx context.Context, req *pb.StartServiceRequest) (*pb.StartServiceResponse, error) {
	slog.Info("StartService request received", "container_id", req.ContainerId, "agent_id", req.AgentId)

	_, err := center.GetCenter().Create(req.ContainerId, req.AgentId, req.Provider, req.Key, req.Url, req.ModelName)
	if err != nil {
		slog.Error("Failed to create agent", "error", err)
		return nil, err
	}

	slog.Info("Agent created successfully", "container_id", req.ContainerId, "agent_id", req.AgentId)
	return &pb.StartServiceResponse{
		Success: true,
	}, nil
}

// RemoveService 移除服务 (补充实现)
func (s *serverImpl) RemoveService(ctx context.Context, req *pb.RemoveServiceRequest) (*pb.RemoveServiceResponse, error) {
	slog.Info("RemoveService request received", "container_id", req.ContainerId, "agent_id", req.AgentId)

	err := center.GetCenter().Remove(req.ContainerId, req.AgentId)
	if err != nil {
		slog.Error("Failed to remove agent", "error", err)
		return nil, err
	}

	slog.Info("Agent removed successfully", "container_id", req.ContainerId, "agent_id", req.AgentId)
	return &pb.RemoveServiceResponse{
		Success: true,
	}, nil
}

// CompressSession 压缩会话 (补充实现)
func (s *serverImpl) CompressSession(ctx context.Context, req *pb.CompressSessionRequest) (*pb.CompressSessionResponse, error) {
	slog.Info("CompressSession request received", "session_id", req.SessionId)
	return &pb.CompressSessionResponse{
		SessionId: req.SessionId,
	}, nil
}

// Start 初始化并启动 gRPC 服务器
func Start(port string) {
	// 初始化 slog
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	// 创建 gRPC server
	grpcServer := grpc.NewServer()

	// 注册服务
	srv := newServer()
	pb.RegisterContainerServiceServer(grpcServer, srv)
	
	// 注册反射服务，方便调试
	reflection.Register(grpcServer)

	// 启动服务并处理优雅退出
	go func() {
		slog.Info("server listening at", "address", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("failed to serve", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	slog.Info("shutting down server...")
	grpcServer.GracefulStop()
	slog.Info("server stopped")
}
