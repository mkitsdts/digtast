package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"digital-labor/internal/center"
	"digital-labor/pkg/conf"
)

func RunLocalREPL() {
	s := &replState{
		scanner: bufio.NewScanner(os.Stdin),
	}
	activeState = s

	s.initAgent()
	printWelcome()
	s.printStatus()

	for {
		fmt.Print(s.prompt())

		if !s.scanner.Scan() {
			break
		}

		input := strings.TrimSpace(s.scanner.Text())
		if input == "" {
			continue
		}

		if strings.HasPrefix(input, "/") {
			// Use Cobra to parse and execute commands
			cmdStr := strings.TrimPrefix(input, "/")
			args := strings.Split(cmdStr, " ")

			// Special case for /help to use cobra's help
			if args[0] == "help" {
				rootCmd.Help()
				continue
			}

			rootCmd.SetArgs(args)
			if err := rootCmd.Execute(); err != nil {
				// Cobra already prints the error if it's a usage error or similar
			}

			if exitREPL {
				fmt.Println("Goodbye!")
				return
			}
			continue
		}

		s.runChat(input)
	}
}

func (s *replState) initAgent() {
	s.agentID = conf.Conf.State.LastUsedAgent

	ag, err := center.AgentManager.GetAgent(s.agentID)
	if err != nil {
		ags := center.AgentManager.GetAgents()
		if len(ags) == 0 {
			warnMsg("[System] No agents found. Create one via /new or config first.")
			return
		}
		ag = ags[0]
	}

	s.agentID = ag.ID
	s.agent = ag
	s.agentName = ag.Name
	conf.Conf.State.LastUsedAgent = ag.ID
	conf.SaveConfig()
}

func (s *replState) printStatus() {
	if s.agent != nil {
		successMsg("[System] Using Agent: %s (%s)", s.agentName, s.agentID)
		s.showHistory()
	}
}

func (s *replState) runChat(input string) {
	if s.agent == nil {
		errorMsg("No agent selected. Use /ls and /use <id> first.")
		return
	}

	ctx := context.Background()
	fmt.Printf("%sAssistant: %s", colorCyan, colorReset)

	ch, err := s.agent.Run(ctx, input, true)
	if err != nil {
		fmt.Printf("\n")
		errorMsg("%v", err)
		return
	}

	for content := range ch {
		fmt.Print(content)
	}
	fmt.Println()
}
