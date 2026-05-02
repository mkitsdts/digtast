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

		filename := fmt.Sprintf("tasks/%s/tasks.jsonl", event.AgentID)
		workspace.AppendFile(filename, data)

		globalTaskManager.listenerMu.RLock()
		for _, ch := range globalTaskManager.streamListeners[event.AgentID] {
			select {
			case ch <- event:
			default:
			}
		}
		globalTaskManager.listenerMu.RUnlock()
	}
}

func init() {
	go persist()
}
