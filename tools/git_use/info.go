package git_use

import (
	"context"
	"digital-labor/pkg/llm"
	toolpkg "digital-labor/pkg/tool"
	"encoding/json"
	"fmt"
	"strconv"
)

type GitTool struct {
}

func (t *GitTool) Definition(context.Context) (llm.ToolDefinition, error) {
	return llm.ToolDefinition{
		Name: "git_use", Description: "Request confirmation or a choice before a Git operation.",
		Parameters: toolpkg.ObjectSchema(map[string]any{
			"question": map[string]any{"type": "string"},
			"kind":     map[string]any{"type": "string", "enum": []string{"confirm", "option"}},
			"options":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		}, "question", "kind"),
	}, nil
}

type GitUseArguments struct {
	Question string   `json:"question"`
	Kind     string   `json:"kind"`
	Options  []string `json:"options"`
}

func (t *GitTool) Execute(ctx context.Context, arguments json.RawMessage) (*toolpkg.Result, error) {
	args := GitUseArguments{}
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, err
	}

	fmt.Println(args.Question)
	input := ""
	// 根据 args.Kind 决定是 confirm 还是 option
	if args.Kind == "confirm" {
		// 请求用户同意
		fmt.Println("choose allow or deny. allow is y, deny is n")
		fmt.Scanln(&input)
		if input == "y" || input == "yes" {
			return &toolpkg.Result{Content: "yes"}, nil
		} else {
			return &toolpkg.Result{Content: "no"}, nil
		}
	} else if args.Kind == "option" {
		// 列出选项给用户选择
		fmt.Println("choose an option:")
		for i, option := range args.Options {
			fmt.Printf("%d. %s\n", i+1, option)
		}
		fmt.Scanln(&input)
		if input == "" {
			return nil, fmt.Errorf("invalid choice")
		}
		choice, err := strconv.Atoi(input)
		if err != nil {
			return nil, fmt.Errorf("invalid choice")
		}
		if choice < 1 || choice > len(args.Options) {
			return nil, fmt.Errorf("invalid choice")
		}
		return &toolpkg.Result{Content: args.Options[choice-1]}, nil
	}

	return nil, fmt.Errorf("unknown git confirmation kind %q", args.Kind)
}
