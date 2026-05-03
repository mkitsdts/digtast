package workspace

import "log/slog"

type PromptImpl interface {
	CreatePromptImpl() error
	GetPromptImpl() (string, error)
	GetPromptName() string
	GetRole() string
}

var promptCreators []PromptImpl

func RegisterPromptCreator(promptImpl PromptImpl) {
	// create template
	promptCreators = append(promptCreators, promptImpl)
	slog.Info("registered prompt creator", "name", promptImpl.GetPromptName())
}

func GetPromptCreators() []PromptImpl {
	return promptCreators
}
