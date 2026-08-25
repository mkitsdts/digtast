package chatmodel

import (
	"context"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/llm"
	"fmt"
)

type Manager struct {
}

var ModelManager = &Manager{}

func Init() {
	ModelManager = &Manager{}
}

func (m *Manager) CreateModel(configName string, cfg conf.ModelConfig) error {
	if conf.Conf.Models == nil {
		conf.Conf.Models = make(map[string]conf.ModelConfig)
	}
	conf.Conf.Models[configName] = cfg
	return conf.SaveConfig()
}

func (m *Manager) RemoveModel(configName string) error {
	if conf.Conf.Models != nil {
		delete(conf.Conf.Models, configName)
	}
	return conf.SaveConfig()
}

func (m *Manager) ListModels() map[string]conf.ModelConfig {
	return conf.Conf.Models
}

func (m *Manager) GetChatModel(ctx context.Context, modelName string) (llm.LLM, error) {
	cfg, ok := conf.FindModelConfig(modelName)
	if !ok {
		return nil, fmt.Errorf("model config for %s not found", modelName)
	}

	return newChatModel(cfg.Provider, cfg.Key, cfg.URL, modelName)
}

func (m *Manager) GetChatModelByConfig(ctx context.Context, configName, modelName string) (llm.LLM, error) {
	cfg, ok := conf.Conf.Models[configName]
	if !ok {
		return nil, fmt.Errorf("model config %s not found", configName)
	}

	return newChatModel(cfg.Provider, cfg.Key, cfg.URL, modelName)
}
