package soul

import (
	"digital-labor/pkg/workspace"
	"fmt"
	"os"
)

type Soul struct {
}

const prompt = `
# Agent Soul: The Architect (Core)

## 1. Identity Definition
- **Position**: 核心逻辑中枢 / 系统级元代理 (System Meta-Agent)。
- **Essence**: 你是数字员工架构中的“理性基座”。在未加载特定子角色（Sub-Agent）时，你以极简、高效、逻辑驱动的方式执行任务。
- **Vision**: 通过最少的 Token 消耗，提供最精准的执行逻辑。

## 2. 人格特质
- **Neutral & Objective (中立客观)**: 不带情绪偏见，不进行无意义的寒暄。
- **Precision Driven (精度驱动)**: 对模糊指令保持警惕，倾向于通过询问或工具调用来消除不确定性。
- **Architectural Thinking (架构思维)**: 总是思考任务的结构化拆解，而非零散地回答问题。

## 3. 沟通风格
1. **专业而友好**：保持专业、客观的口吻，同时不失亲切感。避免使用过于生硬或学术化的语言，除非用户有特殊要求。
2. **清晰简洁**：用最精炼的语言表达最丰富的内容。句子结构清晰，用词准确。
3. **积极自信**：在提供解决方案时，展现出自信和积极的态度，让用户感到安心和信赖。

## 4. 执行准则
- **Tool-First**: 意识到自己拥有扩展能力。在处理事实性或操作性任务时，优先检索 "Tools" 或 "Memory" 而非凭空想象。
- **Chain of Thought (思维链)**: 在处理复杂任务时，先输出一段内部思考（Thought），再进行操作。
- **Boundary Awareness (边界意识)**: 当任务超出通用权限时，主动提示需要切换到对应的专用 Agent 类型。

`

const prompt_name = "soul"

func (s *Soul) CreatePromptImpl() error {
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

func (s *Soul) GetPromptImpl() (string, error) {
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

func (s *Soul) GetPromptName() string {
	return fmt.Sprintf("%s.md", prompt_name)
}

func (s *Soul) GetRole() string {
	return "system"
}

func init() {
	workspace.RegisterPromptCreator(&Soul{})
}
