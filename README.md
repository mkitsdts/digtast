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

默认配置位于 `~/.digtast/config.json`。首次启动会直接进入添加模型的表单；之后启动会恢复上次选择的模型，随时可以在对话框里用 `/model` 切换。Agent 名称由程序自动生成 UUID，无需手动配置。

## 使用

```bash
./bin/digtast             # 打开终端对话框
./bin/digtast --browser   # 同时启用浏览器模式
```

对话框内置命令（输入 `/` 会有提示，不需要记子命令）：

| 命令 | 作用 |
| --- | --- |
| `/model` | 选择模型；添加、编辑、删除模型配置 |
| `/channel` | 配置 QQ / Telegram 渠道，启用、停用、删除 |
| `/new` | 清空历史，开始新会话 |
| `/help` | 查看快捷键和命令 |
| `/exit` | 退出 |

常用按键：`Enter` 发送/确认，`Tab` 补全，`↑/↓` 选择，`Space` 多选，`Esc` 取消或返回，`Ctrl+C` 退出。

添加模型不需要手动输入 Provider 或模型名：先从内置列表选择 Provider（OpenAI、DeepSeek、通义千问、豆包/方舟、自定义），填写 API Key 后，程序会调用 Provider 的 `/models` 接口拉取可用模型，再从列表里选择（支持多选）。只需输入一次 API Key，之后在 `/model` 列表里直接切换模型。

模型配置以 UUID 为键保存，无需手动命名，例如：

```json
{
  "agent": {
    "name": "6f1c0d7e-6a1c-4f6e-9d2b-1a2b3c4d5e6f",
    "model": "8b7a9c0d-1e2f-4a3b-9c8d-7e6f5a4b3c2d",
    "model_kind": "deepseek-chat"
  },
  "model": {
    "8b7a9c0d-1e2f-4a3b-9c8d-7e6f5a4b3c2d": {
      "provider": "deepseek",
      "model_names": ["deepseek-chat", "deepseek-reasoner"],
      "url": "",
      "key": "sk-..."
    }
  }
}
```

使用 `--browser` 可启用浏览器模式。浏览器工具是懒加载的，只有模型真正调用 `browser_use` 时才会启动浏览器：

```bash
./bin/digtast --browser
```

## 验证

```bash
make test
make vet
```
