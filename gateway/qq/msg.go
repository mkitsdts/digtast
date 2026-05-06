package qq

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
