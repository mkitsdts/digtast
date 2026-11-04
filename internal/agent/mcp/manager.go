package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"

	toolpkg "digital-labor/component/tool"
	"digital-labor/internal/agent/workspace"
)

type ClientInfo struct {
	URL     string
	Enabled bool
}

type Manager struct {
	mu      sync.RWMutex
	clients map[string]*Client
	enabled map[string]bool
}

var manager = &Manager{clients: make(map[string]*Client), enabled: make(map[string]bool)}

func NewManager() *Manager { return manager }

func (m *Manager) AddClient(url string) error {
	cli, err := NewMCPConfig(url)
	if err != nil {
		return err
	}
	if err := cli.Start(context.Background()); err != nil {
		return err
	}
	m.mu.Lock()
	old := m.clients[url]
	m.clients[url] = cli
	m.enabled[url] = true
	m.mu.Unlock()
	if old != nil {
		_ = old.Close()
	}
	return workspace.SaveMCPServerConfig(&workspace.MCPServerConfig{URL: url, Enabled: true})
}

func (m *Manager) GetClient(url string) (*Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cli, ok := m.clients[url]
	return cli, ok
}

func (m *Manager) RemoveClient(url string) {
	m.mu.Lock()
	cli := m.clients[url]
	delete(m.clients, url)
	delete(m.enabled, url)
	m.mu.Unlock()
	if cli != nil {
		_ = cli.Close()
	}
	if err := workspace.RemoveMCPServerConfig(url); err != nil {
		slog.Error("failed to remove mcp config", "url", url, "error", err)
	}
}

func (m *Manager) GetAllClients() []ClientInfo {
	m.mu.RLock()
	infos := make([]ClientInfo, 0, len(m.enabled))
	for url := range m.enabled {
		infos = append(infos, ClientInfo{URL: url, Enabled: m.enabled[url]})
	}
	m.mu.RUnlock()
	sort.Slice(infos, func(i, j int) bool { return infos[i].URL < infos[j].URL })
	return infos
}

func (m *Manager) DisableClient(url string) {
	m.mu.Lock()
	m.enabled[url] = false
	m.mu.Unlock()
	if err := workspace.SetMCPServerEnabled(url, false); err != nil {
		slog.Error("failed to disable mcp config", "url", url, "error", err)
	}
}

func (m *Manager) EnableClient(url string) {
	m.mu.RLock()
	_, exists := m.clients[url]
	m.mu.RUnlock()
	if !exists {
		cli, err := NewMCPConfig(url)
		if err != nil {
			slog.Error("failed to create mcp client", "url", url, "error", err)
			return
		}
		if err := cli.Start(context.Background()); err != nil {
			slog.Error("failed to start mcp client", "url", url, "error", err)
			return
		}
		m.mu.Lock()
		m.clients[url] = cli
		m.mu.Unlock()
	}
	m.mu.Lock()
	m.enabled[url] = true
	m.mu.Unlock()
	if err := workspace.SetMCPServerEnabled(url, true); err != nil {
		slog.Error("failed to enable mcp config", "url", url, "error", err)
	}
}

func (m *Manager) LoadPersisted() {
	configs, err := workspace.LoadAllMCPServerConfigs()
	if err != nil {
		slog.Error("failed to load persisted mcp configs", "error", err)
		return
	}
	for url, cfg := range configs {
		if !cfg.Enabled {
			m.mu.Lock()
			m.enabled[url] = false
			m.mu.Unlock()
			continue
		}
		if err := m.AddClient(url); err != nil {
			slog.Error("failed to restore mcp client", "url", url, "error", err)
		}
	}
}

func (m *Manager) LoadTools(ctx context.Context) ([]toolpkg.Tool, error) {
	m.mu.RLock()
	clients := make(map[string]*Client, len(m.clients))
	for url, cli := range m.clients {
		if m.enabled[url] {
			clients[url] = cli
		}
	}
	m.mu.RUnlock()
	urls := make([]string, 0, len(clients))
	for url := range clients {
		urls = append(urls, url)
	}
	sort.Strings(urls)
	var result []toolpkg.Tool
	for _, url := range urls {
		tools, err := clients[url].GetTools(ctx)
		if err != nil {
			return nil, fmt.Errorf("load tools from mcp %q: %w", url, err)
		}
		result = append(result, tools...)
	}
	return result, nil
}

func (m *Manager) GetAllTools() []toolpkg.Tool {
	tools, err := m.LoadTools(context.Background())
	if err != nil {
		slog.Error("failed to load mcp tools", "error", err)
		return nil
	}
	return tools
}
