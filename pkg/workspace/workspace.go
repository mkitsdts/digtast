package workspace

import (
	"log/slog"
	"os"
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

	return workspacePath
}

func InitWorkspace() {
	// TODO:

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
}

var promptCreators []PromptImpl

func RegisterPromptCreator(promptImpl PromptImpl) {
	promptCreators = append(promptCreators, promptImpl)
}

func GetPromptCreators() []PromptImpl {
	return promptCreators
}
