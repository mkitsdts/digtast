package registry

import (
	"context"
	"digital-labor/pkg/model"
	"digital-labor/pkg/task"
	"errors"
	"log/slog"

	"github.com/cloudwego/eino/compose"
)

func Invokable(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
	return func(ctx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
		// 前置处理
		taskId, _ := ctx.Value("task_id").(string)
		sessionId, _ := ctx.Value("session_id").(string)
		if taskId == "" || sessionId == "" {
			slog.Error("missing context values for tool middleware", "task_id", taskId, "session_id", sessionId)
			return nil, errors.New("task_id and session_id must be set in context")
		}

		// TODO: 如果后续加了子任务，需要判断任务名
		step := task.CreateStep(taskId, sessionId, task.StepConfig{
			Status: model.StatusRunning,
			Input:  input,
		})
		if step == nil {
			slog.Error("failed to create step", "task_id", taskId, "session_id", sessionId)
			return nil, errors.New("failed to create step")
		}

		// 调用真正的工具逻辑
		output, err := next(ctx, input)

		// 后置处理
		if err != nil {
			task.UpdateStep(taskId, sessionId, step.ID, task.StepConfig{
				Status: model.StatusFailed,
				Error:  err.Error(),
			})
			return nil, err
		}
		if output != nil {
			task.UpdateStep(taskId, sessionId, step.ID, task.StepConfig{
				Output: output,
				Status: model.StatusDone,
				Error:  output.Result,
			})
		}
		return output, err
	}
}
