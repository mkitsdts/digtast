package ftp

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"

	server "github.com/fclairamb/ftpserverlib"
	"github.com/spf13/afero"
)

type MainDriver struct {
	BaseDir string
	Port    int
}

func (driver *MainDriver) GetSettings() (*server.Settings, error) {
	return &server.Settings{
		ListenAddr: fmt.Sprintf(":%d", driver.Port),
	}, nil
}

func (driver *MainDriver) ClientConnected(cc server.ClientContext) (string, error) {
	slog.Info("client is connected", "addr", cc.RemoteAddr())
	return begin_sentense, nil
}

func (driver *MainDriver) ClientDisconnected(cc server.ClientContext) {
	slog.Info("client is disconnected", "addr", cc.RemoteAddr())
}

func (driver *MainDriver) AuthUser(cc server.ClientContext, user, pass string) (server.ClientDriver, error) {
	os.MkdirAll(driver.BaseDir, 0755)

	// use afero's OsFs to connect local
	baseFs := afero.NewOsFs()

	userFs := afero.NewBasePathFs(baseFs, driver.BaseDir)

	return userFs, nil
}

// TLS config
func (driver *MainDriver) GetTLSConfig() (*tls.Config, error) {
	return nil, nil
}
