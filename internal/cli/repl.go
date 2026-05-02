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
			parts := strings.Split(input, " ")
			cmd := parts[0]
			args := parts[1:]

			if s.dispatch(cmd, args) {
				return
			}
			continue
		}

		s.runChat(input)
	}
}

func (s *replState) dispatch(cmd string, args []string) bool {
	switch cmd {
	case "/exit", "/quit":
		fmt.Println("Goodbye!")
		return true

	case "/help":
		printHelp()

	case "/ls":
		s.cmdLs()

	case "/use":
		s.cmdUse(args)

	case "/new":
		s.cmdNew()

	case "/reset":
		s.cmdReset()

	case "/model":
		s.cmdModel(args)

	case "/channel":
		s.cmdChannel(args)

	default:
		errorMsg("Unknown command: %s", cmd)
	}
	return false
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
