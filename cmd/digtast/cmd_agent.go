package main

import (
	"fmt"

	"digital-labor/config"
	"digital-labor/internal/agent"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Configure the singleton agent interactively",
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := agent.Instance(); err == nil {
			errorMsg("Agent is already configured.")
			return
		}
		fmt.Println(colorCyan + "Configuring Agent (Interactive):" + colorReset)

		name := activeState.promptString("Agent Name (e.g., my-agent)")
		if name == "" {
			errorMsg("Agent name is required.")
			return
		}

		modelName := activeState.promptString("Model Name (e.g., doubao, deepseek)")
		if modelName == "" {
			errorMsg("Model name is required.")
			return
		}

		modelVersion := activeState.promptString("Model Version(optional)")

		description := activeState.promptOptional("Init Prompt(optional)")

		cfg := &config.AgentConfig{
			Name:        name,
			Model:       modelName,
			Description: description,
			ModelKind:   modelVersion,
		}

		ag, err := agent.Initialize(cfg)
		if err != nil {
			errorMsg("Error creating agent: %v", err)
			return
		}

		config.Conf.Agent = cfg
		if err := config.SaveConfig(); err != nil {
			errorMsg("Agent initialized but config could not be saved: %v", err)
		}
		activeState.setAgent(ag)
		successMsg("Agent configured: %s", name)
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Clear current agent's conversation history",
	Run: func(cmd *cobra.Command, args []string) {
		if activeState.agent == nil {
			errorMsg("No agent selected.")
			return
		}
		if err := activeState.agent.ClearHistory(); err != nil {
			errorMsg("Error clearing history: %v", err)
			return
		}
		successMsg("Agent '%s' history cleared.", activeState.agentName)
	},
}

func (s *replState) setAgent(ag *agent.Agent) {
	s.agent = ag
	s.agentName = ag.Name
}

func (s *replState) showHistory() {
	sess, _ := s.agent.GetSession()
	if sess != nil && len(sess.GetMessages()) > 0 {
		successMsg("[System] Loaded %d messages from history.", len(sess.GetMessages()))
	}
}
