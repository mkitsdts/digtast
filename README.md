# Digtast

Digtast 是一个使用 Go 实现的单实例智能体程序。进程内只会初始化一个 `Agent`，通过本地 CLI 或已配置的消息网关与用户交互。

## 目录

```text
internal/agent/ Agent 运行时及其私有实现（loop、memory、state、task、skill、MCP、workspace）
cmd/    程序入口和本地 CLI
config/         配置结构、加载与保存
gateway/        QQ、Telegram 等网关实现
pkg/chatmodel/  ChatModel 抽象和模型适配器
pkg/            可复用的文件系统、工具、沙箱、队列等功能包
prompt/         系统提示词组件
tools/          内置工具实现
```

项目不再包含多 Agent 管理、Center、gRPC Server、Proto/Labor API 或 FTP 服务。

## 构建与运行

需要 Go 1.26.1。

```bash
make build      # 生成 bin/digtast
./bin/digtast
```

也可以直接运行：

```bash
make run        # 等价于 go run ./cmd
```

默认配置位于 `~/.digtast/config.json`。首次启动后，可在 CLI 中使用 `/model add` 添加模型配置，再使用 `/new` 初始化唯一的 Agent。后续启动会从配置自动初始化该实例。

配置中的 Agent 结构如下：

```json
{
  "agent": {
    "name": "assistant",
    "model": "openai",
    "model_kind": "gpt-5",
    "description": "A helpful assistant"
  }
}
```

使用 `--browser` 可启用浏览器工具配置：

```bash
./bin/digtast --browser
```

## 验证

```bash
make test
make vet
```
