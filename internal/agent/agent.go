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
	"log/slog"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/google/uuid"
)

type DigitalAgent struct {
	ID       string
	Name     string
	cm       model.ToolCallingChatModel
	agent    *adk.ChatModelAgent
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

	store, err := mem.NewStore(cfg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory store: %w", err)
	}

	dga := &DigitalAgent{
		prompts: NewPromptBuilder(),
		ID:      cfg.ID,
		Name:    cfg.Name,
		memory:  store,
	}

	mCfg, ok := conf.FindModelConfig(cfg.Model)
	if !ok {
		return nil, fmt.Errorf("model config not found for: %s", cfg.Model)
	}
	slog.Info("model config found", "model", cfg.Model, "modelKind", mCfg.ModelNames)

	ctx := ctxmanager.GetOrCreate(cfg.ID)
	if cfg.ModelKind == "default" {
		cfg.ModelKind = mCfg.ModelNames[0]
	}

	cm, err := newChatModel(ctx, mCfg.Provider, mCfg.Key, mCfg.URL, cfg.ModelKind)
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

// Run starts a conversation turn. The agent uses its own ID as the session key.
func (dga *DigitalAgent) Run(ctx context.Context, content string, isStream bool) (chan string, error) {
	if content == "" {
		return nil, errors.New("content is empty")
	}

	vctx := ctxmanager.GetOrCreate(dga.ID)

	return dga.run(vctx, mmodel.ChatRequest{
		Content:  content,
		IsStream: isStream,
	})
}

// GetSession returns the agent's single session.
func (dga *DigitalAgent) GetSession() (*mem.Session, error) {
	if dga.memory == nil {
		return nil, errors.New("memory store is not initialized")
	}
	return dga.memory.GetOrCreate()
}

// ClearHistory deletes the agent's conversation history and starts fresh.
func (dga *DigitalAgent) ClearHistory() error {
	if dga.memory == nil {
		return errors.New("memory store is not initialized")
	}
	return dga.memory.Delete()
}

func (dga *DigitalAgent) Cancel() error {
	return dga.stop()
}

func (dga *DigitalAgent) UpdateTools() error {
	if dga.agent == nil {
		return nil
	}

	ctx := ctxmanager.GetOrCreate(dga.ID)
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
