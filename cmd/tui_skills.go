package main

import (
	"strings"

	"digital-labor/config"
	"digital-labor/internal/agent/skill"

	tea "github.com/charmbracelet/bubbletea"
)

// /skills 命令：列出 workspace/skills 下扫描到的全部 skill，切换启用状态。
// 状态保存在 config.Conf.Skills；停用的 skill 不会注入提示词，也无法通过
// load_skill 加载。提示词在每次回复前重新组装，切换立即生效。

type skillView struct {
	skills []skill.SkillInfo
	cursor int
}

func (v *skillView) current() (skill.SkillInfo, bool) {
	if len(v.skills) == 0 || v.cursor >= len(v.skills) {
		return skill.SkillInfo{}, false
	}
	return v.skills[v.cursor], true
}

func (m *uiModel) openSkills() {
	manager := skill.NewManager()
	skills, err := manager.GetAllSkills()
	if err != nil {
		m.setStatus(true, "读取 skills 失败: %v", err)
		return
	}
	m.status = ""
	m.skillsView = &skillView{skills: skills}
	m.screen = screenSkills
}

func (m *uiModel) updateSkills(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	v := m.skillsView
	if v == nil {
		m.openSkills()
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.skillsView = nil
		m.screen = screenChat
		m.input.Focus()
	case "up", "k":
		if v.cursor > 0 {
			v.cursor--
		}
	case "down", "j":
		if v.cursor < len(v.skills)-1 {
			v.cursor++
		}
	case "r":
		// 重新扫描 skills 目录。
		m.openSkills()
	case " ", "enter":
		m.toggleSkill()
	}
	return m, nil
}

// toggleSkill 切换当前 skill 的启用状态并落盘。
func (m *uiModel) toggleSkill() {
	v := m.skillsView
	item, ok := v.current()
	if !ok {
		return
	}
	enable := !item.Enabled
	if err := config.SetSkillEnabled(item.Name, enable); err != nil {
		m.setStatus(true, "保存失败: %v", err)
		return
	}
	v.skills[v.cursor].Enabled = enable
	if enable {
		m.setStatus(false, "已启用 %s，下次回复生效", item.Name)
	} else {
		m.setStatus(false, "已停用 %s，下次回复生效", item.Name)
	}
}

func (m *uiModel) viewSkills() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render("Skills"))
	b.WriteString("\n\n")

	v := m.skillsView
	if len(v.skills) == 0 {
		b.WriteString(styleMuted.Render("没有扫描到 skill。"))
		b.WriteString("\n")
	}
	for i, item := range v.skills {
		marker := styleMuted.Render("○ 停用")
		if item.Enabled {
			marker = styleOK.Render("● 启用")
		}
		line := "  " + marker + "  " + item.Name
		if item.Description != "" {
			line += styleMuted.Render("  " + item.Description)
		}
		if i == v.cursor {
			b.WriteString(styleSelected.Render(line))
		} else {
			b.WriteString(styleBody.Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styleMuted.Render("↑/↓ 选择 · Space/Enter 启用停用 · r 重新扫描 · Esc 返回"))
	b.WriteString("\n\n")
	if line := m.statusLine(); line != "" {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString(styleFooter.Render("Skills 目录：~/.digtast/skills/<name>/SKILL.md"))
	return b.String()
}
