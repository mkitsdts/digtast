package task

import (
	"digital-labor/pkg/errs"
	"digital-labor/pkg/model"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TaskEvent struct {
	Type      string `json:"type"`
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
	Data      any    `json:"data"`
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

func CreateTask(cfg Config) (*model.Task, error) {
	if cfg.AgentID == "" {
		return nil, errs.ErrAgentIDRequired
	}
	if cfg.SessionID == "" {
		return nil, errs.ErrSessionIDRequired
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
		AgentID:      cfg.AgentID,
		Status:       model.TaskStatusPending,
		NotifyPolicy: cfg.NotifyPolicy,
		CreatedAt:    now,
		Error:        cfg.Error,
	}

	globalTaskManager.mux.Lock()
	if globalTaskManager.Tasks[cfg.SessionID] == nil {
		globalTaskManager.Tasks[cfg.SessionID] = make(map[string]*model.Task)
	}
	globalTaskManager.Tasks[cfg.SessionID][id] = &task
	globalTaskManager.mux.Unlock()

	globalTaskManager.eventChan <- TaskEvent{
		Type:      "task",
		AgentID:   cfg.AgentID,
		SessionID: cfg.SessionID,
		Data:      task,
	}

	return &task, nil
}

func UpdateStatus(sessionid, taskid string, status string) {
	globalTaskManager.mux.Lock()
	var task *model.Task
	if sess, ok := globalTaskManager.Tasks[sessionid]; ok {
		if t, ok := sess[taskid]; ok {
			t.Status = status
			task = t
		}
	}
	globalTaskManager.mux.Unlock()

	if task != nil {
		globalTaskManager.eventChan <- TaskEvent{
			Type:      "task",
			AgentID:   task.AgentID,
			SessionID: sessionid,
			Data:      *task,
		}
	}
}

func CreateStep(taskid, sessionid string, cfg StepConfig) *model.Step {
	globalTaskManager.mux.Lock()
	if _, ok := globalTaskManager.Tasks[sessionid]; !ok {
		globalTaskManager.mux.Unlock()
		return nil
	}
	step := model.Step{
		ID:        uuid.New().String(),
		SessionID: sessionid,
		TaskID:    taskid,
		Status:    cfg.Status,
		ParentID:  cfg.ParentID,
		Input:     cfg.Input,
		Output:    cfg.Output,
		Error:     cfg.Error,
		CreatedAt: cfg.CreatedAt,
		UpdatedAt: time.Now().UTC(),
	}
	var agentID string
	if task, ok := globalTaskManager.Tasks[sessionid][taskid]; ok {
		task.Steps = append(task.Steps, step)
		agentID = task.AgentID
	}
	globalTaskManager.mux.Unlock()

	if agentID != "" {
		globalTaskManager.eventChan <- TaskEvent{
			Type:      "step",
			AgentID:   agentID,
			SessionID: sessionid,
			Data:      step,
		}
	}
	return &step
}

func UpdateStep(sessionid, taskid, stepid string, cfg StepConfig) *model.Step {
	globalTaskManager.mux.Lock()
	var step *model.Step
	var agentID string
	if sess, ok := globalTaskManager.Tasks[sessionid]; ok {
		if task, ok := sess[taskid]; ok {
			agentID = task.AgentID
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

	if step != nil && agentID != "" {
		globalTaskManager.eventChan <- TaskEvent{
			Type:      "step",
			AgentID:   agentID,
			SessionID: sessionid,
			Data:      *step,
		}
	}
	return step
}

func GetTask(sessionid, taskid string) *model.Task {
	globalTaskManager.mux.RLock()
	defer globalTaskManager.mux.RUnlock()
	if _, ok := globalTaskManager.Tasks[sessionid]; !ok {
		return nil
	}
	if task, ok := globalTaskManager.Tasks[sessionid][taskid]; ok {
		return task
	}
	return nil
}
