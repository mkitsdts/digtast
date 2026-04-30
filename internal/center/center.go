package center

import (
	"digital-labor/internal/agent"
	"digital-labor/internal/vdisplay"
	"digital-labor/pkg/conf"
)

// 汇聚全局变量
var (
	AgentManager *agent.Manager
	Vdisplay     *vdisplay.VisualDisplayManager
)

func Init() {
	AgentManager = agent.NewManager()
	AgentManager.LoadAgents()

	if conf.Conf.VNC.Enabled {
		Vdisplay = vdisplay.NewVisualDisplayManager()
	}
}
