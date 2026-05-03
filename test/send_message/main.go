package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	pb "digital-labor/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	configPath := findConfigPath(os.Args[1:])
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("LLM_API_KEY")
	}

	config := flag.String("config", configPath, "JSON config file path")
	addr := flag.String("addr", cfg.Addr, "gRPC server address")
	containerID := flag.String("container", cfg.ContainerID, "container_id")
	agentID := flag.String("agent", cfg.AgentID, "agent_id")
	message := flag.String("message", cfg.Message, "message to send")
	timeoutText := flag.String("timeout", cfg.Timeout, "request timeout, e.g. 2m or 30s")

	provider := flag.String("provider", cfg.Provider, "model provider for StartService, e.g. deepseek/qwen/doubao")
	apiKey := flag.String("key", cfg.APIKey, "model API key for StartService; defaults to config api_key or LLM_API_KEY")
	baseURL := flag.String("url", cfg.BaseURL, "model base URL for StartService")
	modelName := flag.String("model", cfg.ModelName, "model name for StartService")
	flag.Parse()

	timeout, err := time.ParseDuration(*timeoutText)
	if err != nil {
		log.Fatalf("invalid --timeout: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Println("config:", *config)

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("create grpc client: %v", err)
	}
	defer conn.Close()

	client := pb.NewContainerServiceClient(conn)

	if err := registerAgent(ctx, client, serviceConfig{
		containerID: *containerID,
		agentID:     *agentID,
		provider:    *provider,
		apiKey:      *apiKey,
		baseURL:     *baseURL,
		modelName:   *modelName,
	}); err != nil {
		log.Fatal(err)
	}

	reply, err := sendMessage(ctx, client, messageConfig{
		containerID: *containerID,
		agentID:     *agentID,
		message:     *message,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("SendMessage response:")
	fmt.Println(reply)
}

type appConfig struct {
	Addr        string `json:"addr"`
	ContainerID string `json:"container_id"`
	AgentID     string `json:"agent_id"`
	Message     string `json:"message"`
	Timeout     string `json:"timeout"`
	Provider    string `json:"provider"`
	APIKey      string `json:"api_key"`
	BaseURL     string `json:"base_url"`
	ModelName   string `json:"model_name"`
}

type serviceConfig struct {
	containerID string
	agentID     string
	provider    string
	apiKey      string
	baseURL     string
	modelName   string
}

type messageConfig struct {
	containerID string
	agentID     string
	message     string
}

func defaultConfig() appConfig {
	return appConfig{
		Addr:        "127.0.0.1:10086",
		ContainerID: "local-container",
		AgentID:     "local-agent",
		Message:     "你好",
		Timeout:     "2m",
	}
}

var agentId string

func findConfigPath(args []string) string {
	for i, arg := range args {
		if arg == "-config" || arg == "--config" {
			if i+1 < len(args) {
				return args[i+1]
			}
			return "test/send_message/config.json"
		}

		for _, prefix := range []string{"-config=", "--config="} {
			if strings.HasPrefix(arg, prefix) {
				return strings.TrimPrefix(arg, prefix)
			}
		}
	}

	if _, err := os.Stat("test/send_message/config.json"); err == nil {
		return "test/send_message/config.json"
	}
	return "config.json"
}

func loadConfig(path string) (appConfig, error) {
	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return appConfig{}, fmt.Errorf("read config %q: %w", path, err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return appConfig{}, fmt.Errorf("parse config %q: %w", path, err)
	}

	return cfg, nil
}

func registerAgent(ctx context.Context, client pb.ContainerServiceClient, cfg serviceConfig) error {
	if cfg.provider == "" || cfg.apiKey == "" || cfg.modelName == "" {
		return errors.New("--provider, --model, and --key or LLM_API_KEY are required to register the agent")
	}

	resp1, err := client.CreateChatModel(ctx, &pb.CreateChatModelRequest{
		Provider:  cfg.provider,
		Key:       cfg.apiKey,
		BaseUrl:   cfg.baseURL,
		ModelName: cfg.modelName,
	})
	if err != nil {
		return fmt.Errorf("CreateChatModel failed: %w", err)
	}
	if !resp1.GetSuccess() {
		return errors.New("CreateChatModel returned success=false")
	}

	resp2, err := client.CreateAgent(ctx, &pb.CreateAgentRequest{
		AgentName:   "114514",
		ChatModelId: resp1.GetChatModelId(),
	})
	if err != nil {
		return fmt.Errorf("CreateAgent failed: %w", err)
	}
	if !resp2.GetSuccess() {
		return errors.New("CreateAgent returned success=false")
	}
	agentId = resp2.GetAgentId()

	fmt.Println("StartService: success")
	return nil
}

func sendMessage(ctx context.Context, client pb.ContainerServiceClient, cfg messageConfig) (string, error) {
	respStream, err := client.SendMessage(ctx, &pb.SendMessageRequest{
		AgentId:  agentId,
		Message:  cfg.message,
		IsStream: false,
	})
	if err != nil {
		return "", fmt.Errorf("SendMessage failed: %w", err)
	}

	var reply strings.Builder
	for {
		resp, err := respStream.Recv()
		if errors.Is(err, io.EOF) {
			return reply.String(), nil
		}
		if err != nil {
			return "", fmt.Errorf("receive SendMessage response: %w", err)
		}

		reply.WriteString(resp.GetDeltaContent())
	}
}
