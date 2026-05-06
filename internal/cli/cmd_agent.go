package cli

import (
	"fmt"

	"digital-labor/internal/agent"
	"digital-labor/internal/center"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/model"
)

func (s *replState) cmdLs() {
	ags := center.AgentManager.GetAgents()
	if len(ags) == 0 {
		fmt.Println("No agents available.")
		return
	}
	fmt.Println("Available Agents:")
	for _, a := range ags {
		fmt.Printf("- %s (ID: %s)\n", a.Name, a.ID)
	}
}

func (s *replState) cmdUse(args []string) {
	if len(args) < 1 {
		errorMsg("Usage: /use <agent_id>")
		return
	}
	targetID := args[0]
	ag, err := center.AgentManager.GetAgent(targetID)
	if err != nil {
		errorMsg("Agent '%s' not found.", targetID)
		return
	}
	s.switchAgent(ag)
	successMsg("Switched to agent '%s'.", s.agentName)
	s.showHistory()
}

func (s *replState) cmdNew() {
	fmt.Println("Creating new Agent (Interactive):")

	name := s.promptString("Agent Name (e.g., my-agent)")
	if name == "" {
		errorMsg("Agent name is required.")
		return
	}

	modelName := s.promptString("Model Name (supported: doubao, deepseek)")
	if modelName == "" {
		errorMsg("Model name is required.")
		return
	}

	modelVersion := s.promptString("Model Version (e.g., doubao-seed-2-0-lite-260215)")
	if modelVersion == "" {
		errorMsg("Model version is required.")
		return
	}

	description := s.promptOptional("Description")

	cfg := &model.DigitalAgentConfig{
		Name:        name,
		Model:       modelName,
		Description: description,
		ModelKind:   modelVersion,
	}

	ag, err := center.AgentManager.CreateAgent(name, cfg)
	if err != nil {
		errorMsg("Error creating agent: %v", err)
		return
	}

	s.switchAgent(ag)
	successMsg("Agent created! ID: %s, Name: %s", ag.ID, name)
}

func (s *replState) cmdReset() {
	if s.agent == nil {
		errorMsg("No agent selected.")
		return
	}
	if err := s.agent.ClearHistory(); err != nil {
		errorMsg("Error clearing history: %v", err)
		return
	}
	successMsg("Agent '%s' history cleared.", s.agentName)
}

func (s *replState) switchAgent(ag *agent.DigitalAgent) {
	s.agentID = ag.ID
	s.agent = ag
	s.agentName = ag.Name
	conf.Conf.State.LastUsedAgent = ag.ID
	conf.SaveConfig()
}

func (s *replState) showHistory() {
	sess, _ := s.agent.GetSession()
	if sess != nil && len(sess.GetMessages()) > 0 {
		successMsg("[System] Loaded %d messages from history.", len(sess.GetMessages()))
	}
}
