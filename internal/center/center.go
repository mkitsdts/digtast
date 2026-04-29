package center

import (
	"digital-labor/internal/agent"
	"digital-labor/internal/ftp"
	"digital-labor/internal/vdisplay"
	"digital-labor/pkg/conf"
)

// 汇聚全局变量
var (
	AgentManager *agent.Manager
	FtpServer    *ftp.FTPServer
	Vdisplay     *vdisplay.VisualDisplayManager
)

func Init() {
	AgentManager = agent.NewManager()
	AgentManager.LoadAgents()

	if conf.Conf.FTP.Enabled {
		FtpServer = ftp.NewFTPServer(&ftp.ServerConfig{
			BaseDir: conf.Conf.FTP.RootDir,
			Port:    conf.Conf.FTP.Port,
		})
	}
	if conf.Conf.VNC.Enabled {
		Vdisplay = vdisplay.NewVisualDisplayManager()
	}
}
