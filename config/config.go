package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

type Config struct {
	WorkSpaceDir string                    `json:"workspace_dir"`
	Agent        *AgentConfig              `json:"agent,omitempty"`
	Memory       MemoryConfig              `json:"memory"`
	Models       map[string]ModelConfig    `json:"model"`
	Skills       SkillsConfig              `json:"skills"`
	Channels     map[string]map[string]any `json:"channels"`
	Browser      BrowserConfig             `json:"browser"`
}

// SkillsConfig 记录扫描到的 skill 的启用状态；Disabled 中的 skill 不会注入
// 提示词，也不允许通过 load_skill 加载。
type SkillsConfig struct {
	Disabled []string `json:"disabled,omitempty"`
}

// IsSkillDisabled 报告某个 skill 是否处于停用状态。
func IsSkillDisabled(name string) bool {
	return slices.Contains(Conf.Skills.Disabled, name)
}

// SetSkillEnabled 更新 skill 的启用状态并落盘。
func SetSkillEnabled(name string, enabled bool) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("skill name is empty")
	}
	disabled := make([]string, 0, len(Conf.Skills.Disabled))
	for _, item := range Conf.Skills.Disabled {
		if item != name {
			disabled = append(disabled, item)
		}
	}
	if !enabled {
		disabled = append(disabled, name)
	}
	Conf.Skills.Disabled = disabled
	return SaveConfig()
}

type BrowserConfig struct {
	Enabled bool `json:"enabled"`
}

type AgentConfig struct {
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	Description string `json:"description,omitempty"`
}

type MemoryConfig struct {
	MaxMessagesSize int64 `json:"max_messages_size"`
	TokenLimit      int   `json:"token_limit"`       // 触发对话压缩的 token 阈值
	ChunkTokenLimit int   `json:"chunk_token_limit"` // 提取分片的 token 上限
	CleanInterval   int   `json:"clean_interval"`    // memory.md 清理间隔（轮次）
}

// ModelConfig 保存某个厂商（Provider）的接入配置，Models 的键为 Provider 名
// （与模型注册表一致）。URL 非空时替换厂商的默认基础地址，Models 为启用列表。
type ModelConfig struct {
	Key    string   `json:"key"`
	URL    string   `json:"url,omitempty"`
	Models []string `json:"models,omitempty"`
}

// ModelCredentials 返回某个 Provider 的凭据配置，未配置时返回零值。
func ModelCredentials(provider string) ModelConfig {
	return Conf.Models[provider]
}

// SaveModelCredentials 保存某个 Provider 的凭据并落盘。
func SaveModelCredentials(provider string, cfg ModelConfig) error {
	if Conf.Models == nil {
		Conf.Models = make(map[string]ModelConfig)
	}
	Conf.Models[provider] = cfg
	return SaveConfig()
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

// storedConfig 同时兼容新旧两种落盘格式。旧格式的模型以 uuid 为键并内嵌
// provider/model_names，Agent 记录 model 而非 provider；新格式的模型直接以
// Provider 为键，Agent 记录 provider + model。
type storedConfig struct {
	WorkSpaceDir string                    `json:"workspace_dir"`
	Memory       MemoryConfig              `json:"memory"`
	Skills       SkillsConfig              `json:"skills"`
	Channels     map[string]map[string]any `json:"channels"`
	Browser      BrowserConfig             `json:"browser"`
	Models       map[string]storedModel    `json:"model"`
	Agent        *storedAgent              `json:"agent"`
}

// storedModel 兼容新旧字段：新格式用 models（启用列表），旧格式用 provider +
// model_names，以 provider 是否非空区分。
type storedModel struct {
	Key        string   `json:"key"`
	URL        string   `json:"url,omitempty"`
	Models     []string `json:"models,omitempty"`
	Provider   string   `json:"provider,omitempty"`
	ModelNames []string `json:"model_names,omitempty"`
}

// storedAgent 兼容新旧字段：旧格式的 model 指向模型配置的 uuid 键，真实模型
// 名在 model_kind；新格式直接是 provider + model。
type storedAgent struct {
	Name        string `json:"name"`
	Provider    string `json:"provider,omitempty"`
	Model       string `json:"model"`
	ModelKind   string `json:"model_kind,omitempty"`
	Description string `json:"description"`
}

// toConfig 把落盘配置归一为新结构，返回配置是否来自旧格式（需要回写）。
func (s storedConfig) toConfig() (Config, bool) {
	conf := Config{
		WorkSpaceDir: s.WorkSpaceDir, Memory: s.Memory, Skills: s.Skills,
		Channels: s.Channels, Browser: s.Browser,
		Models: make(map[string]ModelConfig, len(s.Models)),
	}
	legacy := false
	legacyProviderOf := map[string]string{} // 旧格式 uuid 键 → provider，供 Agent 迁移

	// 排序保证同一 Provider 出现多条旧配置时迁移结果确定。
	keys := make([]string, 0, len(s.Models))
	for key := range s.Models {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		model := s.Models[key]
		provider := strings.ToLower(strings.TrimSpace(model.Provider))
		if provider == "" {
			conf.Models[key] = ModelConfig{Key: model.Key, URL: model.URL, Models: model.Models}
			continue
		}
		// 旧格式：按 provider 归并凭据与启用列表。
		legacy = true
		legacyProviderOf[key] = provider
		entry := conf.Models[provider]
		if model.Key != "" && entry.Key == "" {
			entry.Key, entry.URL = model.Key, model.URL
		}
		entry.Models = mergeUnique(entry.Models, model.ModelNames)
		conf.Models[provider] = entry
	}
	if s.Agent != nil {
		agent := AgentConfig{
			Name: s.Agent.Name, Provider: s.Agent.Provider, Model: s.Agent.Model, Description: s.Agent.Description,
		}
		if s.Agent.ModelKind != "" && s.Agent.ModelKind != "default" {
			// 旧格式中 model_kind 才是真实的模型名。
			agent.Model = s.Agent.ModelKind
			legacy = true
		}
		if agent.Provider == "" && s.Agent.Model != "" {
			if provider, ok := legacyProviderOf[s.Agent.Model]; ok {
				agent.Provider = provider
				legacy = true
			} else if len(conf.Models) == 1 {
				// 只配置过一个厂商时可以唯一确定。
				for provider := range conf.Models {
					agent.Provider = provider
				}
			}
		}
		conf.Agent = &agent
	}
	return conf, legacy
}

// mergeUnique 合并两个列表，保序去重。
func mergeUnique(base, extra []string) []string {
	seen := make(map[string]struct{}, len(base)+len(extra))
	merged := make([]string, 0, len(base)+len(extra))
	for _, item := range append(append([]string(nil), base...), extra...) {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		merged = append(merged, item)
	}
	return merged
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

	var stored storedConfig
	if err := json.Unmarshal(data, &stored); err != nil {
		slog.Info("Config file parse failed, using default config", "name", path)
		Conf = DefaultConfig()
		return SaveConfig()
	}
	conf, legacy := stored.toConfig()
	if conf.WorkSpaceDir == "" {
		conf.WorkSpaceDir = Conf.WorkSpaceDir
	}
	Conf = conf
	if legacy {
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
