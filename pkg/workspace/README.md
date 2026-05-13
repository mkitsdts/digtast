#

##

这是管理工作区的代码。

工作区默认位置是 ~/.digtast

运行时会创建目录

- agent
|   |_____ agent.md
|   |_____ soul.md
|   |_____ user.md
|   |_____ rule.md
|
| memory
|   |_____ sessions
|   |         |_____ {agent-name}
|   |                      |_____ {session_id:chunk}.jsonl ...
|   |                      |_____ {session_id:chunk-compress}.jsonl ...
|   |                      |_____ memory.md
|   |_____ memory.json (map agent session to real path)
| workspace (save generate file)
|
| skills (save personal skill)
|   |_____ {skill}
|             |_____ skill.md ...
|
| config.json (save config file)
