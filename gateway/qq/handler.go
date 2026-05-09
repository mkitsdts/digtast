package qq

import (
	"bytes"
	"digital-labor/internal/gateway"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type Attachment struct {
	ContentType  string `json:"content_type"`
	FileName     string `json:"filename"`
	Height       int    `json:"height"`
	Width        int    `json:"width"`
	Size         int    `json:"size"`
	Url          string `json:"url"`
	VoiceWavUrl  string `json:"voice_wav_url"`
	AsrReferText string `json:"asr_refer_text"`
}

type C2CMessage struct {
	ID     string `json:"id"`
	Author struct {
		UserOpenID string `json:"user_openid"`
	} `json:"author"`
	Content     string       `json:"content"`
	TimeStamp   string       `json:"timestamp"`
	Attachments []Attachment `json:"attachments"` // 富消息文本
}

// C2CMessageEventHandler 实现处理 at 消息的回调
func (c *QQChannel) C2CMessageEventHandler() func(msg json.RawMessage) error {
	return func(msg json.RawMessage) error {
		data := C2CMessage{}
		if err := json.Unmarshal(msg, &data); err != nil {
			return err
		}
		slog.Debug("message received", "content", data.Content)

		var multi map[string][]byte
		if len(data.Attachments) > 0 {
			multi = make(map[string][]byte)
			for i := range data.Attachments {
				url := data.Attachments[i].Url
				if data.Attachments[i].ContentType == "voice" && data.Attachments[i].VoiceWavUrl != "" {
					url = data.Attachments[i].VoiceWavUrl
				}
				body, err := c.GetMultiContent(url)
				if err != nil {
					continue
				}
				multi[data.Attachments[i].FileName] = body
			}
		}

		content := data.Content
		gateway.PushUserMessage(&gateway.UserMessage{
			Content:      content,
			NotifyWay:    "qq",
			Params:       map[string]any{"user_id": data.Author.UserOpenID, "msg_id": data.ID},
			MultiContent: multi,
		})
		slog.Info("message pushed", "content", content, "user_id", data.Author.UserOpenID, "msg_id", data.ID)

		return nil
	}
}

func (c *QQChannel) SendMessage(content, user_id, msg_id string) error {
	// /v2/users/{openid}/messages 需要调用 HTTP POST 接口发送消息，珠宝要回家了，由于需要git远程同步，先暂时提交一下
	url := fmt.Sprintf("https://api.sgroup.qq.com/v2/users/%s/messages", user_id)
	payload := map[string]any{
		"content":  content,
		"msg_type": 0,
		"event_id": "C2C_MSG_RECEIVE",
		"msg_id":   msg_id,
	}

	body := bytes.NewBuffer(marshalJSON(payload))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.token())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Error("send message fatal", "status", resp.StatusCode, "user_id", user_id, "msg_id", msg_id, "token", c.token())
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *QQChannel) GetMultiContent(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.token())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("get multi content failed", "status", resp.StatusCode)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body := make([]byte, 0)
	if _, err = resp.Body.Read(body); err != nil {
		return nil, err
	}
	return body, nil
}
