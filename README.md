# Digtast

Digtast 是一个使用 Go 实现的单实例智能体程序。进程内只会初始化一个 `Agent`，通过终端对话框（TUI）或已配置的消息网关与用户交互。

## 目录

```text
schema/          数据定义（消息、流式事件迭代器、工具契约）
component/       组件抽象（prompt、tool、chatmodel、gateway、middleware）
chatmodel/       模型适配器，按 Provider 分包（openai、deepseek）
internal/agent/  Agent 运行时及其私有实现（loop、memory、state、task、skill、MCP、workspace）
internal/sandbox/ Docker 沙箱后端
cmd/             程序入口和终端对话框（TUI）
config/          配置结构、加载与保存
gateway/         QQ、Telegram 等网关实现
middleware/      执行后端实现（lbackend 本地后端）
prompts/         系统提示词
tools/           内置工具实现
pkg/             可复用功能包（queue、ctxmanager）
```

详细的开发约束见 [AGENTS.md](AGENTS.md)，TUI 扩展指南见 [cmd/README.md](cmd/README.md)。

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

默认配置位于 `~/.digtast/config.json`。

## 使用

```bash
./bin/digtast             # 打开终端对话框
./bin/digtast --browser   # 同时启用浏览器模式
```

对话框内置命令（输入 `/` 会有提示，不需要记子命令）：

| 命令 | 作用 |
| --- | --- |
| `/model` | 配置厂商（API Key、可选 Base URL），从 `/models` 拉取列表启用模型，日常切换模型 |
| `/skills` | 查看扫描到的 skills，启用或停用 |
| `/channel` | 配置 QQ / Telegram 渠道，启用、停用、删除 |
| `/new` | 清空历史，开始新会话 |
| `/help` | 查看快捷键和命令 |
| `/exit` | 退出 |

常用按键：`Enter` 发送/确认，`Tab` 补全，`↑/↓` 选择，`Space` 多选/启停，`Esc` 取消或返回，`Ctrl+C` 退出。

### 模型配置

厂商在代码中注册（`chatmodel/deepseek`、`chatmodel/openai`），界面里配置的是厂商：

1. 选择厂商（如 DeepSeek），填写 API Key；
2. 询问是否替换默认 Base URL，不需要就直接用厂商默认地址；
3. 程序调用厂商 `/models` 接口拉取模型，`Space` 多选要启用的模型；
4. 之后在 `/model` 列表里随时切换启用的模型，按 `a` 重新拉取列表。

配置示例（按厂商为键保存凭据与启用列表）：

```json
{
  "agent": { "name": "6f1c...", "provider": "deepseek", "model": "deepseek-chat" },
  "model": {
    "deepseek": {
      "key": "sk-...",
      "models": ["deepseek-chat", "deepseek-reasoner"]
    }
  }
}
```

### Skills

把 skill 放在 `~/.digtast/skills/<name>/` 下（目录内含 `SKILL.md`，可选 `description.txt`），启动时自动扫描。未停用的 skill 名称与描述会注入提示词，模型通过 `load_skill` 工具按需加载完整内容；用 `/skills` 可以随时启停，状态记录在配置文件的 `skills.disabled` 中，停用的 skill 不注入上下文。

使用 `--browser` 可启用浏览器模式。浏览器工具是懒加载的，只有模型真正调用 `browser_use` 时才会启动浏览器。

## 验证

```bash
make test
make vet
```
