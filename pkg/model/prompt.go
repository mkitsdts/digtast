package model

import "github.com/cloudwego/eino/components/tool"

type PromptContext struct {
	UserInstruction string
	UserPreference  string
	Skills          []string
	Tools           []tool.BaseTool
}
