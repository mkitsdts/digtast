package agent

import (
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestExtractSummaryText_Content(t *testing.T) {
	msg := &schema.Message{
		Role:    schema.User,
		Content: "this is a summary",
	}
	result := extractSummaryText(msg)
	if result != "this is a summary" {
		t.Fatalf("expected 'this is a summary', got '%s'", result)
	}
}

func TestExtractSummaryText_MultiContent(t *testing.T) {
	msg := &schema.Message{
		Role: schema.User,
		UserInputMultiContent: []schema.MessageInputPart{
			{Type: schema.ChatMessagePartTypeText, Text: "part one"},
			{Type: schema.ChatMessagePartTypeText, Text: "part two"},
		},
	}
	result := extractSummaryText(msg)
	if result != "part one\npart two" {
		t.Fatalf("expected 'part one\\npart two', got '%s'", result)
	}
}

func TestExtractSummaryText_Nil(t *testing.T) {
	result := extractSummaryText(nil)
	if result != "" {
		t.Fatalf("expected empty, got '%s'", result)
	}
}
