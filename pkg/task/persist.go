package task

import (
	"digital-labor/pkg/workspace"
	"encoding/json"
	"fmt"
	"log/slog"
)

func persist() {
	data := make([]byte, 0, 1024)
	var err error
	for {
		var event TaskEvent
		var ok bool
		select {
		case event, ok = <-globalTaskManager.eventChan:
			if !ok {
				continue
			}
			data, err = json.Marshal(event)
			if err != nil {
				slog.Error("failed to marshal task event", "error", err)
				continue
			}
		default:
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
