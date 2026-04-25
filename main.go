package main

import (
	"digital-labor/internal/server"
	"digital-labor/pkg/workspace"
)

func main() {
	Init()
	server.Start(":10086")
}

func Init() {
	workspace.InitWorkspace()
}
