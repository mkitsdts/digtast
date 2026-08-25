package registry

import (
	"context"
	"errors"
	"sync"

	"digital-labor/pkg/agentloop"
	"digital-labor/pkg/model"
	"digital-labor/pkg/task"
)

type TaskHook struct {
	agentloop.BaseHook
	steps sync.Map
}

func (h *TaskHook) BeforeTool(ctx context.Context, tc *agentloop.ToolContext) error {
	taskID, _ := ctx.Value("task_id").(string)
	agentID, _ := ctx.Value("agent_id").(string)
	if taskID == "" || agentID == "" {
		return nil
	}
	step := task.CreateStep(taskID, agentID, task.StepConfig{Status: model.StatusRunning, Input: tc.Call})
	if step == nil {
		return errors.New("failed to create task step")
	}
	h.steps.Store(taskID+":"+tc.Call.ID, step.ID)
	return nil
}

func (h *TaskHook) AfterTool(ctx context.Context, tc *agentloop.ToolContext) error {
	taskID, _ := ctx.Value("task_id").(string)
	agentID, _ := ctx.Value("agent_id").(string)
	value, ok := h.steps.LoadAndDelete(taskID + ":" + tc.Call.ID)
	if !ok {
		return nil
	}
	stepID, _ := value.(string)
	config := task.StepConfig{Status: model.StatusDone, Output: tc.Result}
	if tc.Error != nil {
		config.Status = model.StatusFailed
		config.Error = tc.Error.Error()
	}
	task.UpdateStep(agentID, taskID, stepID, config)
	return nil
}

func init() { RegisterHandler(&TaskHook{}) }
