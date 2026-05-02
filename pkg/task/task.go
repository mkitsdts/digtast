package task

import (
	"digital-labor/pkg/errs"
	"digital-labor/pkg/model"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TaskEvent struct {
	Type    string `json:"type"`
	AgentID string `json:"agent_id"`
	Data    any    `json:"data"`
}

type TaskManager struct {
	mux             sync.RWMutex
	Tasks           map[string]map[string]*model.Task `json:"tasks"`
	eventChan       chan TaskEvent
	listenerMu      sync.RWMutex
	streamListeners map[string][]chan TaskEvent
}

var globalTaskManager = &TaskManager{
	mux:             sync.RWMutex{},
	Tasks:           map[string]map[string]*model.Task{},
	eventChan:       make(chan TaskEvent, 1000),
	streamListeners: map[string][]chan TaskEvent{},
}

func CreateTask(agentID string, cfg Config) (*model.Task, error) {
	if agentID == "" {
		return nil, errs.ErrAgentIDRequired
	}

	id := uuid.New().String()

	if cfg.ParentFlowID == "" {
		cfg.ParentFlowID = "root"
	}

	if cfg.ParentTaskID == "" {
		cfg.ParentTaskID = "root"
	}

	now := time.Now().UTC().Unix()

	task := model.Task{
		TaskID:       id,
		ParentFlowID: cfg.ParentFlowID,
		ParentTaskID: cfg.ParentTaskID,
		AgentID:      agentID,
		Status:       model.TaskStatusPending,
		NotifyPolicy: cfg.NotifyPolicy,
		Title:        cfg.Title,
		Description:  cfg.Description,
		CreatedAt:    now,
		Error:        cfg.Error,
	}

	globalTaskManager.mux.Lock()
	if globalTaskManager.Tasks[agentID] == nil {
		globalTaskManager.Tasks[agentID] = make(map[string]*model.Task)
	}
	globalTaskManager.Tasks[agentID][id] = &task
	globalTaskManager.mux.Unlock()

	globalTaskManager.eventChan <- TaskEvent{
		Type:    "task",
		AgentID: agentID,
		Data:    task,
	}

	return &task, nil
}

func UpdateStatus(agentID, taskid string, status string) {
	globalTaskManager.mux.Lock()
	var t *model.Task
	if agentTasks, ok := globalTaskManager.Tasks[agentID]; ok {
		if tk, ok := agentTasks[taskid]; ok {
			tk.Status = status
			now := time.Now().UTC().Unix()
			if status == model.TaskStatusRunning && tk.StartedAt == 0 {
				tk.StartedAt = now
			}
			if status == model.TaskStatusCompleted || status == model.TaskStatusFailed {
				tk.EndedAt = now
			}
			t = tk
		}
	}
	globalTaskManager.mux.Unlock()

	if t != nil {
		globalTaskManager.eventChan <- TaskEvent{
			Type:    "task",
			AgentID: agentID,
			Data:    *t,
		}
	}
}

func UpdateTask(agentID, taskid string, cfg UpdateConfig) *model.Task {
	globalTaskManager.mux.Lock()
	var t *model.Task
	if agentTasks, ok := globalTaskManager.Tasks[agentID]; ok {
		if tk, ok := agentTasks[taskid]; ok {
			if cfg.Status != "" {
				tk.Status = cfg.Status
				now := time.Now().UTC().Unix()
				if cfg.Status == model.TaskStatusRunning && tk.StartedAt == 0 {
					tk.StartedAt = now
				}
				if cfg.Status == model.TaskStatusCompleted || cfg.Status == model.TaskStatusFailed {
					tk.EndedAt = now
				}
			}
			if cfg.Title != "" {
				tk.Title = cfg.Title
			}
			if cfg.Description != "" {
				tk.Description = cfg.Description
			}
			if cfg.Error != "" {
				tk.Error = cfg.Error
			}
			t = tk
		}
	}
	globalTaskManager.mux.Unlock()

	if t != nil {
		globalTaskManager.eventChan <- TaskEvent{
			Type:    "task",
			AgentID: agentID,
			Data:    *t,
		}
	}
	return t
}

func CreateStep(taskid, agentID string, cfg StepConfig) *model.Step {
	globalTaskManager.mux.Lock()
	if _, ok := globalTaskManager.Tasks[agentID]; !ok {
		globalTaskManager.mux.Unlock()
		return nil
	}
	step := model.Step{
		ID:        uuid.New().String(),
		AgentID:   agentID,
		TaskID:    taskid,
		Status:    cfg.Status,
		ParentID:  cfg.ParentID,
		Input:     cfg.Input,
		Output:    cfg.Output,
		Error:     cfg.Error,
		CreatedAt: cfg.CreatedAt,
		UpdatedAt: time.Now().UTC(),
	}
	if task, ok := globalTaskManager.Tasks[agentID][taskid]; ok {
		task.Steps = append(task.Steps, step)
	}
	globalTaskManager.mux.Unlock()

	globalTaskManager.eventChan <- TaskEvent{
		Type:    "step",
		AgentID: agentID,
		Data:    step,
	}
	return &step
}

func UpdateStep(agentID, taskid, stepid string, cfg StepConfig) *model.Step {
	globalTaskManager.mux.Lock()
	var step *model.Step
	if agentTasks, ok := globalTaskManager.Tasks[agentID]; ok {
		if task, ok := agentTasks[taskid]; ok {
			for i, s := range task.Steps {
				if s.ID == stepid {
					if cfg.Status != "" {
						task.Steps[i].Status = cfg.Status
					}
					if cfg.Output != nil {
						task.Steps[i].Output = cfg.Output
					}
					if cfg.Error != "" {
						task.Steps[i].Error = cfg.Error
					}
					task.Steps[i].UpdatedAt = time.Now().UTC()
					step = &task.Steps[i]
					break
				}
			}
		}
	}
	globalTaskManager.mux.Unlock()

	if step != nil {
		globalTaskManager.eventChan <- TaskEvent{
			Type:    "step",
			AgentID: agentID,
			Data:    *step,
		}
	}
	return step
}

func GetTask(agentID, taskid string) *model.Task {
	globalTaskManager.mux.RLock()
	defer globalTaskManager.mux.RUnlock()
	if _, ok := globalTaskManager.Tasks[agentID]; !ok {
		return nil
	}
	if task, ok := globalTaskManager.Tasks[agentID][taskid]; ok {
		return task
	}
	return nil
}

func ListTasks(agentID string) []*model.Task {
	globalTaskManager.mux.RLock()
	defer globalTaskManager.mux.RUnlock()
	agentTasks, ok := globalTaskManager.Tasks[agentID]
	if !ok {
		return nil
	}
	result := make([]*model.Task, 0, len(agentTasks))
	for _, t := range agentTasks {
		result = append(result, t)
	}
	return result
}

func SubscribeEvents(agentID string) <-chan TaskEvent {
	ch := make(chan TaskEvent, 200)
	globalTaskManager.listenerMu.Lock()
	globalTaskManager.streamListeners[agentID] = append(
		globalTaskManager.streamListeners[agentID], ch,
	)
	globalTaskManager.listenerMu.Unlock()
	return ch
}

func UnsubscribeEvents(agentID string, ch <-chan TaskEvent) {
	globalTaskManager.listenerMu.Lock()
	listeners := globalTaskManager.streamListeners[agentID]
	for i, c := range listeners {
		if c == ch {
			globalTaskManager.streamListeners[agentID] = append(listeners[:i], listeners[i+1:]...)
			close(c)
			break
		}
	}
	globalTaskManager.listenerMu.Unlock()
}
