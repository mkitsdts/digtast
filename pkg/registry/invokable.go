package registry

import (
	"context"
	"digital-labor/pkg/model"
	"digital-labor/pkg/task"
	"errors"
	"log/slog"

	"github.com/cloudwego/eino/compose"
)

// Tool Call middleware that tracks tool execution steps
func InvokableTool(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
	return func(ctx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
		taskId, _ := ctx.Value("task_id").(string)
		agentId, _ := ctx.Value("agent_id").(string)
		if taskId == "" || agentId == "" {
			slog.Error("missing context values for tool middleware", "task_id", taskId, "agent_id", agentId)
			return nil, errors.New("task_id and agent_id must be set in context")
		}

		step := task.CreateStep(taskId, agentId, task.StepConfig{
			Status: model.StatusRunning,
			Input:  input,
		})
		if step == nil {
			slog.Error("failed to create step", "task_id", taskId, "agent_id", agentId)
			return nil, errors.New("failed to create step")
		}

		output, err := next(ctx, input)

		if err != nil {
			task.UpdateStep(agentId, taskId, step.ID, task.StepConfig{
				Status: model.StatusFailed,
				Error:  err.Error(),
			})
			return nil, err
		}
		if output != nil {
			task.UpdateStep(agentId, taskId, step.ID, task.StepConfig{
				Output: output,
				Status: model.StatusDone,
				Error:  output.Result,
			})
		}
		return output, err
	}
}
