package agentloop

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"digital-labor/pkg/chatmodel"
	"digital-labor/pkg/tool"
)

type scriptedLLM struct {
	mu        sync.Mutex
	responses []*chatmodel.ChatResponse
	requests  []*chatmodel.ChatRequest
}

func (m *scriptedLLM) Chat(_ context.Context, request *chatmodel.ChatRequest) (*chatmodel.ChatResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = append(m.requests, request)
	if len(m.responses) == 0 {
		return nil, errors.New("no scripted response")
	}
	response := m.responses[0]
	m.responses = m.responses[1:]
	return response, nil
}

func (m *scriptedLLM) ChatStream(context.Context, *chatmodel.ChatRequest) (chatmodel.Stream, error) {
	return nil, errors.New("not implemented")
}
func (m *scriptedLLM) GetUsage() chatmodel.Usage { return chatmodel.Usage{} }
func (m *scriptedLLM) CheckStatus(context.Context) (*chatmodel.Status, error) {
	return &chatmodel.Status{Available: true}, nil
}

type echoTool struct{}

func (echoTool) Definition(context.Context) (chatmodel.ToolDefinition, error) {
	return chatmodel.ToolDefinition{Name: "echo", Parameters: tool.ObjectSchema(map[string]any{"value": map[string]any{"type": "string"}}, "value")}, nil
}
func (echoTool) Execute(_ context.Context, arguments json.RawMessage) (*tool.Result, error) {
	var input struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(arguments, &input); err != nil {
		return nil, err
	}
	return &tool.Result{Content: input.Value}, nil
}

type recordingHook struct {
	BaseHook
	mu     sync.Mutex
	events []string
}

func (h *recordingHook) add(event string) {
	h.mu.Lock()
	h.events = append(h.events, event)
	h.mu.Unlock()
}
func (h *recordingHook) BeforeStep(context.Context, *StepContext) error {
	h.add("before_step")
	return nil
}
func (h *recordingHook) AfterStep(context.Context, *StepContext) error {
	h.add("after_step")
	return nil
}
func (h *recordingHook) BeforeModel(context.Context, *ModelContext) error {
	h.add("before_model")
	return nil
}
func (h *recordingHook) AfterModel(context.Context, *ModelContext) error {
	h.add("after_model")
	return nil
}
func (h *recordingHook) BeforeTool(context.Context, *ToolContext) error {
	h.add("before_tool")
	return nil
}
func (h *recordingHook) AfterTool(context.Context, *ToolContext) error {
	h.add("after_tool")
	return nil
}

func TestLoopExecutesToolAndHooks(t *testing.T) {
	model := &scriptedLLM{responses: []*chatmodel.ChatResponse{
		{Message: chatmodel.Message{Role: chatmodel.RoleAssistant, ToolCalls: []chatmodel.ToolCall{{ID: "call-1", Type: "function", Function: chatmodel.FunctionCall{Name: "echo", Arguments: json.RawMessage(`{"value":"ok"}`)}}}}, FinishReason: chatmodel.FinishToolCalls, Usage: chatmodel.Usage{TotalTokens: 2}},
		{Message: chatmodel.Message{Role: chatmodel.RoleAssistant, Content: "done"}, FinishReason: chatmodel.FinishStop, Usage: chatmodel.Usage{TotalTokens: 3}},
	}}
	hook := &recordingHook{}
	loop, err := New(Config{
		LLM: model, PromptLoader: PromptLoaderFunc(func(context.Context) (string, error) { return "system prompt", nil }),
		ToolLoaders: []tool.Loader{tool.StaticLoader{echoTool{}}}, Hooks: []Hook{hook}, MaxSteps: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := loop.Run(context.Background(), Input{Messages: []chatmodel.Message{{Role: chatmodel.RoleUser, Content: "start"}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Response.Message.Content != "done" || result.Steps != 2 || result.Usage.TotalTokens != 5 {
		t.Fatalf("unexpected result %#v", result)
	}
	if len(model.requests) != 2 || model.requests[0].Messages[0].Role != chatmodel.RoleSystem {
		t.Fatalf("prompt was not loaded into model request")
	}
	secondMessages := model.requests[1].Messages
	if secondMessages[len(secondMessages)-1].Role != chatmodel.RoleTool || secondMessages[len(secondMessages)-1].Content != "ok" {
		t.Fatalf("tool result was not appended: %#v", secondMessages)
	}
	want := []string{"before_step", "before_model", "after_model", "before_tool", "after_tool", "after_step", "before_step", "before_model", "after_model", "after_step"}
	if len(hook.events) != len(want) {
		t.Fatalf("unexpected hook events %#v", hook.events)
	}
	for index := range want {
		if hook.events[index] != want[index] {
			t.Fatalf("unexpected hook order %#v", hook.events)
		}
	}
}
