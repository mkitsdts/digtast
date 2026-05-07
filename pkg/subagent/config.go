package subagent

import (
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type Config struct {
	ParentID        string
	ChatModel       model.ToolCallingChatModel
	Messages        []*schema.Message
	TaskDescription string
}

type Result struct {
	Summary      string
	Content      string
	MultiContent [][]byte
	Error        error
	Success      bool
}
