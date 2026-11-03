# Digtast - Digital Labor Agent

Go 实现的单实例智能体程序，使用自研 AgentLoop，支持多种 LLM 后端。

## 项目结构

```text
internal/agent/ 唯一 Agent 的实现
  loop/         AgentLoop
  memory/       会话内存
  state/        提示词状态和长期记忆
  task/         单 Agent 任务追踪
  skill/        Skill 加载
  mcp/          MCP 工具加载
  registry/     工具和 Hook 注册
  workspace/    Agent 工作区持久化

cmd/    CLI 和程序入口
config/         配置文件解析与保存
gateway/        外部消息渠道实现
pkg/chatmodel/  ChatModel 抽象及模型适配器
pkg/            可复用功能包
prompt/         提示词组件
tools/          内置工具
```

## 构建

```bash
make build        # 编译到 bin/digtast
make test         # go test ./...
```

等价的手动命令：

```bash
go build -o bin/digtast ./cmd
go test ./...
```

Go 版本：`1.26.1`（见 `go.mod`）。

## 架构约束

- 进程内只允许一个 Agent，通过 `agent.Initialize` 初始化、`agent.Instance` 获取。
- 不要引入 Agent Manager、Agent ID、Agent 列表或多 Agent 路由。
- 不提供 gRPC Server、Proto/Labor API、Center 或 FTP 服务。
- Agent 私有实现放入 `internal/agent/`，不要放入顶层目录或 `pkg/`。
- `pkg/chatmodel` 只负责模型抽象和适配器；其他 `pkg` 包必须可复用且不依赖具体 Agent 生命周期。
- 配置定义和 JSON 解析统一放在顶层 `config/`。
- 可执行入口和 CLI 命令统一放在 `cmd/`。

## 代码规范

- 写简洁地道的 Go，小函数、显式数据流、清晰校验。
- 导出名使用 `CamelCase`，包内名称使用 `camelCase`，包名保持简短。
- 日志统一使用 `log/slog`，禁止使用 `fmt.Println` 或旧 `log` 包记录运行日志；CLI 输出除外。
- API 边界返回显式 error，禁止 nil panic 和硬编码凭据。
- 注释只说明不明显的行为或约束。

## 沙箱机制

`pkg/sandbox` 提供 Docker 沙箱隔离：

- `sandbox.WithSandbox()` 创建容器并 bind mount 工作区。
- `Agent.Rebuild(ModeSandbox)` 切换到沙箱模式。
- `Agent.Rebuild(ModeNormal)` 切回普通模式。
- 容器生命周期跟随唯一 Agent，`Close()` 时销毁。
