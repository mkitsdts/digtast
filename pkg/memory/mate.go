package memory

// Mate is long-term memory storage built from session records.
type Mate struct {
	Name    string
	Content string

	session SessionReader
}

func NewMateMemory(sessionReader SessionReader) (*Mate, error) {
	mate := &Mate{
		session: sessionReader,
	}

	return mate, nil
}

const maxChunkSize = 4096

func (mate *Mate) Retrieve(sessionID string) error {
	if mate.session == nil {
		return nil
	}

	messages, err := mate.session.Messages(sessionID)
	if err != nil {
		return err
	}

	var chunk string
	for _, record := range messages {
		chunk += record
		if len(chunk) > maxChunkSize {
			if err := mate.retrieve(chunk); err != nil {
				return err
			}
			chunk = ""
		}
	}

	if chunk != "" {
		return mate.retrieve(chunk)
	}

	return nil
}

func (mate *Mate) retrieve(content string) error {
	return nil
}
