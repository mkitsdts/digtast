package agent

import (
	"context"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/ctxmanager"
	mmodel "digital-labor/pkg/model"
	"digital-labor/pkg/registry"
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type DigitalAgent struct {
	ID       string
	cm       model.ToolCallingChatModel // 供临时对话使用
	agent    *adk.Runner                // 智能体
	prompts  *PromptBuilder
	runMu    sync.Mutex
	runStops map[string]context.CancelFunc
}

func NewDigitalAgent(provider, key, url, name, agent_id string) (*DigitalAgent, error) {
	dga := &DigitalAgent{
		runStops: make(map[string]context.CancelFunc),
		prompts:  NewPromptBuilder(conf.Conf.WorkSpaceDir),
		ID:       agent_id,
	}

	ctx := ctxmanager.GetOrCreate("digital_agent")
	cm, err := newChatModel(ctx, provider, key, url, name)
	if err != nil {
		return nil, err
	}
	dga.cm = cm

	planner := newPlanner(ctx, dga.cm)
	executor := newExecutor(ctx, dga.cm)
	replanner := newReplanner(ctx, dga.cm)

	agentRunner, err := planexecute.New(ctx, &planexecute.Config{
		Planner:       planner,
		Executor:      executor,
		Replanner:     replanner,
		MaxIterations: 10,
	})
	if err != nil {
		return nil, err
	}

	dga.agent = adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agentRunner,
		EnableStreaming: true,
	})
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

func newPlanner(ctx context.Context, model model.ToolCallingChatModel) adk.Agent {
	planner, err := planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: model,
		ToolInfo:             &planexecute.PlanToolInfo,
	})
	if err != nil {
		slog.Error("create planner failed", "err", err)
	}
	return planner
}

func newExecutor(ctx context.Context, model model.ToolCallingChatModel) adk.Agent {
	toolsConfig := adk.ToolsConfig{
		ToolsNodeConfig: compose.ToolsNodeConfig{
			Tools: registry.GetTools(),
		},
	}
	executor, err := planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
		Model:         model,
		ToolsConfig:   toolsConfig,
		MaxIterations: 5,
	})
	if err != nil {
		slog.Error("create executor failed", "err", err)
	}
	return executor
}

func newReplanner(ctx context.Context, model model.ToolCallingChatModel) adk.Agent {
	replanner, err := planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel: model,
	})
	if err != nil {
		slog.Error("create replanner failed", "err", err)
	}
	return replanner
}

func buildMessages(content, systemPrompt string) ([]*schema.Message, error) {
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("content is empty")
	}

	msgs := make([]*schema.Message, 0, 2)
	if strings.TrimSpace(systemPrompt) != "" {
		msgs = append(msgs, schema.SystemMessage(systemPrompt))
	}
	msgs = append(msgs, schema.UserMessage(content))
	return msgs, nil
}

func sessionIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	sessionID, _ := ctx.Value("session_id").(string)
	return sessionID
}

const (
	DEEPSEEK = "deepseek"
	QWEN     = "qwen"
	DOUBAO   = "doubao"

	DeepseekDefaultBaseURL = ""
	DeepseekDefaultModel   = ""

	QwenDefaultBaseURL = ""
	QwenDefaultModel   = ""

	DoubaoDefaultBaseURL = ""
	DoubaoDefaultModel   = ""
)

func newChatModel(ctx context.Context, provider, key, url, name string) (model.ToolCallingChatModel, error) {
	provider = strings.ToLower(provider)
	if provider == "" {
		return nil, errors.New("provider is empty")
	}
	if key == "" {
		return nil, errors.New("key is empty")
	}
	if name == "" {
		slog.Warn("name is empty, using default")
	}

	switch provider {
	case DEEPSEEK:
		if url == "" {
			slog.Warn("url is empty, using default")
			url = DeepseekDefaultBaseURL
		}
		if name == "" {
			slog.Warn("name is empty, using default")
			name = DeepseekDefaultModel
		}
		return deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
			APIKey:  key,
			BaseURL: url,
			Model:   name,
		})
	case QWEN:
		if url == "" {
			slog.Warn("url is empty, using default")
			url = QwenDefaultBaseURL
		}
		if name == "" {
			slog.Warn("name is empty, using default")
			name = QwenDefaultModel
		}
		return qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
			APIKey:  key,
			BaseURL: url,
			Model:   name,
		})
	case DOUBAO:
		if url == "" {
			slog.Warn("url is empty, using default")
			url = DoubaoDefaultBaseURL
		}
		if name == "" {
			slog.Warn("name is empty, using default")
			name = DoubaoDefaultModel
		}
		return ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey:  key,
			BaseURL: url,
			Model:   name,
		})
	default:
		return nil, errors.New("unknown provider")
	}
}
