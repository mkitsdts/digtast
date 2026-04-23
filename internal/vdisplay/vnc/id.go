package vnc

import "errors"

func (s *VNCServer) getNextPort() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := 0; i < s.maxUser; i++ {
		if !s.states[i] {
			s.states[i] = true
			return (s.beginId + i)
		}
	}
	return -1
}

func (v *VNCServer) AddVNCService(key string) (int, error) {
	id := v.getNextPort()
	if id < 0 {
		return -1, errors.New("no available id")
	}
	s.mu.Lock()
	v.vncs[key] = id
	s.mu.Unlock()
	return id, nil
}

func (v *VNCServer) GetVNCService(key string) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := v.vncs[key]
	return id, ok
}

func (v *VNCServer) RemoveVNCService(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(v.vncs, key)
	s.states[v.beginId%v.maxUser] = false
}
