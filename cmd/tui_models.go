package main

import (
	"errors"
	"strings"

	"digital-labor/component/chatmodel"
	"digital-labor/config"
	"digital-labor/internal/agent"

	tea "github.com/charmbracelet/bubbletea"
)

// 模型页配置的是厂商（chatmodel 注册表中的 Provider）：先填 API Key，再决定
// 是否替换默认 Base URL，然后从 /models 拉取模型列表选择启用哪些；日常切换
// 只在厂商的启用模型中挑选。凭据与启用列表按 Provider 保存在 config 中。

type providerView struct {
	providers []chatmodel.Provider
	cursor    int
}

func newProviderView() *providerView {
	return &providerView{providers: chatmodel.Providers()}
}

func (v *providerView) current() (chatmodel.Provider, bool) {
	if len(v.providers) == 0 || v.cursor >= len(v.providers) {
		return chatmodel.Provider{}, false
	}
	return v.providers[v.cursor], true
}

// modelListView 展示某个厂商已启用的模型，供日常切换。
type modelListView struct {
	provider chatmodel.Provider
	models   []string
	cursor   int
}

func (v *modelListView) current() (string, bool) {
	if len(v.models) == 0 || v.cursor >= len(v.models) {
		return "", false
	}
	return v.models[v.cursor], true
}

func (m *uiModel) openProviders() {
	if _, err := agent.Instance(); err != nil {
		m.setStatus(false, "选择一个厂商开始使用")
	}
	m.providers = newProviderView()
	m.screen = screenProviders
}

func (m *uiModel) updateProviders(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	v := m.providers
	if v == nil {
		m.openProviders()
		return m, nil
	}
	switch msg.String() {
	case "esc":
		if _, err := agent.Instance(); err != nil {
			m.setStatus(true, "请先配置厂商并启用模型")
			return m, nil
		}
		m.screen = screenChat
		m.input.Focus()
	case "up", "k":
		if v.cursor > 0 {
			v.cursor--
		}
	case "down", "j":
		if v.cursor < len(v.providers)-1 {
			v.cursor++
		}
	case "enter":
		if provider, ok := v.current(); ok {
			if config.ModelCredentials(provider.ID).Key == "" {
				m.startProviderSetup(provider)
			} else {
				m.openProviderModels(provider)
			}
		}
	}
	return m, nil
}

// startProviderSetup 启动厂商配置向导：API Key → 是否替换 Base URL → 拉取
// 模型列表 → 选择启用的模型。
func (m *uiModel) startProviderSetup(provider chatmodel.Provider) {
	m.status = ""
	m.setupProv = provider
	m.setupKey = ""
	m.setupURL = ""
	m.setupForm = newFormView("配置 "+provider.Label, []formFieldSpec{
		{Label: "API Key", Placeholder: "sk-...", Secret: true},
	}, func(values []string) (string, error) {
		key := strings.TrimSpace(values[0])
		if key == "" {
			return "", errors.New("API Key 不能为空")
		}
		m.setupKey = key
		return "", nil
	})
	m.setupForm.setWidth(m.width)
	m.screen = screenProviderKey
}

func (m *uiModel) updateProviderKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := m.setupForm
	if form == nil {
		m.openProviders()
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.setupForm = nil
		m.openProviders()
		return m, nil
	case "enter":
		if err := form.submit(); err != nil {
			return m, nil
		}
		m.setupForm = nil
		m.openBaseURLPick()
		return m, nil
	}
	return m, form.update(msg)
}

// openBaseURLPick 询问是否需要修改默认 Base URL。
func (m *uiModel) openBaseURLPick() {
	picker := newPickerView("是否修改 Base URL？")
	picker.items = []pickerItem{
		{label: "使用默认地址 " + m.setupProv.BaseURL, value: ""},
		{label: "自定义地址…", value: "custom"},
	}
	picker.width = m.width
	m.urlPick = picker
	m.screen = screenProviderURL
}

func (m *uiModel) updateBaseURLPick(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	picker := m.urlPick
	if picker == nil {
		m.openProviders()
		return m, nil
	}
	if picker.handleKey(msg) {
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.urlPick = nil
		m.openProviders()
	case "enter":
		item, ok := picker.current()
		if !ok {
			return m, nil
		}
		m.urlPick = nil
		if item.value == "custom" {
			m.openURLForm()
			return m, nil
		}
		m.setupURL = ""
		return m, m.fetchProviderModels()
	}
	return m, nil
}

func (m *uiModel) openURLForm() {
	m.setupForm = newFormView("自定义 Base URL", []formFieldSpec{
		{Label: "Base URL", Placeholder: "https://...", Help: "替换 " + m.setupProv.Label + " 的默认地址"},
	}, func(values []string) (string, error) {
		url := strings.TrimSpace(values[0])
		if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
			return "", errors.New("Base URL 必须以 http:// 或 https:// 开头")
		}
		m.setupURL = url
		return "", nil
	})
	m.setupForm.setWidth(m.width)
	m.screen = screenProviderURLForm
}

func (m *uiModel) updateURLForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := m.setupForm
	if form == nil {
		m.openBaseURLPick()
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.setupForm = nil
		m.openBaseURLPick()
		return m, nil
	case "enter":
		if err := form.submit(); err != nil {
			return m, nil
		}
		m.setupForm = nil
		return m, m.fetchProviderModels()
	}
	return m, form.update(msg)
}

// fetchProviderModels 用配置中的凭据拉取 /models 列表。
func (m *uiModel) fetchProviderModels() tea.Cmd {
	m.fetchGen++
	gen := m.fetchGen
	provider := m.setupProv
	if provider.ID == "" {
		return nil
	}
	key, url := m.setupKey, m.setupURL
	if key == "" {
		creds := config.ModelCredentials(provider.ID)
		key, url = creds.Key, creds.URL
	}
	picker := newPickerView("选择启用的模型 · " + provider.Label)
	picker.multi = true
	picker.loading = true
	picker.width = m.width
	m.modelPick = picker
	m.screen = screenProviderModels

	return func() tea.Msg {
		names, err := provider.ListModels(m.ctx, chatmodel.ClientOptions{APIKey: key, BaseURL: url})
		return modelsFetchedMsg{gen: gen, names: names, err: err}
	}
}

func (m *uiModel) updateProviderModels(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	picker := m.modelPick
	if picker == nil {
		m.openProviders()
		return m, nil
	}
	if picker.handleKey(msg) {
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.modelPick = nil
		m.openProviders()
	case "ctrl+r":
		if !picker.loading {
			return m, m.fetchProviderModels()
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
		m.confirmModelSelection(item.value)
	}
	return m, nil
}

// confirmModelSelection 保存凭据与启用列表，并激活一个启用的模型：优先保留
// 当前正在使用的模型，否则取列表第一个。
func (m *uiModel) confirmModelSelection(active string) {
	provider := m.setupProv
	enabled := m.modelPick.selectedValues()
	if len(enabled) == 0 {
		enabled = []string{active}
	}
	creds := config.ModelCredentials(provider.ID)
	if m.setupKey != "" {
		creds.Key = m.setupKey
	}
	creds.URL = m.setupURL
	creds.Models = enabled
	if err := config.SaveModelCredentials(provider.ID, creds); err != nil {
		m.setStatus(true, "保存配置失败: %v", err)
		return
	}
	target := enabled[0]
	if current := config.Conf.Agent; current != nil && current.Provider == provider.ID {
		for _, model := range enabled {
			if model == current.Model {
				target = model
				break
			}
		}
	}
	m.modelPick = nil
	m.setupProv = chatmodel.Provider{}
	m.setupKey = ""
	m.setupURL = ""
	m.activateModel(provider, target)
}

// openProviderModels 打开已配置厂商的启用模型列表。
func (m *uiModel) openProviderModels(provider chatmodel.Provider) {
	m.status = ""
	models := config.ModelCredentials(provider.ID).Models
	view := &modelListView{provider: provider, models: models}
	// 当前使用的模型排在光标处。
	if active := config.Conf.Agent; active != nil && active.Provider == provider.ID {
		for index, model := range models {
			if model == active.Model {
				view.cursor = index
			}
		}
	}
	m.provModels = view
	m.screen = screenProviderEnabled
}

func (m *uiModel) updateProviderEnabled(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	v := m.provModels
	if v == nil {
		m.openProviders()
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.provModels = nil
		m.openProviders()
	case "up", "k":
		if v.cursor > 0 {
			v.cursor--
		}
	case "down", "j":
		if v.cursor < len(v.models)-1 {
			v.cursor++
		}
	case "a":
		// 重新拉取 /models 更新启用列表（沿用已保存的凭据）。
		m.setupProv = v.provider
		m.setupKey = ""
		m.setupURL = config.ModelCredentials(v.provider.ID).URL
		return m, m.fetchProviderModels()
	case "enter":
		if model, ok := v.current(); ok {
			m.activateModel(v.provider, model)
		}
	}
	return m, nil
}

// activateModel 切换当前使用的模型。
func (m *uiModel) activateModel(provider chatmodel.Provider, model string) {
	cfg := config.AgentConfig{Provider: provider.ID, Model: model}
	if current := config.Conf.Agent; current != nil {
		cfg.Name = current.Name
		cfg.Description = current.Description
	}
	ag, err := agent.Reconfigure(&cfg)
	if err != nil {
		m.setStatus(true, "切换模型失败: %v", err)
		return
	}
	saved := config.AgentConfig{Name: ag.Name, Provider: provider.ID, Model: model, Description: cfg.Description}
	if err := config.SetAgent(saved); err != nil {
		m.setStatus(true, "模型已生效，但保存配置失败: %v", err)
	} else {
		m.setStatus(false, "已切换模型 %s", modelLabel(provider.ID, model))
	}
	m.modelName = modelLabel(provider.ID, model)
	m.provModels = nil
	m.modelPick = nil
	m.screen = screenChat
	m.input.Focus()
	m.refreshChat(false)
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
	enabled := config.ModelCredentials(m.setupProv.ID).Models
	picker.items = make([]pickerItem, 0, len(msg.names))
	firstSelected := -1
	for _, name := range msg.names {
		selected := false
		for _, existing := range enabled {
			if existing == name {
				selected = true
				break
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

func (m *uiModel) viewProviders() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render("模型厂商"))
	b.WriteString("\n\n")

	if len(m.providers.providers) == 0 {
		b.WriteString(styleMuted.Render("没有注册的厂商。"))
		b.WriteString("\n")
	}
	for i, provider := range m.providers.providers {
		line := "  " + provider.Label
		creds := config.ModelCredentials(provider.ID)
		switch {
		case isCurrentProvider(provider):
			line += styleOK.Render("  ● 使用中")
		case creds.Key != "":
			line += styleMuted.Render("  已配置")
		default:
			line += styleMuted.Render("  未配置")
		}
		if i == m.providers.cursor {
			b.WriteString(styleSelected.Render(line))
		} else {
			b.WriteString(styleBody.Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styleMuted.Render("↑/↓ 选择 · Enter 配置/切换 · Esc 返回"))
	b.WriteString("\n\n")
	if line := m.statusLine(); line != "" {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString(styleFooter.Render("厂商在代码中注册；凭据与启用模型保存在 ~/.digtast/config.json"))
	return b.String()
}

func (m *uiModel) viewProviderEnabled() string {
	var b strings.Builder
	v := m.provModels
	b.WriteString(styleTitle.Render(v.provider.Label + " · 启用的模型"))
	b.WriteString("\n\n")

	if len(v.models) == 0 {
		b.WriteString(styleMuted.Render("还没有启用的模型，按 a 从模型列表添加。"))
		b.WriteString("\n")
	}
	for i, model := range v.models {
		line := "  " + model
		if isCurrentModel(v.provider.ID, model) {
			line += styleOK.Render("  ● 使用中")
		}
		if i == v.cursor {
			b.WriteString(styleSelected.Render(line))
		} else {
			b.WriteString(styleBody.Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styleMuted.Render("Enter 使用 · a 重新获取模型列表 · Esc 返回"))
	b.WriteString("\n\n")
	if line := m.statusLine(); line != "" {
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func (m *uiModel) viewProviderKey() string {
	return m.viewForm(m.setupForm, "Tab 切换字段 · Enter 下一步 · Esc 取消")
}

func (m *uiModel) viewBaseURLPick() string {
	var b strings.Builder
	b.WriteString(m.urlPick.view(m.pickerRows(), m.spinner.View()))
	b.WriteString("\n")
	b.WriteString(styleMuted.Render("Enter 选择 · Esc 取消"))
	b.WriteString("\n\n")
	if line := m.statusLine(); line != "" {
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func (m *uiModel) viewURLForm() string {
	return m.viewForm(m.setupForm, "Enter 获取模型列表 · Esc 返回")
}

func (m *uiModel) viewProviderModels() string {
	var b strings.Builder
	b.WriteString(m.modelPick.view(m.pickerRows(), m.spinner.View()))
	b.WriteString("\n")
	switch {
	case m.modelPick.loading:
		b.WriteString(styleMuted.Render("Esc 取消"))
	case m.modelPick.err != "":
		b.WriteString(styleMuted.Render("Ctrl+R 重试 · Esc 取消"))
	default:
		b.WriteString(styleMuted.Render("Space 启用/取消 · Enter 确认 · 输入过滤 · Ctrl+R 重新获取 · Esc 取消"))
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

func isCurrentProvider(provider chatmodel.Provider) bool {
	cfg := config.Conf.Agent
	return cfg != nil && cfg.Provider == provider.ID
}

func isCurrentModel(provider, model string) bool {
	cfg := config.Conf.Agent
	return cfg != nil && cfg.Provider == provider && cfg.Model == model
}

// modelLabel 返回「厂商 · 模型」展示名。
func modelLabel(provider, model string) string {
	if info, ok := chatmodel.FindProvider(provider); ok {
		return info.Label + " · " + model
	}
	return model
}
