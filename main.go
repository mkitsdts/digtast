package main

import (
	"digital-labor/internal/center"
	"digital-labor/internal/cli"
	"digital-labor/internal/server"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/workspace"
	_ "digital-labor/prompt"
	_ "digital-labor/tools"
	"flag"
	"fmt"
	"path/filepath"
)

func main() {
	localMode := flag.Bool("local", true, "Run in local interactive mode")
	port := flag.String("port", "10086", "gRPC server port")
	flag.Parse()

	// 1. Get workspace path and load config
	wp := workspace.GetWorkspacePath()
	configPath := filepath.Join(wp, "config.json")
	if err := conf.LoadConfig(configPath); err != nil {
		panic(fmt.Sprintf("Failed to load config from %s: %v", configPath, err))
	}

	Init()
	if *localMode {
		conf.Conf.Mode = "local"
		cli.RunLocalREPL()
	} else {
		conf.Conf.Mode = "server"
		server.Start(":" + *port)
	}
}

func Init() {
	workspace.InitWorkspace()
	center.Init()
}
