package agent

import (
	"errors"
	"sync"
)

type Manager struct {
	mu     sync.RWMutex
	agents map[string]*DigitalAgent
}

var (
	manager     *Manager
	managerOnce sync.Once
)

func GetManager() *Manager {
	managerOnce.Do(func() {
		manager = &Manager{
			agents: make(map[string]*DigitalAgent),
		}
	})
	return manager
}

func (m *Manager) Create(containerID, agentID, provider, key, url, modelName string) (*DigitalAgent, error) {
	if containerID == "" {
		return nil, errors.New("container_id is empty")
	}
	if agentID == "" {
		return nil, errors.New("agent_id is empty")
	}

	agent, err := NewDigitalAgent(provider, key, url, modelName)
	if err != nil {
		return nil, err
	}
	agent.ID = agentID

	keyID := buildAgentKey(containerID, agentID)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[keyID] = agent
	return agent, nil
}

func (m *Manager) Get(containerID, agentID string) (*DigitalAgent, error) {
	if containerID == "" {
		return nil, errors.New("container_id is empty")
	}
	if agentID == "" {
		return nil, errors.New("agent_id is empty")
	}

	keyID := buildAgentKey(containerID, agentID)

	m.mu.RLock()
	defer m.mu.RUnlock()
	agent, ok := m.agents[keyID]
	if !ok {
		return nil, errors.New("agent not found")
	}
	return agent, nil
}

func (m *Manager) Pause(containerID, agentID, sessionID string) error {
	agent, err := m.Get(containerID, agentID)
	if err != nil {
		return err
	}
	return agent.PauseSession(sessionID)
}

func (m *Manager) Remove(containerID, agentID string) error {
	if containerID == "" {
		return errors.New("container_id is empty")
	}
	if agentID == "" {
		return errors.New("agent_id is empty")
	}

	keyID := buildAgentKey(containerID, agentID)

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.agents[keyID]; !ok {
		return errors.New("agent not found")
	}
	delete(m.agents, keyID)
	return nil
}

func buildAgentKey(containerID, agentID string) string {
	return containerID + "/" + agentID
}
