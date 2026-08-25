package subagent

import (
	"digital-labor/pkg/llm"
)

type Config struct {
	ParentID        string
	ChatModel       llm.LLM
	Messages        []llm.Message
	TaskDescription string
}

type Result struct {
	Summary      string
	Content      string
	MultiContent [][]byte
	Error        error
	Success      bool
}
