package center

type FTPServerParamater struct {
}

func (m *Center) StartFTPServer(req *FTPServerParamater) error {
	return m.ftpServer.Start()
}

func (m *Center) StopFTPServer() error {
	return m.ftpServer.Stop()
}
