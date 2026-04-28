package task

import (
	"digital-labor/pkg/errs"
	"digital-labor/pkg/model"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TaskManager struct {
	mux      sync.RWMutex
	Tasks    map[string]map[string]*model.Task `json:"tasks"`
	is_dirty bool                              `json:"-"`
}

var globalTaskManager = &TaskManager{
	mux:   sync.RWMutex{},
	Tasks: map[string]map[string]*model.Task{},
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

	globalTaskManager.is_dirty = true

	return &task, nil
}

func UpdateStatus(sessionid, taskid string, status string) {
	globalTaskManager.mux.Lock()
	defer globalTaskManager.mux.Unlock()
	if _, ok := globalTaskManager.Tasks[sessionid]; !ok {
		return
	}
	if task, ok := globalTaskManager.Tasks[sessionid][taskid]; ok {
		task.Status = status
	}
	globalTaskManager.is_dirty = true
}

func CreateStep(sessionid, taskid string, cfg StepConfig) *model.Step {
	globalTaskManager.mux.Lock()
	defer globalTaskManager.mux.Unlock()
	if _, ok := globalTaskManager.Tasks[sessionid]; !ok {
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
	if task, ok := globalTaskManager.Tasks[sessionid][taskid]; ok {
		task.Steps = append(task.Steps, step)
	}
	globalTaskManager.is_dirty = true
	return &step
}

func UpdateStep(sessionid, taskid, stepid string, cfg StepConfig) *model.Step {
	globalTaskManager.mux.Lock()
	defer globalTaskManager.mux.Unlock()
	if _, ok := globalTaskManager.Tasks[sessionid]; !ok {
		return nil
	}
	if task, ok := globalTaskManager.Tasks[sessionid][taskid]; ok {
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
				return &task.Steps[i]
			}
		}
	}
	globalTaskManager.is_dirty = true
	return nil
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
