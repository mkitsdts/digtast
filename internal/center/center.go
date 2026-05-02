package center

import (
	"digital-labor/internal/agent"
)

// 汇聚全局变量
var (
	AgentManager *agent.Manager
)

func Init() {
	AgentManager = agent.NewManager()
	AgentManager.LoadAgents()
}
