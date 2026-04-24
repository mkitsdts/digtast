package center

import (
	"digital-labor/internal/agent"
	"digital-labor/internal/gateway"
	"digital-labor/internal/vdisplay"
	"errors"
	"sync"
)

type Center struct {
	mu             sync.RWMutex
	agents         map[string]*agent.DigitalAgent
	channelGateway gateway.ChannelGateway
	visualDisplay  vdisplay.VisualDisplayManager
}

var (
	center      *Center
	managerOnce sync.Once
)

func GetCenter() *Center {
	managerOnce.Do(func() {
		//TODO: parse config.json
		center = &Center{
			agents: make(map[string]*agent.DigitalAgent),
		}
	})
	return center
}

func (m *Center) CreateAgent(config *agent.DigitalAgentConfig) (*agent.DigitalAgent, error) {
	if config.Key == "" {
		return nil, errors.New("container_id is empty")
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

	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[config.ID] = ag
	return ag, nil
}

func (m *Center) Get(id string) (*agent.DigitalAgent, error) {
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

func (m *Center) Pause(id, sessionID string) error {
	agent, err := m.Get(id)
	if err != nil {
		return err
	}
	return agent.PauseSession(sessionID)
}

func (m *Center) Remove(id string) error {
	if id == "" {
		return errors.New("invalid agent id parameter")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.agents[id]; !ok {
		return errors.New("agent not found")
	}
	delete(m.agents, id)
	return nil
}
