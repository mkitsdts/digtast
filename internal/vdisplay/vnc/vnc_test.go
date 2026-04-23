package vnc

import (
	"digital-labor/pkg/model"
	"testing"
	"time"
)

func TestVncServer(t *testing.T) {
	vncServer := NewVNCServer(10, 6009, "1920x1080", `C:\Program Files\TightVNC\tvnserver.exe`)
	resp, err := vncServer.GetDesktopDisplay(model.GetDesktopDisplayRequest{
		Key: "user-1",
	})
	if err != nil {
		t.Fatalf("failed to get desktop display: %v", err)
	}
	if resp.Port == 0 {
		t.Fatalf("port is not set")
	}
	t.Logf("port: %d", resp.Port)
	time.Sleep(1 * time.Hour)
}
