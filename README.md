# 数字助理容器

digtast 是一个数字助理，可以单独在宿主机工作，也可以接入核心中枢，在云端工作。

为了与核心中枢保持连接，提供有接口

## 容器相关
网关层次，负责服务的启动和暂停

StartService      启动服务，注册 APIKey，并选择是否启用远程桌面
StopService       暂停服务，关闭远程桌面
RestartService    重启服务，打开远程桌面
RemoveService     移除服务，删除所有内容
BackupService     备份服务，备份记忆

## 会话相关
会话仅负责维护上下文

RemoveSession         删除会话
ResumeSession         恢复会话

## 对话相关
仅仅发送对话

SendMessageToSession  会话里发起对话

发起对话需要的参数：
- sessionid：会话id，用于标记历史会话记录。如果没有则自动创建，下次带上
- agentid: 调用的智能体id，用于区分不同助手
- content：用户请求会话内容

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
5. 网关：解决与主节点的通信传输，启动桌面传输
