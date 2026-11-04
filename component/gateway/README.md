# component/gateway 网关抽象

定义外部消息渠道的抽象：

- `MessageChannel`：消息渠道接口（Send / Serve / Init / 注册 / 配置）。
- `Result`：渠道回传给用户的结果结构。
- `UserMessage`：网关投递给 Agent 的用户消息结构。
- 通道注册表（`RegisterChannel` / `NewChannel` / `GetAllChannels`）。
- 用户消息队列（`PushUserMessage` / `PopUserMessage`）。

具体网关实现（QQ、Telegram 等）放在顶层 `gateway/` 目录，只依赖本包；
通道的生命周期管理和消息消费循环放在 `gateway/` 根包。
