# Digital Labor (数字助理容器)

`digital-labor` 是一个数字助理容器服务。既可以作为独立的本地服务运行，也可以接入核心中枢（Center）在云端协同工作。

本项目基于 gRPC 提供了丰富的接口，允许外部程序（或主节点）对数字助理进行全生命周期管理，包括服务的启停、会话上下文的维护、大模型的对话驱动，以及针对工具 (Tool)、技能 (Skill) 和 MCP 的动态挂载与管理。

## 🏗️ 系统架构

数字助理在内部主要划分为五个核心部分：

1. **网关 (Gateway)**：解决与主节点的通信传输，并负责管理视觉显示（如启动 VNC 、FTP 远程桌面传输）。
2. **会话 (Session)**：负责维护用户的对话上下文，保障多轮对话的连贯性。
3. **记忆 (Memory)**：负责整理并持久化对话内容，为智能体提供长期记忆支持。
4. **执行 (Execution)**：核心的行动层，包含工具管理器 (Tool)、MCP 管理器和技能管理器 (Skill)。
5. **任务 (Task)**：将每一次对话或复杂请求抽象为一次任务，记录任务状态并驱动执行层进行工作。

---

## 🚀 快速开始

### 1. 环境准备
- 需要安装 **Go 1.22+** 环境。
- (可选) 推荐使用具有 gRPC 调试功能的客户端（如 Postman、Evans 或 grpcurl）进行接口测试。

### 2. 编译与启动

获取代码后，直接在项目根目录运行 `main.go` 即可启动服务。

```bash
# 启动 gRPC 服务，默认监听 10086 端口
go run main.go
```

启动成功后，控制台会输出如下日志，表示服务正在等待请求：
```text
time=... level=INFO msg="server listening at" address=[::]:10086
```
*(注：服务支持响应 SIGINT/SIGTERM 信号进行优雅退出 Graceful Stop)*

---

## 🔌 gRPC 接口指南

服务通过 gRPC 提供交互能力，所有的接口定义可以在 `proto/container.proto` 中查看。以下是核心接口的简要说明：

### 📦 1. 容器/服务级操作 (Container)
负责应用实例和智能体的生命周期管理：
- `StartService`：启动服务。需要传入 `container_id`, `agent_id`，并注册模型的 `provider`, `key` 等，同时可以选择是否启用 VNC 远程桌面。
- `StopService`：暂停服务，例如关闭相关的远程桌面连接。
- `RestartService`：重启服务，重新建立连接等。
- `RemoveService`：移除服务，将会销毁内部创建的智能体和对应的运行资源。
- `BackupService`：备份当前容器状态与记忆。

### 💬 2. 对话与会话管理 (Dialogue & Session)
负责维护用户与智能体的交流与上下文：
- `SendMessageToSession`：**核心接口**。在指定的会话中发起对话。支持**流式 (Stream)** 返回打字机效果。
  - 参数包含：`session_id` (为空则自动创建并绑定)、`agent_id` 以及用户输入的 `message`。
- `CompressSession`：压缩历史会话上下文，降低 Token 消耗。
- `RemoveSession`：删除并清空指定会话的上下文记忆。

### 📝 3. 任务管理 (Task)
复杂的对话过程会被抽象为任务进行调度：
- `StopTask`：暂停正在执行中的长任务。
- `RestartTask`：恢复或重新启动被暂停的任务。
- `RemoveTask`：中止并移除任务。

### 🛠️ 4. 扩展能力层 (Skill, Tool, MCP)
数字助理支持动态扩展自身能力：
- **Skill 相关**：`CreateSkill`, `DisableSkill`, `RemoveSkill` (管理定制化技能)
- **Tool 相关**：`CreateTool`, `DisableTool`, `RemoveTool` (管理原子工具调用)
- **MCP 相关**：`CreateMCP`, `DisableMCP`, `RemoveMCP` (挂载与管理外部 MCP 协议扩展)

---

## 🛠️ 二次开发说明

本项目的核心逻辑存放在 `internal` 目录下：
- **`internal/server/server.go`**: gRPC 服务的入口包装与启动逻辑。在这里将原有的 API 进行了组装和补充。
- **`internal/server/api/`**: 包含各项 gRPC 接口的默认实现和占位 (TODOs)。
- **`internal/agent/`**: 封装了 `DigitalAgent`，通过 Eino 框架对接了包括 DeepSeek, Qwen, Doubao 等主流大模型。
- **`internal/center/`**: `Center` 维护了多实例、多智能体的全局状态。

请在 `internal/server/server.go` 或 `internal/server/api` 内进一步补充各个功能模块（如 FTP, VNC 等）的具体业务代码。
