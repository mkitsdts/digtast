package model

type DigitalAgentConfig struct {
	ID          string `json:"id"`          // unique key to tag agent
	Key         string `json:"key"`         // llm api key
	Name        string `json:"name"`        // human-readable name of the agent
	Model       string `json:"model"`       // llm model to use
	Description string `json:"description"` // human-readable description of the agent
	URL         string `json:"url"`         // llm base url
	Provider    string `json:"provider"`    // llm provider
}
