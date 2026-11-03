package config

import (
	"os"
	"path"
)

func DefaultConfig() Config {
	workspacePath, _ := os.UserHomeDir()

	if workspacePath == "" {
		workspacePath, _ = os.Getwd()
	}

	workspacePath = path.Join(workspacePath, "/.digtast")

	return Config{
		WorkSpaceDir: workspacePath,
		Memory: MemoryConfig{
			MaxMessagesSize: 10 * 1024 * 1024, // 10MB
			TokenLimit:      32000,
			ChunkTokenLimit: 4000,
			CleanInterval:   10,
		},
		Models: map[string]ModelConfig{},
		Browser: BrowserConfig{
			Enabled: false,
		},
		Channels: map[string]map[string]any{},
	}
}
