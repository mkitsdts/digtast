package main

import (
	"digital-labor/internal/center"
	"digital-labor/internal/server"
	"digital-labor/pkg/workspace"
	_ "digital-labor/prompt"
	_ "digital-labor/tools"
	"flag"
)

func main() {
	localMode := flag.Bool("local", false, "Run in local interactive mode")
	port := flag.String("port", "10086", "gRPC server port")
	flag.Parse()

	Init()

	if *localMode {
		center.GetCenter().RunLocalREPL()
	} else {
		server.Start(":" + *port)
	}
}

func Init() {
	workspace.InitWorkspace()
}
