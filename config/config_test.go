package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndModelCredentials(t *testing.T) {
	previous := Conf
	t.Cleanup(func() { Conf = previous })
	Conf = DefaultConfig()
	Conf.WorkSpaceDir = t.TempDir()

	err := SaveModelCredentials("deepseek", ModelConfig{
		Key: "sk-1", URL: "https://gw.example", Models: []string{"deepseek-chat", "deepseek-reasoner"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := ModelCredentials("deepseek")
	if got.Key != "sk-1" || got.URL != "https://gw.example" || len(got.Models) != 2 {
		t.Fatalf("unexpected credentials %#v", got)
	}
	if got := ModelCredentials("openai"); got.Key != "" {
		t.Fatalf("expected empty credentials, got %#v", got)
	}
}

func TestLoadConfigMigratesLegacyModels(t *testing.T) {
	previous := Conf
	t.Cleanup(func() { Conf = previous })
	Conf = DefaultConfig()
	Conf.WorkSpaceDir = t.TempDir()

	legacy := `{
  "workspace_dir": "",
  "agent": {"name": "a1", "model": "8a2b41c0", "model_kind": "deepseek-chat", "description": "d"},
  "model": {
    "8a2b41c0": {"provider": "deepseek", "url": "", "key": "sk-old", "model_names": ["deepseek-chat", "deepseek-reasoner"]},
    "9f1c22d1": {"provider": "custom", "url": "https://gw.example/v1", "key": "sk-gw", "model_names": ["m1"]}
  },
  "memory": {"token_limit": 32000},
  "channels": {},
  "browser": {"enabled": false}
}`
	if err := os.WriteFile(filepath.Join(Conf.WorkSpaceDir, "config.json"), []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	if err := LoadConfig(); err != nil {
		t.Fatal(err)
	}
	got := Conf.Models["deepseek"]
	if got.Key != "sk-old" || len(got.Models) != 2 {
		t.Fatalf("deepseek config was not migrated: %#v", got)
	}
	if got := Conf.Models["custom"]; got.Key != "sk-gw" || got.URL != "https://gw.example/v1" {
		t.Fatalf("custom config was not migrated: %#v", got)
	}
	if _, exists := Conf.Models["8a2b41c0"]; exists {
		t.Fatal("legacy uuid key should not survive migration")
	}
	if Conf.Agent == nil || Conf.Agent.Provider != "deepseek" || Conf.Agent.Model != "deepseek-chat" {
		t.Fatalf("agent was not migrated: %#v", Conf.Agent)
	}
	// 迁移结果应已回写为新格式。
	data, err := os.ReadFile(filepath.Join(Conf.WorkSpaceDir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	var saved map[string]any
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	models, ok := saved["model"].(map[string]any)
	if !ok {
		t.Fatal("saved config has no model map")
	}
	if _, exists := models["deepseek"]; !exists {
		t.Fatalf("saved models keyed by provider: %#v", models)
	}
}

func TestLoadConfigKeepsNewFormat(t *testing.T) {
	previous := Conf
	t.Cleanup(func() { Conf = previous })
	Conf = DefaultConfig()
	Conf.WorkSpaceDir = t.TempDir()

	content := `{
  "agent": {"name": "a1", "provider": "deepseek", "model": "deepseek-reasoner"},
  "model": {"deepseek": {"key": "sk-1", "models": ["deepseek-chat", "deepseek-reasoner"]}},
  "memory": {"token_limit": 32000},
  "channels": {},
  "browser": {"enabled": false}
}`
	if err := os.WriteFile(filepath.Join(Conf.WorkSpaceDir, "config.json"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	if err := LoadConfig(); err != nil {
		t.Fatal(err)
	}
	got := ModelCredentials("deepseek")
	if got.Key != "sk-1" || len(got.Models) != 2 {
		t.Fatalf("unexpected config %#v", got)
	}
	if Conf.Agent == nil || Conf.Agent.Provider != "deepseek" || Conf.Agent.Model != "deepseek-reasoner" {
		t.Fatalf("unexpected agent %#v", Conf.Agent)
	}
}

func TestLoadConfigInfersProviderFromSingleEntry(t *testing.T) {
	previous := Conf
	t.Cleanup(func() { Conf = previous })
	Conf = DefaultConfig()
	Conf.WorkSpaceDir = t.TempDir()

	content := `{
  "agent": {"name": "a1", "model": "deepseek-chat"},
  "model": {"deepseek": {"key": "sk-1", "models": ["deepseek-chat"]}},
  "memory": {}, "channels": {}, "browser": {}
}`
	if err := os.WriteFile(filepath.Join(Conf.WorkSpaceDir, "config.json"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	if err := LoadConfig(); err != nil {
		t.Fatal(err)
	}
	if Conf.Agent == nil || Conf.Agent.Provider != "deepseek" {
		t.Fatalf("provider was not inferred: %#v", Conf.Agent)
	}
}
