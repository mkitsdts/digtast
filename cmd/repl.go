package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"digital-labor/internal/agent"
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
	ag, err := agent.Instance()
	if err != nil {
		warnMsg("[System] Agent is not configured. Create it via /new.")
		return
	}
	s.agent = ag
	s.agentName = ag.Name
}

func (s *replState) printStatus() {
	if s.agent != nil {
		successMsg("[System] Using Agent: %s", s.agentName)
		s.showHistory()
	}
}

func (s *replState) runChat(input string) {
	if s.agent == nil {
		errorMsg("Agent is not configured. Use /new first.")
		return
	}

	ctx := context.Background()
	fmt.Printf("%sAssistant: %s", colorCyan, colorReset)

	ch, err := s.agent.Run(ctx, input, nil, true)
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
