package vdisplay

import (
	"digital-labor/pkg/model"
)

type VisualDisplayManager struct {
	visualDisplays map[string]VirtualDisplay
}

const (
	vnc = "vnc"

	default_visual_display_kind = "vnc"
)

var (
	vdm *VisualDisplayManager = &VisualDisplayManager{
		visualDisplays: make(map[string]VirtualDisplay),
	}
)

func GetVisualDisplayManager() *VisualDisplayManager {
	return vdm
}

// don't need mutex because the map finish initial before it being read
func (vdm *VisualDisplayManager) GetOrStartVisualDisplay(req model.GetDesktopDisplayRequest) (model.GetDesktopDisplayResponse, error) {
	if _, ok := vdm.visualDisplays[default_visual_display_kind]; !ok {
		return vdm.visualDisplays[req.Kind].GetDesktopDisplay(req)
	}
	return vdm.visualDisplays[req.Kind].GetDesktopDisplay(req)
}

func (vdm *VisualDisplayManager) ShutdownVisualDisplay(req model.ShutdownDesktopDisplayRequest) (model.ShutdownDesktopDisplayResponse, error) {
	if _, ok := vdm.visualDisplays[req.Kind]; ok {
		return vdm.visualDisplays[default_visual_display_kind].ShutdownDesktopDisplay(req)
	}
	return vdm.visualDisplays[req.Kind].ShutdownDesktopDisplay(req)
}

func Register(kind string, vd VirtualDisplay) {
	GetVisualDisplayManager().visualDisplays[kind] = vd
}
