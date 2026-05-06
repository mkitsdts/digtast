package grpc

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"digital-labor/internal/server"
	pb "digital-labor/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Addr        string `json:"addr"`
	ContainerID string `json:"container_id"`
	AgentID     string `json:"agent_id"`
	SessionID   string `json:"session_id"`
	Message     string `json:"message"`
	Timeout     string `json:"timeout"`
	Provider    string `json:"provider"`
	APIKey      string `json:"api_key"`
	BaseURL     string `json:"base_url"`
	ModelName   string `json:"model_name"`
}

var (
	client pb.ContainerServiceClient
	conn   *grpc.ClientConn
	cfg    Config
)

func TestMain(m *testing.M) {
	// 读取配置
	confData, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("failed to read config.json: %v", err)
	}
	if err := json.Unmarshal(confData, &cfg); err != nil {
		log.Fatalf("failed to unmarshal config.json: %v", err)
	}

	// 启动服务器
	// 假设我们在测试中启动本地服务器，或者连接到配置中的地址
	// 如果配置中的地址是远程的，我们可以跳过本地启动
	go func() {
		server.Start(":" + "10086") // 强制使用 10086 或者从 cfg.Addr 提取端口
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)

	// 创建客户端连接
	conn, err = grpc.Dial(cfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client = pb.NewContainerServiceClient(conn)

	// 运行测试
	os.Exit(m.Run())
}

func getContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Minute)
	return ctx
}
