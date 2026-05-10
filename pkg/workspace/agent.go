package workspace

import (
	"digital-labor/pkg/model"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

var (
	agentMu sync.RWMutex
)

func getAgentsConfigPath() string {
	return filepath.Join(GetWorkspacePath(), "agents.json")
}

// AgentDir returns the root directory for files owned by one agent.
func AgentDir(agentID string) string {
	if agentID == "" {
		agentID = DefaultAgentID()
	}
	return filepath.Join(GetWorkspacePath(), "agents", agentID)
}

// AgentSkillsDir returns the directory containing one agent's skill folders.
func AgentSkillsDir(agentID string) string {
	return filepath.Join(AgentDir(agentID), "skills")
}

// SaveAgentConfig saves or updates an agent configuration in the workspace.
func SaveAgentConfig(cfg *model.DigitalAgentConfig) error {
	agentMu.Lock()
	defer agentMu.Unlock()

	configs, err := LoadAllAgentConfigs()
	if err != nil {
		configs = make(map[string]*model.DigitalAgentConfig)
	}

	configs[cfg.ID] = cfg

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getAgentsConfigPath(), data, 0644)
}

// LoadAllAgentConfigs returns all saved agent configurations.
func LoadAllAgentConfigs() (map[string]*model.DigitalAgentConfig, error) {
	data, err := os.ReadFile(getAgentsConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*model.DigitalAgentConfig), nil
		}
		return nil, err
	}

	var configs map[string]*model.DigitalAgentConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}

	if configs == nil {
		configs = make(map[string]*model.DigitalAgentConfig)
	}

	return configs, nil
}

// RemoveAgentConfig deletes an agent configuration from the workspace.
func RemoveAgentConfig(id string) error {
	agentMu.Lock()
	defer agentMu.Unlock()

	configs, err := LoadAllAgentConfigs()
	if err != nil {
		return err
	}

	if _, ok := configs[id]; !ok {
		return nil
	}

	delete(configs, id)
	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getAgentsConfigPath(), data, 0644)
}
