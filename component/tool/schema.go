package tool

type Result struct {
	Content string         `json:"content"`
	Data    map[string]any `json:"data,omitempty"`
	IsError bool           `json:"is_error,omitempty"`
}
