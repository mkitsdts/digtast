package deepseek

import (
	"encoding/json"

	"digital-labor/component/chatmodel/oai"
)

// DeepSeek 流式 chunk 与标准 OpenAI 格式的差异集中在 delta.reasoning_content：
// deepseek-reasoner 会在流式调用中回传思维链，非流式响应则出现在
// message.reasoning_content（由公共的 oai.Message 直接解析）。
type chunk struct {
	ID      string     `json:"id"`
	Model   string     `json:"model"`
	Created int64      `json:"created"`
	Choices []choice   `json:"choices"`
	Usage   *oai.Usage `json:"usage,omitempty"`
}

type choice struct {
	Index        int    `json:"index"`
	Delta        delta  `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

type delta struct {
	Role      string         `json:"role,omitempty"`
	Content   string         `json:"content,omitempty"`
	Reasoning string         `json:"reasoning_content,omitempty"`
	ToolCalls []oai.ToolCall `json:"tool_calls,omitempty"`
}

func decodeChunk(data []byte) (oai.DeltaChunk, error) {
	var wire chunk
	if err := json.Unmarshal(data, &wire); err != nil {
		return oai.DeltaChunk{}, err
	}
	result := oai.DeltaChunk{ID: wire.ID, Model: wire.Model, Created: wire.Created}
	if wire.Usage != nil {
		usage := wire.Usage.ToSchema()
		result.Usage = &usage
	}
	for _, item := range wire.Choices {
		result.Reasoning += item.Delta.Reasoning
		result.Content += item.Delta.Content
		result.ToolCalls = append(result.ToolCalls, item.Delta.ToolCalls...)
		if item.FinishReason != "" {
			result.FinishReason = item.FinishReason
		}
	}
	return result, nil
}
