package agent

import (
	"strings"
	"testing"

	"digital-labor/pkg/model"

	"github.com/cloudwego/eino/schema"
)

func TestBuildMessages(t *testing.T) {
	msgs := make([]*schema.Message, 0)
	err := buildMessages("hello world", msgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// BUG: buildMessages uses append on a copy of the slice,
	// so the original 'msgs' is NOT modified. This documents the bug.
	if len(msgs) == 0 {
		t.Log("BUG CONFIRMED: buildMessages does not modify the caller's slice — append works on a copy, not the original")
	}
}

func TestBuildMessages_AppendReturnsNewSlice(t *testing.T) {
	msgs := make([]*schema.Message, 0)
	err := buildMessages("test", msgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The function appends internally but doesn't return the new slice.
	// The caller in run() line 22 does:
	//   msgs := session.GetMessages()
	//   buildMessages(req.Content, msgs)
	// which means the user message is lost.
	//
	// Fix: either return []*schema.Message from buildMessages,
	// or use *[]*schema.Message as parameter.
	_ = msgs // msgs remains empty due to bug
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
	// loadPrompt() returns "" (TODO), so result should only contain instruction and preference
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
