package center

import (
	"digital-labor/internal/agent"
	"digital-labor/internal/ftp"
	"digital-labor/internal/gateway"
	"digital-labor/internal/vdisplay"
	"sync"
)

type Center struct {
	mu             sync.RWMutex
	agents         map[string]*agent.DigitalAgent
	channelGateway gateway.ChannelGateway
	visualDisplay  vdisplay.VisualDisplayManager
	ftpServer      *ftp.FTPServer
}

var (
	center      *Center
	managerOnce sync.Once
)

func GetCenter() *Center {
	managerOnce.Do(func() {
		// TODO: parse config.json
		// fix concurrency get center may cause nil pointer
		center = &Center{
			agents: make(map[string]*agent.DigitalAgent),
			ftpServer: ftp.NewFTPServer(&ftp.ServerConfig{
				BaseDir: "/",
				Port:    2121,
			}),
		}
	})
	return center
}
