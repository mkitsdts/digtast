package memory

import "digital-labor/internal/agent"

// 长期记忆
type Mate struct {
	Name    string
	Content string
}

func NewMateMemory() (*Mate, error) {
	mate := &Mate{}

	return mate, nil
}

const max_chunk_size = 4096

func (mate *Mate) Retrieve(sessionID string) error {

	m, err := agent.GetDigitalAgent().GetOrCreateSession(sessionID)
	if err != nil {
		return err
	}

	var chunk string
	for _, record := range m.GetMessages() {
		chunk += record.Content
		if len(chunk) > max_chunk_size {
			mate.retrieve(chunk)
			chunk = ""
		}
	}

	return nil
}

func (mate *Mate) retrieve(content string) error {

	return nil
}
