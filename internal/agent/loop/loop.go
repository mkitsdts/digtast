package agentloop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"digital-labor/component/chatmodel"
	"digital-labor/component/prompt"
	"digital-labor/component/tool"
	"digital-labor/schema"
)

var ErrMaxSteps = errors.New("agent loop reached maximum steps")

type Config struct {
	LLM             chatmodel.ChatModel
	PromptLoader    prompt.Loader
	ToolLoaders     []tool.Loader
	Hooks           []Hook
	MaxSteps        int
	MaxModelRetries int
	RetryDelay      time.Duration
}

type Loop struct {
	llm             chatmodel.ChatModel
	promptLoader    prompt.Loader
	toolLoaders     []tool.Loader
	hooks           []Hook
	maxSteps        int
	maxModelRetries int
	retryDelay      time.Duration
}

type EventType string

const (
	EventTextDelta  EventType = "text_delta"
	EventToolCall   EventType = "tool_call"
	EventToolResult EventType = "tool_result"
	EventUsage      EventType = "usage"
	EventDone       EventType = "done"
)

type Event struct {
	Type       EventType
	Delta      string
	ToolCall   *schema.ToolCall
	ToolResult *tool.Result
	Usage      *schema.Usage
	Response   *schema.ChatResponse
}

type Input struct {
	Messages []schema.Message
	Stream   bool
	Emit     func(Event) error
	Metadata map[string]string
}

type Result struct {
	Messages  []schema.Message
	Generated []schema.Message
	Response  *schema.ChatResponse
	Usage     schema.Usage
	Steps     int
}

func New(cfg Config) (*Loop, error) {
	if cfg.LLM == nil {
		return nil, errors.New("agent loop llm is required")
	}
	if cfg.MaxSteps <= 0 {
		cfg.MaxSteps = 20
	}
	if cfg.MaxModelRetries < 0 {
		return nil, errors.New("max model retries cannot be negative")
	}
	if cfg.RetryDelay <= 0 {
		cfg.RetryDelay = 250 * time.Millisecond
	}
	return &Loop{
		llm: cfg.LLM, promptLoader: cfg.PromptLoader, toolLoaders: cfg.ToolLoaders,
		hooks: cfg.Hooks, maxSteps: cfg.MaxSteps, maxModelRetries: cfg.MaxModelRetries, retryDelay: cfg.RetryDelay,
	}, nil
}

func (l *Loop) Run(ctx context.Context, input Input) (result *Result, runErr error) {
	if len(input.Messages) == 0 {
		return nil, errors.New("agent loop messages are required")
	}
	runContext := &RunContext{Input: &input}
	defer func() {
		runContext.Result = result
		runContext.Error = runErr
		if err := after(ctx, l.hooks, func(h Hook) error { return h.AfterRun(ctx, runContext) }); runErr == nil && err != nil {
			runErr = err
		}
	}()
	if err := before(ctx, l.hooks, func(h Hook) error { return h.BeforeRun(ctx, runContext) }); err != nil {
		runContext.Error = err
		return nil, err
	}

	registry, err := tool.Load(ctx, l.toolLoaders...)
	if err != nil {
		return nil, fmt.Errorf("load tools: %w", err)
	}
	_, definitions, err := registry.All(ctx)
	if err != nil {
		return nil, err
	}
	messages := append([]schema.Message(nil), input.Messages...)
	if l.promptLoader != nil {
		prompt, err := l.promptLoader.LoadPrompt(ctx)
		if err != nil {
			return nil, fmt.Errorf("load prompt: %w", err)
		}
		if prompt != "" {
			messages = append([]schema.Message{{Role: schema.RoleSystem, Content: prompt}}, messages...)
		}
	}
	result = &Result{Messages: messages}

	for stepIndex := 0; stepIndex < l.maxSteps; stepIndex++ {
		request := &schema.ChatRequest{Messages: messages, Tools: definitions, Metadata: input.Metadata}
		if len(definitions) > 0 {
			request.ToolChoice = schema.ToolChoiceAuto
		}
		stepContext := &StepContext{Run: runContext, Index: stepIndex, Request: request}
		if err := before(ctx, l.hooks, func(h Hook) error { return h.BeforeStep(ctx, stepContext) }); err != nil {
			stepContext.Error = err
			_ = after(ctx, l.hooks, func(h Hook) error { return h.AfterStep(ctx, stepContext) })
			return nil, err
		}

		response, err := l.callModel(ctx, stepContext, input.Stream, input.Emit)
		stepContext.Response = response
		stepContext.Error = err
		if err != nil {
			_ = after(ctx, l.hooks, func(h Hook) error { return h.AfterStep(ctx, stepContext) })
			return nil, err
		}
		result.Steps = stepIndex + 1
		result.Response = response
		result.Usage.Add(response.Usage)
		messages = append(messages, response.Message)
		result.Generated = append(result.Generated, response.Message)
		result.Messages = messages

		if len(response.Message.ToolCalls) == 0 {
			if err := after(ctx, l.hooks, func(h Hook) error { return h.AfterStep(ctx, stepContext) }); err != nil {
				return nil, err
			}
			if input.Emit != nil {
				if err := input.Emit(Event{Type: EventDone, Response: response}); err != nil {
					return nil, err
				}
			}
			return result, nil
		}

		toolMessages, err := l.executeTools(ctx, stepContext, registry, response.Message.ToolCalls, input.Emit)
		if err != nil {
			stepContext.Error = err
			_ = after(ctx, l.hooks, func(h Hook) error { return h.AfterStep(ctx, stepContext) })
			return nil, err
		}
		messages = append(messages, toolMessages...)
		result.Generated = append(result.Generated, toolMessages...)
		result.Messages = messages
		if err := after(ctx, l.hooks, func(h Hook) error { return h.AfterStep(ctx, stepContext) }); err != nil {
			return nil, err
		}
	}
	return result, ErrMaxSteps
}

func (l *Loop) callModel(ctx context.Context, step *StepContext, stream bool, emit func(Event) error) (*schema.ChatResponse, error) {
	modelContext := &ModelContext{Step: step, Request: step.Request}
	if err := before(ctx, l.hooks, func(h Hook) error { return h.BeforeModel(ctx, modelContext) }); err != nil {
		modelContext.Error = err
		_ = after(ctx, l.hooks, func(h Hook) error { return h.AfterModel(ctx, modelContext) })
		return nil, err
	}
	var response *schema.ChatResponse
	var callErr error
	for attempt := 0; attempt <= l.maxModelRetries; attempt++ {
		if stream {
			response, callErr = l.consumeStream(ctx, modelContext.Request, emit)
		} else {
			response, callErr = l.llm.Chat(ctx, modelContext.Request)
		}
		if callErr == nil || stream || attempt == l.maxModelRetries || ctx.Err() != nil {
			break
		}
		timer := time.NewTimer(l.retryDelay * time.Duration(1<<attempt))
		select {
		case <-ctx.Done():
			timer.Stop()
			callErr = ctx.Err()
		case <-timer.C:
		}
	}
	modelContext.Response = response
	modelContext.Error = callErr
	if afterErr := after(ctx, l.hooks, func(h Hook) error { return h.AfterModel(ctx, modelContext) }); callErr == nil && afterErr != nil {
		callErr = afterErr
	}
	return response, callErr
}

func (l *Loop) consumeStream(ctx context.Context, request *schema.ChatRequest, emit func(Event) error) (*schema.ChatResponse, error) {
	stream, err := l.llm.ChatStream(ctx, request)
	if err != nil {
		return nil, err
	}
	defer stream.Close()
	for {
		event, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil, errors.New("llm stream ended without a done event")
		}
		if err != nil {
			return nil, err
		}
		switch event.Type {
		case schema.StreamEventTextDelta:
			if emit != nil {
				if err := emit(Event{Type: EventTextDelta, Delta: event.Delta}); err != nil {
					return nil, err
				}
			}
		case schema.StreamEventUsage:
			if emit != nil {
				if err := emit(Event{Type: EventUsage, Usage: event.Usage}); err != nil {
					return nil, err
				}
			}
		case schema.StreamEventDone:
			if event.Response == nil {
				return nil, errors.New("llm stream done event has no response")
			}
			return event.Response, nil
		}
	}
}

func (l *Loop) executeTools(ctx context.Context, step *StepContext, registry *tool.Registry, calls []schema.ToolCall, emit func(Event) error) ([]schema.Message, error) {
	type outcome struct {
		message schema.Message
		result  *tool.Result
		err     error
	}
	outcomes := make([]outcome, len(calls))
	if emit != nil {
		for index := range calls {
			call := calls[index]
			if err := emit(Event{Type: EventToolCall, ToolCall: &call}); err != nil {
				return nil, err
			}
		}
	}
	var wg sync.WaitGroup
	for index, call := range calls {
		index, call := index, call
		wg.Add(1)
		go func() {
			defer wg.Done()
			t, ok := registry.Get(call.Function.Name)
			if !ok {
				content := fmt.Sprintf("[tool error] tool %q is not registered", call.Function.Name)
				outcomes[index].message = schema.Message{Role: schema.RoleTool, ToolCallID: call.ID, Content: content}
				outcomes[index].result = &tool.Result{Content: content, IsError: true}
				return
			}
			toolContext := &ToolContext{Step: step, Call: call, Tool: t, Arguments: call.Function.Arguments}
			if err := before(ctx, l.hooks, func(h Hook) error { return h.BeforeTool(ctx, toolContext) }); err != nil {
				toolContext.Error = err
				_ = after(ctx, l.hooks, func(h Hook) error { return h.AfterTool(ctx, toolContext) })
				outcomes[index].err = err
				return
			}
			result, execErr := t.Execute(ctx, call.Function.Arguments)
			toolContext.Result = result
			toolContext.Error = execErr
			if afterErr := after(ctx, l.hooks, func(h Hook) error { return h.AfterTool(ctx, toolContext) }); execErr == nil && afterErr != nil {
				execErr = afterErr
			}
			if execErr != nil {
				result = &tool.Result{Content: "[tool error] " + execErr.Error(), IsError: true}
			}
			if result == nil {
				result = &tool.Result{Content: "[tool error] tool returned no result", IsError: true}
			}
			outcomes[index].result = result
			content := result.Content
			if content == "" && result.Data != nil {
				encoded, err := json.Marshal(result.Data)
				if err != nil {
					outcomes[index].err = err
					return
				}
				content = string(encoded)
			}
			outcomes[index].message = schema.Message{Role: schema.RoleTool, ToolCallID: call.ID, Content: content}
		}()
	}
	wg.Wait()
	messages := make([]schema.Message, 0, len(calls))
	for index, outcome := range outcomes {
		if outcome.err != nil {
			return nil, outcome.err
		}
		if emit != nil {
			call := calls[index]
			if err := emit(Event{Type: EventToolResult, ToolCall: &call, ToolResult: outcome.result}); err != nil {
				return nil, err
			}
		}
		messages = append(messages, outcome.message)
	}
	return messages, nil
}

func before(ctx context.Context, hooks []Hook, call func(Hook) error) error {
	for _, hook := range hooks {
		if hook == nil {
			continue
		}
		if err := call(hook); err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}

func after(ctx context.Context, hooks []Hook, call func(Hook) error) error {
	var result error
	for index := len(hooks) - 1; index >= 0; index-- {
		if hooks[index] == nil {
			continue
		}
		if err := call(hooks[index]); err != nil && result == nil {
			result = err
		}
	}
	if result == nil {
		result = ctx.Err()
	}
	return result
}
