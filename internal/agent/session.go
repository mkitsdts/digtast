package agent

import (
	mem "digital-labor/internal/memory"

	"github.com/google/uuid"
)

func (ag *DigitalAgent) GetOrCreateSession(sessionId string) (string, *mem.Session) {
	//TODO:
	if sessionId == "" {
		sessionId = uuid.New().String()
	}

	sess, err := ag.memory.GetOrCreate(sessionId)
	if err != nil {
		return "", nil
	}

	return sessionId, sess
}
