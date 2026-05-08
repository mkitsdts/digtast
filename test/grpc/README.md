# gRPC 集成测试

## 前置条件

1. `go` 版本 >= 1.26.1
2. `test/grpc/config.json` 配置正确（gRPC 服务端地址、API Key 等）

## 配置文件

`config.json` 字段说明：

| 字段 | 说明 |
|------|------|
| `addr` | gRPC 服务端地址，测试会在此地址启动服务端并连接 |
| `api_key` | 模型提供商 API Key |
| `base_url` | 模型提供商 API 地址 |
| `model_name` | 使用的模型名称 |
| `provider` | 模型提供商（如 doubao） |
| `container_id` | 容器 ID（按需填写） |
| `agent_id` | 测试用 Agent ID（按需填写） |
| `session_id` | 测试用 Session ID（按需填写） |
| `timeout` | 超时时间（按需填写） |

## 运行测试

进入 `test/grpc` 目录执行：

```bash
# 运行所有 gRPC 测试
go test -v ./...

# 运行单个测试文件
go test -v -run TestAgentMethods

# 只运行 CreateAgent 子用例
go test -v -run TestAgentMethods/CreateAgent

# 只运行 agent_test.go 中的测试
go test -v -run TestAgent
```

## 测试流程

`TestMain` 在 `setup_test.go` 中自动完成以下工作：

1. 读取 `config.json`
2. 在 `:10086` 启动 gRPC 服务端
3. 创建 gRPC 客户端连接
4. 执行测试用例
5. 测试结束后关闭连接

无需手动启动服务，直接运行 `go test` 即可。
