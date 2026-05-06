package git_use

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type GitTool struct {
}

func (t *GitTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	properties := orderedmap.New[string, *jsonschema.Schema]()
	properties.Set("question", &jsonschema.Schema{
		Type:        "string",
		Description: "发送给用户的问题内容，应当清晰且具体。",
	})
	properties.Set("kind", &jsonschema.Schema{
		Type:        "string",
		Description: "用户同意的类型，如果是请求就需要用户 confirm，如果是询问意见就是 option 然后在 options 列出选项：confirm or option。",
	})
	properties.Set("options", &jsonschema.Schema{
		Type:        "array",
		Description: "当用户同意类型是option的时候，需要此参数提供给用户选项",
		Items: &jsonschema.Schema{
			Type:        "string",
			Description: "选项内容",
		},
	})

	jsonschema := &jsonschema.Schema{
		Version:    "v0.1",
		ID:         "git_use",
		Type:       "object",
		Properties: properties,
		Required:   []string{"question", "option"},
	}

	return &schema.ToolInfo{
		Name:        "ask_user",
		Desc:        "ask user for detail information",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(jsonschema),
	}, nil
}

type GitUseArguments struct {
	Question string   `json:"question"`
	Kind     string   `json:"kind"`
	Options  []string `json:"options"`
}

func (t *GitTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// TODO: 加入漂亮的交互式
	args := GitUseArguments{}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	fmt.Println(args.Question)
	input := ""
	// 根据 args.Kind 决定是 confirm 还是 option
	if args.Kind == "confirm" {
		// 请求用户同意
		fmt.Println("choose allow or deny. allow is y, deny is n")
		fmt.Scanln(&input)
		if input == "y" || input == "yes" {
			return "yes", nil
		} else {
			return "no", nil
		}
	} else if args.Kind == "option" {
		// 列出选项给用户选择
		fmt.Println("choose an option:")
		for i, option := range args.Options {
			fmt.Printf("%d. %s\n", i+1, option)
		}
		fmt.Scanln(&input)
		if input == "" {
			return "", fmt.Errorf("invalid choice")
		}
		choice, err := strconv.Atoi(input)
		if err != nil {
			return "", fmt.Errorf("invalid choice")
		}
		if choice < 1 || choice > len(args.Options) {
			return "", fmt.Errorf("invalid choice")
		}
		return args.Options[choice-1], nil
	}

	return "unknown error", nil
}
