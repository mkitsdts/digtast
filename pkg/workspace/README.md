#

##

这是管理工作区的代码。

工作区默认位置是 ~/.digtast

运行时会创建目录

- agent
|   |_____ agent.md
|   |_____ memory.md
|   |_____ soul.md
|   |_____ user.md
|
| memory
|   |_____ sessions
|   |         |_____ {agent-name}
|   |         |           |_____ {session_id:chunk}.jsonl ...
|   |_____ extract
|   |         |_____ {date}.jsonl ...
|   |         |_____ memory.db (save memory chunk by sqlite3,be used to RAG)
|   |_____ memory.json (map agent session to real path)
| workspace (save generate file)
|
| skills (save personal skill)
|   |_____ {skill}
|             |_____ skill.md ...
|
| config.json (save config file)
