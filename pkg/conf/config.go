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
	MemoryDir    string       `json:"memory_dir"`
	SkillDir     string       `json:"skill_dir"`
	RootDir      string       `json:"root_dir"`
	WorkSpaceDir string       `json:"workspace_dir"`
	Memory       MemoryConfig `json:"memory"`
	FTP          FTPConfig    `json:"ftp"`
	VNC          VNCConfig    `json:"vnc"`
	Model        ModelConfig  `json:"model"`
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
	DefaultProvider string `json:"default_provider"`
	DefaultModel    string `json:"default_model"`
	DefaultURL      string `json:"default_url"`
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
