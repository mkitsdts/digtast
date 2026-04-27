package center

import (
	"bufio"
	"context"
	"digital-labor/pkg/model"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

func (c *Center) RunLocalREPL() {
	scanner := bufio.NewScanner(os.Stdin)
	currentAgentID := ""
	sessionID := uuid.New().String()

	fmt.Printf("%s=== Digital Labor Local Interactive Mode ===%s\n", colorCyan, colorReset)
	fmt.Println("Type '/ls' to list agents, '/use <id>' to switch, '/exit' to quit.")

	// Auto-select first agent if available
	c.mu.RLock()
	for id := range c.agents {
		currentAgentID = id
		break
	}
	c.mu.RUnlock()

	if currentAgentID == "" {
		fmt.Printf("%s[System] No agents found. Please create one via API or config first.%s\n", colorYellow, colorReset)
	} else {
		fmt.Printf("%s[System] Active Agent: %s%s\n", colorGreen, currentAgentID, colorReset)
	}

	for {
		prompt := ">>> "
		if currentAgentID != "" {
			prompt = fmt.Sprintf("%s[%s]%s >>> ", colorBlue, currentAgentID, colorReset)
		}
		fmt.Print(prompt)

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if strings.HasPrefix(input, "/") {
			parts := strings.Split(input, " ")
			cmd := parts[0]
			switch cmd {
			case "/exit", "/quit":
				fmt.Println("Goodbye!")
				return
			case "/ls":
				c.mu.RLock()
				if len(c.agents) == 0 {
					fmt.Println("No agents available.")
				} else {
					fmt.Println("Available Agents:")
					for id, ag := range c.agents {
						fmt.Printf("- %s (ID: %s)\n", id, ag.ID)
					}
				}
				c.mu.RUnlock()
			case "/new":
				fmt.Println("Creating new Agent (Interactive):")
				fmt.Print("Enter ID (e.g., my-agent): ")
				scanner.Scan()
				id := strings.TrimSpace(scanner.Text())
				fmt.Print("Enter Name: ")
				scanner.Scan()
				name := strings.TrimSpace(scanner.Text())
				fmt.Print("Enter LLM Provider (ark/openai/qwen/deepseek): ")
				scanner.Scan()
				provider := strings.TrimSpace(scanner.Text())
				fmt.Print("Enter API Key: ")
				scanner.Scan()
				key := strings.TrimSpace(scanner.Text())
				fmt.Print("Enter Model: ")
				scanner.Scan()
				modelName := strings.TrimSpace(scanner.Text())

				cfg := &model.DigitalAgentConfig{
					ID:       id,
					Name:     name,
					Provider: provider,
					Key:      key,
					Model:    modelName,
				}
				ag, err := c.CreateAgent(cfg)
				if err != nil {
					fmt.Printf("%sError creating agent: %v%s\n", colorRed, err, colorReset)
				} else {
					currentAgentID = ag.ID
					fmt.Printf("%sAgent '%s' created and selected!%s\n", colorGreen, currentAgentID, colorReset)
				}
			case "/help":
				fmt.Println("Commands:")
				fmt.Println("  /ls          - List all agents")
				fmt.Println("  /new         - Create a new agent interactively")
				fmt.Println("  /use <id>    - Switch to a different agent")
				fmt.Println("  /exit        - Exit the application")
				fmt.Println("  /help        - Show this help message")
			default:
				fmt.Printf("%sUnknown command: %s%s\n", colorRed, cmd, colorReset)
			}
			continue
		}

		if currentAgentID == "" {
			fmt.Printf("%s[Error] No agent selected. Use '/ls' and '/use <id>' first.%s\n", colorRed, colorReset)
			continue
		}

		// Execute Agent
		ag, _ := c.GetAgent(currentAgentID)
		ctx := context.Background()

		fmt.Printf("%sAssistant: %s", colorCyan, colorReset)

		ch, err := ag.Run(ctx, input, true, sessionID)
		if err != nil {
			fmt.Printf("\n%sError: %v%s\n", colorRed, err, colorReset)
			continue
		}

		for content := range ch {
			fmt.Print(content)
		}
		fmt.Println() // New line after stream ends
	}
}
