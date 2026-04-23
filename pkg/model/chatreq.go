package model

type ChatRequest struct {
	SessionID string
	Content   string
	IsStream  bool
	Prompt    PromptContext
}
