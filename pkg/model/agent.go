package model

type DigitalAgentConfig struct {
	ID          string `json:"id"`          // unique key to tag agent
	Name        string `json:"name"`        // human-readable name of the agent
	Model       string `json:"model"`       // key into global Models config (e.g., "deepseek")
	Description string `json:"description"` // human-readable description of the agent
}
