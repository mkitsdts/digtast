package user

import (
	"digital-labor/pkg/workspace"
	"fmt"
	"os"
)

type User struct {
}

const agentUserPrompt = `
# user.md

这是你需要辅助的人的性格，根据性格选择合适的辅助方式

## 内容
`

const prompt_name = "user"

func (s *User) CreatePromptImpl() error {
	workspacePath := workspace.GetWorkspacePath()
	if workspacePath == "" {
		return nil
	}

	file, err := os.OpenFile(fmt.Sprintf("%s/prompt/%s.md", workspacePath, prompt_name), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	_, err = file.WriteString(agentUserPrompt)
	return err
}

func (s *User) GetPromptImpl() (string, error) {
	//TODO:
	workspacePath := workspace.GetWorkspacePath()
	if workspacePath == "" {
		return "", nil
	}

	content, err := os.ReadFile(fmt.Sprintf("%s/prompt/%s.md", workspacePath, prompt_name))
	if err != nil {
		return agentUserPrompt, s.CreatePromptImpl()
	}

	return string(content), nil
}

func (s *User) GetPromptName() string {
	return fmt.Sprintf("%s.md", prompt_name)
}

func (s *User) GetRole() string {
	return "system"
}

func init() {
	workspace.RegisterPromptCreator(&User{})
}
