package config

import (
	"testing"

	"github.com/google/uuid"
)

func TestAddAndRemoveModel(t *testing.T) {
	previous := Conf
	t.Cleanup(func() { Conf = previous })
	Conf = DefaultConfig()
	Conf.WorkSpaceDir = t.TempDir()

	id, err := AddModel(ModelConfig{Provider: "openai", ModelNames: []string{"gpt-4o"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := uuid.Parse(id); err != nil {
		t.Fatalf("expected a uuid model key, got %q", id)
	}
	if _, ok := Conf.Models[id]; !ok {
		t.Fatal("model was not stored under the generated key")
	}

	updated := ModelConfig{Provider: "deepseek", ModelNames: []string{"deepseek-chat"}}
	if err := UpdateModel(id, updated); err != nil {
		t.Fatal(err)
	}
	if got := Conf.Models[id]; got.Provider != "deepseek" {
		t.Fatalf("model was not updated: %#v", got)
	}

	if err := RemoveModel(id); err != nil {
		t.Fatal(err)
	}
	if _, ok := Conf.Models[id]; ok {
		t.Fatal("model was not removed")
	}
}
