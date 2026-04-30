package cli

import (
	"bufio"
	"context"
	"digital-labor/internal/center"
	"digital-labor/internal/gateway"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/ctxmanager"
	"digital-labor/pkg/model"
	"fmt"
	"os"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

func RunLocalREPL() {
	scanner := bufio.NewScanner(os.Stdin)
	currentAgentID := conf.Conf.State.LastUsedAgent
	ag, _ := center.AgentManager.GetAgent(currentAgentID)
	currentAgentName := ""
	if ag != nil {
		currentAgentName = ag.Name
	}

	fmt.Printf("%s=== Digital Labor Local Interactive Mode ===%s\n", colorCyan, colorReset)
	fmt.Println("Type '/ls' to list agents, '/use <id>' to switch, '/exit' to quit.")

	// Auto-select agent if not set or invalid
	ag, err := center.AgentManager.GetAgent(currentAgentID)
	if err != nil {
		ags := center.AgentManager.GetAgents()
		if len(ags) == 0 {
			fmt.Printf("%s[System] No agents found. Please create one via '/new' or config first.%s\n", colorYellow, colorReset)
		} else {
			currentAgentID = ags[0].ID
			conf.Conf.State.LastUsedAgent = currentAgentID
			conf.SaveConfig()
			ag = ags[0]
			currentAgentName = ag.Name
		}
	}

	if currentAgentID != "" && ag != nil {
		fmt.Printf("%s[System] Selected Agent: %s (%s)%s\n", colorGreen, currentAgentName, currentAgentID, colorReset)
		sess, _ := ag.GetSession()
		if sess != nil && len(sess.GetMessages()) > 0 {
			fmt.Printf("%s[System] Loaded %d messages from history.%s\n", colorGreen, len(sess.GetMessages()), colorReset)
		}
	}

	for {
		prompt := ">>> "
		if currentAgentID != "" {
			prompt = fmt.Sprintf("%s[%s]%s >>> ", colorBlue, currentAgentName, colorReset)
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
				ags := center.AgentManager.GetAgents()
				if len(ags) == 0 {
					fmt.Println("No agents available.")
				} else {
					fmt.Println("Available Agents:")
					for _, a := range ags {
						fmt.Printf("- %s (ID: %s)\n", a.Name, a.ID)
					}
				}
			case "/use":
				if len(parts) < 2 {
					fmt.Println("Usage: /use <agent_id>")
					continue
				}
				targetID := parts[1]
				newAg, err := center.AgentManager.GetAgent(targetID)
				if err != nil {
					fmt.Printf("%sError: Agent '%s' not found.%s\n", colorRed, targetID, colorReset)
				} else {
					currentAgentID = targetID
					ag = newAg
					currentAgentName = ag.Name
					conf.Conf.State.LastUsedAgent = currentAgentID
					conf.SaveConfig()
					fmt.Printf("%sSwitched to agent '%s'.%s\n", colorGreen, currentAgentName, colorReset)

					sess, _ := ag.GetSession()
					if sess != nil && len(sess.GetMessages()) > 0 {
						fmt.Printf("%s[System] Loaded %d messages from history.%s\n", colorGreen, len(sess.GetMessages()), colorReset)
					}
				}
			case "/new":
				fmt.Println("Creating new Agent (Interactive):")

				fmt.Print("Enter Agent Name (e.g., my-agent): ")
				scanner.Scan()
				name := strings.TrimSpace(scanner.Text())

				fmt.Print("Enter Model Name (from config.json, e.g., deepseek): ")
				scanner.Scan()
				modelName := strings.TrimSpace(scanner.Text())

				fmt.Print("Enter Model Version (from config.json, e.g., doubao-seed-2-0-lite-260215): ")
				scanner.Scan()
				modelVersion := strings.TrimSpace(scanner.Text())

				fmt.Print("Enter Description(Optional): ")
				scanner.Scan()
				description := strings.TrimSpace(scanner.Text())

				cfg := &model.DigitalAgentConfig{
					Name:        name,
					Model:       modelName,
					Description: description,
					ModelKind:   modelVersion,
				}

				newAg, err := center.AgentManager.CreateAgent(name, cfg)
				if err != nil {
					fmt.Printf("%sError creating agent: %v%s\n", colorRed, err, colorReset)
				} else {
					currentAgentID = newAg.ID
					ag = newAg
					currentAgentName = ag.Name
					conf.Conf.State.LastUsedAgent = currentAgentID
					conf.SaveConfig()
					fmt.Printf("%sAgent created! ID: %s, Name: %s%s\n", colorGreen, ag.ID, name, colorReset)
				}
			case "/reset":
				if ag == nil {
					fmt.Printf("%s[Error] No agent selected.%s\n", colorRed, colorReset)
					continue
				}
				if err := ag.ClearHistory(); err != nil {
					fmt.Printf("%sError clearing history: %v%s\n", colorRed, err, colorReset)
				} else {
					fmt.Printf("%sAgent '%s' history cleared.%s\n", colorGreen, currentAgentName, colorReset)
				}
			case "/channel":
				// TODO：通道持久化
				if len(parts) < 2 {
					fmt.Println("Usage: /channel <subcommand>")
					fmt.Println("Subcommands: ")
					continue
				}
				subCmd := parts[1]
				fmt.Printf("Creating channel: %s\n", subCmd)
				c := gateway.NewChannel(subCmd)
				if c == nil {
					fmt.Printf("%s[Error] Failed to create channel: qq%s\n", colorRed, colorReset)
					continue
				}
				c.Register()
				ctx := ctxmanager.GetOrCreate(subCmd)
				go c.Serve(ctx)
				if conf.Conf.Channels != nil {
					conf.Conf.Channels[subCmd] = c.GetConfig()
				}
				conf.SaveConfig()
			case "/model":
				if len(parts) < 2 {
					fmt.Println("Usage: /model <subcommand>")
					fmt.Println("Subcommands: ls, add")
					continue
				}
				subCmd := parts[1]
				switch subCmd {
				case "ls":
					fmt.Println("Configured Providers & Models:")
					for name, cfg := range conf.Conf.Models {
						fmt.Printf("- Provider Config: %s\n", name)
						fmt.Printf("  Provider: %s\n", cfg.Provider)
						fmt.Printf("  Models: %s\n", strings.Join(cfg.ModelNames, ", "))
					}
				case "add":
					fmt.Print("Enter Provider Config Name (e.g., my-openai): ")
					scanner.Scan()
					configName := strings.TrimSpace(scanner.Text())
					fmt.Print("Enter LLM Provider (ark/openai/qwen/deepseek): ")
					scanner.Scan()
					provider := strings.TrimSpace(scanner.Text())
					fmt.Print("Enter API Key: ")
					scanner.Scan()
					key := strings.TrimSpace(scanner.Text())
					fmt.Print("Enter Model Names (comma separated, e.g., gpt-4,gpt-3.5-turbo): ")
					scanner.Scan()
					modelsInput := strings.TrimSpace(scanner.Text())
					fmt.Print("Enter Base URL (optional): ")
					scanner.Scan()
					baseURL := strings.TrimSpace(scanner.Text())

					modelNames := strings.Split(modelsInput, ",")
					for i := range modelNames {
						modelNames[i] = strings.TrimSpace(modelNames[i])
					}

					if conf.Conf.Models == nil {
						conf.Conf.Models = make(map[string]conf.ModelConfig)
					}
					conf.Conf.Models[configName] = conf.ModelConfig{
						Provider:   provider,
						Key:        key,
						ModelNames: modelNames,
						URL:        baseURL,
					}
					conf.SaveConfig()
					fmt.Printf("%sModel configuration '%s' saved!%s\n", colorGreen, configName, colorReset)
				}
			case "/help":
				fmt.Println("Commands:")
				fmt.Println("  /ls              - List all agents")
				fmt.Println("  /new             - Create a new agent interactively")
				fmt.Println("  /use <id>        - Switch to a different agent")
				fmt.Println("  /reset           - Clear current agent's conversation history")
				fmt.Println("  /model ls|add    - Manage model configurations")
				fmt.Println("  /exit            - Exit the application")
				fmt.Println("  /help            - Show this help message")
			default:
				fmt.Printf("%sUnknown command: %s%s\n", colorRed, cmd, colorReset)
			}
			continue
		}

		if ag == nil {
			fmt.Printf("%s[Error] No agent selected. Use '/ls' and '/use <id>' first.%s\n", colorRed, colorReset)
			continue
		}

		ctx := context.Background()
		fmt.Printf("%sAssistant: %s", colorCyan, colorReset)

		ch, err := ag.Run(ctx, input, true)
		if err != nil {
			fmt.Printf("\n%sError: %v%s\n", colorRed, err, colorReset)
			continue
		}

		for content := range ch {
			fmt.Print(content)
		}
		fmt.Println()
	}
}
