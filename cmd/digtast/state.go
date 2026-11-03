package main

import (
	"bufio"

	"digital-labor/internal/agent"
)

type replState struct {
	scanner   *bufio.Scanner
	agent     *agent.Agent
	agentName string
}

var activeState *replState

func (s *replState) prompt() string {
	if s.agent != nil {
		return colorBlue + "[" + s.agentName + "]" + colorReset + " >>> "
	}
	return ">>> "
}
