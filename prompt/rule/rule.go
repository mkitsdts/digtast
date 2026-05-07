package rule

import (
	"digital-labor/pkg/workspace"
	"fmt"
	"os"
)

type Rule struct {
}

const prompt = `以下是你必须要遵守的规则
---
1. 不可以以绕过断言的方式通过测试
2. 用户安排你创作的内容都保存在 ~/.digtast/workspace 这个目录
3 禁止生成任何未经确认的假代码。
4 严禁在回答中表现出过度的人格化（除非子角色明确要求）。
5 始终保持对工作区文件系统（WorkDir）的尊重，执行写操作前默认遵循“安全第一”原则。
`

const prompt_name = "rule"

func (s *Rule) CreatePromptImpl() error {
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

func (s *Rule) GetPromptImpl() (string, error) {
	//TODO:
	workspacePath := workspace.GetWorkspacePath()
	if workspacePath == "" {
		return "", nil
	}

	content, err := os.ReadFile(fmt.Sprintf("%s/%s.md", workspacePath, prompt_name))
	if err != nil {
		return prompt, s.CreatePromptImpl()
	}

	return string(content), nil
}

func (s *Rule) GetPromptName() string {
	return fmt.Sprintf("%s.md", prompt_name)
}

func (s *Rule) GetRole() string {
	return "system"
}

func init() {
	workspace.RegisterPromptCreator(&Rule{})
}
