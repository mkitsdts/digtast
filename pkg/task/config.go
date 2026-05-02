package task

import "time"

type Config struct {
	ParentFlowID string
	ParentTaskID string
	AgentID      string
	NotifyPolicy string
	Title        string
	Description  string
	Error        string
}

type UpdateConfig struct {
	Status      string
	Title       string
	Description string
	Error       string
}

type StepConfig struct {
	Status    string
	ParentID  string
	Input     any
	Output    any
	Error     string
	CreatedAt time.Time
}
