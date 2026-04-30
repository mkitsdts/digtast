package state

import (
	"crypto/sha256"
	"digital-labor/pkg/workspace"
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

// StateManager manages the system prompt cache and per-agent memory for one agent.
type StateManager struct {
	agentID      string
	mu           sync.RWMutex
	cachedPrompt string
	promptHash   [32]byte
}

// NewStateManager creates a StateManager for the given agent.
func NewStateManager(agentID string) *StateManager {
	return &StateManager{
		agentID: agentID,
	}
}

// Build returns the full system prompt, using cache if the source files have not changed.
// The prompt is composed of all workspace prompt creators plus the per-agent memory.md.
func (sm *StateManager) Build() string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	promptCreators := workspace.GetPromptCreators()
	var parts []string

	for _, creator := range promptCreators {
		content, err := creator.GetPromptImpl()
		if err != nil {
			slog.Error("failed to get prompt content", "name", creator.GetPromptName(), "error", err)
			continue
		}
		if content != "" {
			parts = append(parts, fmt.Sprintf("### %s\n%s", creator.GetPromptName(), content))
		}
	}

	// Append per-agent memory
	if memory := workspace.LoadAgentMemory(sm.agentID); memory != "" {
		parts = append(parts, "### memory.md\n"+memory)
	}

	fullPrompt := strings.Join(parts, "\n\n")
	hash := sha256.Sum256([]byte(fullPrompt))

	if sm.cachedPrompt != "" && hash == sm.promptHash {
		return sm.cachedPrompt
	}

	sm.cachedPrompt = fullPrompt
	sm.promptHash = hash
	return fullPrompt
}

// Invalidate forces the next Build() to re-read from disk.
func (sm *StateManager) Invalidate() {
	sm.mu.Lock()
	sm.cachedPrompt = ""
	sm.promptHash = [32]byte{}
	sm.mu.Unlock()
}

// SaveMemory appends content to the agent's memory.md via the workspace layer.
func (sm *StateManager) SaveMemory(content string) error {
	if err := workspace.SaveAgentMemory(sm.agentID, content); err != nil {
		return err
	}
	sm.Invalidate()
	return nil
}

// LoadMemory reads the agent's memory.md content.
func (sm *StateManager) LoadMemory() string {
	return workspace.LoadAgentMemory(sm.agentID)
}
