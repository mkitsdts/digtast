package center

import (
	"digital-labor/internal/agent"
	"digital-labor/internal/ftp"
	"digital-labor/internal/gateway"
	"digital-labor/internal/vdisplay"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/workspace"
	"log/slog"
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
		center = &Center{
			agents: make(map[string]*agent.DigitalAgent),
			ftpServer: ftp.NewFTPServer(&ftp.ServerConfig{
				BaseDir:    "/",
				Port:       conf.Conf.FTP.Port,
				FTPEnabled: conf.Conf.FTP.Enabled,
			}),
			visualDisplay: *vdisplay.GetVisualDisplayManager(),
		}
		// Load persisted agents
		configs, err := workspace.LoadAllAgentConfigs()
		if err != nil {
			slog.Error("failed to load agent configs from workspace", "error", err)
		} else {
			for id, cfg := range configs {
				ag, err := agent.NewDigitalAgent(cfg)
				if err != nil {
					slog.Error("failed to restore agent", "id", id, "error", err)
					continue
				}
				center.agents[id] = ag
				slog.Info("restored agent from workspace", "id", id)
			}
		}
	})
	return center
}
