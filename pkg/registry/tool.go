package registry

import "github.com/cloudwego/eino/components/tool"

func RegisterTool(t tool.BaseTool) {
	if t == nil {
		return
	}
	toolMu.Lock()
	defer toolMu.Unlock()
	tools = append(tools, t)
}

func GetTools() []tool.BaseTool {
	toolMu.RLock()
	defer toolMu.RUnlock()

	result := make([]tool.BaseTool, len(tools))
	copy(result, tools)
	return result
}
