package agent

import (
	"digital-labor/pkg/errs"
	mmodel "digital-labor/pkg/model"
	"digital-labor/pkg/workspace"
	"errors"
	"sync"
)

type Manager struct {
	mux    sync.RWMutex
	agents map[string]*DigitalAgent
}

func NewManager() *Manager {
	return &Manager{
		agents: make(map[string]*DigitalAgent),
	}
}

func (m *Manager) CreateAgent(name string, cfg *mmodel.DigitalAgentConfig) (*DigitalAgent, error) {
	if cfg.ID == "" {
		return nil, errors.New("id is empty")
	}
	if cfg.Key == "" {
		return nil, errors.New("key is empty")
	}
	if cfg.Name == "" {
		return nil, errors.New("name is empty")
	}
	if cfg.Model == "" {
		return nil, errors.New("model is empty")
	}

	ag, err := newDigitalAgent(cfg)
	if err != nil {
		return nil, err
	}

	// Persist to workspace
	if err := workspace.SaveAgentConfig(cfg); err != nil {
		return ag, err
	}

	m.mux.Lock()
	m.agents[cfg.ID] = ag
	defer m.mux.Unlock()
	return ag, nil
}

func (m *Manager) GetAgent(name string) (*DigitalAgent, error) {
	if name == "" {
		return nil, errs.ErrAgentIDRequired
	}

	m.mux.RLock()
	defer m.mux.RUnlock()
	agent, ok := m.agents[name]
	if !ok {
		return nil, errs.ErrAgentNotFound
	}
	return agent, nil
}

func (m *Manager) GetAgents() []*DigitalAgent {
	m.mux.RLock()
	defer m.mux.RUnlock()
	agents := make([]*DigitalAgent, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	return agents
}

func (m *Manager) RemoveAgent(id string) error {
	if id == "" {
		return errs.ErrAgentIDRequired
	}

	// Remove from workspace first
	if err := workspace.RemoveAgentConfig(id); err != nil {
		return err
	}

	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.agents[id]; !ok {
		return errors.New("agent not found")
	}
	delete(m.agents, id)
	return nil
}

func (m *Manager) LoadAgents() error {
	// TODO: 加载 Agents
	return nil
}
