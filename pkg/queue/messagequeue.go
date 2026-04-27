package queue

import (
	"github.com/cloudwego/eino/schema"
)

type MessageQueue struct {
	q map[string]*Queue[*schema.Message]
}

func NewMessageQueue() *MessageQueue {
	return &MessageQueue{q: make(map[string]*Queue[*schema.Message])}
}

func (mq *MessageQueue) Push(key string, message *schema.Message) {
	mq.q[key].Push(message)
}

func (mq *MessageQueue) Pop(key string) (*schema.Message, bool) {
	return mq.q[key].Pop()
}

func (mq *MessageQueue) Front(key string) (*schema.Message, bool) {
	return mq.q[key].Front()
}
