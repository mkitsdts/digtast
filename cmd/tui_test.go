package main

import (
	"reflect"
	"testing"
)

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
