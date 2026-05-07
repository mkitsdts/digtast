package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"digital-labor/internal/agent"
	"digital-labor/internal/center"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/model"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all agents",
	Run: func(cmd *cobra.Command, args []string) {
		ags := center.AgentManager.GetAgents()
		if len(ags) == 0 {
			infoMsg("No agents available.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, colorCyan+"ID\tName"+colorReset)
		
		for _, a := range ags {
			fmt.Fprintf(w, "%s\t%s\n", a.ID, a.Name)
		}
		fmt.Println(colorCyan + "Available Agents:" + colorReset)
		w.Flush()
	},
}

var useCmd = &cobra.Command{
	Use:   "use <agent_id>",
	Short: "Switch to a different agent",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		targetID := args[0]
		ag, err := center.AgentManager.GetAgent(targetID)
		if err != nil {
			errorMsg("Agent '%s' not found.", targetID)
			return
		}
		activeState.switchAgent(ag)
		successMsg("Switched to agent '%s'.", activeState.agentName)
		activeState.showHistory()
	},
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new agent interactively",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(colorCyan + "Creating new Agent (Interactive):" + colorReset)

		name := activeState.promptString("Agent Name (e.g., my-agent)")
		if name == "" {
			errorMsg("Agent name is required.")
			return
		}

		modelName := activeState.promptString("Model Name (supported: doubao, deepseek)")
		if modelName == "" {
			errorMsg("Model name is required.")
			return
		}

		modelVersion := activeState.promptString("Model Version (e.g., doubao-seed-2-0-lite-260215)")
		if modelVersion == "" {
			errorMsg("Model version is required.")
			return
		}

		description := activeState.promptOptional("Description")

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

		activeState.switchAgent(ag)
		successMsg("Agent created! ID: %s, Name: %s", ag.ID, name)
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
