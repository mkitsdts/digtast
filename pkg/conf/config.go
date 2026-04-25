package conf

const (
	IS_DEBUG = true
)

type Config struct {
	MemoryDir    string `yaml:"memory_dir"`
	SkillDir     string `yaml:"skill_dir"`
	RootDir      string `yaml:"root_dir"`
	WorkSpaceDir string `yaml:"workspace_dir"`
}

var Conf Config

func Init() {

}
