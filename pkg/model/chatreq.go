package model

type ChatRequest struct {
	Content  string
	IsStream bool
}

type CreateChatModelRequest struct {
	Model       string `json:"model"`
	Description string `json:"description"`
	Key         string `json:"key"`
	Provider    string `json:"provider"`
}

type CreateChatModelResponse struct {
	ChatModelId string `json:"chatModelId"`
	Success     bool   `json:"success"`
	Message     string `json:"message"`
}

type RemoveChatModelRequest struct {
	ChatModelId string `json:"chatModelId"`
}

type RemoveChatModelResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
