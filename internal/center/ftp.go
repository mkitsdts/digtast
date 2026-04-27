package center

type FTPServerParams struct {
	Port int `json:"port"`
}

func (m *Center) StartFTPServer(req *FTPServerParams) error {
	return m.ftpServer.Start()
}

func (m *Center) StopFTPServer() error {
	return m.ftpServer.Stop()
}
