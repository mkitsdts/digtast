package time

import (
	"digital-labor/pkg/workspace"
	"fmt"
	"os"
	"time"
)

type Time struct {
}

var prompt = fmt.Sprintf("current time is %s", time.Now().Format(time.RFC3339))

const prompt_name = "user"

func (s *Time) CreatePromptImpl() error {
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

	_, err = file.WriteString(prompt)
	return err
}

func (s *Time) GetPromptImpl() (string, error) {
	//TODO:
	workspacePath := workspace.GetWorkspacePath()
	if workspacePath == "" {
		return "", nil
	}

	content, err := os.ReadFile(fmt.Sprintf("%s/prompt/%s.md", workspacePath, prompt_name))
	if err != nil {
		return prompt, s.CreatePromptImpl()
	}

	return string(content), nil
}

func (s *Time) GetPromptName() string {
	return fmt.Sprintf("%s.md", prompt_name)
}

func (s *Time) GetRole() string {
	return "system"
}

func init() {
	workspace.RegisterPromptCreator(&Time{})
}
