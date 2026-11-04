package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"digital-labor/config"
	"digital-labor/gateway"

	tea "github.com/charmbracelet/bubbletea"
)

type channelFieldSpec struct {
	Key    string
	Label  string
	Help   string
	Secret bool
	Number bool
}

// channelSchemas describes the parameters of the built-in channels. Unknown
// channels can still be configured through a free-form key=value field.
var channelSchemas = map[string][]channelFieldSpec{
	"qq": {
		{Key: "appid", Label: "AppID"},
		{Key: "appsecret", Label: "AppSecret", Secret: true},
		{Key: "port", Label: "端口", Number: true, Help: "例如 8080"},
	},
	"telegram": {
		{Key: "token", Label: "Bot Token", Secret: true},
	},
}

type channelView struct {
	kinds         []string
	cursor        int
	confirmDelete bool
}

func newChannelView() *channelView {
	v := &channelView{kinds: gateway.GetAllChannels()}
	sort.Strings(v.kinds)
	return v
}

func (v *channelView) current() (string, bool) {
	if len(v.kinds) == 0 {
		return "", false
	}
	return v.kinds[v.cursor], true
}

func (m *uiModel) openChannels() {
	m.channelForm = nil
	m.channels = newChannelView()
	m.screen = screenChannels
}

func (m *uiModel) updateChannels(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	v := m.channels
	if v == nil {
		m.openChannels()
		return m, nil
	}
	key := msg.String()
	if v.confirmDelete && key != "d" {
		v.confirmDelete = false
	}
	switch key {
	case "esc":
		m.screen = screenChat
		m.input.Focus()
	case "up", "k":
		if v.cursor > 0 {
			v.cursor--
		}
	case "down", "j":
		if v.cursor < len(v.kinds)-1 {
			v.cursor++
		}
	case "enter":
		kind, ok := v.current()
		if !ok {
			return m, nil
		}
		if _, configured := config.Conf.Channels[kind]; configured {
			m.toggleChannel(kind)
		} else {
			m.openChannelForm(kind)
		}
	case "a", "e":
		if kind, ok := v.current(); ok {
			m.openChannelForm(kind)
		}
	case "d":
		m.deleteChannel()
	}
	return m, nil
}

func (m *uiModel) toggleChannel(kind string) {
	params, ok := config.Conf.Channels[kind]
	if !ok {
		m.openChannelForm(kind)
		return
	}
	enable := !config.IsChannelEnabled(kind)
	params[config.ChannelEnabledKey] = enable
	if err := config.SetChannel(kind, params); err != nil {
		m.setStatus(true, "保存失败: %v", err)
		return
	}
	if enable {
		if err := gateway.EnableChannel(kind); err != nil {
			m.setStatus(true, "渠道已保存，但启动失败: %v", err)
			return
		}
		m.setStatus(false, "渠道 %s 已启用", kind)
		return
	}
	if err := gateway.DisableChannel(kind); err != nil {
		m.setStatus(true, "渠道已保存，但停止失败: %v", err)
		return
	}
	m.setStatus(false, "渠道 %s 已停用", kind)
}

func (m *uiModel) deleteChannel() {
	v := m.channels
	kind, ok := v.current()
	if !ok {
		return
	}
	if _, configured := config.Conf.Channels[kind]; !configured {
		m.setStatus(false, "渠道 %s 还没有配置", kind)
		return
	}
	if !v.confirmDelete {
		v.confirmDelete = true
		return
	}
	v.confirmDelete = false
	if err := gateway.RemoveChannel(kind); err != nil {
		m.setStatus(true, "删除失败: %v", err)
		return
	}
	m.setStatus(false, "已删除渠道 %s", kind)
}

func (m *uiModel) openChannelForm(kind string) {
	current := config.Conf.Channels[kind]
	schema, known := channelSchemas[kind]

	var specs []formFieldSpec
	if known {
		for _, field := range schema {
			value := ""
			if v, ok := current[field.Key]; ok {
				value = fmt.Sprint(v)
			}
			specs = append(specs, formFieldSpec{
				Label:       field.Label,
				Value:       value,
				Placeholder: field.Key,
				Secret:      field.Secret,
				Help:        field.Help,
			})
		}
	} else {
		specs = append(specs, formFieldSpec{
			Label:       "参数",
			Value:       formatParams(current),
			Placeholder: "key=value, key=value",
		})
	}

	title := "配置渠道 " + kind
	m.channelForm = newFormView(title, specs, func(values []string) (string, error) {
		params := make(map[string]any, len(current)+len(values))
		for key, value := range current {
			params[key] = value
		}
		if known {
			for i, field := range schema {
				value := strings.TrimSpace(values[i])
				if value == "" {
					return "", fmt.Errorf("%s 不能为空", field.Label)
				}
				if field.Number {
					number, err := strconv.Atoi(value)
					if err != nil {
						return "", fmt.Errorf("%s 必须是数字", field.Label)
					}
					params[field.Key] = number
				} else {
					params[field.Key] = value
				}
			}
		} else {
			parsed, err := parseParams(values[0])
			if err != nil {
				return "", err
			}
			for key, value := range parsed {
				params[key] = value
			}
		}
		if err := gateway.CreateChannel(kind, params); err != nil {
			return "", err
		}
		if config.IsChannelEnabled(kind) {
			_ = gateway.DisableChannel(kind)
			if err := gateway.EnableChannel(kind); err != nil {
				return "渠道已保存，但启动失败: " + err.Error(), nil
			}
		}
		return "渠道 " + kind + " 已保存", nil
	})
	m.channelForm.setWidth(m.width)
	m.screen = screenChannelForm
}

func (m *uiModel) updateChannelForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := m.channelForm
	if form == nil {
		m.openChannels()
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.channelForm = nil
		m.openChannels()
		return m, nil
	case "enter":
		if err := form.submit(); err != nil {
			return m, nil
		}
		if form.done {
			toast := form.toast
			m.channelForm = nil
			m.openChannels()
			m.setStatus(false, "%s", toast)
		}
		return m, nil
	}
	return m, form.update(msg)
}

func (m *uiModel) viewChannels() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render("消息渠道"))
	b.WriteString("\n\n")

	if len(m.channels.kinds) == 0 {
		b.WriteString(styleMuted.Render("没有可用的渠道实现。"))
		b.WriteString("\n")
	}
	for i, kind := range m.channels.kinds {
		line := "  " + channelStatus(kind) + " " + kind
		if i == m.channels.cursor {
			b.WriteString(styleSelected.Render(line))
		} else {
			b.WriteString(styleBody.Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.channels.confirmDelete {
		if kind, ok := m.channels.current(); ok {
			b.WriteString(styleError.Render(fmt.Sprintf("删除渠道 %s？再次按 d 确认，Esc 取消。", kind)))
		}
	} else {
		b.WriteString(styleMuted.Render("Enter 启用/停用 · a 配置 · d 删除 · Esc 返回"))
	}
	b.WriteString("\n\n")
	if line := m.statusLine(); line != "" {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString(styleFooter.Render("渠道配置保存在 ~/.digtast/config.json"))
	return b.String()
}

func channelStatus(kind string) string {
	if _, ok := config.Conf.Channels[kind]; !ok {
		return styleMuted.Render("· 未配置")
	}
	if !config.IsChannelEnabled(kind) {
		return styleMuted.Render("○ 已停用")
	}
	if gateway.IsChannelRunning(kind) {
		return styleOK.Render("● 运行中")
	}
	return styleError.Render("◐ 已启用 · 未运行")
}

func formatParams(params map[string]any) string {
	if len(params) == 0 {
		return ""
	}
	keys := make([]string, 0, len(params))
	for key := range params {
		if key == config.ChannelEnabledKey {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", key, params[key]))
	}
	return strings.Join(parts, ", ")
}

func parseParams(value string) (map[string]any, error) {
	params := map[string]any{}
	for _, pair := range strings.Split(value, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		key, val, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("参数 %q 需要写成 key=value", pair)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, errors.New("参数名不能为空")
		}
		params[key] = strings.TrimSpace(val)
	}
	return params, nil
}
