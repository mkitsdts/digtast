package agent

import (
	"context"
	"digital-labor/pkg/chatmodel"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/ctxmanager"
	"digital-labor/pkg/errs"
	mem "digital-labor/pkg/memory"
	local "digital-labor/pkg/middleware/lbackend"
	skillmw "digital-labor/pkg/middleware/skill"
	mmodel "digital-labor/pkg/model"
	"digital-labor/pkg/registry"
	"digital-labor/pkg/state"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/google/uuid"
)

const defaultModelDescription = "一位云端数字助理。核心目标是成为用户高效、可靠且易于沟通的智能伙伴。具备卓越的理解能力、严谨的逻辑思维和强大的信息整合能力，旨在帮助用户解决问题、获取知识、激发创意并提升效率。"

type DigitalAgent struct {
	ID       string
	Name     string
	cm       model.ToolCallingChatModel
	agent    adk.ResumableAgent
	state    *state.StateManager
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
		state:  state.NewStateManager(cfg.ID),
		ID:     cfg.ID,
		Name:   cfg.Name,
		memory: store,
	}

	mCfg, ok := conf.FindModelConfig(cfg.Model)
	if !ok {
		return nil, fmt.Errorf("model config not found for: %s", cfg.Model)
	}
	slog.Info("model config found", "model", cfg.Model, "modelKind", mCfg.ModelNames)

	ctx := ctxmanager.GetOrCreate(cfg.ID)
	if cfg.ModelKind == "" || cfg.ModelKind == "default" {
		cfg.ModelKind = mCfg.ModelNames[0]
	}

	cm, err := chatmodel.ModelManager.GetChatModelByConfig(ctx, cfg.Model, cfg.ModelKind)
	if err != nil {
		return nil, err
	}
	dga.cm = cm
	handlers := registry.GetHandlers()
	if skillHandler := skillmw.GetSkillMiddlewareForAgent(cfg.ID); skillHandler != nil {
		handlers = append(handlers, skillHandler)
	}

	dga.agent, err = deep.New(ctx, &deep.Config{
		Name:      cfg.Name,
		ChatModel: cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: registry.GetTools(),
				ToolCallMiddlewares: []compose.ToolMiddleware{
					{Invokable: registry.InvokableTool},
				},
			},
		},
		Backend:        local.GetBackend(),
		StreamingShell: local.GetBackend(),
		Description:    cfg.Description,
		Handlers:       handlers,
		ModelRetryConfig: &adk.ModelRetryConfig{
			MaxRetries: 5,
		},
		WithoutWriteTodos: true,
	})

	if err != nil {
		return nil, err
	}
	return dga, nil
}

func (dga *DigitalAgent) GetModel() model.ToolCallingChatModel {
	return dga.cm
}

// Run starts a conversation turn. The agent uses its own ID as the session key.
func (dga *DigitalAgent) Run(ctx context.Context, content string, resources []mmodel.MultiModalResource, isStream bool) (chan string, error) {
	if content == "" && len(resources) == 0 {
		return nil, errors.New("content is empty")
	}

	vctx := ctxmanager.GetOrCreate(dga.ID)

	return dga.run(vctx, mmodel.ChatRequest{
		Content:             content,
		MultiModalResources: resources,
		IsStream:            isStream,
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

// Compress triggers manual memory compression for the agent's session.
func (dga *DigitalAgent) Compress() error {
	dga.runMu.Lock()
	defer dga.runMu.Unlock()
	session, err := dga.memory.GetOrCreate()
	if err != nil {
		return err
	}
	ctx := ctxmanager.GetOrCreate(dga.ID)
	return state.Compress(ctx, dga.cm, session, dga.state)
}

func (dga *DigitalAgent) UpdateTools() error {
	if dga.agent == nil {
		return nil
	}

	ctx := ctxmanager.GetOrCreate(dga.ID)
	handlers := registry.GetHandlers()
	if skillHandler := skillmw.GetSkillMiddlewareForAgent(dga.ID); skillHandler != nil {
		handlers = append(handlers, skillHandler)
	}

	ag, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:  dga.agent.Name(context.Background()),
		Model: dga.cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: registry.GetTools(),
				ToolCallMiddlewares: []compose.ToolMiddleware{
					{Invokable: registry.InvokableTool},
				},
			},
		},
		Description: dga.agent.Description(ctx),
		Handlers:    handlers,
	})
	if err != nil {
		return err
	}

	dga.runMu.Lock()
	dga.agent = ag
	dga.runMu.Unlock()

	return nil
}
