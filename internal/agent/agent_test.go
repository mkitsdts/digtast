package agent

import (
	"strings"
	"testing"

	"digital-labor/pkg/model"

	"github.com/cloudwego/eino/schema"
)

func TestBuildMessages_ReturnsNewSlice(t *testing.T) {
	msgs := make([]*schema.Message, 0)
	result, err := buildMessages("hello world", msgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 message in result, got %d", len(result))
	}
	if result[0].Role != schema.User {
		t.Fatalf("expected User role, got %s", result[0].Role)
	}
	if result[0].Content != "hello world" {
		t.Fatalf("expected 'hello world', got '%s'", result[0].Content)
	}
}

func TestBuildMessages_AppendsToExisting(t *testing.T) {
	existing := []*schema.Message{
		{Role: schema.User, Content: "prev"},
	}
	result, err := buildMessages("new msg", existing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result))
	}
	if result[1].Content != "new msg" {
		t.Fatalf("expected 'new msg' as last message, got '%s'", result[1].Content)
	}
}

func TestRenderTools_Empty(t *testing.T) {
	result := renderTools(nil)
	if result != "" {
		t.Fatalf("expected empty for no tools, got '%s'", result)
	}
}

func TestRenderSkills_Empty(t *testing.T) {
	result := renderSkills(nil)
	if result != "" {
		t.Fatalf("expected empty for no skills, got '%s'", result)
	}
}

func TestRenderSkills_Single(t *testing.T) {
	result := renderSkills([]string{"web-search"})
	expected := "## Skills\n- web-search"
	if result != expected {
		t.Fatalf("expected '%s', got '%s'", expected, result)
	}
}

func TestRenderSkills_Multiple(t *testing.T) {
	result := renderSkills([]string{"search", "browser"})
	expected := "## Skills\n- search\n- browser"
	if result != expected {
		t.Fatalf("expected '%s', got '%s'", expected, result)
	}
}

func TestRenderSkills_EmptyStringsIgnored(t *testing.T) {
	result := renderSkills([]string{"", "  ", "valid"})
	expected := "## Skills\n- valid"
	if result != expected {
		t.Fatalf("expected '%s', got '%s'", expected, result)
	}
}

func TestPromptBuilder_Build(t *testing.T) {
	b := NewPromptBuilder()
	ctx := model.PromptContext{
		UserInstruction: "be helpful",
		UserPreference:  "speak English",
	}

	result := b.Build(ctx)
	// loadPrompt() returns "" (TODO), so result should contain instruction and preference
	if !strings.Contains(result, "be helpful") {
		t.Fatal("expected Build to contain user instruction")
	}
	if !strings.Contains(result, "speak English") {
		t.Fatal("expected Build to contain user preference")
	}
}

func TestPromptBuilder_Build_Empty(t *testing.T) {
	b := NewPromptBuilder()
	ctx := model.PromptContext{}
	result := b.Build(ctx)
	// loadPrompt returns "", and all ctx fields are empty
	if result != "" {
		t.Fatalf("expected empty build result, got '%s'", result)
	}
}
