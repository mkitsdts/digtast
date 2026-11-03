package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
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

// AddModel stores a new model configuration under a generated identifier and
// returns that identifier.
func AddModel(model ModelConfig) (string, error) {
	if Conf.Models == nil {
		Conf.Models = make(map[string]ModelConfig)
	}
	id := uuid.NewString()
	Conf.Models[id] = model
	return id, SaveConfig()
}

func UpdateModel(id string, model ModelConfig) error {
	if _, ok := Conf.Models[id]; !ok {
		return fmt.Errorf("model config not found: %s", id)
	}
	Conf.Models[id] = model
	return SaveConfig()
}

func RemoveModel(id string) error {
	delete(Conf.Models, id)
	return SaveConfig()
}

// FindModelByProvider returns a stored model config for a provider, preferring
// the lexicographically smallest key so the result is deterministic.
func FindModelByProvider(provider string) (ModelConfig, bool) {
	ids := make([]string, 0, len(Conf.Models))
	for id := range Conf.Models {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		cfg := Conf.Models[id]
		if normalizeProvider(cfg.Provider) == normalizeProvider(provider) {
			return cfg, true
		}
	}
	return ModelConfig{}, false
}

func normalizeProvider(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "ark" {
		return "doubao"
	}
	return provider
}

// SetAgent stores the selected model as the singleton agent configuration.
func SetAgent(cfg AgentConfig) error {
	Conf.Agent = &cfg
	return SaveConfig()
}

// ChannelEnabledKey is the channel parameter that controls whether a channel
// is started when the agent runs.
const ChannelEnabledKey = "enabled"

// SetChannel stores channel parameters, normalizing their value types through
// JSON so in-memory and reloaded configs behave the same (numbers become
// float64).
func SetChannel(kind string, params map[string]any) error {
	if Conf.Channels == nil {
		Conf.Channels = make(map[string]map[string]any)
	}
	data, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("marshal channel config: %w", err)
	}
	var normalized map[string]any
	if err := json.Unmarshal(data, &normalized); err != nil {
		return fmt.Errorf("normalize channel config: %w", err)
	}
	Conf.Channels[kind] = normalized
	return SaveConfig()
}

func RemoveChannel(kind string) error {
	delete(Conf.Channels, kind)
	return SaveConfig()
}

// IsChannelEnabled reports whether a configured channel should be started.
// Channels configured before the enabled flag existed default to enabled.
func IsChannelEnabled(kind string) bool {
	params, ok := Conf.Channels[kind]
	if !ok {
		return false
	}
	enabled, ok := params[ChannelEnabledKey].(bool)
	if !ok {
		return true
	}
	return enabled
}

func SetChannelEnabled(kind string, enabled bool) error {
	params, ok := Conf.Channels[kind]
	if !ok {
		return fmt.Errorf("channel config not found: %s", kind)
	}
	params[ChannelEnabledKey] = enabled
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
