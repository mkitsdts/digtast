package state

import (
	"context"
	"digital-labor/config"
	mem "digital-labor/internal/agent/memory"
	"digital-labor/pkg/chatmodel"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

const extractPrompt = `你是一个对话提取专家。请从以下对话中提取关键信息，保留：
1. 用户的需求、偏好、决策
2. 助手的重要回复和建议
3. 双方达成的共识或结论

删除：
- 工具调用的技术细节
- 重复的确认语句
- 不影响后续对话的中间过程

输出格式：JSON 数组，每个元素格式为 {"role":"user"|"assistant","content":"..."}。
只输出 JSON，不要有其他内容。`

const promotePrompt = `分析以下短期记忆和现有长期记忆，找出短期记忆中反复提及或重要的内容。
这些内容应该被提升为长期记忆。

短期记忆：
%s

现有长期记忆：
%s

输出格式：JSON 数组，每个元素格式为 {"content":"需要记住的关键信息"}。
如果没有需要提升的内容，输出空数组 []。
只输出 JSON，不要有其他内容。`

const cleanPrompt = `以下是一份长期记忆文件（memory.md），请去重和精简：
1. 合并重复或相似的内容
2. 优先删除较旧的信息（日期更早的）
3. 保留最近和最重要的信息
4. 如果内容过多，只保留最近 %d 条

当前内容：
%s

输出格式：清理后的完整 memory.md 内容（markdown 格式）。直接输出内容，不要有其他说明。`

// FilterUserAssistant filters out Tool messages, keeping only User and Assistant messages.
func FilterUserAssistant(msgs []*chatmodel.Message) []*chatmodel.Message {
	var result []*chatmodel.Message
	for _, msg := range msgs {
		if msg == nil {
			continue
		}
		if msg.Role == chatmodel.RoleUser || msg.Role == chatmodel.RoleAssistant {
			result = append(result, msg)
		}
	}
	return result
}

// ShardMessages splits messages into chunks where each chunk's estimated tokens <= tokenLimit.
func ShardMessages(msgs []*chatmodel.Message, tokenLimit int) [][]*chatmodel.Message {
	if tokenLimit <= 0 {
		tokenLimit = 4000
	}

	var shards [][]*chatmodel.Message
	var currentShard []*chatmodel.Message
	currentTokens := 0

	for _, msg := range msgs {
		if msg == nil {
			continue
		}
		msgTokens := estimateTokens(msg.Content)
		for _, part := range msg.Parts {
			msgTokens += estimateTokens(part.Text)
		}

		if currentTokens+msgTokens > tokenLimit && len(currentShard) > 0 {
			shards = append(shards, currentShard)
			currentShard = nil
			currentTokens = 0
		}

		currentShard = append(currentShard, msg)
		currentTokens += msgTokens
	}

	if len(currentShard) > 0 {
		shards = append(shards, currentShard)
	}

	return shards
}

type indexedMessage struct {
	line int
	msg  *chatmodel.Message
}

func shardIndexedMessages(msgs []indexedMessage, tokenLimit int) [][]indexedMessage {
	if tokenLimit <= 0 {
		tokenLimit = 4000
	}

	var shards [][]indexedMessage
	var currentShard []indexedMessage
	currentTokens := 0

	for _, item := range msgs {
		if item.msg == nil {
			continue
		}
		msgTokens := messageTokens(item.msg)
		if currentTokens+msgTokens > tokenLimit && len(currentShard) > 0 {
			shards = append(shards, currentShard)
			currentShard = nil
			currentTokens = 0
		}

		currentShard = append(currentShard, item)
		currentTokens += msgTokens
	}

	if len(currentShard) > 0 {
		shards = append(shards, currentShard)
	}

	return shards
}

func messageTokens(msg *chatmodel.Message) int {
	if msg == nil {
		return 0
	}
	msgTokens := estimateTokens(msg.Content)
	for _, part := range msg.Parts {
		msgTokens += estimateTokens(part.Text)
	}
	return msgTokens
}

// ExtractChunk calls the LLM to extract key information from a conversation chunk.
func ExtractChunk(ctx context.Context, m chatmodel.ChatModel, chunk []*chatmodel.Message) ([]*chatmodel.Message, error) {
	var conversation strings.Builder
	for _, msg := range chunk {
		if msg == nil {
			continue
		}
		role := string(msg.Role)
		if msg.Content != "" {
			conversation.WriteString(fmt.Sprintf("[%s] %s\n", role, msg.Content))
		}
		for _, part := range msg.Parts {
			if part.Text != "" {
				conversation.WriteString(fmt.Sprintf("[%s] %s\n", role, part.Text))
			}
		}
	}

	resp, err := m.Chat(ctx, &chatmodel.ChatRequest{Messages: []chatmodel.Message{
		{Role: chatmodel.RoleUser, Content: extractPrompt + "\n\n对话内容：\n" + conversation.String()},
	}})
	if err != nil {
		return nil, fmt.Errorf("llm extract failed: %w", err)
	}

	var extracted []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(resp.Message.Content), &extracted); err != nil {
		// If parsing fails, return the raw response as a single assistant message
		return []*chatmodel.Message{
			{Role: chatmodel.RoleAssistant, Content: resp.Message.Content},
		}, nil
	}

	var result []*chatmodel.Message
	for _, item := range extracted {
		if item.Content == "" {
			continue
		}
		role := chatmodel.RoleAssistant
		if item.Role == "user" {
			role = chatmodel.RoleUser
		}
		result = append(result, &chatmodel.Message{Role: role, Content: item.Content})
	}
	return result, nil
}

// Compress executes the full memory compression pipeline
func Compress(ctx context.Context, m chatmodel.ChatModel, session *mem.Session, sm *StateManager) error {
	msgs := session.GetMessages()
	if len(msgs) == 0 {
		return nil
	}

	compressedThrough := session.CompressedThrough()
	if compressedThrough >= len(msgs) {
		return nil
	}

	var filtered []indexedMessage
	for idx, msg := range msgs[compressedThrough:] {
		if msg == nil {
			continue
		}
		if msg.Role == chatmodel.RoleUser || msg.Role == chatmodel.RoleAssistant {
			filtered = append(filtered, indexedMessage{
				line: compressedThrough + idx + 1,
				msg:  msg,
			})
		}
	}
	if len(filtered) == 0 {
		return nil
	}

	// Slice messages into shards
	chunkLimit := config.Conf.Memory.ChunkTokenLimit
	if chunkLimit <= 0 {
		chunkLimit = 4000
	}
	shards := shardIndexedMessages(filtered, chunkLimit)

	//LLM extract each shard
	var extracted []*chatmodel.Message
	for _, shard := range shards {
		chunk := make([]*chatmodel.Message, 0, len(shard))
		for _, item := range shard {
			chunk = append(chunk, item.msg)
		}
		chunkResult, err := ExtractChunk(ctx, m, chunk)
		if err != nil {
			slog.Error("failed to extract chunk", "err", err)
			// Fallback: keep the shard as-is
			chunkResult = chunk
		}
		if len(chunkResult) == 0 {
			continue
		}
		startLine := shard[0].line
		endLine := shard[len(shard)-1].line
		if err := session.AppendCompression(startLine, endLine, chunkResult); err != nil {
			return fmt.Errorf("failed to write compressed session: %w", err)
		}
		extracted = append(extracted, chunkResult...)
	}

	if len(extracted) == 0 {
		return nil
	}

	// Promote to long-term memory
	if err := PromoteMessagesToMemory(ctx, m, extracted, sm); err != nil {
		slog.Error("failed to promote to memory", "err", err)
	}

	// Archive original chunks and clear from disk
	if err := session.DeleteChunks(); err != nil {
		slog.Error("failed to archive session chunks", "err", err)
	}

	// Periodic cleanup
	if shouldCleanMemory() {
		if err := CleanMemory(ctx, m, sm); err != nil {
			slog.Error("failed to clean memory", "err", err)
		}
	}

	return nil
}

var compressCount int

func shouldCleanMemory() bool {
	compressCount++
	interval := config.Conf.Memory.CleanInterval
	if interval <= 0 {
		interval = 10
	}
	return compressCount%interval == 0
}

// PromoteToMemory analyzes short-term memory and existing memory.md,
// promoting frequently mentioned or important items to long-term memory.
func PromoteMessagesToMemory(ctx context.Context, m chatmodel.ChatModel, msgs []*chatmodel.Message, sm *StateManager) error {
	existing := sm.LoadMemory()

	// Build short-term memory context from session messages
	var stmBuilder strings.Builder
	for _, msg := range msgs {
		if msg == nil || (msg.Role != chatmodel.RoleUser && msg.Role != chatmodel.RoleAssistant) {
			continue
		}
		if msg.Content != "" {
			stmBuilder.WriteString(fmt.Sprintf("[%s] %s\n", string(msg.Role), msg.Content))
		}
	}
	shortTermMemory := stmBuilder.String()
	if shortTermMemory == "" {
		return nil
	}

	resp, err := m.Chat(ctx, &chatmodel.ChatRequest{Messages: []chatmodel.Message{
		{Role: chatmodel.RoleUser, Content: fmt.Sprintf(promotePrompt, shortTermMemory, existing)},
	}})
	if err != nil {
		return fmt.Errorf("llm promote failed: %w", err)
	}

	var items []struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(resp.Message.Content), &items); err != nil {
		return nil // No items to promote
	}

	for _, item := range items {
		if item.Content == "" {
			continue
		}
		entry := fmt.Sprintf("\n\n## %s\n%s\n", time.Now().Format("2006-01-02 15:04"), item.Content)
		if err := sm.SaveMemory(entry); err != nil {
			slog.Error("failed to save promoted memory", "err", err)
		}
	}

	return nil
}

// CleanMemory deduplicates and cleans memory.md using LLM.
func CleanMemory(ctx context.Context, m chatmodel.ChatModel, sm *StateManager) error {
	existing := sm.LoadMemory()
	if existing == "" {
		return nil
	}

	resp, err := m.Chat(ctx, &chatmodel.ChatRequest{Messages: []chatmodel.Message{
		{Role: chatmodel.RoleUser, Content: fmt.Sprintf(cleanPrompt, 20, existing)},
	}})
	if err != nil {
		return fmt.Errorf("llm clean failed: %w", err)
	}

	cleaned := strings.TrimSpace(resp.Message.Content)
	if cleaned == "" || cleaned == existing {
		return nil
	}

	// Replace memory.md with cleaned content
	if err := sm.ReplaceMemory(cleaned); err != nil {
		return fmt.Errorf("failed to replace memory: %w", err)
	}

	return nil
}

func estimateTokens(text string) int {
	return len(text) / 4
}
