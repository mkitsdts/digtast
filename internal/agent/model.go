package agent

import (
	"fmt"

	"digital-labor/component/chatmodel"
	"digital-labor/config"
)

// newChatModel 根据注册表中的厂商构建指定模型的客户端，凭据来自
// config.Conf.Models 中按 Provider 保存的配置；URL 非空时替换默认地址。
func newChatModel(provider, model string) (chatmodel.ChatModel, error) {
	info, ok := chatmodel.FindProvider(provider)
	if !ok {
		return nil, fmt.Errorf("provider %q is not registered", provider)
	}
	creds := config.ModelCredentials(provider)
	if creds.Key == "" {
		return nil, fmt.Errorf("provider %s has no api key configured", provider)
	}
	client, err := info.NewClient(chatmodel.ClientOptions{APIKey: creds.Key, BaseURL: creds.URL}, model)
	if err != nil {
		return nil, fmt.Errorf("build %s client: %w", provider, err)
	}
	return client, nil
}
