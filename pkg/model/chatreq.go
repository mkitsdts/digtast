package model

type ChatRequest struct {
	SessionID string
	Content   string
	IsStream  bool
	Prompt    PromptContext
}

type CreateChatModelRequest struct {
	Model       string `json:"model"`       // key into global Models config (e.g., "deepseek")
	Description string `json:"description"` // human-readable description of the model
	Key         string `json:"key"`         // api-Key
	Provider    string `json:"provider"`    // provider of the model (e.g., "doubao")
}

type CreateChatModelResponse struct {
	ChatModelId string `json:"chatModelId"` //
	Success     bool   `json:"success"`     //
	Message     string `json:"message"`     //
}

type RemoveChatModelRequest struct {
	ChatModelId string `json:"chatModelId"` //
}

type RemoveChatModelResponse struct {
	Success bool   `json:"success"` //
	Message string `json:"message"` //
}
