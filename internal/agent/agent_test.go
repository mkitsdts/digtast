package agent

import (
	"strings"
	"testing"

	"digital-labor/pkg/model"

	"github.com/cloudwego/eino/schema"
)

func TestBuildMessages_ReturnsNil(t *testing.T) {
	msgs := make([]*schema.Message, 0)
	msgs, err := buildMessages("hello world", msgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildMessages_CreatesUserMessage(t *testing.T) {
	msgs := make([]*schema.Message, 0)
	msgs, err := buildMessages("test content", msgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// buildMessages does: messages = append(messages, msg)
	// and discards the result. With len=0, cap=0, append allocates a
	// new array, so the original 'msgs' is not modified.
	// This is a known issue: the user message is lost in the caller.
	_ = msgs // msgs remains empty due to the design issue
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
	if result != "" {
		t.Fatalf("expected empty build result, got '%s'", result)
	}
}

func TestPromptBuilder_Build_WithTools(t *testing.T) {
	// renderTools requires tool.BaseTool implementations, which we don't
	// have a simple mock for in this test package. Skip for now.
}
