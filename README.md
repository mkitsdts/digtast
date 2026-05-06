# Digtast (数字助理容器)

`Digtast` 是一个数字助理容器服务。既可以作为独立的本地服务运行，也可以接入核心中枢（Center）在云端协同工作。

本项目目标是基于 gRPC 提供大量接口，允许外部程序（或主节点）对数字助理进行全生命周期管理，包括服务的启停、会话上下文的维护、大模型的对话驱动。但目前主要是开发本地服务，优先确保智能体的正常工作

## 🏗️ 系统架构

数字助理在内部主要划分为五个核心部分：

1. **网关 (Gateway)**：解决与主节点的通信传输，并负责管理视觉显示（如启动 VNC 、FTP）。
2. **会话 (Session)**：负责维护用户的对话上下文，保障多轮对话的连贯性。
3. **记忆 (Memory)**：负责整理并持久化对话内容，为智能体提供长期记忆支持。
4. **执行 (Execution)**：核心的行动层，包含工具管理器 (Tool)、MCP 管理器和技能管理器 (Skill)。
5. **任务 (Task)**：将每一次对话或复杂请求抽象为一次任务，记录任务状态并驱动执行层进行工作。

---

## 🚀 快速开发

### 1. 环境准备
- 需要安装 **Go 1.22+** 环境。
- (可选) 推荐使用具有 gRPC 调试功能的客户端（如 Postman、Evans 或 grpcurl）进行接口测试。

### 2. 编译与启动

获取代码后，直接在项目根目录运行 `go build ` 即可完成编译，得到纯二进制文件。

```bash
./digast --local # 本地启动
./digast --port 10086 # 在 10086 端口启动 grpc 服务器，与远程中心配合
```

启动成功后，控制台会输出如下日志，表示服务正在等待请求：
```text
time=... level=INFO msg="server listening at" address=[::]:10086
```
*(注：服务支持响应 SIGINT/SIGTERM 信号进行 Graceful Stop)* 暂时不支持多模态

---

## 🔌 gRPC 接口指南

服务通过 gRPC 提供交互能力，所有的接口定义可以在 `proto/container.proto` 中查看。以下是核心接口的简要说明：

### 📦 1. 容器/服务级操作
负责应用实例和智能体的生命周期管理：
- `StartService`：启动服务。需要传入 `container_id`, `agent_id`。
- `StopService`：暂停服务。
- `RestartService`：重启服务，重新建立连接等。
- `RemoveService`：移除服务，将会销毁内部创建的智能体和对应的运行资源。
- `BackupService`：备份当前容器状态与记忆。

### 💬 2. 对话与会话管理
负责维护用户与智能体的交流与上下文：
- `SendMessageToSession`：**核心接口**。在指定的会话中发起对话。支持**流式 (Stream)** 返回打字机效果。
  - 参数包含：`session_id` (为空则自动创建并绑定)、`agent_id` 以及用户输入的 `message`。
- `CompressSession`：压缩历史会话上下文，降低 Token 消耗。
- `RemoveSession`：删除并清空指定会话的上下文记忆。

### 📝 3. 任务管理
复杂的对话过程会被抽象为任务进行调度：
- `StopTask`：暂停正在执行中的长任务。

###    4. API 管理
负责模型管理
- `CreateModel`: 核心接口，创建模型配置
- `RemoveChatModel`:删除模型

###    5. 智能体管理
负责智能体管理
- `CreateAgent`:核心接口，创建智能体。方便不同智能体启用不同的 Skill 和工具
- `RemoveAgent`:删除智能体

###    6. Skill 和 MCP 管理
负责 Skill 和 MCP 的启用与禁用。目前在考虑是否需要自定义安装 Skill。初步思考，用户应该通过对话创建 Skill。实在有需要，通过远程文件手动管理 Skill,不提供接口
- `GetAllSkills`:获取全部 Skill
- `DisableSkill`:禁用某个 Skill
- `EnableSkill`:启用某个 Skill
- `AddMCP`:添加 MCP
- `GetAllMCPs`:获取全部 MCP
- `DisableMCP`:禁用某个 MCP
- `EnableMCP`:启用某个 MCP

###    7. 通道管理
- `GetAllChannels`:获取全部通道
- `CreateChannel`:创建通道
- `RemoveChannel`:删除通道
- `EnableChannel`:启用通道
- `DisableChannel`:禁用通道

## 🔌 终端指南

可通过终端在本地运行 digital ，后续会通过命令行参数区分启用 grpc 接口或终端。

终端可发送命令和显示输出，其余内容一样

---

## 🛠️ 功能开发清单 (TODO List)

### 🧠 记忆与上下文管理
- [ X ] **完善持久化存储**：修复 `internal/memory` 中的消息合并逻辑，支持工具调用（Tool Calls）和结果的结构化存储。
- [ X ] **实现记忆压缩**：开发 `CompressSession` 逻辑，支持通过 LLM 总结摘要或固定窗口裁剪方式压缩历史上下文。
- [ X ] **长期记忆检索**

### 📋 任务管理系统
- [ X ] **任务中心**：建立独立的任务追踪模块，支持对异步执行的 Agent 任务进行编号和状态管理。
- [ X ] **状态查询接口**：实现 `GetTaskStatus`，允许用户实时查看 Agent 的行动轨迹（如：正在调用某工具、正在思考）。
- [ X ] **任务控制增强**：完善 `StopTask` 功能。

### 🔌 技能与扩展性
- [ X ] **动态工具加载**：实现 `CreateTool` 接口，支持动态为 Agent 挂载远程工具。
- [ X ] **技能包 (Skill) 支持**：定义 Skill 规范，允许将一组 Prompt + Tools 封装为特定技能（如：翻译专家、代码审计员）。
- [ ] **MCP 协议对接**：实现 Model Context Protocol (MCP)，支持挂载标准化的外部上下文服务器。

###    通道内容交互
- [ X ] **QQ**
- [ X ] **Telegram**
