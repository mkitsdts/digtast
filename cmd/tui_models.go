package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"digital-labor/config"
	"digital-labor/internal/agent"
	"digital-labor/pkg/chatmodel/openai"

	tea "github.com/charmbracelet/bubbletea"
)

type modelChoice struct {
	id       string
	provider string
	kind     string
}

func (c modelChoice) label() string {
	label := openai.ProviderLabel(c.provider)
	if c.kind != "" {
		label += " · " + c.kind
	}
	return label
}

type modelView struct {
	choices       []modelChoice
	cursor        int
	confirmDelete bool
}

func newModelView() *modelView {
	v := &modelView{}
	v.reload()
	return v
}

func (v *modelView) reload() {
	v.choices = v.choices[:0]
	ids := make([]string, 0, len(config.Conf.Models))
	for id := range config.Conf.Models {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		mc := config.Conf.Models[id]
		if len(mc.ModelNames) == 0 {
			v.choices = append(v.choices, modelChoice{id: id, provider: mc.Provider})
			continue
		}
		for _, kind := range mc.ModelNames {
			v.choices = append(v.choices, modelChoice{id: id, provider: mc.Provider, kind: kind})
		}
	}
	sort.SliceStable(v.choices, func(i, j int) bool {
		return v.choices[i].label() < v.choices[j].label()
	})
	if v.cursor >= len(v.choices) {
		v.cursor = len(v.choices) - 1
	}
	if v.cursor < 0 {
		v.cursor = 0
	}
}

func (v *modelView) current() (modelChoice, bool) {
	if len(v.choices) == 0 {
		return modelChoice{}, false
	}
	return v.choices[v.cursor], true
}

func (m *uiModel) openModels() {
	if _, err := agent.Instance(); err != nil {
		m.setStatus(false, "还没有可用模型，按 a 添加")
	}
	m.models = newModelView()
	m.screen = screenModels
}

func (m *uiModel) updateModels(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	v := m.models
	key := msg.String()
	if v.confirmDelete && key != "d" {
		v.confirmDelete = false
	}
	switch key {
	case "esc":
		if _, err := agent.Instance(); err != nil {
			m.setStatus(true, "请先选择一个模型")
			return m, nil
		}
		m.screen = screenChat
		m.input.Focus()
		return m, nil
	case "up", "k":
		if v.cursor > 0 {
			v.cursor--
		}
	case "down", "j":
		if v.cursor < len(v.choices)-1 {
			v.cursor++
		}
	case "a":
		m.startModelWizard("")
	case "e":
		if c, ok := v.current(); ok {
			m.startModelWizard(c.id)
		}
	case "d":
		m.deleteModel()
	case "enter":
		if c, ok := v.current(); ok {
			m.applyModel(c.id, c.kind)
		}
	}
	return m, nil
}

func (m *uiModel) deleteModel() {
	v := m.models
	c, ok := v.current()
	if !ok {
		return
	}
	if !v.confirmDelete {
		v.confirmDelete = true
		return
	}
	v.confirmDelete = false
	if err := config.RemoveModel(c.id); err != nil {
		m.setStatus(true, "删除失败: %v", err)
		return
	}
	if config.Conf.Agent != nil && config.Conf.Agent.Model == c.id {
		config.Conf.Agent = nil
		_ = config.SaveConfig()
		m.setStatus(true, "当前模型已删除，请重新选择模型")
	} else {
		m.setStatus(false, "已删除模型 %s", c.label())
	}
	v.reload()
	if len(v.choices) == 0 {
		m.setStatus(true, "还没有模型配置，按 a 添加")
	}
}

func (m *uiModel) applyModel(id, kind string) {
	cfg := config.AgentConfig{Model: id, ModelKind: kind}
	if config.Conf.Agent != nil {
		cfg.Name = config.Conf.Agent.Name
	}
	ag, err := agent.Reconfigure(&cfg)
	if err != nil {
		m.setStatus(true, "切换模型失败: %v", err)
		return
	}
	saved := config.AgentConfig{Name: ag.Name, Model: id, ModelKind: kind}
	if err := config.SetAgent(saved); err != nil {
		m.setStatus(true, "模型已生效，但保存配置失败: %v", err)
	} else {
		m.setStatus(false, "已切换模型 %s", agentLabel(id, kind))
	}
	m.modelName = agentLabel(id, kind)
	m.screen = screenChat
	m.input.Focus()
	m.refreshChat(false)
}

// startModelWizard walks the user through provider -> credentials -> model
// selection. Model names are always fetched from the provider.
func (m *uiModel) startModelWizard(id string) {
	m.status = ""
	m.editingModel = id
	m.draft = config.ModelConfig{}
	if id != "" {
		m.draft = config.Conf.Models[id]
	}
	m.openProviderPicker()
}

func (m *uiModel) openProviderPicker() {
	current := strings.ToLower(strings.TrimSpace(m.draft.Provider))
	if current == "ark" {
		current = "doubao"
	}
	picker := newPickerView("选择 Provider")
	cursor := 0
	for i, provider := range openai.Providers() {
		label := provider.Label
		if _, ok := config.FindModelByProvider(provider.ID); ok {
			label += "  (已配置)"
		}
		picker.items = append(picker.items, pickerItem{label: label, value: provider.ID})
		if provider.ID == current {
			cursor = i
		}
	}
	picker.cursor = cursor
	picker.width = m.width
	m.providerPick = picker
	m.screen = screenProvider
}

func (m *uiModel) updateProviderPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	picker := m.providerPick
	if picker == nil {
		m.openModels()
		return m, nil
	}
	if picker.handleKey(msg) {
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.providerPick = nil
		m.openModels()
	case "enter":
		item, ok := picker.current()
		if !ok {
			return m, nil
		}
		previous := strings.ToLower(strings.TrimSpace(m.draft.Provider))
		if previous == "ark" {
			previous = "doubao"
		}
		m.draft.Provider = item.value
		if previous != item.value || m.draft.URL == "" {
			m.draft.URL = openai.ProviderBaseURL(item.value)
		}
		if previous != item.value {
			m.draft.ModelNames = nil
		}
		if m.editingModel == "" && m.draft.Key == "" {
			if existing, ok := config.FindModelByProvider(item.value); ok {
				m.draft.Key = existing.Key
				if existing.URL != "" {
					m.draft.URL = existing.URL
				}
			}
		}
		m.openCredsForm()
	}
	return m, nil
}

func (m *uiModel) openCredsForm() {
	title := "配置 " + openai.ProviderLabel(m.draft.Provider)
	specs := []formFieldSpec{
		{Label: "API Key", Value: m.draft.Key, Placeholder: "sk-...", Secret: true},
		{Label: "Base URL", Value: m.draft.URL, Placeholder: "https://.../v1", Help: "留空使用 Provider 默认地址；自定义 Provider 必填"},
	}
	m.credsForm = newFormView(title, specs, func(values []string) (string, error) {
		key := strings.TrimSpace(values[0])
		baseURL := strings.TrimSpace(values[1])
		if key == "" {
			return "", errors.New("API Key 不能为空")
		}
		if openai.IsCustomProvider(m.draft.Provider) && baseURL == "" {
			return "", errors.New("自定义 Provider 需要填写 Base URL")
		}
		m.draft.Key = key
		m.draft.URL = baseURL
		return "", nil
	})
	m.credsForm.setWidth(m.width)
	m.screen = screenModelCreds
}

func (m *uiModel) updateModelCreds(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := m.credsForm
	if form == nil {
		m.openProviderPicker()
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.credsForm = nil
		m.openProviderPicker()
		return m, nil
	case "enter":
		if err := form.submit(); err != nil {
			return m, nil
		}
		if form.done {
			m.credsForm = nil
			return m, m.fetchModels()
		}
		return m, nil
	}
	return m, form.update(msg)
}

func (m *uiModel) fetchModels() tea.Cmd {
	m.fetchGen++
	gen := m.fetchGen
	picker := newPickerView("选择模型")
	picker.multi = true
	picker.loading = true
	picker.width = m.width
	m.modelPick = picker
	m.screen = screenModelPick

	draft := m.draft
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()
		names, err := openai.ListModels(ctx, openai.Config{APIKey: draft.Key, BaseURL: draft.URL})
		return modelsFetchedMsg{gen: gen, names: names, err: err}
	}
}

func (m *uiModel) updateModelPick(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	picker := m.modelPick
	if picker == nil {
		m.openModels()
		return m, nil
	}
	if picker.handleKey(msg) {
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.modelPick = nil
		m.openCredsForm()
	case "ctrl+r":
		if !picker.loading {
			return m, m.fetchModels()
		}
	case " ":
		if !picker.loading && picker.err == "" {
			picker.toggle()
		}
	case "enter":
		if picker.loading || picker.err != "" {
			return m, nil
		}
		item, ok := picker.current()
		if !ok {
			return m, nil
		}
		names := picker.selectedValues()
		if len(names) == 0 {
			names = []string{item.value}
		}
		m.saveModelSelection(picker, item.value, names)
	}
	return m, nil
}

func (m *uiModel) saveModelSelection(picker *pickerView, active string, names []string) {
	cfg := config.ModelConfig{
		Provider:   m.draft.Provider,
		URL:        m.draft.URL,
		Key:        m.draft.Key,
		ModelNames: names,
	}
	id := m.editingModel
	if id != "" {
		if err := config.UpdateModel(id, cfg); err != nil {
			picker.err = err.Error()
			return
		}
	} else {
		newID, err := config.AddModel(cfg)
		if err != nil {
			picker.err = err.Error()
			return
		}
		id = newID
	}

	m.modelPick = nil
	m.credsForm = nil
	m.providerPick = nil
	m.editingModel = ""
	m.draft = config.ModelConfig{}
	m.applyModel(id, active)
	m.setStatus(false, "已保存并使用 %s", agentLabel(id, active))
}

func (m *uiModel) viewModels() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render("模型"))
	b.WriteString("\n\n")

	if len(m.models.choices) == 0 {
		b.WriteString(styleMuted.Render("还没有模型配置，按 a 添加。"))
		b.WriteString("\n")
	}
	for i, c := range m.models.choices {
		line := "  " + c.label()
		if isCurrentModel(c) {
			line += styleOK.Render("  ● 使用中")
		}
		if i == m.models.cursor {
			b.WriteString(styleSelected.Render(line))
		} else {
			b.WriteString(styleBody.Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.models.confirmDelete {
		if c, ok := m.models.current(); ok {
			b.WriteString(styleError.Render(fmt.Sprintf("删除 %s？再次按 d 确认，Esc 取消。", c.label())))
		}
	} else {
		b.WriteString(styleMuted.Render("Enter 使用 · a 添加 · e 编辑 · d 删除 · Esc 返回"))
	}
	b.WriteString("\n\n")
	if line := m.statusLine(); line != "" {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString(styleFooter.Render("配置保存在 ~/.digtast/config.json"))
	return b.String()
}

func (m *uiModel) viewProviderPicker() string {
	var b strings.Builder
	b.WriteString(m.providerPick.view(m.pickerRows(), m.spinner.View()))
	b.WriteString("\n")
	b.WriteString(styleMuted.Render("Enter 选择 · 输入过滤 · Esc 返回"))
	b.WriteString("\n\n")
	if line := m.statusLine(); line != "" {
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func (m *uiModel) viewModelCreds() string {
	return m.viewForm(m.credsForm, "Tab 切换字段 · Enter 获取模型列表 · Esc 返回")
}

func (m *uiModel) viewModelPick() string {
	var b strings.Builder
	b.WriteString(m.modelPick.view(m.pickerRows(), m.spinner.View()))
	b.WriteString("\n")
	switch {
	case m.modelPick.loading:
		b.WriteString(styleMuted.Render("Esc 返回"))
	case m.modelPick.err != "":
		b.WriteString(styleMuted.Render("Ctrl+R 重试 · Esc 返回"))
	default:
		b.WriteString(styleMuted.Render("Space 多选 · Enter 确认 · 输入过滤 · Ctrl+R 重新获取 · Esc 返回"))
	}
	b.WriteString("\n\n")
	if line := m.statusLine(); line != "" {
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func (m *uiModel) pickerRows() int {
	rows := m.height - 12
	if rows < 5 {
		rows = 5
	}
	return rows
}

func isCurrentModel(c modelChoice) bool {
	cfg := config.Conf.Agent
	if cfg == nil {
		return false
	}
	if cfg.Model != c.id {
		return false
	}
	return cfg.ModelKind == "" || cfg.ModelKind == c.kind
}

func agentLabel(id, kind string) string {
	mc, ok := config.Conf.Models[id]
	if !ok {
		return kind
	}
	label := openai.ProviderLabel(mc.Provider)
	if kind != "" {
		label += " · " + kind
	}
	return label
}

func (m *uiModel) handleModelsFetched(msg modelsFetchedMsg) {
	picker := m.modelPick
	if picker == nil || msg.gen != m.fetchGen {
		return
	}
	picker.loading = false
	if msg.err != nil {
		picker.err = msg.err.Error()
		return
	}
	picker.err = ""
	picker.items = make([]pickerItem, 0, len(msg.names))
	firstSelected := -1
	for _, name := range msg.names {
		selected := false
		if m.editingModel != "" {
			for _, existing := range m.draft.ModelNames {
				if existing == name {
					selected = true
					break
				}
			}
		}
		if selected && firstSelected < 0 {
			firstSelected = len(picker.items)
		}
		picker.items = append(picker.items, pickerItem{label: name, value: name, selected: selected})
	}
	picker.filter = ""
	picker.cursor = 0
	if firstSelected >= 0 {
		picker.cursor = firstSelected
	}
}
