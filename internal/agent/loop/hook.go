package agentloop

import (
	"context"

	"digital-labor/component/tool"
	"digital-labor/schema"
)

type RunContext struct {
	Input  *Input
	Result *Result
	Error  error
}

type StepContext struct {
	Run      *RunContext
	Index    int
	Request  *schema.ChatRequest
	Response *schema.ChatResponse
	Error    error
}

type ModelContext struct {
	Step     *StepContext
	Request  *schema.ChatRequest
	Response *schema.ChatResponse
	Error    error
}

type ToolContext struct {
	Step      *StepContext
	Call      schema.ToolCall
	Tool      tool.Tool
	Arguments []byte
	Result    *tool.Result
	Error     error
}

type Hook interface {
	BeforeRun(context.Context, *RunContext) error
	AfterRun(context.Context, *RunContext) error
	BeforeStep(context.Context, *StepContext) error
	AfterStep(context.Context, *StepContext) error
	BeforeModel(context.Context, *ModelContext) error
	AfterModel(context.Context, *ModelContext) error
	BeforeTool(context.Context, *ToolContext) error
	AfterTool(context.Context, *ToolContext) error
}

type BaseHook struct{}

func (BaseHook) BeforeRun(context.Context, *RunContext) error     { return nil }
func (BaseHook) AfterRun(context.Context, *RunContext) error      { return nil }
func (BaseHook) BeforeStep(context.Context, *StepContext) error   { return nil }
func (BaseHook) AfterStep(context.Context, *StepContext) error    { return nil }
func (BaseHook) BeforeModel(context.Context, *ModelContext) error { return nil }
func (BaseHook) AfterModel(context.Context, *ModelContext) error  { return nil }
func (BaseHook) BeforeTool(context.Context, *ToolContext) error   { return nil }
func (BaseHook) AfterTool(context.Context, *ToolContext) error    { return nil }
