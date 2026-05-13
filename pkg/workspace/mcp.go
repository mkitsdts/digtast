package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// MCPServerConfig holds the persistent configuration for one MCP server.
type MCPServerConfig struct {
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

var mcpMu sync.RWMutex

func getMCPConfigPath() string {
	return filepath.Join(GetWorkspacePath(), "mcp_servers.json")
}

// SaveMCPServerConfig saves or updates an MCP server configuration.
func SaveMCPServerConfig(cfg *MCPServerConfig) error {
	mcpMu.Lock()
	defer mcpMu.Unlock()

	configs, err := loadAllMCPServerConfigsLocked()
	if err != nil {
		configs = make(map[string]*MCPServerConfig)
	}

	configs[cfg.URL] = cfg

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getMCPConfigPath(), data, 0644)
}

// LoadAllMCPServerConfigs returns all saved MCP server configurations.
func LoadAllMCPServerConfigs() (map[string]*MCPServerConfig, error) {
	mcpMu.RLock()
	defer mcpMu.RUnlock()

	return loadAllMCPServerConfigsLocked()
}

func loadAllMCPServerConfigsLocked() (map[string]*MCPServerConfig, error) {
	data, err := os.ReadFile(getMCPConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*MCPServerConfig), nil
		}
		return nil, err
	}

	var configs map[string]*MCPServerConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}

	if configs == nil {
		configs = make(map[string]*MCPServerConfig)
	}

	return configs, nil
}

// RemoveMCPServerConfig deletes an MCP server configuration by URL.
func RemoveMCPServerConfig(url string) error {
	mcpMu.Lock()
	defer mcpMu.Unlock()

	configs, err := loadAllMCPServerConfigsLocked()
	if err != nil {
		return err
	}

	if _, ok := configs[url]; !ok {
		return nil
	}

	delete(configs, url)

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getMCPConfigPath(), data, 0644)
}

// SetMCPServerEnabled toggles the enabled state of an MCP server.
func SetMCPServerEnabled(url string, enabled bool) error {
	mcpMu.Lock()
	defer mcpMu.Unlock()

	configs, err := loadAllMCPServerConfigsLocked()
	if err != nil {
		return err
	}

	cfg, ok := configs[url]
	if !ok {
		cfg = &MCPServerConfig{URL: url}
		configs[url] = cfg
	}
	cfg.Enabled = enabled

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getMCPConfigPath(), data, 0644)
}
