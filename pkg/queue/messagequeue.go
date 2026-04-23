package queue

type MessageQueue struct {
	q map[string]*Queue[string]
}

func NewMessageQueue() *MessageQueue {
	return &MessageQueue{q: make(map[string]*Queue[string])}
}

func (mq *MessageQueue) Push(key, message string) {
	mq.q[key].Push(message)
}

func (mq *MessageQueue) Pop(userID string) (string, bool) {
	return mq.q[userID].Pop()
}

func (mq *MessageQueue) Front(userID string) (string, bool) {
	return mq.q[userID].Front()
}
