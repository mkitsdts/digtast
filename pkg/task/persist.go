package task

import (
	"digital-labor/pkg/workspace"
	"encoding/json"
	"log/slog"
)

func persist() {
	for {
		if !globalTaskManager.is_dirty {
			return
		}

		data, err := json.Marshal(globalTaskManager.Tasks)
		if err != nil {
			slog.Error("failed to persist tasks")
			return
		}

		go workspace.PersistFile("tasks.json", data)
		globalTaskManager.is_dirty = false
	}
}

func init() {
	go persist()
}
