package skill

import (
	"digital-labor/pkg/workspace"
	"os"
	"path/filepath"
)

type SkillInfo struct {
	Name        string
	Description string
	Enabled     bool
}

type Manager struct {
	baseDir string
}

func NewManager() *Manager {
	return &Manager{
		baseDir: filepath.Join(workspace.GetWorkspacePath(), "skills"),
	}
}

func (m *Manager) GetAllSkills() ([]SkillInfo, error) {
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
			
			skills = append(skills, SkillInfo{
				Name:    name,
				Enabled: enabled,
			})
		}
	}
	return skills, nil
}

func (m *Manager) DisableSkill(name string) error {
	return os.WriteFile(filepath.Join(m.baseDir, name, ".disabled"), []byte(""), 0644)
}

func (m *Manager) EnableSkill(name string) error {
	return os.Remove(filepath.Join(m.baseDir, name, ".disabled"))
}

func (m *Manager) AddSkill(name string, description string) error {
	path := filepath.Join(m.baseDir, name)
	return os.MkdirAll(path, 0755)
}
