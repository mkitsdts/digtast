package webfetch

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type FetchTool struct {
	client *http.Client
	mux    *sync.Mutex
}

func (t *FetchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	properties := orderedmap.New[string, *jsonschema.Schema]()
	properties.Set("url", &jsonschema.Schema{
		Type:        "string",
		Description: "http 请求的远程地址。",
	})
	properties.Set("method", &jsonschema.Schema{
		Type:        "string",
		Description: "http 请求的方法。可以是 GET ， POST ， PUT ， DELETE 等。",
	})
	properties.Set("options", &jsonschema.Schema{
		Type:        "string",
		Description: "http 请求的选项。当 method 是 POST 或 PUT 时，需要此参数提供请求体内容。请求体必须是 json 格式的",
	})

	jsonschema := &jsonschema.Schema{
		Version:    "draft",
		ID:         "test_user_ask",
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

type WebFetchArguments struct {
	Method string `json:"method"`
	URL    string `json:"url"`
	Body   string
}

func (t *FetchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	t.mux.Lock()
	defer t.mux.Unlock()

	args := WebFetchArguments{}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	req, err := http.NewRequest(args.Method, args.URL, strings.NewReader(args.Body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
