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
	mux       sync.RWMutex
	Tasks     map[string]map[string]*model.Task `json:"tasks"`
	eventChan chan TaskEvent
}

var globalTaskManager = &TaskManager{
	mux:       sync.RWMutex{},
	Tasks:     map[string]map[string]*model.Task{},
	eventChan: make(chan TaskEvent, 1000),
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
