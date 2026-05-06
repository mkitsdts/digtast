package gateway

import "digital-labor/pkg/model"

type UserMessage struct {
	Content      string
	MultiContent map[string][]byte // title: data
	NotifyWay    string
	Params       map[string]any // key-value pairs for notify way
	Result       model.Result
}
