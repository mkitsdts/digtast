package server

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

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
