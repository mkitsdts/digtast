package agent

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"digital-labor/pkg/agentloop"
	"digital-labor/pkg/chatmodel"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/errs"
	"digital-labor/pkg/filesystem"
	"digital-labor/pkg/llm"
	"digital-labor/pkg/mcp"
	mem "digital-labor/pkg/memory"
	local "digital-labor/pkg/middleware/lbackend"
	mmodel "digital-labor/pkg/model"
	"digital-labor/pkg/registry"
	"digital-labor/pkg/sandbox"
	"digital-labor/pkg/skill"
	"digital-labor/pkg/state"
	"digital-labor/pkg/tool"
	"digital-labor/pkg/workspace"

	"github.com/google/uuid"
)

const defaultModelDescription = "一位云端数字助理。核心目标是成为用户高效、可靠且易于沟通的智能伙伴。具备卓越的理解能力、严谨的逻辑思维和强大的信息整合能力，旨在帮助用户解决问题、获取知识、激发创意并提升效率。"

type AgentMode string

const (
	ModeNormal  AgentMode = "normal"
	ModeSandbox AgentMode = "sandbox"
)

type DigitalAgent struct {
	ID          string
	Name        string
	Description string
	cm          llm.LLM
	loop        *agentloop.Loop
	state       *state.StateManager
	runMu       sync.Mutex
	stopMu      sync.Mutex
	runStop     context.CancelFunc
	memory      *mem.Store
	mode        AgentMode
	sandbox     *sandbox.SandboxedBackend
	files       filesystem.ToolLoader
}

func newDigitalAgent(cfg *mmodel.DigitalAgentConfig) (*DigitalAgent, error) {
	if cfg == nil {
		return nil, errors.New("agent config is required")
	}
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
		return nil, fmt.Errorf("create memory store: %w", err)
	}
	mCfg, ok := conf.FindModelConfig(cfg.Model)
	if !ok {
		return nil, fmt.Errorf("model config not found for: %s", cfg.Model)
	}
	if cfg.ModelKind == "" || cfg.ModelKind == "default" {
		if len(mCfg.ModelNames) == 0 {
			return nil, fmt.Errorf("model config %q contains no model names", cfg.Model)
		}
		cfg.ModelKind = mCfg.ModelNames[0]
	}
	cm, err := chatmodel.ModelManager.GetChatModelByConfig(context.Background(), cfg.Model, cfg.ModelKind)
	if err != nil {
		return nil, err
	}
	dga := &DigitalAgent{
		ID: cfg.ID, Name: cfg.Name, Description: cfg.Description, cm: cm, state: state.NewStateManager(cfg.ID), memory: store, mode: ModeNormal,
		files: filesystem.ToolLoader{Backend: local.GetBackend(), Shell: local.GetBackend()},
	}
	if err := dga.rebuildLoop(); err != nil {
		return nil, err
	}
	return dga, nil
}

func (dga *DigitalAgent) rebuildLoop() error {
	skillManager := skill.NewManagerForAgent(dga.ID)
	loadSkill, err := skill.NewLoadSkillTool(skillManager, nil)
	if err != nil {
		return err
	}
	promptLoader := agentloop.MultiPromptLoader{
		agentloop.PromptLoaderFunc(func(context.Context) (string, error) {
			return "### Agent identity\nName: " + dga.Name + "\nDescription: " + dga.Description, nil
		}),
		dga.state,
		agentloop.PromptLoaderFunc(skillManager.Prompt),
	}
	loop, err := agentloop.New(agentloop.Config{
		LLM:          dga.cm,
		PromptLoader: promptLoader,
		ToolLoaders: []tool.Loader{
			registry.Loader{},
			mcp.NewManager(),
			dga.files,
			tool.StaticLoader{loadSkill},
		},
		Hooks: registry.GetHandlers(), MaxSteps: 20, MaxModelRetries: 5,
	})
	if err != nil {
		return err
	}
	dga.loop = loop
	return nil
}

func (dga *DigitalAgent) GetModel() llm.LLM { return dga.cm }

func (dga *DigitalAgent) Run(ctx context.Context, content string, resources []mmodel.MultiModalResource, isStream bool) (chan string, error) {
	if content == "" && len(resources) == 0 {
		return nil, errors.New("content is empty")
	}
	return dga.run(ctx, mmodel.ChatRequest{Content: content, MultiModalResources: resources, IsStream: isStream})
}

func (dga *DigitalAgent) GetSession() (*mem.Session, error) {
	if dga.memory == nil {
		return nil, errors.New("memory store is not initialized")
	}
	return dga.memory.GetOrCreate()
}

func (dga *DigitalAgent) ClearHistory() error {
	if dga.memory == nil {
		return errors.New("memory store is not initialized")
	}
	return dga.memory.Delete()
}

func (dga *DigitalAgent) Cancel() error { return dga.stop() }

func (dga *DigitalAgent) Close() error {
	if dga.sandbox != nil {
		return dga.sandbox.Close()
	}
	return nil
}

func (dga *DigitalAgent) RebuildAgent(mode AgentMode) error {
	dga.runMu.Lock()
	defer dga.runMu.Unlock()
	if mode != ModeNormal && mode != ModeSandbox {
		return fmt.Errorf("unknown agent mode %q", mode)
	}
	if mode == ModeSandbox && dga.sandbox == nil {
		hostWorkPath := filepath.Join(workspace.GetWorkspacePath(), "workspace")
		sb, err := sandbox.WithSandbox(context.Background(), hostWorkPath)
		if err != nil {
			return err
		}
		dga.sandbox = sb
		dga.files = filesystem.ToolLoader{Backend: sb, Shell: sb}
	}
	if mode == ModeNormal && dga.sandbox != nil {
		if err := dga.sandbox.Close(); err != nil {
			return err
		}
		dga.sandbox = nil
		dga.files = filesystem.ToolLoader{Backend: local.GetBackend(), Shell: local.GetBackend()}
	}
	dga.mode = mode
	return dga.rebuildLoop()
}

func (dga *DigitalAgent) dream(ctx context.Context) error {
	if !dga.runMu.TryLock() {
		return nil
	}
	defer dga.runMu.Unlock()
	session, err := dga.memory.GetOrCreate()
	if err != nil {
		return err
	}
	return state.Compress(ctx, dga.cm, session, dga.state)
}

func (dga *DigitalAgent) Compress() error {
	dga.runMu.Lock()
	defer dga.runMu.Unlock()
	session, err := dga.memory.GetOrCreate()
	if err != nil {
		return err
	}
	return state.Compress(context.Background(), dga.cm, session, dga.state)
}

func (dga *DigitalAgent) UpdateTools() error {
	dga.runMu.Lock()
	defer dga.runMu.Unlock()
	return dga.rebuildLoop()
}

func (dga *DigitalAgent) stop() error {
	dga.stopMu.Lock()
	defer dga.stopMu.Unlock()
	if dga.runStop == nil {
		return errors.New("task not exist")
	}
	dga.runStop()
	dga.runStop = nil
	return nil
}

func (dga *DigitalAgent) setStop(cancel context.CancelFunc) {
	dga.stopMu.Lock()
	dga.runStop = cancel
	dga.stopMu.Unlock()
}
