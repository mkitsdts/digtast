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
	Mode         string                  `json:"-"` // 本地运行还是云端运行
	Memory       MemoryConfig            `json:"memory"`
	FTP          FTPConfig               `json:"ftp"`
	VNC          VNCConfig               `json:"vnc"`
	Models       map[string]ModelConfig  `json:"model"`
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
	ModelNames []string `json:"model_names"` // 具体的模型名称列表，如 ["gpt-3.5-turbo", "gpt-4"]
	Provider   string   `json:"provider"`
	URL        string   `json:"url"`
	Key        string   `json:"key"`
}

// FindModelConfig searches all configured models for the given model name.
func FindModelConfig(modelName string) (ModelConfig, bool) {
	for _, mCfg := range Conf.Models {
		for _, name := range mCfg.ModelNames {
			if name == modelName {
				return mCfg, true
			}
		}
	}
	return ModelConfig{}, false
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

func SaveConfig(name string) error {
	configPath := filepath.Join(Conf.WorkSpaceDir, name, "config.json")

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
