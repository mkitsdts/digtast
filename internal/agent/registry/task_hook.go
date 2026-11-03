package registry

import (
	"context"
	"errors"
	"sync"

	"digital-labor/internal/agent/loop"
	"digital-labor/internal/agent/task"
)

type TaskHook struct {
	agentloop.BaseHook
	steps sync.Map
}

func (h *TaskHook) BeforeTool(ctx context.Context, tc *agentloop.ToolContext) error {
	taskID := task.IDFromContext(ctx)
	if taskID == "" {
		return nil
	}
	step := task.CreateStep(taskID, task.StepConfig{Status: task.StatusRunning, Input: tc.Call})
	if step == nil {
		return errors.New("failed to create task step")
	}
	h.steps.Store(taskID+":"+tc.Call.ID, step.ID)
	return nil
}

func (h *TaskHook) AfterTool(ctx context.Context, tc *agentloop.ToolContext) error {
	taskID := task.IDFromContext(ctx)
	value, ok := h.steps.LoadAndDelete(taskID + ":" + tc.Call.ID)
	if !ok {
		return nil
	}
	stepID, _ := value.(string)
	config := task.StepConfig{Status: task.StatusCompleted, Output: tc.Result}
	if tc.Error != nil {
		config.Status = task.StatusFailed
		config.Error = tc.Error.Error()
	}
	task.UpdateStep(taskID, stepID, config)
	return nil
}

func init() { RegisterHandler(&TaskHook{}) }
