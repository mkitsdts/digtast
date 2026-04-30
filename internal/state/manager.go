package state

import (
	"crypto/sha256"
	"digital-labor/pkg/workspace"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// StateManager manages the system prompt cache and per-agent memory for one agent.
type StateManager struct {
	agentID      string
	mu           sync.RWMutex
	cachedPrompt string
	promptHash   [32]byte
	memoryPath   string
}

// NewStateManager creates a StateManager for the given agent.
func NewStateManager(agentID string) *StateManager {
	memDir := filepath.Join(workspace.GetWorkspacePath(), "memory", agentID)
	return &StateManager{
		agentID:    agentID,
		memoryPath: filepath.Join(memDir, "memory.md"),
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
	if memory := sm.loadMemory(); memory != "" {
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

// SaveMemory appends content to the agent's memory.md file.
func (sm *StateManager) SaveMemory(content string) error {
	dir := filepath.Dir(sm.memoryPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create memory dir: %w", err)
	}

	f, err := os.OpenFile(sm.memoryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open memory file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("write memory: %w", err)
	}

	sm.Invalidate()
	return nil
}

// LoadMemory reads the agent's memory.md content. Returns empty string if not found.
func (sm *StateManager) LoadMemory() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.loadMemory()
}

func (sm *StateManager) loadMemory() string {
	data, err := os.ReadFile(sm.memoryPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
