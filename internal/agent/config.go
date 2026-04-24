package agent

import (
	"digital-labor/pkg/workspace"
	"encoding/json"
	"log/slog"
	"os"
)

type DigitalAgentConfig struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Model       string `json:"model"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Provider    string `json:"provider"`
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
