package task

import (
	"digital-labor/internal/agent/workspace"
	"encoding/json"
	"log/slog"
)

func persist() {
	for event := range tasks.events {
		data, err := json.Marshal(event)
		if err != nil {
			slog.Error("failed to marshal task event", "error", err)
			continue
		}

		workspace.AppendFile("tasks/tasks.jsonl", data)
	}
}

func init() {
	go persist()
}
