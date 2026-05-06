package conf

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
		Mode:         "local",
		Memory: MemoryConfig{
			MaxMessagesSize: 10 * 1024 * 1024, // 10MB
			TokenLimit:      32000,
			ChunkTokenLimit: 4000,
			CleanInterval:   10,
		},
		FTP: FTPConfig{
			Port:    2121,
			Enabled: false,
		},
		Models: map[string]ModelConfig{},
		State: LocalRunningStateConfig{
			LastUsedAgent: "",
		},
	}
}
