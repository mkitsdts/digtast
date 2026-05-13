# Digtast - Digital Labor Agent

Go 实现的智能体服务，基于 CloudWeGo eino ADK 框架，支持多种 LLM 后端。

## 项目结构

```
internal/
  agent/       智能体核心：创建、对话、会话管理、提示词组装
  center/      活跃智能体注册中心
  cli/         CLI 入口
  gateway/     外部服务集成（飞书、Telegram 等）
  server/      gRPC 服务端

pkg/
  chatmodel/   LLM 模型管理
  conf/        配置加载
  ctxmanager/  上下文管理
  memory/      会话持久化
  middleware/
    lbackend/  本地文件系统后端（filesystem.Backend 实现）
    safetool/  工具调用错误兜底
    skill/     技能中间件
  model/       数据模型定义
  registry/    工具和处理器全局注册
  sandbox/     Docker 沙箱隔离
  state/       状态管理与压缩
  workspace/   工作区初始化与持久化

tools/         内置工具（ask_user, browser_use, subagent, web_search, web_fetch, memory）
prompt/        提示词模板
proto/         gRPC 生成代码（*.pb.go, *_grpc.pb.go，勿手动编辑）
labor/v1/      Proto 源文件
```

## 构建

```bash
go build ./...
```

Go 版本：`1.26.1`（见 go.mod）

## 代码规范

- 写简洁地道的 Go，小函数、显式数据流、清晰校验
- 导出名 `CamelCase`，包内 `camelCase`，包名简短
- 日志统一用 `log/slog`，禁止 `fmt.Println` / `log` / `logrus`
- API 和 manager 边界必须返回显式 error，禁止 nil panic 和硬编码凭据
- 注释只写不明显的行为或约束，不写显而易见的内容

## Proto 变更

修改 `labor/v1/container.proto` 后用 `buf generate` 重新生成。未实现的 RPC 返回 `codes.Unimplemented`，不要返回假的成功响应。

## 沙箱机制

`pkg/sandbox` 提供 Docker 沙箱隔离：
- `sandbox.New()` 创建容器，bind mount 工作区
- `sandbox.WithSandbox()` 包装 `filesystem.Backend`，所有操作（文件读写、命令执行）都在容器内完成
- 通过 `RebuildAgent(ModeSandbox)` 切换到沙箱模式，`RebuildAgent(ModeNormal)` 切回普通模式
- 容器生命周期跟随智能体，`Close()` 时销毁

## 当前状态

早期开发阶段，接口频繁变更。主要已知问题：
- Proto 生成路径和包名不完全一致，提交前需检查
- 部分外部集成（飞书、Telegram）仍是 TODO 占位
- 会话生命周期管理分散在 ctxmanager 和 DigitalAgent 之间，待收敛
