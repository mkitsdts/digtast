package main

import (
	"digital-labor/config"
	_ "digital-labor/gateway"
	"digital-labor/internal/agent"
	"digital-labor/internal/agent/workspace"
	gw "digital-labor/internal/gateway"
	"digital-labor/pkg/gateway"
	_ "digital-labor/prompt"
	_ "digital-labor/tools"
	"flag"
	"fmt"
	"path/filepath"
)

func main() {
	browserMode := flag.Bool("browser", false, "Run in browser mode")
	flag.Parse()

	// Get workspace path and load config
	wp := workspace.GetWorkspacePath()
	configPath := filepath.Join(wp, "config.json")
	if err := config.LoadConfig(); err != nil {
		panic(fmt.Sprintf("Failed to load config from %s: %v", configPath, err))
	}
	if *browserMode {
		config.Conf.Browser.Enabled = true
	}
	if err := Init(); err != nil {
		panic(err)
	}
	RunLocalREPL()
}

func Init() error {
	workspace.InitWorkspace()
	if config.Conf.Agent != nil {
		if _, err := agent.Initialize(config.Conf.Agent); err != nil {
			return fmt.Errorf("initialize agent: %w", err)
		}
	}
	gateway.LoadChannels()
	gw.Start()
	return nil
}
