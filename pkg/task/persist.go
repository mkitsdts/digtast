package task

import (
	"digital-labor/pkg/workspace"
	"encoding/json"
	"fmt"
	"log/slog"
)

func persist() {
	for event := range globalTaskManager.eventChan {
		data, err := json.Marshal(event)
		if err != nil {
			slog.Error("failed to marshal task event", "error", err)
			continue
		}

		// Use .jsonl extension for incremental append
		filename := fmt.Sprintf("tasks/%s/%s.jsonl", event.AgentID, event.SessionID)
		workspace.AppendFile(filename, data)
	}
}

func init() {
	go persist()
}
