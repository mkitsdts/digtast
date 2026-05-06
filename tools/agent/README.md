#

##

这个是生成子 agent 的模块。调用需要的参数有

- `description`: 任务的简短描述（3-5 个单词）
- `prompt`: 要执行的任务
- `subagent_type`: 子 agent 的类型（必选）：区分可以调用的工具类型。子 agent 有： research, execute 等
- `run_in_background`: 是否在后台运行（可选）
