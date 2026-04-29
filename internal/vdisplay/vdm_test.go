package vdisplay

import (
	"digital-labor/pkg/model"
	"testing"
)

func TestVisualDisplayManager(t *testing.T) {
	vdm := NewVisualDisplayManager()
	if vdm == nil {
		t.Log("vdm is nil")
	} else {
		t.Log("vdm test success")
	}
}

func TestVisualDisplayManagerDisplays(t *testing.T) {
	vdm := NewVisualDisplayManager()
	if vdm == nil {
		t.Log("vdm is nil")
	} else {
		vdm.GetOrStartVisualDisplay(model.GetDesktopDisplayRequest{
			Key:  "114514",
			Kind: "vnc",
		})
	}
}
