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
	WorkSpaceDir string                    `json:"workspace_dir"`
	Mode         string                    `json:"-"` // 本地运行还是云端运行
	Memory       MemoryConfig              `json:"memory"`
	FTP          FTPConfig                 `json:"ftp"`
	VNC          VNCConfig                 `json:"vnc"`
	Models       map[string]ModelConfig    `json:"model"`
	State        LocalRunningStateConfig   `json:"state"`
	Channels     map[string]map[string]any `json:"channels"`
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
	// 1. Try exact match on map key
	if mCfg, ok := Conf.Models[modelName]; ok {
		return mCfg, true
	}

	// 2. Try match within ModelNames list
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
}

var Conf Config

func init() {
	homedir, _ := os.UserHomeDir()
	Conf.WorkSpaceDir = filepath.Join(homedir, ".digtast")
}

func LoadConfig() error {
	path := filepath.Join(Conf.WorkSpaceDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("Config file not found, using default config", "name", path)
			Conf = DefaultConfig()
			return SaveConfig()
		}
		return err
	}

	err = json.Unmarshal(data, &Conf)
	if err != nil {
		slog.Info("Config file parse failed, using default config", "name", path)
		Conf = DefaultConfig()
		return SaveConfig()
	}
	slog.Info("config loaded", "config", Conf)
	return nil
}

func SaveConfig() error {
	configPath := filepath.Join(Conf.WorkSpaceDir, "config.json")

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(&Conf, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
