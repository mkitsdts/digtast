package qq

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	opDispatch       = 0
	opHeartbeat      = 1
	opIdentify       = 2
	opResume         = 6
	opReconnect      = 7
	opInvalidSession = 9
	opHello          = 10
	opHeartbeatACK   = 11

	intentGuilds             = 1 << 0
	intentGuildMembers       = 1 << 1
	intentGuildMessages      = 1 << 9
	intentDirectMessage      = 1 << 12
	intentGroupAndC2C        = 1 << 25
	intentInteraction        = 1 << 26
	intentPublicGuildMessage = 1 << 30

	maxReconnect     = 10
	baseReconnectSec = 2
)

type baseMessage struct {
	ID string `json:"id"`
	Op int    `json:"op"`
	D  any    `json:"d"`
	S  int    `json:"s"`
	T  string `json:"t"`
}

type helloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type readyData struct {
	SessionID string `json:"session_id"`
}

type resumeData struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Seq       int    `json:"seq"`
}

type identifyData struct {
	Token   string         `json:"token"`
	Intents int            `json:"intents"`
	Shard   [2]int         `json:"shard"`
	Props   map[string]any `json:"properties"`
}

type wsConn struct {
	mu            sync.Mutex
	conn          *websocket.Conn
	seq           int
	sessionID     string
	lastAck       time.Time
	refresh_token string
}

func (w *wsConn) send(msg baseMessage) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteJSON(msg)
}

func (w *wsConn) updateSeq(s int) {
	if s > 0 {
		w.mu.Lock()
		w.seq = s
		w.mu.Unlock()
	}
}

func (w *wsConn) markAck() {
	w.mu.Lock()
	w.lastAck = time.Now()
	w.mu.Unlock()
}

func (w *wsConn) snapshot() (seq int, sessionID string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.seq, w.sessionID
}

// Handler 建立并维护与 QQ 网关的 WebSocket 连接，处理断线重连和会话恢复。
func (c *QQChannel) Handler(ctx context.Context, gatewayURL string) {
	for attempt := 0; attempt < maxReconnect; attempt++ {
		if ctx.Err() != nil {
			return
		}

		if attempt > 0 {
			backoff := baseReconnectSec * (1 << (attempt - 1))
			if backoff > 60 {
				backoff = 60
			}
			slog.Info("reconnecting to qq gateway", "attempt", attempt+1, "backoff_sec", backoff)
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(backoff) * time.Second):
			}
		}

		resumed := c.connectAndServe(ctx, gatewayURL, attempt > 0)
		if resumed {
			attempt = -1 // 重置重试计数
		}
	}
	slog.Error("exceeded max reconnect attempts to qq gateway")
}

// connectAndServe 建立一次 WebSocket 连接并运行完整的协议流程。
// 返回 true 表示连接正常结束（服务端要求重连），可立即重试。
func (c *QQChannel) connectAndServe(ctx context.Context, gatewayURL string, tryResume bool) bool {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, gatewayURL, nil)
	if err != nil {
		slog.Error("failed to dial qq gateway", "error", err)
		return false
	}

	wc := &wsConn{conn: conn, lastAck: time.Now()}
	defer conn.Close()

	// 等待 Hello 消息（op=10）
	helloCtx, helloCancel := context.WithTimeout(ctx, 10*time.Second)
	defer helloCancel()
	interval, err := c.waitHello(helloCtx, wc)
	if err != nil {
		slog.Error("failed to receive hello", "error", err)
		return false
	}

	// 尝试恢复会话或重新鉴权
	if tryResume && c.sessionID != "" && c.lastSeq > 0 {
		if err := c.sendResume(wc); err != nil {
			slog.Warn("failed to send resume, falling back to identify", "error", err)
			if err := c.sendIdentify(wc); err != nil {
				slog.Error("failed to send identify", "error", err)
				return false
			}
		}
	} else {
		if err := c.sendIdentify(wc); err != nil {
			slog.Error("failed to send identify", "error", err)
			return false
		}
	}

	// 启动心跳
	heartCtx, heartCancel := context.WithCancel(ctx)
	defer heartCancel()
	go c.heartbeatLoop(heartCtx, wc, interval)

	// 进入消息循环
	return c.messageLoop(ctx, wc)
}

// waitHello 等待服务端发送 op=10 Hello 消息，返回心跳间隔（毫秒）。
func (c *QQChannel) waitHello(ctx context.Context, wc *wsConn) (int, error) {
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}

		_, raw, err := wc.conn.ReadMessage()
		if err != nil {
			return 0, fmt.Errorf("read hello: %w", err)
		}

		var msg baseMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			slog.Warn("failed to parse hello message", "error", err)
			continue
		}

		if msg.Op != opHello {
			slog.Warn("unexpected op before hello", "op", msg.Op)
			continue
		}

		var hello helloData
		if err := json.Unmarshal(marshalJSON(msg.D), &hello); err != nil {
			return 0, fmt.Errorf("parse hello data: %w", err)
		}
		return hello.HeartbeatInterval, nil
	}
}

// sendIdentify 发送 op=2 鉴权消息。
func (c *QQChannel) sendIdentify(wc *wsConn) error {
	msg := baseMessage{
		Op: opIdentify,
		D: identifyData{
			Token:   fmt.Sprintf("QQBot %s", c.token()),
			Intents: intentGuilds | intentPublicGuildMessage | intentGuildMembers | intentDirectMessage | intentGroupAndC2C | intentInteraction,
		},
	}
	return wc.send(msg)
}

// sendResume 发送 op=6 会话恢复消息。
func (c *QQChannel) sendResume(wc *wsConn) error {
	return wc.send(baseMessage{
		Op: opResume,
		D: resumeData{
			Token:     fmt.Sprintf("QQBot %s", c.token()),
			SessionID: c.sessionID,
			Seq:       c.lastSeq,
		},
	})
}

// heartbeatLoop 按网关指定的间隔发送心跳，并检测 ACK 超时。
func (c *QQChannel) heartbeatLoop(ctx context.Context, wc *wsConn, intervalMs int) {
	ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
	defer ticker.Stop()

	deadline := time.Duration(intervalMs*2) * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 检查上次 ACK 是否超时
			wc.mu.Lock()
			stale := time.Since(wc.lastAck) > deadline
			wc.mu.Unlock()
			if stale {
				slog.Warn("heartbeat ack timeout, connection may be dead")
				wc.conn.Close()
				return
			}

			seq, _ := wc.snapshot()
			if err := wc.send(baseMessage{Op: opHeartbeat, D: seq}); err != nil {
				slog.Error("failed to send heartbeat", "error", err)
				return
			}
		}
	}
}

// messageLoop 读取并处理所有网关消息，返回值指示是否应立即重连。
func (c *QQChannel) messageLoop(ctx context.Context, wc *wsConn) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		_, raw, err := wc.conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return false
			}
			slog.Error("websocket read error", "error", err)
			return false
		}

		var msg baseMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			slog.Warn("failed to parse message", "error", err)
			continue
		}

		wc.updateSeq(msg.S)

		switch msg.Op {
		case opDispatch:
			c.handleDispatch(wc, msg)

		case opReconnect:
			slog.Info("server requested reconnect")
			return true

		case opInvalidSession:
			slog.Warn("invalid session, will re-identify")
			c.sessionID = ""
			c.lastSeq = 0
			return true

		case opHeartbeatACK:
			wc.markAck()

		case opHeartbeat:
			// 服务端主动请求心跳
			seq, _ := wc.snapshot()
			wc.send(baseMessage{Op: opHeartbeat, D: seq})

		default:
			slog.Debug("unhandled op", "op", msg.Op)
		}
	}
}

// handleDispatch 处理 op=0 事件分发。
func (c *QQChannel) handleDispatch(wc *wsConn, msg baseMessage) {
	switch msg.T {
	case "READY":
		var ready readyData
		if err := json.Unmarshal(marshalJSON(msg.D), &ready); err != nil {
			slog.Error("failed to parse READY data", "error", err)
			return
		}
		wc.mu.Lock()
		wc.sessionID = ready.SessionID
		wc.mu.Unlock()
		c.sessionID = ready.SessionID
		slog.Info("qq gateway ready", "session_id", ready.SessionID)

	case "RESUMED":
		slog.Info("qq gateway session resumed")

	default:
		c.dispatchEvent(msg)
	}
}

// dispatchEvent 将事件交给注册的处理器处理。
func (c *QQChannel) dispatchEvent(msg baseMessage) {
	c.mu.Lock()
	handler, ok := c.handlers[msg.T]
	c.mu.Unlock()

	if !ok {
		slog.Debug("no handler for event", "event", msg.T)
		return
	}

	data, err := json.Marshal(msg.D)
	if err != nil {
		slog.Error("failed to marshal event data", "error", err)
		return
	}

	if err := handler(data); err != nil {
		slog.Error("event handler error", "event", msg.T, "error", err)
	}
}

// RegisterHandler 注册指定事件类型的处理器。
func (c *QQChannel) RegisterHandler(eventType string, handler func(json.RawMessage) error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.handlers == nil {
		c.handlers = make(map[string]func(json.RawMessage) error)
	}
	c.handlers[eventType] = handler
}

// token 生成 QQ Bot API 格式的 access token。
func (c *QQChannel) token() string {
	for c.refreshToken == "" {
		c.refreshMux.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	c.refreshMux.Lock()
	defer c.refreshMux.Unlock()
	return fmt.Sprintf("%s.%s", c.refreshToken)
}

// generateNonce 生成随机 nonce 字符串。
func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// getWebSocketUrl 从 QQ 网关获取 WebSocket 连接地址。
func (c *QQChannel) getWebSocketUrl() (string, error) {
	url := "https://api.sgroup.qq.com/gateway"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("QQBot %s", c.token()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gateway returned status %d", resp.StatusCode)
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode gateway response: %w", err)
	}
	if result.URL == "" {
		return "", fmt.Errorf("gateway returned empty url")
	}
	return result.URL, nil
}

func (c *QQChannel) refreshTokenLoop(ctx context.Context) error {
	url := "https://bots.qq.com/app/getAppAccessToken"

	params := map[string]string{
		"appId":        c.AppID,
		"clientSecret": c.AppSecret,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalJSON(params)))
	if err != nil {
		return err
	}

	for {
		c.refreshMux.Lock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req.Header.Set("Authorization", fmt.Sprintf("QQBot %s", c.token()))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("gateway returned status %d", resp.StatusCode)
		}

		var result struct {
			RefreshToken string `json:"refresh_token"`
			ExpireIn     int    `json:"expire_in"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("decode gateway response: %w", err)
		}
		c.refreshToken = result.RefreshToken
		c.refreshMux.Unlock()
		time.Sleep(time.Duration(result.ExpireIn/5*4) * time.Millisecond)
	}
}

// marshalJSON 将任意值序列化再反序列化为 json.RawMessage，用于结构体转换。
func marshalJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
