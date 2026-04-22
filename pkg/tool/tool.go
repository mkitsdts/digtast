package tool

import "github.com/cloudwego/eino/components/tool"

var tools = make([]tool.BaseTool, 0)

func GetTools() []tool.BaseTool {
	return tools
}

func RegisterTool(tool tool.BaseTool) {
	tools = append(tools, tool)
}
