package agent

import (
	"context"
	"os"
	"testing"

	_ "digital-labor/chatmodel/deepseek"
	"digital-labor/config"
)

func TestChat(t *testing.T) {
	if os.Getenv("TEST_INTEGRATION") == "" {
		t.Skip("set TEST_INTEGRATION=1 to run")
	}
	config.Conf.Models = map[string]config.ModelConfig{
		"deepseek": {Key: os.Getenv("DEEPSEEK_API_KEY")},
	}

	agent, err := newAgent(&config.AgentConfig{
		Name:        "test-agent",
		Model:       "deepseek-chat",
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
