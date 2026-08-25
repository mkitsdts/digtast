package grpc

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

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
	if os.Getenv("TEST_GRPC_INTEGRATION") == "" {
		os.Exit(0)
	}
	// 读取配置
	confData, err := os.ReadFile("config.json")
	if err != nil {
		panic("failed to read config.json: " + err.Error())
	}
	if err := json.Unmarshal(confData, &cfg); err != nil {
		panic("failed to unmarshal config.json: " + err.Error())
	}

	// 启动服务器
	// 假设我们在测试中启动本地服务器，或者连接到配置中的地址
	// 如果配置中的地址是远程的，我们可以跳过本地启动
	// 等待服务器启动
	time.Sleep(1 * time.Second)

	// 创建客户端连接
	conn, err = grpc.NewClient(cfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("did not connect: " + err.Error())
	}
	defer conn.Close()
	client = pb.NewContainerServiceClient(conn)

	// 运行测试
	os.Exit(m.Run())
}

func getContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	context.AfterFunc(ctx, cancel)
	return ctx
}
