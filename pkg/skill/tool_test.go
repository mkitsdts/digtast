package skill

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"digital-labor/pkg/conf"
)

func TestSkillPromptAndControlledLoad(t *testing.T) {
	conf.Conf.WorkSpaceDir = t.TempDir()
	manager := NewManagerForAgent("agent-1")
	if err := manager.AddSkill("review", "Review code carefully"); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: review\ndescription: Review code carefully\n---\n\nSECRET INSTRUCTIONS"
	if err := os.WriteFile(filepath.Join(manager.BaseDir(), "review", "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	prompt, err := manager.Prompt(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prompt, "Review code carefully") || strings.Contains(prompt, "SECRET INSTRUCTIONS") {
		t.Fatalf("skill prompt disclosed the wrong content: %s", prompt)
	}
	loadTool, err := NewLoadSkillTool(manager, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(context.Background(), "agent_id", "agent-1")
	result, err := loadTool.Execute(ctx, json.RawMessage(`{"name":"review"}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Content, "SECRET INSTRUCTIONS") {
		t.Fatalf("skill body was not loaded")
	}
	wrongAgent := context.WithValue(context.Background(), "agent_id", "agent-2")
	if _, err := loadTool.Execute(wrongAgent, json.RawMessage(`{"name":"review"}`)); err == nil {
		t.Fatal("expected cross-agent load to be denied")
	}
	if err := manager.DisableSkill("review"); err != nil {
		t.Fatal(err)
	}
	if _, err := loadTool.Execute(ctx, json.RawMessage(`{"name":"review"}`)); err == nil {
		t.Fatal("expected disabled skill load to be denied")
	}
}
