package main

import (
	"reflect"
	"testing"

	"digital-labor/component/chatmodel"
	"digital-labor/config"
	"digital-labor/internal/agent/skill"
)

func TestProviderViewListsRegisteredProviders(t *testing.T) {
	view := newProviderView()
	if len(view.providers) == 0 {
		t.Fatal("expected registered providers in the provider view")
	}
	for i := 1; i < len(view.providers); i++ {
		if view.providers[i-1].ID > view.providers[i].ID {
			t.Fatalf("providers are not sorted: %#v", view.providers)
		}
	}
	if _, ok := view.current(); !ok {
		t.Fatal("expected a current selection")
	}
}

func TestProviderStatusMarksCurrent(t *testing.T) {
	previous := config.Conf
	t.Cleanup(func() { config.Conf = previous })
	config.Conf.Agent = nil

	provider, ok := chatmodel.FindProvider("deepseek")
	if !ok {
		t.Skip("deepseek provider is not registered in this test binary")
	}
	if isCurrentProvider(provider) {
		t.Fatal("no selected provider should not be current")
	}
	config.Conf.Agent = &config.AgentConfig{Provider: "deepseek", Model: "deepseek-chat"}
	if !isCurrentProvider(provider) {
		t.Fatal("selected provider should be current")
	}
	if !isCurrentModel("deepseek", "deepseek-chat") || isCurrentModel("deepseek", "deepseek-reasoner") {
		t.Fatal("unexpected current model check")
	}
}

func TestModelLabel(t *testing.T) {
	if got := modelLabel("deepseek", "deepseek-chat"); got != "DeepSeek · deepseek-chat" {
		t.Fatalf("unexpected label %q", got)
	}
	if got := modelLabel("unknown", "m"); got != "m" {
		t.Fatalf("unexpected fallback label %q", got)
	}
}

func TestSkillViewToggle(t *testing.T) {
	previous := config.Conf
	t.Cleanup(func() { config.Conf = previous })
	config.Conf = config.DefaultConfig()
	config.Conf.WorkSpaceDir = t.TempDir()

	manager := skill.NewManager()
	if err := manager.AddSkill("review", "Review code carefully"); err != nil {
		t.Fatal(err)
	}
	view := &skillView{skills: []skill.SkillInfo{
		{Name: "review", Description: "Review code carefully", Enabled: true},
	}}
	m := &uiModel{skillsView: view}
	m.toggleSkill()
	if !config.IsSkillDisabled("review") {
		t.Fatal("expected review to be disabled after toggle")
	}
	if view.skills[0].Enabled {
		t.Fatal("expected the view state to flip")
	}
	m.toggleSkill()
	if config.IsSkillDisabled("review") {
		t.Fatal("expected review to be re-enabled")
	}
}

func TestPickerFiltersAndSelects(t *testing.T) {
	picker := newPickerView("models")
	picker.items = []pickerItem{
		{label: "gpt-4o", value: "gpt-4o"},
		{label: "gpt-4o-mini", value: "gpt-4o-mini"},
		{label: "deepseek-chat", value: "deepseek-chat"},
	}
	picker.filter = "gpt"
	if got := len(picker.filtered()); got != 2 {
		t.Fatalf("expected 2 filtered items, got %d", got)
	}
	picker.filter = "deepseek"
	if item, ok := picker.current(); !ok || item.value != "deepseek-chat" {
		t.Fatalf("unexpected current item: %#v", item)
	}
	picker.multi = true
	picker.toggle()
	if got := picker.selectedValues(); !reflect.DeepEqual(got, []string{"deepseek-chat"}) {
		t.Fatalf("unexpected selection: %v", got)
	}
}

func TestParseParams(t *testing.T) {
	got, err := parseParams("appid=1020, port=8080")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"appid": "1020", "port": "8080"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if _, err := parseParams("broken"); err == nil {
		t.Fatal("expected an error for a parameter without '='")
	}
}
