package conf

const (
	IS_DEBUG = true
)

type Config struct {
	MemoryDir    string `yaml:"memory_dir"`
	SkillDir     string `yaml:"skill_dir"`
	RootDir      string `yaml:"root_dir"`
	WorkSpaceDir string `yaml:"workspace_dir"`
	Memory
}

type Memory struct {
	MaxMessagesSize int64 `yaml:"max_messages_size"`
}

var Conf Config

func Init() {

}
