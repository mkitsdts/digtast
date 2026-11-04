# Digtast - Digital Labor Agent

Go 实现的单实例智能体程序，使用自研 AgentLoop，支持多种 LLM 后端。

## 项目结构

```text
schema/          具体数据定义（对话消息、非流式 ChatRequest/Response、流式 Stream 迭代器、ToolDefinition）
component/       各类组件的抽象定义（接口与数据契约）
  prompt/        提示词 Loader 抽象
  tool/          Tool、Tool Loader、Registry 抽象
  chatmodel/     ChatModel/StreamChatModel 抽象、模型注册表
    oai/         OpenAI 兼容协议公共底座（HTTP、SSE 装配、编解码、用量）
  gateway/       MessageChannel 抽象、通道注册表、用户消息队列
  middleware/    执行后端抽象（Backend、StreamingShell）

prompts/         提示词具体实现
tools/           内置工具具体实现
chatmodel/       模型适配器具体实现，按 Provider 分包（openai、deepseek），各自在 init 中注册厂商
gateway/         外部消息渠道具体实现（QQ、Telegram）及通道管理/消息消费
middleware/      执行后端具体实现（lbackend 本地后端）

internal/agent/  唯一 Agent 的实现
  loop/          AgentLoop
  memory/        会话内存
  state/         提示词状态和长期记忆
  task/          单 Agent 任务追踪
  skill/         Skill 加载
  mcp/           MCP 工具加载
  registry/      工具和 Hook 注册
  workspace/     Agent 工作区持久化
internal/sandbox/ Docker 沙箱后端

cmd/             CLI 和程序入口
config/          配置文件解析与保存
pkg/             可复用功能包（queue、ctxmanager）
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
- 分层规则：`schema/` 只放数据定义；`component/` 只放抽象定义，不依赖 `config/`、`internal/` 和顶层实现包；具体实现放顶层 `prompts/`、`tools/`、`chatmodel/`、`gateway/`、`middleware/`。
- 模型适配器按 Provider 分包：公共协议实现放在 `component/chatmodel/oai`，各 Provider 组合它并用自己的默认基础配置与线格式差异（如 DeepSeek 的 `reasoning_content`、`max_tokens`）。不提供「一个 openai 客户端 + 多 Provider 预设」的模式。
- 厂商通过 `chatmodel.RegisterProvider` 在 init 中注册（与 prompts 相同的注册模式）。TUI 配置厂商凭据（API Key、可选替换 Base URL），从 /models 拉取列表选择启用的模型；凭据与启用列表按 Provider 保存在 `config.Conf.Models`。
- 流式对话使用迭代器 `schema.Stream`（`iter.Seq2`），不要新增 `Recv`/`Close` 式的拉取接口。
- `pkg/` 包必须可复用且不依赖具体 Agent 生命周期。
- 配置定义和 JSON 解析统一放在顶层 `config/`。
- 可执行入口和 CLI 命令统一放在 `cmd/`。

## 代码规范

- 写简洁地道的 Go，小函数、显式数据流、清晰校验。
- 导出名使用 `CamelCase`，包内名称使用 `camelCase`，包名保持简短。
- 日志统一使用 `log/slog`，禁止使用 `fmt.Println` 或旧 `log` 包记录运行日志；CLI 输出除外。
- API 边界返回显式 error，禁止 nil panic 和硬编码凭据。
- 注释只说明不明显的行为或约束。

## 沙箱机制

`internal/sandbox` 提供 Docker 沙箱隔离（实现 `component/middleware` 的 Backend 抽象）：

- `sandbox.WithSandbox()` 创建容器并 bind mount 工作区。
- `Agent.Rebuild(ModeSandbox)` 切换到沙箱模式。
- `Agent.Rebuild(ModeNormal)` 切回普通模式。
- 容器生命周期跟随唯一 Agent，`Close()` 时销毁。
