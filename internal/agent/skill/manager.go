package skill

import (
	"context"
	"digital-labor/internal/agent/workspace"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type SkillInfo struct {
	Name        string
	Description string
	Enabled     bool
}

type Authorizer func(ctx context.Context, skill SkillInfo) error

type Manager struct {
	baseDir string
}

func NewManager() *Manager {
	return &Manager{baseDir: workspace.SkillsDir()}
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

			description, _ := m.description(name)
			skills = append(skills, SkillInfo{
				Name:        name,
				Description: description,
				Enabled:     enabled,
			})
		}
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })
	return skills, nil
}

func (m *Manager) LoadSkill(ctx context.Context, name string, authorize Authorizer) (string, error) {
	if err := validateSkillName(name); err != nil {
		return "", err
	}
	path := filepath.Join(m.baseDir, name)
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		if err == nil {
			err = errors.New("not a directory")
		}
		return "", fmt.Errorf("skill %q not found: %w", name, err)
	}
	if _, err := os.Stat(filepath.Join(path, ".disabled")); err == nil {
		return "", fmt.Errorf("skill %q is disabled", name)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("check skill %q state: %w", name, err)
	}
	description, _ := m.description(name)
	info := SkillInfo{Name: name, Description: description, Enabled: true}
	if authorize != nil {
		if err := authorize(ctx, info); err != nil {
			return "", fmt.Errorf("skill %q is not authorized: %w", name, err)
		}
	}
	content, err := os.ReadFile(filepath.Join(path, "SKILL.md"))
	if err != nil {
		return "", fmt.Errorf("read skill %q: %w", name, err)
	}
	return string(content), nil
}

func (m *Manager) Prompt(context.Context) (string, error) {
	skills, err := m.GetAllSkills()
	if err != nil {
		return "", err
	}
	var lines []string
	for _, skill := range skills {
		if !skill.Enabled {
			continue
		}
		line := "- " + skill.Name
		if skill.Description != "" {
			line += ": " + skill.Description
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return "", nil
	}
	return "### Available skills\nThe following skills are available. Their full instructions are not loaded. Call `load_skill` with the exact skill name before following a skill.\n" + strings.Join(lines, "\n"), nil
}

func (m *Manager) description(name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(m.baseDir, name, "description.txt"))
	if err == nil && strings.TrimSpace(string(data)) != "" {
		return strings.TrimSpace(string(data)), nil
	}
	skillData, skillErr := os.ReadFile(filepath.Join(m.baseDir, name, "SKILL.md"))
	if skillErr != nil {
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		return "", skillErr
	}
	content := string(skillData)
	if !strings.HasPrefix(content, "---\n") {
		return "", nil
	}
	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return "", nil
	}
	var metadata struct {
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal([]byte(content[4:4+end]), &metadata); err != nil {
		return "", err
	}
	return strings.TrimSpace(metadata.Description), nil
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
