package memory

type SessionReader interface {
	Messages(sessionID string) ([]string, error)
}

type Manager struct {
	mate    *Mate
	session SessionReader
}

func NewManager(sessionReader SessionReader) (*Manager, error) {
	var err error

	mg := &Manager{
		session: sessionReader,
	}

	if mg.mate, err = NewMateMemory(sessionReader); err != nil {
		return nil, err
	}

	return mg, nil
}

func (mg *Manager) RetrieveMemory(sessionID string) error {
	return mg.mate.Retrieve(sessionID)
}
