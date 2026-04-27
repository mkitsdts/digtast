package vnc

import (
	"digital-labor/pkg/model"
	"errors"
	"sync"
)

const default_beginId = 6901
const default_maxUser = 10
const default_geometry = "1920x1080"

type VNCServer struct {
	vncs     map[string]int
	ids      []int
	states   []bool
	mu       sync.Mutex
	maxUser  int
	beginId  int
	geometry string
	path     string
}

var s *VNCServer

// geometry is the desktop geometry in the format "widthxheight" (e.g. "1920x1080").
func NewVNCServer(maxUser int, beginId int, geometry string, path string) *VNCServer {
	if beginId == 0 {
		beginId = default_beginId
	}
	if maxUser <= 0 {
		maxUser = default_maxUser
	}
	s = &VNCServer{
		vncs:     make(map[string]int),
		ids:      make([]int, maxUser),
		states:   make([]bool, maxUser),
		maxUser:  maxUser,
		beginId:  beginId,
		geometry: geometry,
		path:     path,
		mu:       sync.Mutex{},
	}
	for i := 0; i < maxUser; i++ {
		s.ids[i] = beginId + i
	}
	return s
}

func GetVNCServer() *VNCServer {
	return s
}

// GetDesktopDisplay start a connection to the VNC server and returns the connection details.
func (s *VNCServer) GetDesktopDisplay(req model.GetDesktopDisplayRequest) (model.GetDesktopDisplayResponse, error) {
	port, err := s.AddVNCService(req.Key)
	if err != nil {
		return model.GetDesktopDisplayResponse{}, err
	}

	return model.GetDesktopDisplayResponse{Port: port}, nil
}

func (s *VNCServer) ShutdownDesktopDisplay(req model.ShutdownDesktopDisplayRequest) (model.ShutdownDesktopDisplayResponse, error) {
	id, ok := s.GetVNCService(req.Key)
	if !ok {
		return model.ShutdownDesktopDisplayResponse{}, errors.New("vnc service not found")
	}

	err := stopVNC(id)
	if err != nil {
		return model.ShutdownDesktopDisplayResponse{}, err
	}

	s.RemoveVNCService(req.Key)
	return model.ShutdownDesktopDisplayResponse{}, nil
}
