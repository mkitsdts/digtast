package memory

type Manager struct {
	mate *Mate
}

func NewManager(userID string) (*Manager, error) {
	var err error

	mg := &Manager{}

	if mg.mate, err = NewMateMemory(); err != nil {
		return nil, err
	}

	return mg, nil
}

func (mg *Manager) RetrieveMemory(sessionID string) error {
	return mg.mate.Retrieve(sessionID)
}
