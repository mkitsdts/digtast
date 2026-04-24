package agent

import (
	"context"
	mem "digital-labor/internal/memory"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/ctxmanager"
	mmodel "digital-labor/pkg/model"
	tooll "digital-labor/pkg/tool"
	"errors"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

type DigitalAgent struct {
	ID       string
	cm       model.ToolCallingChatModel // 供临时对话使用
	agent    *adk.ChatModelAgent        // 智能体
	prompts  *PromptBuilder
	runMu    sync.Mutex
	runStops map[string]context.CancelFunc
	memory   *mem.Store
}

func NewDigitalAgent(provider, key, url, name, agent_id string) (*DigitalAgent, error) {
	dga := &DigitalAgent{
		runStops: make(map[string]context.CancelFunc),
		prompts:  NewPromptBuilder(conf.Conf.WorkSpaceDir),
		ID:       agent_id,
		memory:   mem.NewStore(),
	}

	ctx := ctxmanager.GetOrCreate("digital_agent")
	cm, err := newChatModel(ctx, provider, key, url, name)
	if err != nil {
		return nil, err
	}
	dga.cm = cm

	if err != nil {
		return nil, err
	}

	dga.agent, err = adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:  agent_id,
		Model: cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tooll.GetTools(),
			},
		},
		Description: "你是一位名为“云端数字助理”的AI助手。你的核心目标是成为用户高效、可靠且易于沟通的智能伙伴。你应具备卓越的理解能力、严谨的逻辑思维和强大的信息整合能力，旨在帮助用户解决问题、获取知识、激发创意并提升效率。你的回答应始终体现专业性、准确性和用户友好性。",
	})
	if err != nil {
		return nil, err
	}
	return dga, nil
}

func (dga *DigitalAgent) Run(ctx context.Context, content string, isStream bool) (chan string, error) {
	if content == "" {
		return nil, errors.New("content is empty")
	}
	return dga.run(ctx, mmodel.ChatRequest{
		SessionID: sessionIDFromContext(ctx),
		Content:   content,
		IsStream:  isStream,
	})
}

func (dga *DigitalAgent) PauseSession(sessionID string) error {
	if sessionID == "" {
		return errors.New("session_id is empty")
	}

	dga.runMu.Lock()
	cancel, ok := dga.runStops[sessionID]
	dga.runMu.Unlock()
	if !ok {
		return errors.New("session is not running")
	}

	cancel()
	return nil
}

func (dga *DigitalAgent) bindRun(sessionID string, cancel context.CancelFunc) {
	dga.runMu.Lock()
	defer dga.runMu.Unlock()

	if previous, ok := dga.runStops[sessionID]; ok {
		previous()
	}
	dga.runStops[sessionID] = cancel
}

func (dga *DigitalAgent) unbindRun(sessionID string) {
	dga.runMu.Lock()
	defer dga.runMu.Unlock()
	delete(dga.runStops, sessionID)
}
