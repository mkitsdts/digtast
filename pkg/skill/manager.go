package skill

import (
	"digital-labor/pkg/workspace"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SkillInfo struct {
	Name        string
	Description string
	Enabled     bool
}

type Manager struct {
	agentID string
	baseDir string
}

func NewManager() *Manager {
	return NewManagerForAgent(workspace.DefaultAgentID())
}

func NewManagerForAgent(agentID string) *Manager {
	if agentID == "" {
		agentID = workspace.DefaultAgentID()
	}
	return &Manager{
		agentID: agentID,
		baseDir: workspace.AgentSkillsDir(agentID),
	}
}

func (m *Manager) GetAllSkills() ([]SkillInfo, error) {
	if err := os.MkdirAll(m.baseDir, 0755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		return nil, err
	}

	var skills []SkillInfo
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			enabled := true
			// Check if a .disabled file exists in the directory
			if _, err := os.Stat(filepath.Join(m.baseDir, name, ".disabled")); err == nil {
				enabled = false
			}

			description, _ := os.ReadFile(filepath.Join(m.baseDir, name, "description.txt"))
			skills = append(skills, SkillInfo{
				Name:        name,
				Description: strings.TrimSpace(string(description)),
				Enabled:     enabled,
			})
		}
	}
	return skills, nil
}

func (m *Manager) DisableSkill(name string) error {
	if err := validateSkillName(name); err != nil {
		return err
	}
	path := filepath.Join(m.baseDir, name)
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(path, ".disabled"), []byte(""), 0644)
}

func (m *Manager) EnableSkill(name string) error {
	if err := validateSkillName(name); err != nil {
		return err
	}
	err := os.Remove(filepath.Join(m.baseDir, name, ".disabled"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (m *Manager) AddSkill(name string, description string) error {
	if err := validateSkillName(name); err != nil {
		return err
	}
	path := filepath.Join(m.baseDir, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	if description != "" {
		if err := os.WriteFile(filepath.Join(path, "description.txt"), []byte(description), 0644); err != nil {
			return err
		}
	}
	skillFile := filepath.Join(path, "SKILL.md")
	if _, err := os.Stat(skillFile); errors.Is(err, os.ErrNotExist) {
		content := fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n\n", name, description)
		return os.WriteFile(skillFile, []byte(content), 0644)
	}
	return nil
}

func (m *Manager) BaseDir() string {
	return m.baseDir
}

func validateSkillName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("skill name is empty")
	}
	if strings.ContainsAny(name, `/\`) {
		return errors.New("skill name contains path separator")
	}
	return nil
}
