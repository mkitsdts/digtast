package task

import "time"

type Config struct {
	ParentFlowID string
	ParentTaskID string
	AgentID      string
	NotifyPolicy string
	Error        string
}

type StepConfig struct {
	Status    string
	ParentID  string
	Input     any
	Output    any
	Error     string
	CreatedAt time.Time
}
