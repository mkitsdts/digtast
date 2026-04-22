package agent

import (
	"context"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/ctxmanager"
	"digital-labor/pkg/tool"
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/cloudwego/eino-examples/quickstart/chatwitheino/mem"
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

type DigitalAgent struct {
	UsedID   string
	cm       model.ToolCallingChatModel
	agent    *adk.Runner
	sessions *mem.Store
}

var digitalAgent *DigitalAgent
var once sync.Once

// 初始化 Agent 需要完成 APIKey 的配置，并创建会话存储
func InitDigitalAgent(provider, key, url, name string) error {
	ch := make(chan error)
	once.Do(func() {
		dga, err := newDigitalLabor(provider, key, url, name)
		if err != nil {
			slog.Error("failed to create digital agent", "err", err)
			ch <- err
			return
		}
		digitalAgent = dga
		close(ch)
	})
	return <-ch
}

// 暂停 Agent 运行
func PauseDigitalAgent() {
	// TODO:
}

func GetDigitalAgent() *DigitalAgent {
	return digitalAgent
}

func newDigitalLabor(provider, key, url, name string) (*DigitalAgent, error) {
	dga := &DigitalAgent{}
	var err error

	if dga.sessions, err = mem.NewStore(conf.Conf.WorkSpaceDir + "/sessions"); err != nil {
		return nil, err
	}

	ctx := ctxmanager.GetOrCreate("digital_agent")
	dga.cm, err = newChatModel(ctx, provider, key, url, name)
	if err != nil {
		return nil, err
	}

	planner := newPlanner(ctx, dga.cm)     // 创建计划器
	executor := newExecutor(ctx, dga.cm)   // 创建执行器
	replanner := newReplanner(ctx, dga.cm) // 创建重新计划器

	// MaxIterations 表示最大迭代 10 次
	agent, err := planexecute.New(ctx, &planexecute.Config{
		Planner:       planner,
		Executor:      executor,
		Replanner:     replanner,
		MaxIterations: 10,
	})
	if err != nil {
		return nil, err
	}

	dga.agent = adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent, EnableStreaming: true})

	return dga, nil
}

func newPlanner(ctx context.Context, model model.ToolCallingChatModel) adk.Agent {
	planner, err := planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: model,                     // 使用工具调用模型生成计划
		ToolInfo:             &planexecute.PlanToolInfo, // 默认 Plan 工具 schema
	})
	if err != nil {
		slog.Error("创建 Planner 失败", "err", err)
	}
	return planner
}

func newExecutor(ctx context.Context, model model.ToolCallingChatModel) adk.Agent {
	// 配置 Executor 工具集（仅包含搜索工具）
	toolsConfig := adk.ToolsConfig{
		ToolsNodeConfig: compose.ToolsNodeConfig{
			Tools: tool.GetTools(),
		},
	}
	executor, err := planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
		Model:         model,
		ToolsConfig:   toolsConfig,
		MaxIterations: 5, // ChatModel 最多运行 5 次
	})
	if err != nil {
		slog.Error("创建 Executor 失败", "err", err)
	}
	return executor
}

func newReplanner(ctx context.Context, model model.ToolCallingChatModel) adk.Agent {
	replanner, err := planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel: model, // 使用工具调用模型评估进度
	})
	if err != nil {
		slog.Error("创建 Replanner 失败", "err", err)
	}
	return replanner
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
		// it's impossible to reach here
		return nil, errors.New("unknown error")
	}
}
