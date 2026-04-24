package agent

import (
	"digital-labor/pkg/workspace"
	"encoding/json"
	"log/slog"
	"os"
)

type DigitalAgentConfig struct {
	ID          string `json:"id"`          // unique key to tag agent
	Key         string `json:"key"`         // llm api key
	Name        string `json:"name"`        // human-readable name of the agent
	Model       string `json:"model"`       // llm model to use
	Description string `json:"description"` // human-readable description of the agent
	URL         string `json:"url"`         // llm base url
	Provider    string `json:"provider"`    // llm provider
}

func (dac *DigitalAgentConfig) persistDigitalAgentConfig(times int) {
	if times > 10 {
		return
	}
	ch := make(chan error)
	go func() {
		data, err := json.Marshal(dac)
		if err != nil {
			ch <- err
			return
		}
		path := workspace.GetWorkspacePath() + "/config.json"

		err = os.WriteFile(path, data, 0644)
		if err != nil {
			ch <- err
			return
		}
		ch <- nil
	}()

	go func() {
		err := <-ch
		if err != nil {
			slog.Error("failed to persist digital agent config", "error", err)
			dac.persistDigitalAgentConfig(times + 1)
		}
	}()

}
