package main

import (
	"context"
	"fmt"
	"strings"

	"digital-labor/config"
	"digital-labor/gateway"
	"digital-labor/internal/agent"
	"digital-labor/schema"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type screenID int

const (
	screenChat screenID = iota
	screenModels
	screenProvider
	screenModelCreds
	screenModelPick
	screenChannels
	screenChannelForm
	screenHelp
)

type chatEntry struct {
	role string
	text string
}

type command struct {
	name string
	desc string
}

var chatCommands = []command{
	{name: "model", desc: "选择、添加、编辑、删除模型（自动获取模型列表）"},
	{name: "channel", desc: "配置消息渠道，启用或停用"},
	{name: "new", desc: "清空历史，开始新会话"},
	{name: "help", desc: "查看快捷键和命令"},
	{name: "exit", desc: "退出 digtast"},
}

type deltaMsg struct {
	gen  int
	text string
}

type runStartedMsg struct {
	gen int
	ch  <-chan string
}

type runDoneMsg struct {
	gen int
}

type runErrMsg struct {
	gen int
	err error
}

type modelsFetchedMsg struct {
	gen   int
	names []string
	err   error
}

type uiModel struct {
	ctx    context.Context
	width  int
	height int
	screen screenID

	entries  []chatEntry
	viewport viewport.Model
	input    textinput.Model
	spinner  spinner.Model

	paletteCursor   int
	lastPaletteText string

	streaming bool
	cancel    context.CancelFunc
	streamCh  <-chan string
	streamGen int

	status    string
	statusErr bool
	modelName string

	models       *modelView
	providerPick *pickerView
	modelPick    *pickerView
	credsForm    *formView
	channels     *channelView
	channelForm  *formView

	draft        config.ModelConfig
	editingModel string
	fetchGen     int
}

// RunTUI starts the full-screen terminal UI.
func RunTUI(ctx context.Context) error {
	model := newUIModel(ctx)
	gateway.LoadChannels()
	gateway.Start()

	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithContext(ctx))
	_, err := program.Run()
	if ag, ierr := agent.Instance(); ierr == nil {
		_ = ag.Close()
	}
	return err
}

func newUIModel(ctx context.Context) *uiModel {
	ti := textinput.New()
	ti.Prompt = "› "
	ti.Placeholder = "输入消息，输入 / 查看命令"
	ti.PromptStyle = styleCursor
	ti.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot

	m := &uiModel{
		ctx:      ctx,
		viewport: viewport.New(0, 0),
		input:    ti,
		spinner:  sp,
	}
	m.bootstrap()
	return m
}

func (m *uiModel) bootstrap() {
	if cfg := config.Conf.Agent; cfg != nil {
		if _, ok := config.FindModelConfig(cfg.Model); !ok {
			config.Conf.Agent = nil
			_ = config.SaveConfig()
			m.setStatus(true, "已保存的模型配置不存在，请重新选择模型")
		} else if ag, err := agent.Initialize(cfg); err != nil {
			config.Conf.Agent = nil
			_ = config.SaveConfig()
			m.setStatus(true, "初始化模型失败: %v", err)
		} else {
			if err := config.SetAgent(*cfg); err != nil {
				m.setStatus(true, "保存配置失败: %v", err)
			}
			m.modelName = agentLabel(cfg.Model, cfg.ModelKind)
			m.loadHistory(ag)
			m.screen = screenChat
			return
		}
	}
	m.openModels()
	if m.models != nil && len(m.models.choices) == 0 {
		m.startModelWizard("")
	}
}

func (m *uiModel) loadHistory(ag *agent.Agent) {
	sess, err := ag.GetSession()
	if err != nil {
		return
	}
	for _, msg := range sess.GetMessages() {
		if msg == nil || strings.TrimSpace(msg.Content) == "" {
			continue
		}
		switch msg.Role {
		case schema.RoleUser:
			m.entries = append(m.entries, chatEntry{role: "user", text: msg.Content})
		case schema.RoleAssistant:
			m.entries = append(m.entries, chatEntry{role: "assistant", text: msg.Content})
		}
	}
}

func (m *uiModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m *uiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.credsForm != nil {
			m.credsForm.setWidth(m.width)
		}
		if m.channelForm != nil {
			m.channelForm.setWidth(m.width)
		}
		m.layoutChat()
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			if m.streaming {
				m.cancelRun()
				m.setStatus(false, "已取消当前回复")
				return m, nil
			}
			return m, tea.Quit
		}
		switch m.screen {
		case screenChat:
			return m.updateChat(msg)
		case screenModels:
			return m.updateModels(msg)
		case screenProvider:
			return m.updateProviderPicker(msg)
		case screenModelCreds:
			return m.updateModelCreds(msg)
		case screenModelPick:
			return m.updateModelPick(msg)
		case screenChannels:
			return m.updateChannels(msg)
		case screenChannelForm:
			return m.updateChannelForm(msg)
		case screenHelp:
			if key := msg.String(); key == "esc" || key == "enter" || key == "q" {
				m.screen = screenChat
			}
			return m, nil
		}

	case deltaMsg:
		if msg.gen != m.streamGen {
			return m, nil
		}
		m.appendDelta(msg.text)
		return m, waitDelta(m.streamCh, msg.gen)

	case runStartedMsg:
		if msg.gen != m.streamGen {
			return m, nil
		}
		m.streamCh = msg.ch
		return m, waitDelta(msg.ch, msg.gen)

	case runDoneMsg:
		if msg.gen != m.streamGen {
			return m, nil
		}
		m.finishRun()
		return m, nil

	case runErrMsg:
		if msg.gen != m.streamGen {
			return m, nil
		}
		m.failRun(msg.err)
		return m, nil

	case modelsFetchedMsg:
		m.handleModelsFetched(msg)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	switch m.screen {
	case screenChat:
		m.input, cmd = m.input.Update(msg)
	case screenModelCreds:
		if m.credsForm != nil {
			cmd = m.credsForm.update(msg)
		}
	case screenChannelForm:
		if m.channelForm != nil {
			cmd = m.channelForm.update(msg)
		}
	}
	return m, cmd
}

func (m *uiModel) updateChat(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if m.paletteVisible() {
		candidates := m.visibleCommands()
		switch key {
		case "up":
			if m.paletteCursor > 0 {
				m.paletteCursor--
			}
			return m, nil
		case "down":
			if m.paletteCursor < len(candidates)-1 {
				m.paletteCursor++
			}
			return m, nil
		case "tab":
			if len(candidates) > 0 {
				m.input.SetValue("/" + candidates[m.paletteCursor].name)
				m.input.CursorEnd()
				m.layoutChat()
			}
			return m, nil
		case "esc":
			m.input.SetValue("")
			m.layoutChat()
			return m, nil
		case "enter":
			if len(candidates) > 0 {
				m.input.SetValue("")
				m.layoutChat()
				return m, m.runCommand(candidates[m.paletteCursor].name)
			}
		}
	}

	switch key {
	case "enter":
		text := strings.TrimSpace(m.input.Value())
		if text == "" {
			return m, nil
		}
		m.input.SetValue("")
		m.layoutChat()
		if strings.HasPrefix(text, "/") {
			return m, m.runCommand(strings.TrimPrefix(text, "/"))
		}
		return m, m.send(text)
	case "esc":
		if m.streaming {
			m.cancelRun()
			m.setStatus(false, "已取消当前回复")
		} else {
			m.input.SetValue("")
			m.layoutChat()
		}
		return m, nil
	case "up", "down", "pgup", "pgdown", "home", "end":
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.layoutChat()
	return m, cmd
}

func (m *uiModel) runCommand(name string) tea.Cmd {
	name = strings.TrimPrefix(strings.TrimSpace(name), "/")
	if m.streaming && name != "help" && name != "exit" && name != "quit" {
		m.setStatus(true, "正在回复中，按 Esc 可以取消")
		return nil
	}
	switch name {
	case "model", "models":
		m.openModels()
	case "channel", "channels":
		m.openChannels()
	case "new", "reset":
		m.newSession()
	case "help":
		m.screen = screenHelp
	case "exit", "quit":
		return tea.Quit
	case "":
		return nil
	default:
		m.setStatus(true, "未知命令 /%s，输入 / 查看全部命令", name)
	}
	return nil
}

func (m *uiModel) newSession() {
	ag, err := agent.Instance()
	if err != nil {
		m.setStatus(true, "还没有可用模型，请先选择模型")
		m.openModels()
		return
	}
	if err := ag.ClearHistory(); err != nil {
		m.setStatus(true, "清空历史失败: %v", err)
		return
	}
	m.entries = nil
	m.refreshChat(true)
	m.setStatus(false, "已开始新会话")
}

func (m *uiModel) send(text string) tea.Cmd {
	if m.streaming {
		m.setStatus(true, "正在回复中，按 Esc 可以取消")
		return nil
	}
	ag, err := agent.Instance()
	if err != nil {
		m.setStatus(true, "还没有可用模型，请先选择模型")
		m.openModels()
		return nil
	}

	m.status = ""
	m.entries = append(m.entries, chatEntry{role: "user", text: text}, chatEntry{role: "assistant"})
	m.refreshChat(true)

	ctx, cancel := context.WithCancel(m.ctx)
	m.cancel = cancel
	m.streaming = true
	m.streamGen++
	gen := m.streamGen

	return func() tea.Msg {
		ch, err := ag.Run(ctx, text, nil, true)
		if err != nil {
			return runErrMsg{gen: gen, err: err}
		}
		return runStartedMsg{gen: gen, ch: ch}
	}
}

func waitDelta(ch <-chan string, gen int) tea.Cmd {
	return func() tea.Msg {
		text, ok := <-ch
		if !ok {
			return runDoneMsg{gen: gen}
		}
		return deltaMsg{gen: gen, text: text}
	}
}

func (m *uiModel) appendDelta(text string) {
	if text == "" {
		return
	}
	if len(m.entries) > 0 && m.entries[len(m.entries)-1].role == "assistant" {
		m.entries[len(m.entries)-1].text += text
	} else {
		m.entries = append(m.entries, chatEntry{role: "assistant", text: text})
	}
	m.refreshChat(true)
}

func (m *uiModel) finishRun() {
	m.streaming = false
	m.cancel = nil
	m.streamCh = nil
	if len(m.entries) > 0 {
		last := m.entries[len(m.entries)-1]
		if last.role == "assistant" && strings.TrimSpace(last.text) == "" {
			m.entries = m.entries[:len(m.entries)-1]
		}
	}
	m.refreshChat(true)
}

func (m *uiModel) failRun(err error) {
	m.streaming = false
	m.cancel = nil
	m.streamCh = nil
	if len(m.entries) > 0 {
		last := m.entries[len(m.entries)-1]
		if last.role == "assistant" && strings.TrimSpace(last.text) == "" {
			m.entries = m.entries[:len(m.entries)-1]
		}
	}
	m.entries = append(m.entries, chatEntry{role: "error", text: err.Error()})
	m.refreshChat(true)
}

func (m *uiModel) cancelRun() {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.streaming = false
	m.streamCh = nil
	m.streamGen++
}

func (m *uiModel) setStatus(isErr bool, format string, args ...any) {
	m.status = fmt.Sprintf(format, args...)
	m.statusErr = isErr
}

func (m *uiModel) statusLine() string {
	if m.status == "" {
		return ""
	}
	if m.statusErr {
		return styleError.Render("✗ " + m.status)
	}
	return styleOK.Render("✓ " + m.status)
}

func (m *uiModel) layoutChat() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	paletteLines := 0
	if m.paletteVisible() {
		paletteLines = len(m.visibleCommands())
	}
	height := m.height - 6 - paletteLines
	if height < 3 {
		height = 3
	}
	m.viewport.Width = m.width
	m.viewport.Height = height
	m.viewport.SetContent(m.renderChat())
	inputWidth := m.width - 8
	if inputWidth < 20 {
		inputWidth = 20
	}
	m.input.Width = inputWidth

	if m.input.Value() != m.lastPaletteText {
		m.lastPaletteText = m.input.Value()
		m.paletteCursor = 0
	}
	if candidates := m.visibleCommands(); len(candidates) > 0 && m.paletteCursor >= len(candidates) {
		m.paletteCursor = len(candidates) - 1
	}
}

func (m *uiModel) refreshChat(gotoBottom bool) {
	m.viewport.SetContent(m.renderChat())
	if gotoBottom {
		m.viewport.GotoBottom()
	}
}

func (m *uiModel) renderChat() string {
	if len(m.entries) == 0 {
		return styleMuted.Render("直接输入消息开始对话，输入 / 查看命令。")
	}
	width := m.viewport.Width - 2
	if width < 20 {
		width = 20
	}
	var b strings.Builder
	for i, entry := range m.entries {
		switch entry.role {
		case "user":
			b.WriteString(styleUser.Render("你"))
		case "assistant":
			b.WriteString(styleAssistant.Render("digtast"))
		case "error":
			b.WriteString(styleError.Render("错误"))
		default:
			b.WriteString(styleSystem.Render("系统"))
		}
		b.WriteString("\n")
		text := entry.text
		if entry.role == "assistant" && m.streaming && i == len(m.entries)-1 {
			text += styleCursor.Render("▌")
		}
		if text != "" {
			b.WriteString(styleBody.Width(width).Render(text))
		}
		b.WriteString("\n\n")
	}
	return b.String()
}

func (m *uiModel) paletteVisible() bool {
	value := m.input.Value()
	if !strings.HasPrefix(value, "/") {
		return false
	}
	return !strings.Contains(strings.TrimPrefix(value, "/"), " ")
}

func (m *uiModel) visibleCommands() []command {
	if !m.paletteVisible() {
		return nil
	}
	prefix := strings.TrimPrefix(m.input.Value(), "/")
	matches := make([]command, 0, len(chatCommands))
	for _, cmd := range chatCommands {
		if strings.HasPrefix(cmd.name, prefix) {
			matches = append(matches, cmd)
		}
	}
	return matches
}

func (m *uiModel) viewPalette() string {
	candidates := m.visibleCommands()
	if len(candidates) == 0 {
		return ""
	}
	var b strings.Builder
	for i, cmd := range candidates {
		line := fmt.Sprintf("/%-9s %s", cmd.name, cmd.desc)
		if i == m.paletteCursor {
			b.WriteString(styleSelected.Render(line))
		} else {
			b.WriteString(styleMuted.Render(line))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m *uiModel) View() string {
	switch m.screen {
	case screenModels:
		if m.models == nil {
			m.models = newModelView()
		}
		return m.viewModels()
	case screenProvider:
		if m.providerPick == nil {
			m.screen = screenChat
			return m.viewChat()
		}
		return m.viewProviderPicker()
	case screenModelCreds:
		if m.credsForm == nil {
			m.screen = screenChat
			return m.viewChat()
		}
		return m.viewModelCreds()
	case screenModelPick:
		if m.modelPick == nil {
			m.screen = screenChat
			return m.viewChat()
		}
		return m.viewModelPick()
	case screenChannels:
		if m.channels == nil {
			m.channels = newChannelView()
		}
		return m.viewChannels()
	case screenChannelForm:
		if m.channelForm == nil {
			m.screen = screenChat
			return m.viewChat()
		}
		return m.viewForm(m.channelForm, "Tab 切换字段 · Enter 保存 · Esc 取消")
	case screenHelp:
		return m.viewHelp()
	default:
		return m.viewChat()
	}
}

func (m *uiModel) viewChat() string {
	header := "digtast"
	if m.modelName != "" {
		header += "  ·  " + m.modelName
	}
	var b strings.Builder
	b.WriteString(styleHeader.Width(m.width).Render(header))
	b.WriteString("\n")
	b.WriteString(m.viewport.View())
	b.WriteString("\n")
	b.WriteString(m.viewPalette())
	b.WriteString(m.chatStatus())
	b.WriteString("\n")
	b.WriteString(m.inputBox())
	b.WriteString("\n")
	b.WriteString(styleFooter.Width(m.width).Render("/ 命令 · Enter 发送 · Esc 取消/清空 · Ctrl+C 退出 · PgUp/PgDn 滚动"))
	return b.String()
}

func (m *uiModel) inputBox() string {
	width := m.width - 2
	if width < 24 {
		width = 24
	}
	return styleBorder.Width(width).Render(m.input.View())
}

func (m *uiModel) chatStatus() string {
	if m.streaming {
		return m.spinner.View() + " " + styleMuted.Render("正在生成…（Esc 取消）")
	}
	if line := m.statusLine(); line != "" {
		return line
	}
	return " "
}

func (m *uiModel) viewForm(form *formView, footer string) string {
	var b strings.Builder
	b.WriteString(form.view())
	if m.status != "" {
		b.WriteString(m.statusLine())
		b.WriteString("\n\n")
	}
	b.WriteString(styleFooter.Render(footer))
	return b.String()
}

func (m *uiModel) viewHelp() string {
	lines := []string{
		styleTitle.Render("帮助"),
		"",
		styleFocus.Render("快捷键"),
		"  Enter      发送消息；以 / 开头则执行命令",
		"  Tab        补全命令；在表单中切换字段",
		"  ↑ / ↓      选择命令或列表项；在对话中滚动",
		"  PgUp/PgDn  滚动对话",
		"  Esc        取消回复 / 返回上一级",
		"  Ctrl+C     取消回复；空闲时退出",
		"",
		styleFocus.Render("命令"),
		"  /model     选择、添加、编辑、删除模型（自动获取模型列表）",
		"  /channel   配置消息渠道，启用或停用",
		"  /new       清空历史，开始新会话",
		"  /help      查看本帮助",
		"  /exit      退出",
		"",
		styleMuted.Render("配置保存在 ~/.digtast/config.json"),
		"",
	}
	return strings.Join(lines, "\n")
}
