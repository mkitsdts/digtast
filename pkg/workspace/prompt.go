package workspace

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
}

func GetPromptCreators() []PromptImpl {
	return promptCreators
}
