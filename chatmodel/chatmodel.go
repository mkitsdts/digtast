// Package chatmodel 注册全部内置模型适配器。
// 新增 Provider：在 chatmodel/<provider> 目录组合 oai 底座实现客户端，在
// init 中调用 chatmodel.Register 注册模型，最后在这里导入包。
package chatmodel

import (
	_ "digital-labor/chatmodel/deepseek"
	_ "digital-labor/chatmodel/openai"
)
