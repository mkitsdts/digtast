package state

import (
	"fmt"
	"log/slog"
	"time"
)

// ExtractAndSave appends a summarization result to the agent's memory.md.
func ExtractAndSave(sm *StateManager, summary string) {
	if summary == "" {
		return
	}

	entry := fmt.Sprintf("\n\n## %s\n%s\n", time.Now().Format("2006-01-02 15:04"), summary)

	if err := sm.SaveMemory(entry); err != nil {
		slog.Error("failed to save memory", "agent_id", sm.agentID, "error", err)
	}
}
