package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type formFieldSpec struct {
	Label       string
	Help        string
	Value       string
	Placeholder string
	Secret      bool
}

type formField struct {
	spec  formFieldSpec
	input textinput.Model
}

// formView is a small keyboard-driven form used by the model and channel
// configuration screens.
type formView struct {
	title    string
	fields   []formField
	cursor   int
	err      string
	done     bool
	toast    string
	width    int
	onSubmit func(values []string) (string, error)
}

func newFormView(title string, specs []formFieldSpec, onSubmit func(values []string) (string, error)) *formView {
	f := &formView{title: title, onSubmit: onSubmit}
	for _, spec := range specs {
		ti := textinput.New()
		ti.Prompt = "› "
		ti.Placeholder = spec.Placeholder
		ti.SetValue(spec.Value)
		ti.CursorEnd()
		if spec.Secret {
			ti.EchoMode = textinput.EchoPassword
		}
		f.fields = append(f.fields, formField{spec: spec, input: ti})
	}
	f.focus(0)
	return f
}

func (f *formView) focus(index int) {
	if len(f.fields) == 0 {
		return
	}
	if index < 0 {
		index = len(f.fields) - 1
	}
	if index >= len(f.fields) {
		index = 0
	}
	for i := range f.fields {
		if i == index {
			f.fields[i].input.Focus()
		} else {
			f.fields[i].input.Blur()
		}
	}
	f.cursor = index
}

func (f *formView) update(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if ok {
		switch key.String() {
		case "tab", "down":
			f.focus(f.cursor + 1)
			return nil
		case "shift+tab", "up":
			f.focus(f.cursor - 1)
			return nil
		}
	}
	if len(f.fields) == 0 {
		return nil
	}
	var cmd tea.Cmd
	f.fields[f.cursor].input, cmd = f.fields[f.cursor].input.Update(msg)
	return cmd
}

func (f *formView) submit() error {
	values := make([]string, len(f.fields))
	for i := range f.fields {
		values[i] = strings.TrimSpace(f.fields[i].input.Value())
	}
	f.err = ""
	toast, err := f.onSubmit(values)
	if err != nil {
		f.err = err.Error()
		return err
	}
	f.toast = toast
	f.done = true
	return nil
}

func (f *formView) setWidth(width int) {
	f.width = width
	fieldWidth := width - 8
	if fieldWidth < 24 {
		fieldWidth = 24
	}
	if fieldWidth > 72 {
		fieldWidth = 72
	}
	for i := range f.fields {
		f.fields[i].input.Width = fieldWidth
	}
}

func (f *formView) view() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render(f.title))
	b.WriteString("\n\n")
	for i := range f.fields {
		field := &f.fields[i]
		label := field.spec.Label
		if i == f.cursor {
			label = styleFocus.Render("▸ " + label)
		} else {
			label = styleMuted.Render("  " + label)
		}
		b.WriteString(label)
		b.WriteString("\n")
		b.WriteString(field.input.View())
		b.WriteString("\n")
		if field.spec.Help != "" {
			b.WriteString(styleMuted.Render("  " + field.spec.Help))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	if f.err != "" {
		text := "✗ " + f.err
		if f.width > 4 {
			b.WriteString(styleError.Width(f.width).Render(text))
		} else {
			b.WriteString(styleError.Render(text))
		}
		b.WriteString("\n\n")
	}
	return b.String()
}
