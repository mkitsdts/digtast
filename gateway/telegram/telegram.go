package telegram

type TelegramChannel struct {
	Key string `yaml:"key"`
}

func NewTelegramChannel(key string) *TelegramChannel {
	return &TelegramChannel{
		Key: key,
	}
}

func (c *TelegramChannel) Send(content string) error {
	// TODO:
	return nil
}

func (c *TelegramChannel) IsActive() bool {
	// TODO:
	return false
}

func (c *TelegramChannel) Register(key string) error {
	// TODO:
	return nil
}

func (c *TelegramChannel) Serve() error {
	// TODO:
	return nil
}

func init() {

}
