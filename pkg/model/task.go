package model

import "time"

const (
	StatusRunning   string = "running"
	StatusStreaming string = "streaming"
	StatusDone      string = "done"
	StatusFailed    string = "failed"
)

type Step struct {
	ID        string
	TaskID    string
	AgentID   string
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
	AgentID      string
	Status       string
	NotifyPolicy string
	CreatedAt    int64
	StartedAt    int64
	EndedAt      int64
	Error        string
	Steps        []Step
}

const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)
