package conf

func DefaultConfig() Config {
	return Config{
		Memory: MemoryConfig{
			MaxMessagesSize: 10 * 1024 * 1024, // 10MB
		},
		FTP: FTPConfig{
			Port:    2121,
			Enabled: false,
		},
		VNC: VNCConfig{
			BeginPort: 6901,
			MaxUser:   10,
			Enabled:   false,
		},
		Models: map[string]ModelConfig{},
		Agents: map[string]AgentConfig{},
	}
}
