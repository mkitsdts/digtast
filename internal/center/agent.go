package center

import (
	"digital-labor/internal/agent"
	"digital-labor/pkg/model"
	"digital-labor/pkg/workspace"
	"errors"
)

func (m *Center) CreateAgent(config *model.DigitalAgentConfig) (*agent.DigitalAgent, error) {
	if config.ID == "" {
		return nil, errors.New("id is empty")
	}
	if config.Key == "" {
		return nil, errors.New("key is empty")
	}
	if config.Name == "" {
		return nil, errors.New("name is empty")
	}
	if config.Model == "" {
		return nil, errors.New("model is empty")
	}

	ag, err := agent.NewDigitalAgent(config)
	if err != nil {
		return nil, err
	}

	// Persist to workspace
	if err := workspace.SaveAgentConfig(config); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[config.ID] = ag
	return ag, nil
}

func (m *Center) GetAgent(id string) (*agent.DigitalAgent, error) {
	if id == "" {
		return nil, errors.New("invalid agent id paramater")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	agent, ok := m.agents[id]
	if !ok {
		return nil, errors.New("agent not found")
	}
	return agent, nil
}

func (m *Center) RemoveAgent(id string) error {
	if id == "" {
		return errors.New("invalid agent id parameter")
	}

	// Remove from workspace first
	if err := workspace.RemoveAgentConfig(id); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.agents[id]; !ok {
		return errors.New("agent not found")
	}
	delete(m.agents, id)
	return nil
}
