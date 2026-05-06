package agent

import (
	"context"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/model"
	"testing"
)

func TestChat(t *testing.T) {
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
	ch, err := agent.Run(ctx, "帮我查一下广州的天气", true)
	if err != nil {
		t.Fatal(err)
	}
	for result := range ch {
		t.Log(result)
	}
}
