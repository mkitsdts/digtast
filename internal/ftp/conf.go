package ftp

type ServerConfig struct {
	BaseDir string
	Port    int
	Enabled bool
}

const (
	begin_sentense = "welcome to digtast"
	end_sentense   = "welcome comming back to digtast"
)
