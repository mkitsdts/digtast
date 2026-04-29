package agent

import (
	"context"
	mem "digital-labor/internal/memory"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/ctxmanager"
	"digital-labor/pkg/errs"
	mmodel "digital-labor/pkg/model"
	"digital-labor/pkg/registry"
	"errors"
	"fmt"
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
	runStops context.CancelFunc
	memory   *mem.Store
}

func newDigitalAgent(cfg *mmodel.DigitalAgentConfig) (*DigitalAgent, error) {
	if cfg.Name == "" {
		return nil, errs.ErrAgentNameRequired
	}

	if cfg.Description == "" {
		cfg.Description = defaultModelDescription
	}

	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
	}

	dga := &DigitalAgent{
		prompts: NewPromptBuilder(),
		ID:      cfg.ID,
		memory:  mem.NewStore(cfg.Name),
	}

	// Fetch model config from global config
	mCfg, ok := conf.FindModelConfig(cfg.Model)
	if !ok {
		return nil, fmt.Errorf("model config not found for: %s", cfg.Model)
	}

	ctx := ctxmanager.GetOrCreate(cfg.Name)
	cm, err := newChatModel(ctx, mCfg.Provider, mCfg.Key, mCfg.URL, cfg.Model)
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
				ToolCallMiddlewares: []compose.ToolMiddleware{
					{Invokable: registry.Invokable},
				},
			},
		},
		Description: cfg.Description,
		Handlers:    []adk.ChatModelAgentMiddleware{registry.GetBackendMiddleware()},
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

func (dga *DigitalAgent) GetSession(sessionID string) (*mem.Session, error) {
	if dga.memory == nil {
		return nil, errors.New("memory store is not initialized")
	}
	return dga.memory.GetOrCreate(sessionID)
}

func (dga *DigitalAgent) RemoveSession(sessionID string) error {
	if dga.memory == nil {
		return errors.New("memory store is not initialized")
	}
	return dga.memory.Delete(sessionID)
}

func (dga *DigitalAgent) Cancel() error {
	return dga.stop()
}

func (dga *DigitalAgent) UpdateTools() error {
	if dga.agent == nil {
		return nil
	}

	ctx := ctxmanager.GetOrCreate(dga.agent.Name(context.Background()))
	ag, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:  dga.agent.Name(context.Background()),
		Model: dga.cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: registry.GetTools(),
				ToolCallMiddlewares: []compose.ToolMiddleware{
					{Invokable: registry.Invokable},
				},
			},
		},
		Description: dga.agent.Description(ctx),
		Handlers:    []adk.ChatModelAgentMiddleware{registry.GetBackendMiddleware()},
	})
	if err != nil {
		return err
	}

	dga.runMu.Lock()
	dga.agent = ag
	dga.runMu.Unlock()

	return nil
}
