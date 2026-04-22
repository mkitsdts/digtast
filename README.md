# 数字助理容器

数字助理容器需要与核心中枢保持连接，目前提供的接口有

## 容器相关
网关层次，只负责 Agent 服务的启动和暂停

StopService       暂停服务
RestartService    重启服务
RemoveService     移除服务
BackupService     备份服务

## 会话相关
会话仅负责维护上下文，仅此而已。

RemoveSession         删除会话
ResumeSession         恢复会话

## 对话相关
仅仅发送对话

SendMessageToSession  会话里发起对话

## 任务相关
任务管理器，每一次对话都抽象成一次任务。

StopTask      暂停任务
RestartTask    恢复任务
RemoveTask    移除任务

## Skill相关
CreateSkill   在容器里创建 Skill
DisableSkill  在容器里禁用 Skill
RemoveSkill   在容器里移除 Skill

## Tool相关
CreateTool  在容器里创建 Tool
DisableTool 在容器里禁用 Tool
RemoveTool  在容器里移除 Tool

---
数字助理总共有四部分，加上网关总共有五部分：
1. 会话：负责管理会话
2. 记忆：负责整理对话内容
3. 执行：包含工具管理器，MCP管理器，Skill管理器
4. 任务：负责记录任务状态，驱动执行层
5. 网关：解决与主节点的通信传输
