package main

import (
	"digital-labor/internal/server"
)

func main() {
	server.Start(":10086")
}
