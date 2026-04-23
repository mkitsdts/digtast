package vdisplay

import "digital-labor/pkg/model"

type VirtualDisplay interface {
	GetDesktopDisplay(model.GetDesktopDisplayRequest) (model.GetDesktopDisplayResponse, error)
	ShutdownDesktopDisplay(model.ShutdownDesktopDisplayRequest) (model.ShutdownDesktopDisplayResponse, error)
}
