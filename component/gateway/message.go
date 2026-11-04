package gateway

import "digital-labor/pkg/queue"

// UserMessage 是网关收到用户消息后投递给 Agent 的统一结构。
type UserMessage struct {
	Content      string
	MultiContent map[string][]byte // title: data
	NotifyWay    string
	Params       map[string]any // key-value pairs for notify way
}

var que *queue.Queue[*UserMessage] = queue.NewQueue[*UserMessage]()

// PushUserMessage 由网关实现调用，把用户消息放入待处理队列。
func PushUserMessage(msg *UserMessage) {
	que.Push(msg)
}

// PopUserMessage 取出一条待处理的用户消息。
func PopUserMessage() (*UserMessage, bool) {
	return que.Pop()
}
