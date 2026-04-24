package center

// func (c *Center) FTPRunningState() bool {
// 	c.mu.RLock()
// 	defer c.mu.RUnlock()
// 	return c.
// }

func (c *Center) GatewayRunningState() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.channelGateway.GatewayRunningState()
}
