package agent

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"digital-labor/component/chatmodel"
	"digital-labor/component/prompt"
	"digital-labor/component/tool"
	"digital-labor/config"
	agentloop "digital-labor/internal/agent/loop"
	"digital-labor/internal/agent/mcp"
	mem "digital-labor/internal/agent/memory"
	"digital-labor/internal/agent/registry"
	"digital-labor/internal/agent/skill"
	"digital-labor/internal/agent/state"
	"digital-labor/internal/agent/workspace"
	"digital-labor/internal/sandbox"
	"digital-labor/middleware/lbackend"
	"digital-labor/tools/filesystem"

	"github.com/google/uuid"
)

const defaultModelDescription = "一位云端数字助理。核心目标是成为用户高效、可靠且易于沟通的智能伙伴。具备卓越的理解能力、严谨的逻辑思维和强大的信息整合能力，旨在帮助用户解决问题、获取知识、激发创意并提升效率。"

type AgentMode string

const (
	ModeNormal  AgentMode = "normal"
	ModeSandbox AgentMode = "sandbox"
)

type Agent struct {
	Name        string
	Description string
	cm          chatmodel.ChatModel
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

var (
	instanceMu sync.RWMutex
	instance   *Agent
)

// Initialize creates the process-wide Agent. Repeated calls return the same instance.
func Initialize(cfg *config.AgentConfig) (*Agent, error) {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	if instance != nil {
		return instance, nil
	}
	ag, err := newAgent(cfg)
	if err != nil {
		return nil, err
	}
	instance = ag
	return instance, nil
}

// Instance returns the process-wide Agent.
func Instance() (*Agent, error) {
	instanceMu.RLock()
	defer instanceMu.RUnlock()
	if instance == nil {
		return nil, errors.New("agent is not initialized")
	}
	return instance, nil
}

// Reconfigure replaces the process-wide Agent with one built from cfg. Used
// when the user picks a different model at runtime.
func Reconfigure(cfg *config.AgentConfig) (*Agent, error) {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	ag, err := newAgent(cfg)
	if err != nil {
		return nil, err
	}
	if instance != nil {
		_ = instance.Close()
	}
	instance = ag
	return instance, nil
}

func newAgent(cfg *config.AgentConfig) (*Agent, error) {
	if cfg == nil {
		return nil, errors.New("agent config is required")
	}
	if cfg.Name == "" {
		cfg.Name = uuid.NewString()
	}
	if cfg.Provider == "" {
		return nil, errors.New("model provider is required")
	}
	if cfg.Model == "" {
		// 未指定模型时使用该厂商启用的第一个模型。
		models := config.ModelCredentials(cfg.Provider).Models
		if len(models) == 0 {
			return nil, fmt.Errorf("provider %s has no enabled models", cfg.Provider)
		}
		cfg.Model = models[0]
	}
	if cfg.Description == "" {
		cfg.Description = defaultModelDescription
	}
	store, err := mem.NewStore()
	if err != nil {
		return nil, fmt.Errorf("create memory store: %w", err)
	}
	cm, err := newChatModel(cfg.Provider, cfg.Model)
	if err != nil {
		return nil, err
	}
	files, err := newLocalFiles()
	if err != nil {
		return nil, err
	}
	dga := &Agent{
		Name: cfg.Name, Description: cfg.Description, cm: cm, state: state.NewStateManager(), memory: store, mode: ModeNormal,
		files: files,
	}
	if err := dga.rebuildLoop(); err != nil {
		return nil, err
	}
	return dga, nil
}

func (dga *Agent) rebuildLoop() error {
	skillManager := skill.NewManager()
	loadSkill, err := skill.NewLoadSkillTool(skillManager, nil)
	if err != nil {
		return err
	}
	promptLoader := prompt.MultiLoader{
		prompt.LoaderFunc(func(context.Context) (string, error) {
			return "### Agent identity\nName: " + dga.Name + "\nDescription: " + dga.Description, nil
		}),
		dga.state,
		prompt.LoaderFunc(skillManager.Prompt),
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

func (dga *Agent) GetModel() chatmodel.ChatModel { return dga.cm }

func (dga *Agent) Run(ctx context.Context, content string, resources []MultiModalResource, isStream bool) (chan string, error) {
	if content == "" && len(resources) == 0 {
		return nil, errors.New("content is empty")
	}
	return dga.run(ctx, ChatRequest{Content: content, MultiModalResources: resources, IsStream: isStream})
}

func (dga *Agent) GetSession() (*mem.Session, error) {
	if dga.memory == nil {
		return nil, errors.New("memory store is not initialized")
	}
	return dga.memory.GetOrCreate()
}

func (dga *Agent) ClearHistory() error {
	if dga.memory == nil {
		return errors.New("memory store is not initialized")
	}
	return dga.memory.Delete()
}

func (dga *Agent) Cancel() error { return dga.stop() }

func (dga *Agent) Close() error {
	if dga.sandbox != nil {
		return dga.sandbox.Close()
	}
	return nil
}

func (dga *Agent) Rebuild(mode AgentMode) error {
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
		files, err := newLocalFiles()
		if err != nil {
			return err
		}
		dga.files = files
	}
	dga.mode = mode
	return dga.rebuildLoop()
}

func newLocalFiles() (filesystem.ToolLoader, error) {
	backend, err := local.NewBackend(context.Background(), &local.Config{
		WorkPath: filepath.Join(workspace.GetWorkspacePath(), "workspace"),
	})
	if err != nil {
		return filesystem.ToolLoader{}, fmt.Errorf("create local backend: %w", err)
	}
	return filesystem.ToolLoader{Backend: backend, Shell: backend}, nil
}

func (dga *Agent) dream(ctx context.Context) error {
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

func (dga *Agent) Compress() error {
	dga.runMu.Lock()
	defer dga.runMu.Unlock()
	session, err := dga.memory.GetOrCreate()
	if err != nil {
		return err
	}
	return state.Compress(context.Background(), dga.cm, session, dga.state)
}

func (dga *Agent) UpdateTools() error {
	dga.runMu.Lock()
	defer dga.runMu.Unlock()
	return dga.rebuildLoop()
}

func (dga *Agent) stop() error {
	dga.stopMu.Lock()
	defer dga.stopMu.Unlock()
	if dga.runStop == nil {
		return errors.New("task not exist")
	}
	dga.runStop()
	dga.runStop = nil
	return nil
}

func (dga *Agent) setStop(cancel context.CancelFunc) {
	dga.stopMu.Lock()
	dga.runStop = cancel
	dga.stopMu.Unlock()
}
