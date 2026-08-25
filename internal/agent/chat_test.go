package agent

import (
	"context"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/model"
	"os"
	"testing"
)

func TestChat(t *testing.T) {
	if os.Getenv("TEST_INTEGRATION") == "" {
		t.Skip("set TEST_INTEGRATION=1 to run")
	}
	conf.Conf.Models = map[string]conf.ModelConfig{
		"doubao": {
			Provider:   "doubao",
			Key:        "ark-api-key",
			URL:        "https://ark.cn-beijing.volces.com/api/v3",
			ModelNames: []string{"doubao-seed-2-0-lite-260215"},
		},
	}

	agent, err := newDigitalAgent(&model.DigitalAgentConfig{
		Name:        "test-agent",
		Model:       "doubao",
		Description: "test agent",
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	ch, err := agent.Run(ctx, "帮我查一下广州的天气", nil, true)
	if err != nil {
		t.Fatal(err)
	}
	for result := range ch {
		t.Log(result)
	}
}
