package skill

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"gopkg.in/yaml.v2"
)

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Skill) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: s.Name,
		Desc: s.Description,
	}, nil
}

var skills = make([]tool.BaseTool, 0)

func GetSkills() []tool.BaseTool {
	return skills
}

func ScanSkills(path string) {
	files, err := os.ReadDir(path)
	if err != nil {
		return
	}
	for _, file := range files {
		// if not dir, skip because skill is a dir
		if !file.IsDir() {
			continue
		}
		// parse the skill.md
		skillPath := path + "/" + file.Name()
		result, err := registerSkill(skillPath)
		if err != nil {
			continue
		}
		skills = append(skills, result)
	}
}

func registerSkill(path string) (tool.BaseTool, error) {
	// try to open the skill.md
	f, err := os.ReadFile(path + "/SKILL.md")
	if err != nil {
		f, err = os.ReadFile(path + "/skill.md")
		if err != nil {
			return nil, err
		}
	}
	result, err := parseSkill(string(f))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// parse the front yaml head in markdown
func parseSkill(content string) (tool.BaseTool, error) {
	const delimiter = "---"
	if !strings.HasPrefix(content, delimiter) {
		return nil, errors.New("invalid skill format")
	}

	parts := strings.SplitN(content, delimiter, 3)

	var result *Skill

	if err := yaml.Unmarshal([]byte(parts[1]), result); err != nil {
		return nil, err
	}

	if result.Name == "" || result.Description == "" {
		return nil, errors.New("invalid skill format")
	}

	return result, nil
}
