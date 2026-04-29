package vdisplay

import "digital-labor/pkg/model"

type VirtualDisplay interface {
	GetDesktopDisplay(model.GetDesktopDisplayRequest) (model.GetDesktopDisplayResponse, error)
	ShutdownDesktopDisplay(model.ShutdownDesktopDisplayRequest) (model.ShutdownDesktopDisplayResponse, error)
}

var visualDisplays = make(map[string]VirtualDisplay)

func Register(kind string, vd VirtualDisplay) {
	visualDisplays[kind] = vd
}

func GetVisualDisplay() map[string]VirtualDisplay {
	return visualDisplays
}
