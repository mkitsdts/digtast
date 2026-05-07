package prompt

import (
	_ "digital-labor/prompt/root"
	_ "digital-labor/prompt/rule"
	_ "digital-labor/prompt/soul"
	_ "digital-labor/prompt/time"
	_ "digital-labor/prompt/user"
)

/*
 * 这里存放的都是系统提示词
 * 如果要新增系统提示词，只需要加一个目录，然后在 var prompt = `` 中定义提示词内容，最后在这里导入包
 * 程序启动的时候会自动写入文件，后续修改提示词内容不会覆盖
 */

/*
 * 模板
package name

 import (
	"digital-labor/pkg/workspace"
	"fmt"
	"os"
 )

 type Name struct {
 }

 var prompt = fmt.Sprintf("current time is %s", time.Now().Format(time.RFC3339))

 const prompt_name = "name"

 func (s *Name) CreatePromptImpl() error {
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

 func (s *Name) GetPromptImpl() (string, error) {
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

 func (s *Name) GetPromptName() string {
	return fmt.Sprintf("%s.md", prompt_name)
 }

 func (s *Name) GetRole() string {
	return "system"
 }

 func init() {
	workspace.RegisterPromptCreator(&Name{})
 }
*/
