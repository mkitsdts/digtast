package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
)

type Config struct {
	WorkSpaceDir string                    `json:"workspace_dir"`
	Agent        *AgentConfig              `json:"agent,omitempty"`
	Memory       MemoryConfig              `json:"memory"`
	Models       map[string]ModelConfig    `json:"model"`
	Channels     map[string]map[string]any `json:"channels"`
	Browser      BrowserConfig             `json:"browser"`
}

type BrowserConfig struct {
	Enabled bool `json:"enabled"`
}

type AgentConfig struct {
	Name        string `json:"name"`
	Model       string `json:"model"`
	ModelKind   string `json:"model_kind,omitempty"`
	Description string `json:"description,omitempty"`
}

type MemoryConfig struct {
	MaxMessagesSize int64 `json:"max_messages_size"`
	TokenLimit      int   `json:"token_limit"`       // 触发对话压缩的 token 阈值
	ChunkTokenLimit int   `json:"chunk_token_limit"` // 提取分片的 token 上限
	CleanInterval   int   `json:"clean_interval"`    // memory.md 清理间隔（轮次）
}

type ModelConfig struct {
	ModelNames []string `json:"model_names"` // 具体的模型名称列表，如 ["gpt-3.5-turbo", "gpt-4"]
	Provider   string   `json:"provider"`
	URL        string   `json:"url"`
	Key        string   `json:"key"`
}

func FindModelConfig(modelName string) (ModelConfig, bool) {
	if mCfg, ok := Conf.Models[modelName]; ok {
		return mCfg, true
	}

	for _, mCfg := range Conf.Models {
		for _, name := range mCfg.ModelNames {
			if name == modelName {
				return mCfg, true
			}
		}
	}
	return ModelConfig{}, false
}

func SetModel(name string, model ModelConfig) error {
	if Conf.Models == nil {
		Conf.Models = make(map[string]ModelConfig)
	}
	Conf.Models[name] = model
	return SaveConfig()
}

func RemoveModel(name string) error {
	delete(Conf.Models, name)
	return SaveConfig()
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
	slog.Debug("config loaded", "path", path)
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
