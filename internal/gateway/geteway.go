package gateway

type ChannelGateway struct {
	channels map[string]func(kind string) MessageChannel
}

var register = make(map[string]func(kind string) MessageChannel)
var channelGateway = &ChannelGateway{
	channels: register,
}

func (cg *ChannelGateway) NewChannel(kind string) MessageChannel {
	return cg.channels[kind](kind)
}

func NewChannelGeteway() *ChannelGateway {
	return channelGateway
}

// 提供通道实现注册
func registerChannel(kind string, f func(kind string) MessageChannel) {
	register[kind] = f
}
