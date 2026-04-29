package model

type DigitalAgentConfig struct {
	ID          string `json:"id"`          // unique key to tag agent
	Name        string `json:"name"`        // human-readable name of the agent
	Model       string `json:"model"`       // key into global Models config (e.g., "deepseek")
	ModelKind   string `json:"modelKind"`   // kind of model (e.g., "doubao-seed-2-0-lite-260215")
	Description string `json:"description"` // human-readable description of the agent
}
