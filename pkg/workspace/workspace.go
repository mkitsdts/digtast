package workspace

import (
	"log/slog"
	"os"
	"path"
)

func GetWorkspacePath() string {
	workspacePath, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	if workspacePath == "" {
		workspacePath, err = os.Getwd()
		if err != nil {
			return ""
		}
	}

	workspacePath = path.Join(workspacePath, "/.digtast")

	return workspacePath
}

func InitWorkspace() {
	// TODO:
	path := GetWorkspacePath()
	if err := os.MkdirAll(path, 0755); err != nil {
		slog.Error("Failed to create workspace", "path", path, "err", err)
		return
	}

	for _, promptCreator := range promptCreators {
		err := promptCreator.CreatePromptImpl()
		if err != nil {
			slog.Error("Failed create prompt impl", "name", promptCreator.GetPromptName())
		}
	}
}

type PromptImpl interface {
	CreatePromptImpl() error
	GetPromptImpl() (string, error)
	GetPromptName() string
	GetRole() string
}

var promptCreators []PromptImpl

func RegisterPromptCreator(promptImpl PromptImpl) {
	promptCreators = append(promptCreators, promptImpl)
}

func GetPromptCreators() []PromptImpl {
	return promptCreators
}
