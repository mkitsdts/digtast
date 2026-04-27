package ftp

import (
	"errors"
	"log/slog"

	server "github.com/fclairamb/ftpserverlib"
)

type FTPServer struct {
	server *server.FtpServer
	ch     chan error
}

func NewFTPServer(cfg *ServerConfig) *FTPServer {
	if !cfg.FTPEnabled {
		return nil
	}

	s := &FTPServer{}
	// default base dir is root dir
	if cfg.BaseDir == "" {
		cfg.BaseDir = "/"
	}

	driver := &MainDriver{
		BaseDir: cfg.BaseDir,
		Port:    cfg.Port,
	}

	s.server = server.NewFtpServer(driver)
	s.ch = make(chan error)

	return s
}

func (s *FTPServer) Start() error {

	if s.ch == nil {
		s.ch = make(chan error)
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			slog.Error("FTP Server start falied", "error", err)
			s.ch <- err
		}
		s.ch <- nil
	}()

	go func() {
		err := <-s.ch
		if err != nil {
			slog.Error("ftp server stop", "error", err)
			return
		}
		slog.Info("ftp server stop")
	}()

	return nil
}

func (s *FTPServer) Stop() error {
	if s.ch == nil {
		return errors.New("ftp not start")
	}

	s.ch <- errors.New("stop")

	return nil
}
