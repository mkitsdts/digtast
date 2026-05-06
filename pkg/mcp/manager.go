package mcp

import (
	"digital-labor/pkg/ctxmanager"
	"digital-labor/pkg/registry"
	"log/slog"

	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
)

type Manager struct {
	clients map[string]*Client
}

var manager *Manager = &Manager{
	clients: make(map[string]*Client),
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
	m.clients[url] = &Client{Client: cli}
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
}

func (m *Manager) GetAllTools() []tool.BaseTool {
	var tools []tool.BaseTool
	for _, client := range m.clients {
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
