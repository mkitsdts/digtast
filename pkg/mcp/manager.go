package mcp

import (
	"digital-labor/pkg/ctxmanager"
	"digital-labor/pkg/registry"
	"log/slog"

	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
)

type ClientInfo struct {
	URL     string
	Enabled bool
}

type Manager struct {
	clients map[string]*Client
	enabled map[string]bool
}

var manager *Manager = &Manager{
	clients: make(map[string]*Client),
	enabled: make(map[string]bool),
}

func NewManager() *Manager {
	return manager
}

func (m *Manager) AddClient(url string) error {
	cli, err := client.NewSSEMCPClient(url)
	if err != nil {
		return err
	}
	ctx := ctxmanager.GetOrCreate(url)
	m.clients[url] = &Client{Client: cli, url: url}
	m.enabled[url] = true
	m.clients[url].Start(ctx)
	tools, _ := m.clients[url].GetTools(ctx)
	registerMCPTools(tools)
	return nil
}

func (m *Manager) GetClient(url string) (*Client, bool) {
	cli, ok := m.clients[url]
	return cli, ok
}

func (m *Manager) RemoveClient(url string) {
	delete(m.clients, url)
	delete(m.enabled, url)
}

func (m *Manager) GetAllClients() []ClientInfo {
	var infos []ClientInfo
	for url := range m.clients {
		infos = append(infos, ClientInfo{
			URL:     url,
			Enabled: m.enabled[url],
		})
	}
	return infos
}

func (m *Manager) DisableClient(url string) {
	m.enabled[url] = false
}

func (m *Manager) EnableClient(url string) {
	m.enabled[url] = true
}

func (m *Manager) GetAllTools() []tool.BaseTool {
	var tools []tool.BaseTool
	for url, client := range m.clients {
		if !m.enabled[url] {
			continue
		}
		ctx := ctxmanager.GetOrCreate(client.url)
		result, err := client.GetTools(ctx)
		if err != nil {
			slog.Error("failed to get tools", "url", client.url, "error", err)
			continue
		}
		tools = append(tools, result...)
	}
	return tools
}

func registerMCPTools(tools []tool.BaseTool) {
	for _, tool := range tools {
		registry.RegisterTool(tool)
	}
}
