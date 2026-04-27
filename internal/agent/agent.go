package agent

import (
	"context"
	mem "digital-labor/internal/memory"
	"digital-labor/pkg/ctxmanager"
	mmodel "digital-labor/pkg/model"
	"digital-labor/pkg/registry"
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

func NewDigitalAgent(cfg *mmodel.DigitalAgentConfig) (*DigitalAgent, error) {
	if cfg.Name == "" {
		cfg.Name = uuid.New().String()
	}

	if cfg.Description == "" {
		cfg.Description = defaultModelDescription
	}

	dga := &DigitalAgent{
		runStops: make(map[string]context.CancelFunc),
		prompts:  NewPromptBuilder(),
		ID:       cfg.ID,
		memory:   mem.NewStore(cfg.Name),
	}

	ctx := ctxmanager.GetOrCreate(cfg.Name)
	cm, err := newChatModel(ctx, cfg.Provider, cfg.Key, cfg.URL, cfg.Model)
	if err != nil {
		return nil, err
	}
	dga.cm = cm

	dga.agent, err = adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:  cfg.Name,
		Model: cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: registry.GetTools(),
			},
		},
		Description: cfg.Description,
	})

	if err != nil {
		return nil, err
	}
	return dga, nil
}

func (dga *DigitalAgent) Run(ctx context.Context, content string, isStream bool, sessionID string) (chan string, error) {
	if content == "" {
		return nil, errors.New("content is empty")
	}

	if sessionID == "" {
		return nil, errors.New("session_id is empty")
	}

	// to control the context lifecycle
	vctx := ctxmanager.GetOrCreate(sessionID)

	return dga.run(vctx, mmodel.ChatRequest{
		SessionID: sessionID,
		Content:   content,
		IsStream:  isStream,
	})
}

func (dga *DigitalAgent) Cancel(sessionID string) error {
	if sessionID == "" {
		return errors.New("session_id is empty")
	}

	return dga.stop(sessionID)
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
