package conf

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	IS_DEBUG = true
)

type Config struct {
	WorkSpaceDir string                  `json:"workspace_dir"`
	Memory       MemoryConfig            `json:"memory"`
	FTP          FTPConfig               `json:"ftp"`
	VNC          VNCConfig               `json:"vnc"`
	Models       map[string]ModelConfig  `json:"model"`
	Agents       map[string]AgentConfig  `json:"agents"`
	State        LocalRunningStateConfig `json:"state"`
}

type AgentConfig struct {
	Name      string `json:"name"`
	ModelName string `json:"model_name"`
}

type MemoryConfig struct {
	MaxMessagesSize int64 `json:"max_messages_size"`
}

type FTPConfig struct {
	Port    int    `json:"port"`
	RootDir string `json:"root_dir"`
	Enabled bool   `json:"enabled"`
}

type VNCConfig struct {
	BeginPort int  `json:"begin_port"`
	MaxUser   int  `json:"max_user"`
	Enabled   bool `json:"enabled"`
}

type ModelConfig struct {
	ModelName string `json:"model_name"`
	Provider  string `json:"provider"`
	URL       string `json:"url"`
	Key       string `json:"key"`
}

type LocalRunningStateConfig struct {
	LastUsedAgent string `json:"last_used_agent"` // 上次使用的 AgentID
	LastSession   string `json:"last_session"`    // 上次使用的 Session
}

var Conf Config

func LoadConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("Config file not found, using default config", "path", configPath)
			Conf = DefaultConfig()
			return SaveConfig(configPath)
		}
		return err
	}

	err = json.Unmarshal(data, &Conf)
	if err != nil {
		return err
	}
	return nil
}

func SaveConfig(configPath string) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(Conf, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
