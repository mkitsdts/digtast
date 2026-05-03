package cli

import (
	"bufio"

	"digital-labor/internal/agent"
)

type replState struct {
	scanner     *bufio.Scanner
	agentID     string
	agent       *agent.DigitalAgent
	agentName   string
}

func (s *replState) prompt() string {
	if s.agentID != "" {
		return colorBlue + "[" + s.agentName + "]" + colorReset + " >>> "
	}
	return ">>> "
}
