package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type pickerItem struct {
	label    string
	value    string
	selected bool
}

// pickerView is a selectable list with an inline type-to-filter field. It is
// used for provider and model selection so users never have to type names.
type pickerView struct {
	title   string
	items   []pickerItem
	cursor  int
	filter  string
	loading bool
	err     string
	multi   bool
	width   int
}

func newPickerView(title string) *pickerView {
	return &pickerView{title: title}
}

func (p *pickerView) filtered() []int {
	indices := make([]int, 0, len(p.items))
	filter := strings.ToLower(strings.TrimSpace(p.filter))
	for i, item := range p.items {
		if filter == "" || strings.Contains(strings.ToLower(item.label), filter) {
			indices = append(indices, i)
		}
	}
	return indices
}

func (p *pickerView) clamp() {
	indices := p.filtered()
	if len(indices) == 0 {
		p.cursor = 0
		return
	}
	if p.cursor >= len(indices) {
		p.cursor = len(indices) - 1
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
}

func (p *pickerView) move(delta int) {
	p.cursor += delta
	p.clamp()
}

func (p *pickerView) current() (pickerItem, bool) {
	indices := p.filtered()
	if len(indices) == 0 || p.cursor >= len(indices) {
		return pickerItem{}, false
	}
	return p.items[indices[p.cursor]], true
}

func (p *pickerView) toggle() {
	indices := p.filtered()
	if len(indices) == 0 || p.cursor >= len(indices) {
		return
	}
	item := &p.items[indices[p.cursor]]
	item.selected = !item.selected
}

func (p *pickerView) selectedValues() []string {
	values := make([]string, 0, len(p.items))
	for _, item := range p.items {
		if item.selected {
			values = append(values, item.value)
		}
	}
	return values
}

// handleKey consumes navigation and filter typing. It returns false when the
// caller should handle the key itself.
func (p *pickerView) handleKey(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "up":
		p.move(-1)
		return true
	case "down":
		p.move(1)
		return true
	case "backspace":
		if p.filter != "" {
			runes := []rune(p.filter)
			p.filter = string(runes[:len(runes)-1])
			p.cursor = 0
		}
		return true
	case "esc":
		if p.filter != "" {
			p.filter = ""
			p.cursor = 0
			return true
		}
		return false
	}
	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && string(msg.Runes) != " " {
		p.filter += string(msg.Runes)
		p.cursor = 0
		return true
	}
	return false
}

func (p *pickerView) wrap(style lipgloss.Style, text string) string {
	if p.width > 4 {
		return style.Width(p.width).Render(text)
	}
	return style.Render(text)
}

func (p *pickerView) view(maxRows int, spinner string) string {
	var b strings.Builder
	b.WriteString(styleTitle.Render(p.title))
	b.WriteString("\n\n")

	if p.loading {
		b.WriteString(spinner + " " + styleMuted.Render("正在获取模型列表…"))
		b.WriteString("\n")
	} else if p.err != "" {
		b.WriteString(p.wrap(styleError, "✗ "+p.err))
		b.WriteString("\n")
	}

	if p.filter != "" {
		b.WriteString(p.wrap(styleMuted, "过滤: "+p.filter))
		b.WriteString("\n")
	}

	indices := p.filtered()
	if len(indices) == 0 && !p.loading && p.err == "" {
		b.WriteString(styleMuted.Render("没有匹配的条目。"))
		b.WriteString("\n")
	}

	if maxRows < 1 {
		maxRows = 1
	}
	start := 0
	if p.cursor >= maxRows {
		start = p.cursor - maxRows + 1
	}
	end := start + maxRows
	if end > len(indices) {
		end = len(indices)
	}
	for i := start; i < end; i++ {
		item := p.items[indices[i]]
		prefix := "  "
		if i == p.cursor {
			prefix = "▸ "
		}
		if p.multi {
			checked := "[ ]"
			if item.selected {
				checked = "[x]"
			}
			prefix += checked + " "
		}
		line := prefix + item.label
		if i == p.cursor {
			b.WriteString(styleSelected.Render(line))
		} else {
			b.WriteString(styleBody.Render(line))
		}
		b.WriteString("\n")
	}
	if len(indices) > maxRows {
		b.WriteString(styleMuted.Render(fmt.Sprintf("%d / %d", p.cursor+1, len(indices))))
		b.WriteString("\n")
	}
	return b.String()
}
