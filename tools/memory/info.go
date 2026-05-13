package memory

import (
	"context"
	"digital-labor/internal/center"
	"digital-labor/pkg/workspace"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type MemorySearchTool struct {
}

func (t *MemorySearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	properties := orderedmap.New[string, *jsonschema.Schema]()
	properties.Set("query", &jsonschema.Schema{
		Type:        "string",
		Description: "搜索关键词，用于在记忆（对话历史和长期记忆）中查找相关信息。",
	})

	jsonschema := &jsonschema.Schema{
		Version:    "v0.1",
		ID:         "memory_search",
		Type:       "object",
		Properties: properties,
		Required:   []string{"query"},
	}

	return &schema.ToolInfo{
		Name:        "memory_search",
		Desc:        "search in agent's memory (history and long-term memory)",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(jsonschema),
	}, nil
}

type MemorySearchArguments struct {
	Query string `json:"query"`
}

func (t *MemorySearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	args := MemorySearchArguments{}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	agentId, _ := ctx.Value("agent_id").(string)
	if agentId == "" {
		agentId = workspace.DefaultAgentID()
	}

	// 搜索短期记忆 (Session history & compressed records)
	da, err := center.AgentManager.GetAgent(agentId)
	var sessionResults []string
	if err == nil {
		sess, err := da.GetSession()
		if err == nil {
			// 搜索压缩记录
			compressedMsgs, _ := sess.Search(args.Query)
			for _, m := range compressedMsgs {
				sessionResults = append(sessionResults, fmt.Sprintf("[%s]: %s", m.Role, m.Content))
			}

			// 搜索原始消息
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
		output += "### Session History Results:\n" + strings.Join(finalSessionResults, "\n")
	}

	if output == "" {
		return "No relevant information found in memory.", nil
	}

	return output, nil
}
