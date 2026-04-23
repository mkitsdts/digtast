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
		center = &Center{
			agents: make(map[string]*agent.DigitalAgent),
		}
	})
	return center
}

func (m *Center) Create(containerID, agentID, provider, key, url, modelName string) (*agent.DigitalAgent, error) {
	if containerID == "" {
		return nil, errors.New("container_id is empty")
	}
	if agentID == "" {
		return nil, errors.New("agent_id is empty")
	}

	agent, err := agent.NewDigitalAgent(provider, key, url, modelName, agentID)
	if err != nil {
		return nil, err
	}

	keyID := buildAgentKey(containerID, agentID)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[keyID] = agent
	return agent, nil
}

func (m *Center) Get(containerID, agentID string) (*agent.DigitalAgent, error) {
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

func (m *Center) Pause(containerID, agentID, sessionID string) error {
	agent, err := m.Get(containerID, agentID)
	if err != nil {
		return err
	}
	return agent.PauseSession(sessionID)
}

func (m *Center) Remove(containerID, agentID string) error {
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
