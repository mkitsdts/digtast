package model

// GWMessage 网关消息
type Result struct {
	Success      bool
	Error        error
	Content      string
	MultiContent map[string][]byte // title: data
}
