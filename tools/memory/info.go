package memory

import (
	"context"
	toolpkg "digital-labor/component/tool"
	"digital-labor/internal/agent"
	"digital-labor/schema"
	"encoding/json"
	"fmt"
	"strings"
)

type MemorySearchTool struct {
}

func (t *MemorySearchTool) Definition(context.Context) (schema.ToolDefinition, error) {
	return schema.ToolDefinition{
		Name:        "memory_search",
		Description: "Search the agent's conversation history and compressed memory.",
		Parameters: toolpkg.ObjectSchema(map[string]any{
			"query": map[string]any{"type": "string", "description": "Search query."},
		}, "query"),
	}, nil
}

type MemorySearchArguments struct {
	Query string `json:"query"`
}

func (t *MemorySearchTool) Execute(ctx context.Context, arguments json.RawMessage) (*toolpkg.Result, error) {
	args := MemorySearchArguments{}
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, err
	}

	da, _ := agent.Instance()

	// 搜索短期记忆 (Session history & compressed records)
	var sessionResults []string
	if da != nil {
		sess, err := da.GetSession()
		if err == nil {
			// 搜索压缩记录 (短期记忆)
			compressedMsgs, _ := sess.Search(args.Query)
			for _, m := range compressedMsgs {
				sessionResults = append(sessionResults, fmt.Sprintf("[%s]: %s", m.Role, m.Content))
			}

			// 搜索原始消息 (对话历史)
			msgs := sess.GetMessages()
			for _, m := range msgs {
				if strings.Contains(strings.ToLower(m.Content), strings.ToLower(args.Query)) {
					sessionResults = append(sessionResults, fmt.Sprintf("[%s]: %s", m.Role, m.Content))
				}
			}
		}
	}

	// 合并结果
	output := ""
	if len(sessionResults) > 0 {
		// 去重
		uniqueSessionResults := make(map[string]bool)
		var finalSessionResults []string
		for _, r := range sessionResults {
			if !uniqueSessionResults[r] {
				uniqueSessionResults[r] = true
				finalSessionResults = append(finalSessionResults, r)
			}
		}
		output += "### Session History & Short-term Memory Results:\n" + strings.Join(finalSessionResults, "\n")
	}

	if output == "" {
		return &toolpkg.Result{Content: "No relevant information found in memory."}, nil
	}

	return &toolpkg.Result{Content: output}, nil
}
