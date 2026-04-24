package agent

import (
	"context"
	mem "digital-labor/internal/memory"
	"digital-labor/pkg/ctxmanager"
	mmodel "digital-labor/pkg/model"
	tooll "digital-labor/pkg/tool"
	"errors"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/google/uuid"
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

func NewDigitalAgent(cfg *DigitalAgentConfig) (*DigitalAgent, error) {
	id := uuid.New().String()

	if cfg.Name == "" {
		cfg.Name = id
	}

	if cfg.Description == "" {
		cfg.Description = defaultModelDescription
	}

	dga := &DigitalAgent{
		runStops: make(map[string]context.CancelFunc),
		prompts:  NewPromptBuilder(),
		ID:       id,
		memory:   mem.NewStore(),
	}

	ctx := ctxmanager.GetOrCreate(id)
	cm, err := newChatModel(ctx, cfg.Provider, cfg.Key, cfg.URL, cfg.Name)
	if err != nil {
		return nil, err
	}
	dga.cm = cm

	dga.agent, err = adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:  cfg.Name,
		Model: cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tooll.GetTools(),
			},
		},
		Description: cfg.Description,
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
