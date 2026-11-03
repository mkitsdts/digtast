package task

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

type Step struct {
	ID        string
	TaskID    string
	Status    string
	ParentID  string
	Input     any
	Output    any
	Error     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Task struct {
	TaskID       string
	ParentFlowID string
	ParentTaskID string
	Status       string
	NotifyPolicy string
	Title        string
	Description  string
	CreatedAt    int64
	StartedAt    int64
	EndedAt      int64
	Error        string
	Steps        []Step
}

type TaskEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type runtime struct {
	mu     sync.RWMutex
	tasks  map[string]*Task
	events chan TaskEvent
}

var tasks = &runtime{tasks: make(map[string]*Task), events: make(chan TaskEvent, 1000)}

type contextKey struct{}

func WithContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

func IDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(contextKey{}).(string)
	return id
}

func Create(cfg Config) *Task {
	now := time.Now().UTC().Unix()
	if cfg.ParentFlowID == "" {
		cfg.ParentFlowID = "root"
	}
	if cfg.ParentTaskID == "" {
		cfg.ParentTaskID = "root"
	}
	t := &Task{
		TaskID: uuid.NewString(), ParentFlowID: cfg.ParentFlowID, ParentTaskID: cfg.ParentTaskID,
		Status: StatusPending, NotifyPolicy: cfg.NotifyPolicy, Title: cfg.Title,
		Description: cfg.Description, CreatedAt: now, Error: cfg.Error,
	}
	tasks.mu.Lock()
	tasks.tasks[t.TaskID] = t
	tasks.mu.Unlock()
	tasks.events <- TaskEvent{Type: "task", Data: *t}
	return t
}

func UpdateStatus(id, status string) { Update(id, UpdateConfig{Status: status}) }

func Update(id string, cfg UpdateConfig) *Task {
	tasks.mu.Lock()
	t := tasks.tasks[id]
	if t != nil {
		now := time.Now().UTC().Unix()
		if cfg.Status != "" {
			t.Status = cfg.Status
			if cfg.Status == StatusRunning && t.StartedAt == 0 {
				t.StartedAt = now
			}
			if cfg.Status == StatusCompleted || cfg.Status == StatusFailed {
				t.EndedAt = now
			}
		}
		if cfg.Title != "" {
			t.Title = cfg.Title
		}
		if cfg.Description != "" {
			t.Description = cfg.Description
		}
		if cfg.Error != "" {
			t.Error = cfg.Error
		}
	}
	tasks.mu.Unlock()
	if t != nil {
		tasks.events <- TaskEvent{Type: "task", Data: *t}
	}
	return t
}

func CreateStep(taskID string, cfg StepConfig) *Step {
	tasks.mu.Lock()
	t := tasks.tasks[taskID]
	if t == nil {
		tasks.mu.Unlock()
		return nil
	}
	step := Step{
		ID: uuid.NewString(), TaskID: taskID, Status: cfg.Status, ParentID: cfg.ParentID,
		Input: cfg.Input, Output: cfg.Output, Error: cfg.Error,
		CreatedAt: cfg.CreatedAt, UpdatedAt: time.Now().UTC(),
	}
	t.Steps = append(t.Steps, step)
	tasks.mu.Unlock()
	tasks.events <- TaskEvent{Type: "step", Data: step}
	return &step
}

func UpdateStep(taskID, stepID string, cfg StepConfig) *Step {
	tasks.mu.Lock()
	t := tasks.tasks[taskID]
	var step *Step
	if t != nil {
		for i := range t.Steps {
			if t.Steps[i].ID != stepID {
				continue
			}
			if cfg.Status != "" {
				t.Steps[i].Status = cfg.Status
			}
			if cfg.Output != nil {
				t.Steps[i].Output = cfg.Output
			}
			if cfg.Error != "" {
				t.Steps[i].Error = cfg.Error
			}
			t.Steps[i].UpdatedAt = time.Now().UTC()
			step = &t.Steps[i]
			break
		}
	}
	tasks.mu.Unlock()
	if step != nil {
		tasks.events <- TaskEvent{Type: "step", Data: *step}
	}
	return step
}
